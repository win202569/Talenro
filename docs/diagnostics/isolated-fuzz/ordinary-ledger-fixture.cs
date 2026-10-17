using System;
using System.Collections.Generic;
// Test-only boundary. No P/Invoke, process creation, sleeps or real handles.
public sealed class OrdinaryFakeClock : IRetainedLifetimeClock {
    public long Now=100, Hertz=1000;
    public long Ticks {get{return Now;}} public long Frequency {get{return Hertz;}}
    public string UtcNow {get{return "2026-10-17T00:00:00Z";}}
}
public sealed class OrdinaryFakeApi : IOrdinaryLedgerApi {
    public OrdinaryFakeClock Clock;
    public ulong[] Candidates=new ulong[]{2};
    public Queue<ulong[]> Lists=new Queue<ulong[]>();
    public int OpenError,MembershipError,ReadError,OpenCalls,MembershipCalls,ReleaseCalls,ReadCalls;
    public bool Owned=true,ThrowClose,CloseFails,Partial,Exited,MissingExit,MissingImage,ThrowRead;
    public long AdvanceOnOpen,AdvanceOnQuery,AdvanceOnRead;
    public uint Total=2,Active,LimitTerminated;
    public string ReadRunIdentity;
    public readonly List<string> Calls=new List<string>();
    public readonly List<uint> Waits=new List<uint>();
    public readonly HashSet<long> Closed=new HashSet<long>();
    T Stamp<T>(T value) where T:LifetimeApiEvidence {value.StartTimestamp=Clock.Now;value.EndTimestamp=Clock.Now;value.Frequency=Clock.Frequency;value.StartUTC=Clock.UtcNow;value.EndUTC=Clock.UtcNow;return value;}
    public LifetimeCandidates QueryMembers(IntPtr job,long deadline){Calls.Add("list");Clock.Now+=AdvanceOnQuery;var ids=Lists.Count>0?Lists.Dequeue():Candidates;int capacity=Math.Max(32,ids.Length);return Stamp(new LifetimeCandidates{Succeeded=!Partial,Error=Partial?234:0,Ids=ids,Assigned=(uint)ids.Length,Returned=(uint)ids.Length,Capacity=capacity,BufferBytes=8+8*capacity});}
    public LifetimeOpen Open(uint pid,uint access,bool inherit){if(access!=0x101000 || inherit)throw new Exception("rights changed");Calls.Add("open:"+pid);OpenCalls++;Clock.Now+=AdvanceOnOpen;return Stamp(new LifetimeOpen{Succeeded=OpenError==0,Error=OpenError,Handle=OpenError==0?new IntPtr(900000+pid):IntPtr.Zero});}
    public LifetimeMembership IsOwned(IntPtr handle,IntPtr job){Calls.Add("membership");MembershipCalls++;return Stamp(new LifetimeMembership{Succeeded=MembershipError==0,Error=MembershipError,Owned=Owned});}
    public OrdinaryRead Read(IntPtr handle,uint waitMs,long deadline){Calls.Add("read:"+handle.ToInt64());ReadCalls++;Waits.Add(waitMs);if(ThrowRead)throw new Exception("read fixture");long start=Clock.Now;Clock.Now+=AdvanceOnRead;return new OrdinaryRead{RunIdentity=ReadRunIdentity,Attempted=true,Signaled=Exited,State=ReadError!=0?"unknown":Exited?"exited":"running",Error=ReadError,ExitCodeRead=!MissingExit,ExitCode=Exited?0u:259u,ImageError=MissingImage?31:0,ImagePath=MissingImage?null:"fixture.exe",StartTimestamp=start,EndTimestamp=Clock.Now,Frequency=Clock.Frequency,StartUTC=Clock.UtcNow,EndUTC=Clock.UtcNow};}
    public LifetimeRelease Close(IntPtr handle){Calls.Add("close:"+handle.ToInt64());ReleaseCalls++;if(!Closed.Add(handle.ToInt64()))throw new Exception("double close");if(ThrowClose)throw new Exception("close fixture");return new LifetimeRelease{Succeeded=!CloseFails,Error=CloseFails?6:0};}
    public OrdinaryAccounting QueryAccounting(IntPtr job,long deadline){Calls.Add("accounting");return Stamp(new OrdinaryAccounting{Succeeded=true,TotalProcesses=Total,ActiveProcesses=Active,TotalTerminatedProcesses=LimitTerminated});}
}
public sealed class OrdinaryLedgerFixture {
    public OrdinaryFakeClock Clock;public OrdinaryFakeApi Api;public OrdinaryMemberLedger Ledger;
    public static OrdinaryLedgerFixture Create(){var f=new OrdinaryLedgerFixture();f.Clock=new OrdinaryFakeClock();f.Api=new OrdinaryFakeApi{Clock=f.Clock};f.Ledger=new OrdinaryMemberLedger("run-a",new IntPtr(1),1,f.Api,f.Clock);return f;}
}
