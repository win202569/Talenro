using System;
using System.Collections.Generic;

public interface IRetainedLifetimeClock {long Ticks{get;} long Frequency{get;} string UtcNow{get;}}
public interface IRetainedLifetimeApi {
    LifetimeCandidates QueryMembers(IntPtr job,long deadlineTicks);
    LifetimeOpen Open(uint pid,uint access,bool inherit);
    LifetimeMembership IsOwned(IntPtr handle,IntPtr job);
    LifetimeRead Read(IntPtr handle,long deadlineTicks);
    LifetimeRelease Close(IntPtr handle);
}
public class LifetimeApiEvidence {
    public bool Succeeded; public int Error; public long StartTimestamp,EndTimestamp,Frequency;
    public string StartUTC,EndUTC;
}
public class LifetimeQueryEvidence:LifetimeApiEvidence {public uint Assigned,Returned;public int Capacity,BufferBytes;}
public sealed class LifetimeCandidates:LifetimeQueryEvidence {public ulong[] Ids=new ulong[0];}
// Transient adapter reply only. Never attach this object to a report.
public sealed class LifetimeOpen:LifetimeApiEvidence {public IntPtr Handle;}
public sealed class LifetimeMembership:LifetimeApiEvidence {public bool Owned;}
public sealed class LifetimeRead {
    public bool Attempted,ExitCodeRead,TerminationRequested,SameEntry;
    public uint ExitCode;public int Error,ImageError;public string State="unknown",ImagePath,SkipReason,StartUTC,EndUTC;
    public long StartTimestamp,EndTimestamp,Frequency;
}
public sealed class LifetimeRelease {public bool Succeeded;public int Error;}
public sealed class LifetimeEntry {
    public int EntryId;public ulong Pid;public int SourceQuery;public LifetimeQueryEvidence SourceQueryEvidence;
    public LifetimeApiEvidence Open;public LifetimeMembership Membership;
    public LifetimeRead Admission,RootWaitEnd,Final;
}
public sealed class LifetimeReport {
    public int Schema=1,Attempts,Admitted,QueryCount,ReleaseAttempts,ReleaseFailures;
    public string Coverage="sampled";
    public bool AdmissionBudgetExhausted,Incomplete,RootExited,TimedOut,TeardownAttempted;
    public List<string> Errors=new List<string>();public List<LifetimeEntry> Entries=new List<LifetimeEntry>();
    public LifetimeQueryEvidence FirstQuery,LastQuery;
}
public sealed class RetainedLifetimeRegistry:IDisposable {
    readonly IntPtr job;readonly uint root;readonly IRetainedLifetimeApi api;readonly IRetainedLifetimeClock clock;
    readonly HashSet<ulong> seen=new HashSet<ulong>();readonly List<Owned> owned=new List<Owned>();
    bool stopped,disposed,rootSnapshot,finalSnapshot;long lastTick;
    public LifetimeReport Report{get;private set;}
    sealed class Owned {public IntPtr Handle;public LifetimeEntry Entry;public bool Released;}
    public RetainedLifetimeRegistry(IntPtr job,uint rootId,IRetainedLifetimeApi api,IRetainedLifetimeClock clock) {
        if(job==IntPtr.Zero || job==new IntPtr(-1) || rootId==0 || api==null || clock==null || clock.Frequency<=0)throw new ArgumentException("Registry requires own job, root, API and monotonic clock");
        this.job=job;root=rootId;this.api=api;this.clock=clock;lastTick=clock.Ticks;Report=new LifetimeReport();
    }
    void Error(string stage,int code) {Report.Incomplete=true;stopped=true;if(Report.Errors.Count<96)Report.Errors.Add(stage+":"+code);}
    bool Budget(long deadline) {
        long now=clock.Ticks;if(now<0 || now<lastTick){Error("clock",-1);return false;}lastTick=now;
        if(deadline<=now){Error("budget",-1);return false;}return true;
    }
    static LifetimeApiEvidence Evidence(LifetimeApiEvidence v) {
        return v==null?null:new LifetimeApiEvidence{Succeeded=v.Succeeded,Error=v.Error,StartTimestamp=v.StartTimestamp,EndTimestamp=v.EndTimestamp,Frequency=v.Frequency,StartUTC=v.StartUTC,EndUTC=v.EndUTC};
    }
    static LifetimeQueryEvidence QueryEvidence(LifetimeCandidates v) {
        return v==null?null:new LifetimeQueryEvidence{Succeeded=v.Succeeded,Error=v.Error,StartTimestamp=v.StartTimestamp,EndTimestamp=v.EndTimestamp,Frequency=v.Frequency,StartUTC=v.StartUTC,EndUTC=v.EndUTC,Assigned=v.Assigned,Returned=v.Returned,Capacity=v.Capacity,BufferBytes=v.BufferBytes};
    }
    void Release(Owned value) {
        if(value.Released)return;value.Released=true;Report.ReleaseAttempts++;
        try {var r=api.Close(value.Handle);if(r==null || !r.Succeeded){Report.ReleaseFailures++;Error("close",r==null?-1:r.Error);}}
        catch(Exception){Report.ReleaseFailures++;Error("close-exception",-1);}
    }
    LifetimeRead Read(Owned value,long deadline,bool termination) {
        var unknown=new LifetimeRead{SameEntry=true,TerminationRequested=termination,SkipReason="budget-exhausted"};
        if(!Budget(deadline))return unknown;
        try {
            var r=api.Read(value.Handle,deadline);
            if(r==null){Error("read-null",-1);return new LifetimeRead{SameEntry=true,TerminationRequested=termination,SkipReason="missing-read"};}
            r.SameEntry=true;r.TerminationRequested=termination;
            if(!r.Attempted || r.Error!=0 || !r.ExitCodeRead || (r.State!="running" && r.State!="exited"))Error("read",r.Error);
            if(clock.Ticks>=deadline)Error("read-overrun",-1);
            return r;
        }catch(Exception){Error("read-exception",-1);return new LifetimeRead{SameEntry=true,TerminationRequested=termination,SkipReason="read-exception"};}
    }
    public void Poll(long runtimeDeadlineTicks) {
        if(disposed || stopped || Report.AdmissionBudgetExhausted)return;
        try {
            if(!Budget(runtimeDeadlineTicks))return;
            var list=api.QueryMembers(job,runtimeDeadlineTicks);Report.QueryCount++;
            Report.LastQuery=QueryEvidence(list);if(Report.FirstQuery==null)Report.FirstQuery=QueryEvidence(list);
            if(list==null || !list.Succeeded || list.Error!=0 || list.Ids==null || list.Assigned!=list.Returned || list.Returned!=list.Ids.Length){Error("list",list==null?-1:list.Error);return;}
            foreach(ulong id in list.Ids) {
                if(Report.Attempts>=32){Report.AdmissionBudgetExhausted=true;break;}
                if(id==root)continue;
                if(id==0 || id>UInt32.MaxValue){Report.Incomplete=true;continue;}
                if(!seen.Add(id))continue;
                if(!Budget(runtimeDeadlineTicks))break;
                var row=new LifetimeEntry{EntryId=Report.Attempts+1,Pid=id,SourceQuery=Report.QueryCount,SourceQueryEvidence=QueryEvidence(list)};
                Report.Attempts++;Report.Entries.Add(row);if(Report.Attempts==32)Report.AdmissionBudgetExhausted=true;
                Owned temporary=null;
                try {
                    var open=api.Open((uint)id,0x101000,false);row.Open=Evidence(open);
                    if(open!=null && open.Handle!=IntPtr.Zero && open.Handle!=new IntPtr(-1))temporary=new Owned{Handle=open.Handle,Entry=row};
                    if(open==null || !open.Succeeded || open.Error!=0 || temporary==null){Error("open",open==null?-1:open.Error);break;}
                    if(!Budget(runtimeDeadlineTicks))break;
                    row.Membership=api.IsOwned(temporary.Handle,job);
                    if(row.Membership==null || !row.Membership.Succeeded || row.Membership.Error!=0 || !row.Membership.Owned){Error("membership",row.Membership==null?-1:row.Membership.Error);break;}
                    var retained=temporary;owned.Add(retained);temporary=null;Report.Admitted++;
                    row.Admission=Read(retained,runtimeDeadlineTicks,false);
                }finally{if(temporary!=null)Release(temporary);}
                if(stopped)break;
            }
        }catch(Exception){Error("admission-exception",-1);}
    }
    public void Snapshot(string phase,long shutdownDeadlineTicks,bool rootExited,bool timedOut,bool terminationRequested) {
        if(disposed)return;
        if(phase!="root-wait-end" && phase!="post-cleanup")throw new ArgumentException("Unknown lifetime phase");
        if(phase=="root-wait-end"){if(rootSnapshot)return;rootSnapshot=true;}else{if(finalSnapshot)return;finalSnapshot=true;}
        Report.RootExited=rootExited;Report.TimedOut=timedOut;
        foreach(var value in owned) {
            var read=Read(value,shutdownDeadlineTicks,terminationRequested);
            if(phase=="root-wait-end")value.Entry.RootWaitEnd=read;else value.Entry.Final=read;
        }
    }
    public void Dispose() {
        if(disposed)return;disposed=true;Report.TeardownAttempted=true;
        foreach(var value in owned)Release(value);
    }
}
public static class RetainedLifetimeTeardown {
    public static void Run(Action coreCleanup,RetainedLifetimeRegistry registry) {
        try{coreCleanup();}finally{if(registry!=null)registry.Dispose();}
    }
}
