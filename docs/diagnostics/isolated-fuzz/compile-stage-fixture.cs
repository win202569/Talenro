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
    public static byte[] Padding(int length){var bytes=new byte[length];for(int i=0;i<length;i++)bytes[i]=(byte)((i%65536==65535 || i==length-1)?10:120);return bytes;}
}
