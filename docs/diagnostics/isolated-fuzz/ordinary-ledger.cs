using System;
using System.Collections.Generic;

public interface IOrdinaryLedgerApi {
    LifetimeCandidates QueryMembers(IntPtr job,long deadlineTicks);
    LifetimeOpen Open(uint pid,uint access,bool inherit);
    LifetimeMembership IsOwned(IntPtr handle,IntPtr job);
    OrdinaryRead Read(IntPtr handle,uint waitMs,long deadlineTicks);
    LifetimeRelease Close(IntPtr handle);
    OrdinaryAccounting QueryAccounting(IntPtr job,long deadlineTicks);
}
public sealed class OrdinaryRead {
    public string RunIdentity,State="unknown",ImagePath,SkipReason,StartUTC,EndUTC;
    public int EntryId,Error,ImageError;
    public bool SameHandleVerified,Attempted,Signaled,ExitCodeRead,TerminationRequested;
    public uint ExitCode;
    public long StartTimestamp,EndTimestamp,Frequency;
}
public sealed class OrdinaryAccounting:LifetimeApiEvidence {
    public uint TotalProcesses,ActiveProcesses,TotalTerminatedProcesses;
}
public sealed class OrdinaryEntry {
    public string RunIdentity;public int EntryId,SourceQuery;public ulong Pid;
    public LifetimeQueryEvidence SourceQueryEvidence;public LifetimeApiEvidence Open;
    public LifetimeMembership Membership;public OrdinaryRead Initial,Final;
    public bool ReleaseAttempted,ReleaseSucceeded;public int ReleaseError;
}
public sealed class OrdinaryLedgerReport {
    public int Schema=1,Attempts,Admitted,QueryCount,ReleaseAttempts,ReleaseFailures;
    public string RunIdentity,Mode="ordinary-identity-ledger-v1",Coverage="sampled";
    public uint RootId;
    public bool Incomplete,Ambiguous,CapExhausted,TeardownAttempted,Sealed;
    public List<string> Errors=new List<string>();
    public LifetimeQueryEvidence FirstQuery,LastQuery;
    public List<OrdinaryEntry> Entries=new List<OrdinaryEntry>();
}
public sealed class OrdinaryMemberLedger:IDisposable {
    readonly IntPtr job;readonly uint root;readonly IOrdinaryLedgerApi api;
    readonly IRetainedLifetimeClock clock;readonly HashSet<ulong> seen=new HashSet<ulong>();
    readonly List<Owner> owners=new List<Owner>();long lastTick;bool disposed;
    sealed class Owner {public IntPtr Handle;public OrdinaryEntry Entry;public bool Released;}
    public OrdinaryLedgerReport Report {get;private set;}
    internal ulong[] LastMembers=new ulong[0];
    internal LifetimeQueryEvidence CurrentPollQuery;
    public OrdinaryMemberLedger(string runIdentity,IntPtr job,uint rootId,IOrdinaryLedgerApi api,IRetainedLifetimeClock clock){
        if(String.IsNullOrWhiteSpace(runIdentity) || job==IntPtr.Zero || job==new IntPtr(-1) || rootId==0 || api==null || clock==null || clock.Frequency<=0 || clock.Ticks<0)throw new ArgumentException("Own job, run, root, API and clock required");
        this.job=job;root=rootId;this.api=api;this.clock=clock;lastTick=clock.Ticks;
        Report=new OrdinaryLedgerReport{RunIdentity=runIdentity,RootId=rootId};
    }
    internal void Fail(string stage,int error){Report.Incomplete=true;if(Report.Errors.Count<96)Report.Errors.Add(stage+":"+error);}
    internal bool Budget(long deadline){long now=clock.Ticks;if(now<0 || now<lastTick){Fail("clock",-1);return false;}lastTick=now;if(now>=deadline){Fail("budget",-1);return false;}return true;}
    static LifetimeApiEvidence Copy(LifetimeApiEvidence e){return e==null?null:new LifetimeApiEvidence{Succeeded=e.Succeeded,Error=e.Error,StartTimestamp=e.StartTimestamp,EndTimestamp=e.EndTimestamp,Frequency=e.Frequency,StartUTC=e.StartUTC,EndUTC=e.EndUTC};}
    internal static LifetimeQueryEvidence CopyQuery(LifetimeCandidates e){return e==null?null:new LifetimeQueryEvidence{Succeeded=e.Succeeded,Error=e.Error,StartTimestamp=e.StartTimestamp,EndTimestamp=e.EndTimestamp,Frequency=e.Frequency,StartUTC=e.StartUTC,EndUTC=e.EndUTC,Assigned=e.Assigned,Returned=e.Returned,Capacity=e.Capacity,BufferBytes=e.BufferBytes};}
    internal static bool ValidApi(LifetimeApiEvidence e){return e!=null && e.Succeeded && e.Error==0 && e.Frequency>0 && e.StartTimestamp>=0 && e.EndTimestamp>=e.StartTimestamp;}
    internal static bool ValidList(LifetimeCandidates e){return ValidApi(e) && e.Ids!=null && e.Assigned==e.Returned && e.Returned==e.Ids.Length && e.Capacity>=e.Ids.Length && e.BufferBytes>0;}
    void Release(Owner o){
        if(o.Released)return;o.Released=true;o.Entry.ReleaseAttempted=true;Report.ReleaseAttempts++;
        try{var r=api.Close(o.Handle);o.Entry.ReleaseSucceeded=r!=null && r.Succeeded && r.Error==0;o.Entry.ReleaseError=r==null?-1:r.Error;}
        catch(Exception){o.Entry.ReleaseError=-1;}
        if(!o.Entry.ReleaseSucceeded){Report.ReleaseFailures++;Fail("close",o.Entry.ReleaseError);}
    }
    OrdinaryRead Read(Owner owner,uint wait,long deadline,bool termination){
        var unknown=new OrdinaryRead{RunIdentity=Report.RunIdentity,EntryId=owner.Entry.EntryId,TerminationRequested=termination,SkipReason="budget"};
        if(!Budget(deadline))return unknown;
        try{
            var read=api.Read(owner.Handle,wait,deadline);
            if(read==null){Fail("read-null",-1);return unknown;}
            if((!String.IsNullOrEmpty(read.RunIdentity) && read.RunIdentity!=Report.RunIdentity) || (read.EntryId!=0 && read.EntryId!=owner.Entry.EntryId)){
                Report.Ambiguous=true;Fail("read-identity",-1);return unknown;
            }
            read.RunIdentity=Report.RunIdentity;read.EntryId=owner.Entry.EntryId;read.SameHandleVerified=true;read.TerminationRequested=termination;
            if(!read.Attempted || read.Error!=0 || !read.ExitCodeRead || (read.State!="running" && read.State!="exited") || read.Signaled!=(read.State=="exited") || read.Frequency!=clock.Frequency || read.StartTimestamp<0 || read.EndTimestamp<read.StartTimestamp){Fail("read",read.Error);}
            var prior=owner.Entry.Initial;
            if(prior!=null && prior.State=="exited" && (read.State!="exited" || read.ExitCode!=prior.ExitCode)){Report.Ambiguous=true;Fail("read-regression",-1);}
            Budget(deadline);return read;
        }catch(Exception){Fail("read-exception",-1);return unknown;}
    }
    public void Poll(long deadlineTicks){
        // These are facts for this invocation only, not the last successful poll.
        LastMembers=new ulong[0];CurrentPollQuery=null;
        if(disposed || Report.Sealed || Report.Incomplete)return;
        try{
            if(!Budget(deadlineTicks))return;
            var list=api.QueryMembers(job,deadlineTicks);Report.QueryCount++;
            Report.LastQuery=CopyQuery(list);if(Report.FirstQuery==null)Report.FirstQuery=CopyQuery(list);
            CurrentPollQuery=Report.LastQuery;
            if(!Budget(deadlineTicks))return;
            if(!ValidList(list)){Fail("list",list==null?-1:list.Error);return;}
            LastMembers=(ulong[])list.Ids.Clone();
            foreach(ulong id in list.Ids){
                if(id==root)continue;
                if(id==0 || id>UInt32.MaxValue){Fail("candidate",-1);break;}
                if(seen.Contains(id))continue;
                if(Report.Attempts>=32){Report.CapExhausted=true;Fail("cap",-1);break;}
                if(!Budget(deadlineTicks))break;
                seen.Add(id);
                var entry=new OrdinaryEntry{RunIdentity=Report.RunIdentity,EntryId=Report.Attempts+1,Pid=id,SourceQuery=Report.QueryCount,SourceQueryEvidence=CopyQuery(list)};
                Report.Attempts++;Report.Entries.Add(entry);Owner temporary=null;
                try{
                    var open=api.Open((uint)id,0x101000,false);entry.Open=Copy(open);
                    if(open!=null && open.Handle!=IntPtr.Zero && open.Handle!=new IntPtr(-1))temporary=new Owner{Handle=open.Handle,Entry=entry};
                    if(!ValidApi(open) || temporary==null){Fail("open",open==null?-1:open.Error);break;}
                    if(!Budget(deadlineTicks))break;
                    var membership=api.IsOwned(temporary.Handle,job);entry.Membership=membership;
                    if(!ValidApi(membership) || !membership.Owned){Fail("membership",membership==null?-1:membership.Error);break;}
                    if(!Budget(deadlineTicks))break;
                    var owned=temporary;owners.Add(owned);temporary=null;Report.Admitted++;
                    entry.Initial=Read(owned,0,deadlineTicks,false);
                }finally{if(temporary!=null)Release(temporary);}
                if(Report.Attempts==32){Report.CapExhausted=true;Fail("cap",-1);}
                if(Report.Incomplete)break;
            }
        }catch(Exception){Fail("admission-exception",-1);}
    }
    public void ObserveFinal(long naturalDeadlineTicks,bool terminationRequested){
        if(disposed || Report.Sealed || Report.Incomplete)return;
        foreach(var owner in owners){
            if(owner.Entry.Final!=null)continue;
            if(!Budget(naturalDeadlineTicks))break;
            long left=naturalDeadlineTicks-clock.Ticks;
            uint wait=(uint)Math.Min(500,Math.Max(0,left*1000.0/clock.Frequency));
            owner.Entry.Final=Read(owner,owner.Entry.Initial!=null && owner.Entry.Initial.State=="exited"?0:wait,naturalDeadlineTicks,terminationRequested);
            if(terminationRequested)Fail("post-termination",-1);
            var final=owner.Entry.Final;
            if(!final.Attempted || !final.Signaled || final.State!="exited" || !final.ExitCodeRead || final.Error!=0)Fail("final-exit",final.Error);
            if(Report.Incomplete)break;
        }
    }
    public void SealAndRelease(){Report.Sealed=true;Dispose();}
    public void Dispose(){if(disposed)return;disposed=true;Report.TeardownAttempted=true;foreach(var owner in owners)Release(owner);}
}
