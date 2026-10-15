using System;
using System.Collections;
using System.Collections.Generic;
using System.Diagnostics;
using System.Runtime.InteropServices;
using System.Text;
using System.Threading;

// Own-process diagnostic job only: suspended launch prevents child escape before
// assignment; whitelisted inherited handles prevent leaking unrelated handles.
public static class IsolatedFuzzProcess {
    [StructLayout(LayoutKind.Sequential)] struct Security { public int Size; public IntPtr Descriptor; public int Inherit; }
    [StructLayout(LayoutKind.Sequential, CharSet=CharSet.Unicode)] struct Startup {
        public int Size; public string Reserved,Desktop,Title;
        public uint X,Y,XSize,YSize,XChars,YChars,Fill,Flags;
        public ushort Show,ReservedSize; public IntPtr ReservedPtr,Input,Output,Error;
    }
    [StructLayout(LayoutKind.Sequential)] struct StartupEx {public Startup Info;public IntPtr Attributes;}
    [StructLayout(LayoutKind.Sequential)] struct ProcessInfo {public IntPtr Process,Thread;public uint Id,ThreadId;}
    [StructLayout(LayoutKind.Sequential)] struct BasicLimit {
        public long ProcessTime,JobTime; public uint Flags; public UIntPtr MinWorking,MaxWorking;
        public uint Active; public UIntPtr Affinity; public uint Priority,Scheduling;
    }
    [StructLayout(LayoutKind.Sequential)] struct IoCounts {public ulong ReadOps,WriteOps,OtherOps,ReadBytes,WriteBytes,OtherBytes;}
    [StructLayout(LayoutKind.Sequential)] struct ExtendedLimit {public BasicLimit Basic;public IoCounts Io;public UIntPtr ProcessMemory,JobMemory,PeakProcess,PeakJob;}
    [StructLayout(LayoutKind.Sequential)] struct Accounting {public long User,Kernel,PeriodUser,PeriodKernel;public uint Faults,Total,Active,Terminated;}
    [DllImport("kernel32.dll",SetLastError=true)] static extern IntPtr CreateJobObject(IntPtr attributes,string name);
    [DllImport("kernel32.dll",SetLastError=true)] static extern bool SetInformationJobObject(IntPtr job,int cls,ref ExtendedLimit info,int size);
    [DllImport("kernel32.dll",SetLastError=true)] static extern bool QueryInformationJobObject(IntPtr job,int cls,out Accounting info,int size,IntPtr length);
    [DllImport("kernel32.dll",SetLastError=true)] static extern bool QueryInformationJobObject(IntPtr job,int cls,IntPtr info,int size,IntPtr length);
    [DllImport("kernel32.dll",SetLastError=true)] static extern bool AssignProcessToJobObject(IntPtr job,IntPtr process);
    [DllImport("kernel32.dll",SetLastError=true)] static extern bool TerminateJobObject(IntPtr job,uint exitCode);
    [DllImport("kernel32.dll",SetLastError=true)] static extern bool TerminateProcess(IntPtr process,uint exitCode);
    [DllImport("kernel32.dll",CharSet=CharSet.Unicode,SetLastError=true)] static extern IntPtr CreateFile(string name,uint access,uint share,ref Security security,uint disposition,uint flags,IntPtr template);
    [DllImport("kernel32.dll",SetLastError=true)] static extern bool InitializeProcThreadAttributeList(IntPtr list,int count,int flags,ref IntPtr size);
    [DllImport("kernel32.dll",SetLastError=true)] static extern bool UpdateProcThreadAttribute(IntPtr list,uint flags,IntPtr attr,IntPtr value,IntPtr size,IntPtr previous,IntPtr returned);
    [DllImport("kernel32.dll")] static extern void DeleteProcThreadAttributeList(IntPtr list);
    [DllImport("kernel32.dll",CharSet=CharSet.Unicode,SetLastError=true)] static extern bool CreateProcess(string application,StringBuilder command,IntPtr pa,IntPtr ta,bool inherit,uint flags,IntPtr env,string cwd,ref StartupEx startup,out ProcessInfo process);
    [DllImport("kernel32.dll",SetLastError=true)] static extern uint ResumeThread(IntPtr thread);
    [DllImport("kernel32.dll",SetLastError=true)] static extern uint WaitForSingleObject(IntPtr handle,uint milliseconds);
    [DllImport("kernel32.dll",SetLastError=true)] static extern bool GetExitCodeProcess(IntPtr process,out uint exitCode);
    [DllImport("kernel32.dll",SetLastError=true)] static extern IntPtr OpenProcess(uint access,bool inherit,uint id);
    [DllImport("kernel32.dll",SetLastError=true)] static extern bool IsProcessInJob(IntPtr process,IntPtr job,out bool owned);
    [DllImport("kernel32.dll",CharSet=CharSet.Unicode,SetLastError=true)] static extern bool QueryFullProcessImageName(IntPtr process,uint flags,StringBuilder image,ref uint length);
    [DllImport("kernel32.dll")] static extern bool CloseHandle(IntPtr handle);
    static readonly IntPtr Invalid=new IntPtr(-1);
    static void Check(bool ok) {if(!ok)throw new System.ComponentModel.Win32Exception(Marshal.GetLastWin32Error());}
    static void Close(IntPtr h) {if(h!=IntPtr.Zero && h!=Invalid)CloseHandle(h);}
    static string EnvironmentString(object value) {
        if(value==null)return "";
        for(int i=0;i<4 && value.GetType().FullName=="System.Management.Automation.PSObject";i++)
            value=value.GetType().GetProperty("BaseObject").GetValue(value,null);
        string text=value as string;
        if(text==null)throw new ArgumentException("Environment values must be strings");
        return text;
    }
    static string Quote(string value) {
        var b=new StringBuilder("\""); int slashes=0;
        foreach(char c in value) {
            if(c=='\\'){slashes++;continue;}
            if(c=='\"'){b.Append('\\',slashes*2+1);b.Append(c);slashes=0;continue;}
            b.Append('\\',slashes);slashes=0;b.Append(c);
        }
        b.Append('\\',slashes*2);b.Append('"');return b.ToString();
    }
    static ulong[] Members(IntPtr job) {return MembersWithEvidence(job,null);}
    static ulong[] MembersWithEvidence(IntPtr job,List<MemberListQueryObservation> evidence) {
        int capacity=16;
        for(int attempt=0;attempt<8;attempt++) {
            int size=checked(8+capacity*IntPtr.Size);
            IntPtr buffer=Marshal.AllocHGlobal(size);
            try {
                // Initialize the header so a failed query never reads garbage.
                Marshal.WriteInt32(buffer,0,0);Marshal.WriteInt32(buffer,4,0);
                var query=evidence==null?null:new MemberListQueryObservation{API="QueryInformationJobObject/ProcessIdList",Attempt=attempt+1,Capacity=capacity,BufferBytes=size,StartUTC=DateTime.UtcNow.ToString("o"),StartTimestamp=Stopwatch.GetTimestamp()};
                if(query!=null)evidence.Add(query);
                bool ok=QueryInformationJobObject(job,3,buffer,size,IntPtr.Zero);
                int error=ok?0:Marshal.GetLastWin32Error();
                if(query!=null){query.EndTimestamp=Stopwatch.GetTimestamp();query.EndUTC=DateTime.UtcNow.ToString("o");query.Succeeded=ok;query.Error=error;}
                if(!ok && error!=234)throw new System.ComponentModel.Win32Exception(error);
                uint assigned=unchecked((uint)Marshal.ReadInt32(buffer,0));
                uint count=unchecked((uint)Marshal.ReadInt32(buffer,4));
                if(query!=null){query.HeaderRead=true;query.Assigned=assigned;query.Returned=count;}
                if(ok && assigned==count && count<=capacity) {
                    var ids=new ulong[count];
                    for(int i=0;i<ids.Length;i++)ids[i]=unchecked((ulong)Marshal.ReadIntPtr(buffer,8+i*IntPtr.Size).ToInt64());
                    return ids;
                }
                if(assigned>65536)throw new InvalidOperationException("Job member list exceeds diagnostic bound");
                capacity=checked(Math.Max(capacity*2,(int)assigned));
            } finally {Marshal.FreeHGlobal(buffer);}
        }
        throw new InvalidOperationException("Job member list could not be verified");
    }
    public static uint CountDescendants(uint root,ulong[] members) {
        uint count=0;foreach(ulong id in members)if(id!=root)count++;return count;
    }
    public class ShutdownObservation {public bool Attempted,ExitCodeRead,SameHandleVerified;public ulong MemberId;public uint ExitCode;public int QueryError;public long WaitBudgetMs,StartElapsedMs,EndElapsedMs;public string State="unknown",SkipReason,StartUTC,EndUTC;}
    public class MemberQueryObservation {public string API,StartUTC,EndUTC;public long StartTimestamp,EndTimestamp,Frequency=Stopwatch.Frequency;public bool Succeeded;public int Error;}
    public class AccountingQueryObservation:MemberQueryObservation {public bool ValuesRead;public uint TotalProcesses,ActiveProcesses,TotalTerminatedProcesses;}
    public class MemberListQueryObservation:MemberQueryObservation {public int Attempt,Capacity,BufferBytes;public bool HeaderRead;public uint Assigned,Returned;}
    static Accounting ReadInitialJobQueries(IntPtr job,Result result) {
        var query=new AccountingQueryObservation{API="QueryInformationJobObject/Accounting",StartUTC=DateTime.UtcNow.ToString("o"),StartTimestamp=Stopwatch.GetTimestamp()};
        result.InitialAccountingQuery=query;
        Accounting accounting;
        bool ok=QueryInformationJobObject(job,1,out accounting,Marshal.SizeOf(typeof(Accounting)),IntPtr.Zero);
        int error=ok?0:Marshal.GetLastWin32Error();
        query.EndTimestamp=Stopwatch.GetTimestamp();query.EndUTC=DateTime.UtcNow.ToString("o");query.Succeeded=ok;query.Error=error;
        if(!ok)throw new System.ComponentModel.Win32Exception(error);
        query.ValuesRead=true;query.TotalProcesses=accounting.Total;query.ActiveProcesses=accounting.Active;query.TotalTerminatedProcesses=accounting.Terminated;
        result.MembersAtRootWait=MembersWithEvidence(job,result.InitialMemberQueries);
        return accounting;
    }
    public class MemberObservation {public ulong Id;public bool Owned,ExitCodeRead,InitialExitCodeRead,ConfirmationAttempted;public uint ExitCode,InitialExitCode;public int QueryError,ImageError;public string ImagePath,State,ObservedUTC,CoreObservedUTC,InitialState,ConfirmationState,ConfirmationUTC;public ShutdownObservation DiagnosticShutdown,FinalShutdown;public MemberQueryObservation ProcessOpen,JobMembership;}
    public static void VerifyMemberFinalExit(IntPtr handle,MemberObservation row,Stopwatch shutdownClock) {
        if(row==null || shutdownClock==null || !shutdownClock.IsRunning || row.FinalShutdown!=null || row.DiagnosticShutdown!=null)
            throw new InvalidOperationException("Final exit requires unused final evidence, never diagnostic data");
        var final=new ShutdownObservation{MemberId=row.Id,StartUTC=DateTime.UtcNow.ToString("o"),StartElapsedMs=shutdownClock.ElapsedMilliseconds};
        row.FinalShutdown=final;
        try {
            if(row.Id==0 || !row.Owned || row.QueryError!=0 || !row.ExitCodeRead || row.State!="running" || !row.ConfirmationAttempted || row.ConfirmationState!="running") {final.SkipReason="unverified-state";return;}
            long remaining=500-shutdownClock.ElapsedMilliseconds;
            if(remaining<=0) {final.SkipReason="window-expired";return;}
            if(handle==IntPtr.Zero || handle==Invalid) {final.SkipReason="invalid-handle";return;}
            final.SameHandleVerified=true;final.WaitBudgetMs=remaining;final.Attempted=true;
            // Still the retained job-verified handle; preserve all prior snapshots.
            // Every member shares the one shutdown clock, including this wait.
            uint status=WaitForSingleObject(handle,(uint)remaining);Check(status!=UInt32.MaxValue);
            final.State=status==0?"exited":status==258?"running":"unknown";
            if(status==0) {uint code;Check(GetExitCodeProcess(handle,out code));final.ExitCode=code;final.ExitCodeRead=true;}
        } catch(System.ComponentModel.Win32Exception e) {final.QueryError=e.NativeErrorCode==0?-1:e.NativeErrorCode;final.State="unknown";}
        finally {final.EndElapsedMs=shutdownClock.ElapsedMilliseconds;final.EndUTC=DateTime.UtcNow.ToString("o");}
    }
    public static bool IsFinalShutdownStateVerified(MemberObservation row) {
        if(row==null || row.Id==0 || !row.Owned || row.QueryError!=0 || !row.ExitCodeRead || row.DiagnosticShutdown!=null)return false;
        if(row.State=="exited")return true;
        if(row.State!="running" || !row.ConfirmationAttempted || row.ConfirmationState!="running")return false;
        var final=row.FinalShutdown;
        return final!=null && final.MemberId==row.Id && final.SameHandleVerified && final.Attempted && final.State=="exited" && final.QueryError==0 && final.ExitCodeRead && final.WaitBudgetMs>0 && final.WaitBudgetMs<=500 && final.StartElapsedMs>=0 && final.StartElapsedMs<500 && final.EndElapsedMs>=final.StartElapsedMs && final.EndElapsedMs<=500 && final.StartElapsedMs+final.WaitBudgetMs<=500;
    }
    public static void ObserveMemberShutdown(IntPtr handle,MemberObservation row,Stopwatch shutdownClock) {
        if(row==null || shutdownClock==null || !shutdownClock.IsRunning || row.DiagnosticShutdown!=null)
            throw new InvalidOperationException("Diagnostic observation requires one unused row and the shared running clock");
        var observation=new ShutdownObservation{StartUTC=DateTime.UtcNow.ToString("o"),StartElapsedMs=shutdownClock.ElapsedMilliseconds};
        row.DiagnosticShutdown=observation;
        try {
            if(!row.Owned || row.QueryError!=0 || !row.ExitCodeRead || row.State!="running" || !row.ConfirmationAttempted || row.ConfirmationState!="running") {observation.SkipReason="unverified-state";return;}
            if(row.ImageError==0 && !String.IsNullOrEmpty(row.ImagePath)) {observation.SkipReason="not-needed";return;}
            long remaining=500-shutdownClock.ElapsedMilliseconds;
            if(remaining<=0) {observation.SkipReason="window-expired";return;}
            if(handle==IntPtr.Zero || handle==Invalid) {observation.SkipReason="invalid-handle";return;}
            observation.WaitBudgetMs=remaining;observation.Attempted=true;
            // This is diagnostic evidence only. Never mutate core/confirm reads,
            // give each member a fresh clock, reopen a PID or repeat this wait.
            uint status=WaitForSingleObject(handle,(uint)remaining);Check(status!=UInt32.MaxValue);
            observation.State=status==0?"exited":status==258?"running":"unknown";
            if(status==0) {uint code;Check(GetExitCodeProcess(handle,out code));observation.ExitCode=code;observation.ExitCodeRead=true;}
        } catch(System.ComponentModel.Win32Exception e) {observation.QueryError=e.NativeErrorCode==0?-1:e.NativeErrorCode;observation.State="unknown";}
        finally {observation.EndElapsedMs=shutdownClock.ElapsedMilliseconds;observation.EndUTC=DateTime.UtcNow.ToString("o");}
    }
    public static void ConfirmMemberExit(IntPtr handle,MemberObservation row) {
        // Caller retains the same handle already verified against its live job.
        // Never reopen a PID or turn this single zero-wait into a retry loop.
        if(handle==IntPtr.Zero || handle==Invalid || row==null || !row.Owned || row.QueryError!=0 || !row.ExitCodeRead || row.ConfirmationAttempted)
            throw new InvalidOperationException("Member exit confirmation requires an unconfirmed verified handle");
        row.InitialState=row.State;row.InitialExitCode=row.ExitCode;row.InitialExitCodeRead=row.ExitCodeRead;
        row.ConfirmationAttempted=true;
        try {
            uint status=WaitForSingleObject(handle,0);Check(status!=UInt32.MaxValue);
            row.ConfirmationState=status==0?"exited":status==258?"running":"unknown";
            row.State=row.ConfirmationState;
            if(status==0) {
                // The earlier exit-code read may predate the signaled state.
                row.ExitCodeRead=false;
                uint code;Check(GetExitCodeProcess(handle,out code));row.ExitCode=code;row.ExitCodeRead=true;
            }
        } catch(System.ComponentModel.Win32Exception e) {row.QueryError=e.NativeErrorCode==0?-1:e.NativeErrorCode;row.State="unknown";}
        finally {row.ConfirmationUTC=DateTime.UtcNow.ToString("o");}
    }
    public static bool IsShutdownStateVerified(MemberObservation row) {
        if(row==null || !row.Owned || row.QueryError!=0 || !row.ExitCodeRead)return false;
        // Exited means the exact owned handle was signaled, not merely code0.
        // Image-name loss after exit remains metadata loss, never an exemption
        // for an unknown or still-running process.
        if(row.State=="exited")return true;
        return row.State=="running" && row.ImageError==0 && !String.IsNullOrEmpty(row.ImagePath);
    }
    static MemberObservation[] InspectMembers(IntPtr job,ulong[] members,Stopwatch diagnosticClock=null,Stopwatch finalClock=null) {
        var observations=new List<MemberObservation>();
        var handles=new List<IntPtr>();
        try {
        // Evidence gathering has a fixed cap; it never changes member counting.
        for(int i=0;i<Math.Min(members.Length,32);i++) {
            var row=new MemberObservation{Id=members[i],State="unknown"};
            IntPtr handle=IntPtr.Zero;
            try {
                if(row.Id==0 || row.Id>UInt32.MaxValue)throw new InvalidOperationException("Invalid live job member ID");
                var open=new MemberQueryObservation{API="OpenProcess",StartUTC=DateTime.UtcNow.ToString("o"),StartTimestamp=Stopwatch.GetTimestamp()};
                row.ProcessOpen=open;
                handle=OpenProcess(0x101000,false,(uint)row.Id); // QUERY_LIMITED_INFORMATION | SYNCHRONIZE only.
                // Capture last-error before timing or serialization can overwrite it.
                int openError=handle!=IntPtr.Zero?0:Marshal.GetLastWin32Error();
                open.EndTimestamp=Stopwatch.GetTimestamp();open.EndUTC=DateTime.UtcNow.ToString("o");open.Succeeded=handle!=IntPtr.Zero;open.Error=openError;
                if(!open.Succeeded)throw new System.ComponentModel.Win32Exception(openError);
                var membership=new MemberQueryObservation{API="IsProcessInJob",StartUTC=DateTime.UtcNow.ToString("o"),StartTimestamp=Stopwatch.GetTimestamp()};
                row.JobMembership=membership;
                bool owned;bool membershipOK=IsProcessInJob(handle,job,out owned);
                int membershipError=membershipOK?0:Marshal.GetLastWin32Error();
                membership.EndTimestamp=Stopwatch.GetTimestamp();membership.EndUTC=DateTime.UtcNow.ToString("o");membership.Succeeded=membershipOK;membership.Error=membershipError;
                if(!membershipOK)throw new System.ComponentModel.Win32Exception(membershipError);
                row.Owned=owned;
                if(!owned){row.State="ownership-lost";continue;}
                // Identity/state are read only after this exact handle is verified
                // against our live job. Never query a PID from an archived receipt.
                uint status=WaitForSingleObject(handle,0);Check(status!=UInt32.MaxValue);
                row.State=status==0?"exited":status==258?"running":"unknown";
                uint code;Check(GetExitCodeProcess(handle,out code));row.ExitCode=code;row.ExitCodeRead=true;
                var image=new StringBuilder(32768);uint length=(uint)image.Capacity;
                if(QueryFullProcessImageName(handle,0,image,ref length))row.ImagePath=image.ToString();
                else row.ImageError=Marshal.GetLastWin32Error();
                row.InitialState=row.State;row.InitialExitCode=row.ExitCode;row.InitialExitCodeRead=row.ExitCodeRead;
                if(row.State=="running")ConfirmMemberExit(handle,row);
                row.CoreObservedUTC=DateTime.UtcNow.ToString("o");
            } catch(System.ComponentModel.Win32Exception e) {row.QueryError=e.NativeErrorCode;}
            finally {row.ObservedUTC=DateTime.UtcNow.ToString("o");observations.Add(row);handles.Add(handle);}
        }
        // Acquire/verify the entire bounded batch before waiting on any member.
        // Otherwise a later PID may disappear while the first handle is waited.
        for(int i=0;i<observations.Count;i++) {
            var row=observations[i];
            try {
                if(row.CoreObservedUTC!=null) {
                    if(diagnosticClock!=null)ObserveMemberShutdown(handles[i],row,diagnosticClock);
                    else if(finalClock!=null && row.State=="running")VerifyMemberFinalExit(handles[i],row,finalClock);
                }
            } finally {row.ObservedUTC=DateTime.UtcNow.ToString("o");}
        }
        return observations.ToArray();
        } finally {foreach(IntPtr handle in handles)Close(handle);}
    }
    public class LaunchBoundary {public string Name,UTC;public long Timestamp,ElapsedTicks,Frequency;}
    public class Result {public long ExitCode,ShutdownElapsedMs,NaturalObservationMs;public int ShutdownSchema=3;public string ShutdownStatePolicy="owned-final-signaled-exit-image-optional-v2";public bool TimedOut,CleanupVerified,RootExited,TerminationRequested,MemberEvidenceTruncated,LiveSampling,NaturalShutdownVerified,ShutdownEvidenceComplete,ShutdownImageEvidenceComplete,ShutdownDiagnosticOnly;public uint RootId,DescendantsRemaining,ActiveAtRootWait,ActiveBeforeTermination,ActiveAfterCleanup,TotalProcesses;public ulong[] MembersAtRootWait,MembersBeforeTermination,MembersAfterCleanup;public MemberObservation[] MemberEvidence,MembersDuringRun,ShutdownMemberEvidence;public string StartUTC,EndUTC;public List<LaunchBoundary> LaunchBoundaries=new List<LaunchBoundary>();public AccountingQueryObservation InitialAccountingQuery;public List<MemberListQueryObservation> InitialMemberQueries=new List<MemberListQueryObservation>();}
    public static void RecordLaunchBoundary(Result result,string name,Stopwatch watch) {
        result.LaunchBoundaries.Add(new LaunchBoundary{Name=name,UTC=DateTime.UtcNow.ToString("o"),Timestamp=Stopwatch.GetTimestamp(),ElapsedTicks=watch.ElapsedTicks,Frequency=Stopwatch.Frequency});
    }
    public static Result Run(string executable,string[] args,IDictionary overrides,string cwd,string stdout,string stderr,int limitSeconds,bool captureLiveMembers) {
        return Run(executable,args,overrides,cwd,stdout,stderr,limitSeconds,captureLiveMembers,false);
    }
    public static Result Run(string executable,string[] args,IDictionary overrides,string cwd,string stdout,string stderr,int limitSeconds,bool captureLiveMembers,bool captureShutdownDiagnostic) {
        IntPtr job=IntPtr.Zero,output=IntPtr.Zero,error=IntPtr.Zero,input=IntPtr.Zero,attributes=IntPtr.Zero,handles=IntPtr.Zero,environment=IntPtr.Zero;
        ProcessInfo pi=new ProcessInfo();bool attributesReady=false,assigned=false;
        var result=new Result{StartUTC=DateTime.UtcNow.ToString("o")};
        result.ShutdownDiagnosticOnly=captureShutdownDiagnostic;
        if(captureShutdownDiagnostic) {result.ShutdownSchema=2;result.ShutdownStatePolicy="owned-signaled-exit-image-optional-v1";}
        var watch=Stopwatch.StartNew();
        try {
            job=CreateJobObject(IntPtr.Zero,null);Check(job!=IntPtr.Zero);
            var limits=new ExtendedLimit();limits.Basic.Flags=0x2000; // KILL_ON_JOB_CLOSE, no breakaway permission.
            Check(SetInformationJobObject(job,9,ref limits,Marshal.SizeOf(typeof(ExtendedLimit))));
            var sa=new Security{Size=Marshal.SizeOf(typeof(Security)),Inherit=1};
            output=CreateFile(stdout,0x40000000,3,ref sa,2,0x80,IntPtr.Zero);Check(output!=Invalid);
            error=CreateFile(stderr,0x40000000,3,ref sa,2,0x80,IntPtr.Zero);Check(error!=Invalid);
            input=CreateFile("NUL",0x80000000,3,ref sa,3,0x80,IntPtr.Zero);Check(input!=Invalid);
            IntPtr attributeSize=IntPtr.Zero;
            InitializeProcThreadAttributeList(IntPtr.Zero,1,0,ref attributeSize);
            Check(attributeSize!=IntPtr.Zero);
            attributes=Marshal.AllocHGlobal(attributeSize);Check(InitializeProcThreadAttributeList(attributes,1,0,ref attributeSize));attributesReady=true;
            handles=Marshal.AllocHGlobal(IntPtr.Size*3);
            Marshal.WriteIntPtr(handles,0,input);Marshal.WriteIntPtr(handles,IntPtr.Size,output);Marshal.WriteIntPtr(handles,IntPtr.Size*2,error);
            Check(UpdateProcThreadAttribute(attributes,0,new IntPtr(0x20002),handles,new IntPtr(IntPtr.Size*3),IntPtr.Zero,IntPtr.Zero));
            var env=new SortedDictionary<string,string>(StringComparer.OrdinalIgnoreCase);
            foreach(DictionaryEntry row in Environment.GetEnvironmentVariables())env[(string)row.Key]=(string)row.Value;
            foreach(DictionaryEntry row in overrides) {
                string key=EnvironmentString(row.Key),value=EnvironmentString(row.Value);
                if(key.Length==0 || key.IndexOf('=')>=0 || key.IndexOf('\0')>=0 || value.IndexOf('\0')>=0)throw new ArgumentException("Invalid environment entry");
                env[key]=value;
            }
            var block=new StringBuilder();foreach(var row in env)block.Append(row.Key).Append('=').Append(row.Value).Append('\0');block.Append('\0');
            environment=Marshal.StringToHGlobalUni(block.ToString());
            var command=new StringBuilder(Quote(executable));foreach(string arg in args)command.Append(' ').Append(Quote(arg));
            var startup=new StartupEx();startup.Info.Size=Marshal.SizeOf(typeof(StartupEx));startup.Info.Flags=0x101;startup.Info.Show=0;
            startup.Info.Input=input;startup.Info.Output=output;startup.Info.Error=error;startup.Attributes=attributes;
            RecordLaunchBoundary(result,"BeforeCreateProcess",watch);
            Check(CreateProcess(executable,command,IntPtr.Zero,IntPtr.Zero,true,0x08080404,environment,cwd,ref startup,out pi));
            RecordLaunchBoundary(result,"AfterCreateProcess",watch);
            result.RootId=pi.Id;
            RecordLaunchBoundary(result,"BeforeAssignProcessToJobObject",watch);
            Check(AssignProcessToJobObject(job,pi.Process));assigned=true;
            RecordLaunchBoundary(result,"AfterAssignProcessToJobObject",watch);
            RecordLaunchBoundary(result,"BeforeResumeThread",watch);
            Check(ResumeThread(pi.Thread)!=UInt32.MaxValue);
            RecordLaunchBoundary(result,"AfterResumeThread",watch);
            bool rootExited=false;Accounting accounting;
            var live=new List<MemberObservation>();var seen=new HashSet<ulong>();
            result.LiveSampling=captureLiveMembers;
            while(true) {
                double left=limitSeconds*1000-watch.Elapsed.TotalMilliseconds;
                if(left<=0){result.TimedOut=true;break;}
                if(captureLiveMembers && live.Count<32) {
                    var fresh=new List<ulong>();
                    foreach(ulong id in Members(job))if(!seen.Contains(id) && fresh.Count+live.Count<32){seen.Add(id);fresh.Add(id);}
                    live.AddRange(InspectMembers(job,fresh.ToArray()));
                    left=limitSeconds*1000-watch.Elapsed.TotalMilliseconds;
                    if(left<=0){result.TimedOut=true;break;}
                }
                uint status=WaitForSingleObject(pi.Process,(uint)Math.Min(captureLiveMembers?10:100,Math.Ceiling(left)));
                Check(status!=UInt32.MaxValue);
                Check(QueryInformationJobObject(job,1,out accounting,Marshal.SizeOf(typeof(Accounting)),IntPtr.Zero));
                if(status==0){rootExited=true;break;}
            }
            // One shared five-second shutdown clock includes evidence gathering,
            // natural observation (at most500ms) and any forced-job cleanup.
            var cleanup=Stopwatch.StartNew();
            accounting=ReadInitialJobQueries(job,result);
            result.TotalProcesses=accounting.Total;
            result.MembersDuringRun=live.ToArray();
            result.RootExited=rootExited;result.ActiveAtRootWait=accounting.Active;
            result.DescendantsRemaining=CountDescendants(pi.Id,result.MembersAtRootWait);
            result.MemberEvidence=InspectMembers(job,result.MembersAtRootWait,captureShutdownDiagnostic && rootExited && !result.TimedOut?cleanup:null,!captureShutdownDiagnostic && rootExited && !result.TimedOut?cleanup:null);
            result.MemberEvidenceTruncated=result.MembersAtRootWait.Length>result.MemberEvidence.Length;
            var shutdownEvidence=new List<MemberObservation>(result.MemberEvidence);
            var observed=new HashSet<ulong>(result.MembersAtRootWait);
            result.ShutdownEvidenceComplete=!result.MemberEvidenceTruncated;
            foreach(var row in shutdownEvidence)if(row.Id!=pi.Id && !(captureShutdownDiagnostic?IsShutdownStateVerified(row):IsFinalShutdownStateVerified(row)))result.ShutdownEvidenceComplete=false;
            ulong[] remaining=result.MembersAtRootWait;
            if(rootExited && !result.TimedOut && result.ShutdownEvidenceComplete) {
                while(cleanup.ElapsedMilliseconds<=500) {
                    Check(QueryInformationJobObject(job,1,out accounting,Marshal.SizeOf(typeof(Accounting)),IntPtr.Zero));
                    remaining=Members(job);
                    var fresh=new List<ulong>();
                    foreach(ulong id in remaining)if(!observed.Contains(id)){observed.Add(id);fresh.Add(id);}
                    if(fresh.Count>0) {
                        if(shutdownEvidence.Count+fresh.Count>32){result.ShutdownEvidenceComplete=false;break;}
                        var added=InspectMembers(job,fresh.ToArray(),captureShutdownDiagnostic?cleanup:null,captureShutdownDiagnostic?null:cleanup);shutdownEvidence.AddRange(added);
                        foreach(var row in added)if(row.Id!=pi.Id && !(captureShutdownDiagnostic?IsShutdownStateVerified(row):IsFinalShutdownStateVerified(row)))result.ShutdownEvidenceComplete=false;
                        if(!result.ShutdownEvidenceComplete)break;
                    }
                    if(accounting.Active==0 && remaining.Length==0 && cleanup.ElapsedMilliseconds<=500){result.NaturalShutdownVerified=true;break;}
                    long left=500-cleanup.ElapsedMilliseconds;if(left<=0)break;
                    Thread.Sleep((int)Math.Min(10,left));
                }
            }
            result.NaturalObservationMs=cleanup.ElapsedMilliseconds;
            result.ShutdownMemberEvidence=shutdownEvidence.ToArray();
            result.ShutdownImageEvidenceComplete=!result.MemberEvidenceTruncated;
            foreach(var row in shutdownEvidence)if(row.ImageError!=0 || String.IsNullOrEmpty(row.ImagePath))result.ShutdownImageEvidenceComplete=false;
            result.ActiveBeforeTermination=accounting.Active;result.MembersBeforeTermination=remaining;
            if(!result.NaturalShutdownVerified) {
                result.TerminationRequested=true;
                Check(TerminateJobObject(job,0xE001));
                do {
                    Check(QueryInformationJobObject(job,1,out accounting,Marshal.SizeOf(typeof(Accounting)),IntPtr.Zero));remaining=Members(job);
                    if(accounting.Active==0 && remaining.Length==0)break;
                    long left=5000-cleanup.ElapsedMilliseconds;if(left<=0)break;
                    Thread.Sleep((int)Math.Min(10,left));
                }while(true);
            }
            result.MembersAfterCleanup=Members(job);
            result.ActiveAfterCleanup=accounting.Active;result.CleanupVerified=accounting.Active==0 && result.MembersAfterCleanup.Length==0;
            result.ShutdownElapsedMs=cleanup.ElapsedMilliseconds;
            uint code;Check(GetExitCodeProcess(pi.Process,out code));result.ExitCode=code;
            result.EndUTC=DateTime.UtcNow.ToString("o");return result;
        } finally {
            // If job assignment failed, terminate only the still-suspended root
            // through the exact creation handle. Never enumerate/kill other PIDs.
            if(pi.Process!=IntPtr.Zero && !assigned)TerminateProcess(pi.Process,0xE002);
            Close(job);Close(pi.Thread);Close(pi.Process);Close(input);Close(error);Close(output);
            if(attributesReady)DeleteProcThreadAttributeList(attributes);
            if(attributes!=IntPtr.Zero)Marshal.FreeHGlobal(attributes);
            if(handles!=IntPtr.Zero)Marshal.FreeHGlobal(handles);
            if(environment!=IntPtr.Zero)Marshal.FreeHGlobal(environment);
        }
    }
}
