using System;
using System.IO;
using System.Text;
// Portable test utilities only; no native calls or process handles.
public sealed class FakeCompileStageClock : IRetainedLifetimeClock {
    public long Now=1000;
    public long Ticks {get{return Now;}}
    public long Frequency {get{return 1000;}}
    public string UtcNow {get{return new DateTime(2026,10,16,0,0,0,DateTimeKind.Utc).AddMilliseconds(Now).ToString("o");}}
}
public sealed class FakeCompileStageBytes : ICompileStageByteSource {
    readonly MemoryStream data=new MemoryStream();
    public int ReadCalls,DisposeCalls,MaxRead=Int32.MaxValue;public long VisibleLength=-1;
    public bool ThrowRead,ThrowDispose;
    public long Length {get{return VisibleLength<0?data.Length:VisibleLength;}}
    public void Append(byte[] bytes){long at=data.Position;data.Position=data.Length;data.Write(bytes,0,bytes.Length);data.Position=at;}
    public int Read(byte[] buffer,int offset,int count){ReadCalls++;if(ThrowRead)throw new IOException("fixture read");return data.Read(buffer,offset,Math.Min(count,MaxRead));}
    public void Dispose(){DisposeCalls++;if(ThrowDispose)throw new IOException("fixture dispose");data.Dispose();}
    public static byte[] Padding(int length){var bytes=new byte[length];for(int i=0;i<length;i++)bytes[i]=(byte)((i%32768==32767 || i==length-1)?10:120);return bytes;}
}
public sealed class FakeCompileStageRoot:ICompileStageRootProbe {
    public FakeCompileStageClock Clock;public uint WaitResult=258;public int Error,Calls;
    public CompileStageRootCheck Check(){Calls++;return new CompileStageRootCheck{Succeeded=Error==0,Error=Error,WaitResult=WaitResult,StartTimestamp=Clock.Now,EndTimestamp=Clock.Now,Frequency=Clock.Frequency,StartUTC=Clock.UtcNow,EndUTC=Clock.UtcNow};}
}
public sealed class ClockedCompileLifetimeApi:IRetainedLifetimeApi {
    readonly FakeLifetimeApi api;readonly FakeCompileStageClock clock;
    public ClockedCompileLifetimeApi(FakeLifetimeApi api,FakeCompileStageClock clock){this.api=api;this.clock=clock;}
    void Stamp(LifetimeApiEvidence e){e.StartTimestamp=clock.Now;e.EndTimestamp=clock.Now;e.Frequency=clock.Frequency;e.StartUTC=clock.UtcNow;e.EndUTC=clock.UtcNow;}
    public LifetimeCandidates QueryMembers(IntPtr job,long deadline){var r=api.QueryMembers(job,deadline);Stamp(r);r.Capacity=32;r.BufferBytes=264;return r;}
    public LifetimeOpen Open(uint pid,uint access,bool inherit){var r=api.Open(pid,access,inherit);Stamp(r);return r;}
    public LifetimeMembership IsOwned(IntPtr handle,IntPtr job){var r=api.IsOwned(handle,job);Stamp(r);return r;}
    public LifetimeRead Read(IntPtr handle,long deadline){var r=api.Read(handle,deadline);r.StartTimestamp=clock.Now;r.EndTimestamp=clock.Now;r.Frequency=clock.Frequency;r.StartUTC=clock.UtcNow;r.EndUTC=clock.UtcNow;return r;}
    public LifetimeRelease Close(IntPtr handle){return api.Close(handle);}
}
public sealed class FakeCompileStageHarness {
    public readonly FakeCompileStageClock Clock=new FakeCompileStageClock();public readonly FakeCompileStageBytes Bytes=new FakeCompileStageBytes();
    public readonly FakeLifetimeApi Api=new FakeLifetimeApi();public readonly FakeCompileStageRoot Root;
    public readonly CompileStageRetainedSession Session;public int Factories;
    public FakeCompileStageHarness(){Root=new FakeCompileStageRoot{Clock=Clock};Session=new CompileStageRetainedSession(new CompileStageTrigger(Bytes,Clock,"C:/SDK/pkg/tool/windows_amd64/compile.exe","C:/Repo/internal/nodecontrol/nodeleafbinding"),Root,Create);}
    RetainedLifetimeRegistry Create(){Factories++;return new RetainedLifetimeRegistry(new IntPtr(1),1,new ClockedCompileLifetimeApi(Api,Clock),Clock);}
}
