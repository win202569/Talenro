using System;
using System.Collections.Generic;
// Test-only dependency boundary: never passes a fake handle to kernel32.
public sealed class FakeLifetimeClock : IRetainedLifetimeClock {
    public long Now=100; public long Ticks {get{return Now;}}
    public long Frequency {get{return 1000;}} public string UtcNow {get{return "2026-10-15T00:00:00Z";}}
}
public sealed class FakeLifetimeApi : IRetainedLifetimeApi {
    public ulong[] Ids=new ulong[]{2}; public int OpenError,MembershipError,ReadError;
    public uint FailPid; public bool Owned=true,ThrowMembership,ThrowRead,ThrowClose,CloseFails,Partial,Exited,MissingExit,MissingImage;
    public int QueryCalls,OpenCalls,MembershipCalls,ReadCalls,ReleaseCalls,DoubleCloseCalls,ReopenedHandles,PositiveWaitCalls;
    public long AdvanceOnOpen,AdvanceOnQuery; public FakeLifetimeClock Clock;
    public readonly List<string> Calls=new List<string>();
    readonly HashSet<long> closed=new HashSet<long>(); readonly HashSet<uint> opened=new HashSet<uint>();
    public LifetimeCandidates QueryMembers(IntPtr job,long deadline) {
        QueryCalls++;Calls.Add("list");if(Clock!=null)Clock.Now+=AdvanceOnQuery;return new LifetimeCandidates{Succeeded=!Partial,Error=Partial?234:0,Ids=Ids,Assigned=(uint)Ids.Length,Returned=(uint)Ids.Length};
    }
    public LifetimeOpen Open(uint pid,uint access,bool inherit) {
        if(access!=0x101000 || inherit)throw new Exception("wrong access/inheritance");
        OpenCalls++;Calls.Add("open:"+pid);if(!opened.Add(pid))ReopenedHandles++;
        if(Clock!=null)Clock.Now+=AdvanceOnOpen;
        int e=(FailPid==0 || FailPid==pid)?OpenError:0;
        return new LifetimeOpen{Handle=e==0?new IntPtr(900000+pid):IntPtr.Zero,Succeeded=e==0,Error=e};
    }
    public LifetimeMembership IsOwned(IntPtr handle,IntPtr job) {
        MembershipCalls++;Calls.Add("membership");if(ThrowMembership)throw new Exception("membership fixture");
        return new LifetimeMembership{Succeeded=MembershipError==0,Owned=Owned,Error=MembershipError};
    }
    public LifetimeRead Read(IntPtr handle,long deadline) {
        ReadCalls++;Calls.Add("read:"+handle.ToInt64());if(ThrowRead)throw new Exception("read fixture");
        return new LifetimeRead{Attempted=true,State=ReadError!=0?"unknown":Exited?"exited":"running",Error=ReadError,ExitCodeRead=!MissingExit && ReadError==0,ExitCode=Exited?0u:259u,ImagePath=MissingImage?null:"fixture.exe",ImageError=MissingImage?5:0};
    }
    public LifetimeRelease Close(IntPtr handle) {
        ReleaseCalls++;Calls.Add("close:"+handle.ToInt64());if(!closed.Add(handle.ToInt64()))DoubleCloseCalls++;
        if(ThrowClose)throw new Exception("close fixture");return new LifetimeRelease{Succeeded=!CloseFails,Error=CloseFails?6:0};
    }
}
public static class LifetimeCoreFixture {
    public static Exception Expected;public static FakeLifetimeApi Api;
    public static void ThrowCore(){throw Expected;}
    public static void MarkCleanup(){Api.Calls.Add("cleanup");}
}
