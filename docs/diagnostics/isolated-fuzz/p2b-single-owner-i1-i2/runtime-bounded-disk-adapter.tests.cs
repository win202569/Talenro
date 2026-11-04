using System;
using System.IO;
using System.Linq;
using System.Text;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;
using System.Diagnostics;
using System.Security.Cryptography;
using System.Text.Json;

namespace Talenro.Diagnostics.BoundedDiskV1
{
    public sealed class CaseResult {
        public string Id {get;set;} public string Name {get;set;}
        public bool Passed {get;set;} public string Error {get;set;}
        public bool Truncated {get;set;} public List<string> Observed {get;set;} = new List<string>();
    }
    public sealed class FileResult {
        public string RelativePath {get;set;} public long Length {get;set;} public string SHA256 {get;set;}
    }
    public sealed class VerificationReport {
        public int Schema {get;set;} = 1; public string Status {get;set;} = "ADAPTER_FAILED";
        public List<CaseResult> CaseResults {get;set;} = new List<CaseResult>();
        public List<string> Unexecuted {get;set;} = new List<string>();
        public List<FileResult> Files {get;set;} = new List<FileResult>();
        public List<string> UnavailablePaths {get;set;} = new List<string>();
        public int PendingOperations {get;set;} public string Error {get;set;} public bool Truncated {get;set;}
        public int Computation {get;set;} = 30;
        public bool SourceVerified {get;set;} public bool LiveActionsPerformed {get;set;}
        public bool RuntimeApproved {get;set;} public bool FormalAcceptance {get;set;} public bool RealGateImplemented {get;set;}
    }
    // Controlled faults test adapter accounting, not filesystem failure probability.
    internal sealed class ControlledSink : Stream {
        private readonly byte[] backing = new byte[65536];
        private int count;
        internal int Writes, Flushes, Closes;
        internal string Mode = "";
        internal readonly TaskCompletionSource<bool> Entered = new TaskCompletionSource<bool>(TaskCreationOptions.RunContinuationsAsynchronously);
        internal readonly TaskCompletionSource<bool> Release = new TaskCompletionSource<bool>(TaskCreationOptions.RunContinuationsAsynchronously);
        internal byte[] Bytes => backing.Take(count).ToArray();
        private void Store(ReadOnlySpan<byte> data) {
            if (data.Length > 65536-count) throw new IOException("Fake capacity");
            data.CopyTo(backing.AsSpan(count)); count += data.Length;
        }
        public override bool CanWrite => true;
        public override bool CanRead => false;
        public override bool CanSeek => false;
        public override long Length => count;
        public override long Position {get => count; set => throw new NotSupportedException();}
        public override int Read(byte[] b,int o,int n) => throw new NotSupportedException();
        public override long Seek(long o,SeekOrigin origin) => throw new NotSupportedException();
        public override void SetLength(long n) => throw new NotSupportedException();
        public override void Write(byte[] b,int o,int n) {
            Writes++;
            if (Mode=="sync-fault") throw new IOException("injected sync fault");
            Store(b.AsSpan(o,n));
        }
        public override void Write(ReadOnlySpan<byte> data) { Writes++; Store(data); }
        public override Task WriteAsync(byte[] b,int o,int n,CancellationToken ct) => WriteAsync(b.AsMemory(o,n),ct).AsTask();
        public override async ValueTask WriteAsync(ReadOnlyMemory<byte> data,CancellationToken ct=default) {
            Writes++;
            if (Mode=="partial-fault") { Store(data.Span.Slice(0,3)); throw new IOException("injected partial fault"); }
            if (Mode=="partial-cancel") {
                Store(data.Span.Slice(0,3)); Entered.TrySetResult(true);
                await Release.Task.ConfigureAwait(false); ct.ThrowIfCancellationRequested(); return;
            }
            if (Mode=="hold") { Entered.TrySetResult(true); await Release.Task.ConfigureAwait(false); }
            ct.ThrowIfCancellationRequested(); Store(data.Span);
        }
        public override void Flush() { Flushes++; if(Mode=="flush-fault") throw new IOException("injected flush fault"); }
        public override Task FlushAsync(CancellationToken ct) { ct.ThrowIfCancellationRequested(); Flush(); return Task.CompletedTask; }
        protected override void Dispose(bool disposing) { if(disposing){Closes++; if(Mode=="close-fault")throw new IOException("injected close fault");} }
        public override async ValueTask DisposeAsync() {
            Closes++;
            if(Mode=="close-pending"){Entered.TrySetResult(true);await Release.Task.ConfigureAwait(false);}
            if(Mode=="close-fault")throw new IOException("injected close fault");
        }
    }
    public static class Verification {
        private static readonly string[] Names = {
            "ArrayWrite","SpanWrite","WriteByte","ArrayAsyncToken","MemoryAsyncToken","ArrayAsyncConvenience","BeginEndWrite",
            "OversizedFirst","CopyExact","CopyOverflow","InvalidArguments","UnsupportedOperations","PreCanceled",
            "PartialFault","PartialCancellation","SynchronousFault","ConcurrentWrite","ConcurrentFlush","ConcurrentClose",
            "CloseFault","ClosePending","ClosedAndZero","ActualFiles","ExclusiveCreation"
        };
        private sealed class RunState {
            internal readonly List<Task> Tasks = new List<Task>();
            internal readonly HashSet<string> Stable = new HashSet<string>(StringComparer.OrdinalIgnoreCase);
            internal readonly Stopwatch Clock = Stopwatch.StartNew();
            internal readonly string Root, Journal;
            internal int Seq=2;
            internal CaseResult Current;
            internal RunState(string root) { Root=Path.GetFullPath(root); Journal=Path.Combine(Path.GetDirectoryName(Root),"stages.jsonl"); }
            internal void Need(bool ok,string label) {
                if(!ok)throw new InvalidOperationException("Assertion: "+label);
                Current.Observed.Add(label);
            }
            internal async Task Wait(Task t) { Tasks.Add(t); await t.WaitAsync(TimeSpan.FromMilliseconds(2000)).ConfigureAwait(false); }
            internal async Task Expect<T>(Func<Task> body,string label) where T:Exception {
                try { await Wait(body()).ConfigureAwait(false); }
                catch(T) { Current.Observed.Add(label); return; }
                throw new InvalidOperationException("Expected "+typeof(T).Name+": "+label);
            }
            internal void ExpectSync<T>(Action body,string label) where T:Exception {
                try { body(); } catch(T) { Current.Observed.Add(label); return; }
                throw new InvalidOperationException("Expected "+typeof(T).Name+": "+label);
            }
            internal void Event(string stage,string kind,string id) {
                var line=JsonSerializer.Serialize(new {Schema=1,Seq=++Seq,Stage=stage,Kind=kind,Case=id,ElapsedTicks=Stopwatch.GetTimestamp(),Frequency=Stopwatch.Frequency})+"\n";
                var bytes=Encoding.UTF8.GetBytes(line);
                using(var f=new FileStream(Journal,FileMode.Append,FileAccess.Write,FileShare.Read)){
                    if(Seq>256 || f.Length+bytes.Length>262144)throw new IOException("Journal bound");
                    f.Write(bytes,0,bytes.Length); f.Flush();
                }
            }
            internal async Task Close(BoundedFileSink s,string path=null) {
                await Wait(s.CloseAsync()).ConfigureAwait(false); Need(s.Closed,"closed");
                if(path!=null)Stable.Add(path);
            }
        }
        private static byte[] Pattern(int n) {var a=new byte[n];for(int i=0;i<n;i++)a[i]=(byte)(i%251);return a;}
        private static async Task WriteRoute(BoundedFileSink s,int route,byte[] bytes) {
            switch(route) {
                case 1:s.Write(bytes,0,bytes.Length);break;
                case 2:s.Write(bytes.AsSpan());break;
                case 3:foreach(byte b in bytes)s.WriteByte(b);break;
                case 4:await s.WriteAsync(bytes,0,bytes.Length,CancellationToken.None).ConfigureAwait(false);break;
                case 5:await s.WriteAsync(bytes.AsMemory(),CancellationToken.None).ConfigureAwait(false);break;
                case 6:await s.WriteAsync(bytes,0,bytes.Length).ConfigureAwait(false);break;
                case 7:
                    await Task.Factory.FromAsync((cb,state)=>s.BeginWrite(bytes,0,bytes.Length,cb,state),s.EndWrite,null).ConfigureAwait(false);break;
                default:throw new InvalidOperationException("Route");
            }
        }
        private static async Task Case(RunState r,int id,string dir) {
            string a=Path.Combine(dir,"a.bin"),b=Path.Combine(dir,"b.bin");
            // Cases01-10 catch bypassed capacity checks using real files and real copy calls.
            if(id<=7) {
                var s=BoundedFileSink.OpenNew(a);var bytes=Pattern(65536);
                if(id==3){s.Write(bytes,0,65535);s.WriteByte(bytes[65535]);}
                else await r.Wait(WriteRoute(s,id,bytes)).ConfigureAwait(false);
                await r.Expect<CapacityExceededException>(()=>WriteRoute(s,id,new byte[]{9}),"over-limit rejected").ConfigureAwait(false);
                r.Need(s.ReservedBytes==65536 && s.ConfirmedBytes==65536 && s.Faulted,"bounded accounting");
                await r.Close(s,a).ConfigureAwait(false);
                r.Need(File.ReadAllBytes(a).SequenceEqual(bytes),"exact65536");return;
            }
            if(id==8) {
                var f=new ControlledSink();var s=new BoundedFileSink(f);
                r.ExpectSync<CapacityExceededException>(()=>s.Write(new byte[65537],0,65537),"whole chunk rejected");
                r.Need(f.Writes==0 && s.ReservedBytes==0 && s.Faulted,"no delegation");await r.Close(s);return;
            }
            if(id==9 || id==10) {
                var bytes=Pattern(id==9?65536:65537);var s=BoundedFileSink.OpenNew(a);
                using(var source=new MemoryStream(bytes,false)){
                    if(id==9)await r.Wait(source.CopyToAsync(s,8192,CancellationToken.None));
                    else await r.Expect<CapacityExceededException>(()=>source.CopyToAsync(s,8192,CancellationToken.None),"copy capacity fault");
                }
                await r.Close(s,a);var got=File.ReadAllBytes(a);
                if(id==9)r.Need(got.SequenceEqual(bytes) && !s.Faulted,"exact copy");
                else r.Need(got.Length<=65536 && got.SequenceEqual(bytes.Take(got.Length)) && s.Faulted,"bounded failed prefix");
                return;
            }
            if(id==11) {
                var f=new ControlledSink();var s=new BoundedFileSink(f);
                r.ExpectSync<ArgumentNullException>(()=>s.Write(null,0,1),"null");
                r.ExpectSync<ArgumentOutOfRangeException>(()=>s.Write(new byte[1],-1,1),"negative offset");
                r.ExpectSync<ArgumentOutOfRangeException>(()=>s.Write(new byte[1],0,-1),"negative count");
                r.ExpectSync<ArgumentException>(()=>s.Write(new byte[1],1,1),"outside array");
                await r.Expect<ArgumentException>(()=>s.WriteAsync(new byte[1],1,1,CancellationToken.None),"async invalid");
                r.Need(s.ReservedBytes==0 && f.Writes==0 && !s.Faulted,"invalid unchanged");
                s.WriteByte(42);r.Need(f.Bytes.SequenceEqual(new byte[]{42}),"valid after invalid");await r.Close(s);return;
            }
            if(id==12) {
                var f=new ControlledSink();var s=new BoundedFileSink(f);var buf=new byte[1];
                r.Need(!s.CanRead && !s.CanSeek,"capabilities");
                r.ExpectSync<NotSupportedException>(()=>s.Read(buf,0,1),"array read");
                r.ExpectSync<NotSupportedException>(()=>s.Read(buf.AsSpan()),"span read");
                r.ExpectSync<NotSupportedException>(()=>s.ReadByte(),"byte read");
                await r.Expect<NotSupportedException>(()=>s.ReadAsync(buf,0,1,CancellationToken.None),"array async read");
                await r.Expect<NotSupportedException>(()=>s.ReadAsync(buf.AsMemory(),CancellationToken.None).AsTask(),"memory async read");
                await r.Expect<NotSupportedException>(()=>s.ReadAsync(buf,0,1),"convenience async read");
                await r.Expect<NotSupportedException>(()=>Task.Factory.FromAsync((cb,state)=>s.BeginRead(buf,0,1,cb,state),s.EndRead,null),"APM read");
                r.ExpectSync<NotSupportedException>(()=>s.Seek(0,SeekOrigin.Begin),"seek");
                r.ExpectSync<NotSupportedException>(()=>s.SetLength(1),"set length");
                r.ExpectSync<NotSupportedException>(()=>s.Position=1,"set position");
                r.Need(f.Writes==0 && s.ReservedBytes==0,"unsupported unchanged");await r.Close(s);return;
            }
            if(id==13) {
                var f=new ControlledSink();var s=new BoundedFileSink(f);
                using(var ct=new CancellationTokenSource()){ct.Cancel();await r.Expect<OperationCanceledException>(()=>s.WriteAsync(new byte[8],0,8,ct.Token),"pre-cancel");}
                r.Need(s.ReservedBytes==0 && f.Writes==0 && s.Faulted,"pre-cancel no reservation");
                r.ExpectSync<IOException>(()=>s.WriteByte(1),"latched");await r.Close(s);return;
            }
            if(id==14 || id==15 || id==16) {
                var f=new ControlledSink{Mode=id==14?"partial-fault":id==15?"partial-cancel":"sync-fault"};
                var s=new BoundedFileSink(f);var data=Pattern(8);
                if(id==14)await r.Expect<IOException>(()=>s.WriteAsync(data,0,8,CancellationToken.None),"partial fault");
                if(id==16)r.ExpectSync<IOException>(()=>s.Write(data,0,8),"sync fault");
                if(id==15) {
                    using(var ct=new CancellationTokenSource()){
                        Task t=s.WriteAsync(data,0,8,ct.Token);r.Tasks.Add(t);
                        await r.Wait(f.Entered.Task);ct.Cancel();f.Release.TrySetResult(true);
                        await r.Expect<OperationCanceledException>(()=>t,"in-flight cancel");
                    }
                }
                r.Need(s.ReservedBytes==8 && s.ConfirmedBytes==0 && s.Faulted,"no refund");
                r.Need(f.Bytes.Length==(id==16?0:3),"actual partial length");
                r.ExpectSync<IOException>(()=>s.WriteByte(1),"latched refusal");await r.Close(s);return;
            }
            if(id>=17 && id<=19) {
                var f=new ControlledSink{Mode="hold"};var s=new BoundedFileSink(f);
                Task t=s.WriteAsync(Pattern(8),0,8,CancellationToken.None);r.Tasks.Add(t);await r.Wait(f.Entered.Task);
                if(id==17)await r.Expect<InvalidOperationException>(()=>s.WriteAsync(new byte[1],0,1,CancellationToken.None),"concurrent write");
                if(id==18){
                    r.ExpectSync<InvalidOperationException>(()=>s.Flush(),"concurrent sync flush");
                    await r.Expect<InvalidOperationException>(()=>s.FlushAsync(CancellationToken.None),"concurrent async flush");
                    r.Need(f.Flushes==0,"no concurrent flush");
                }
                if(id==19){
                    r.ExpectSync<InvalidOperationException>(()=>s.Dispose(),"concurrent dispose");
                    await r.Expect<InvalidOperationException>(()=>s.CloseAsync(),"concurrent close");
                    r.Need(f.Closes==0,"no concurrent close");
                }
                r.Need(s.ReservedBytes==8 && !s.Faulted,"owner unpoisoned");
                f.Release.TrySetResult(true);await r.Wait(t);
                r.Need(s.ConfirmedBytes==8 && f.Bytes.SequenceEqual(Pattern(8)),"owner completed");
                s.Flush();await r.Close(s);return;
            }
            if(id==20) {
                foreach(string mode in new[]{"flush-fault","close-fault"}){
                    var f=new ControlledSink{Mode=mode};var s=new BoundedFileSink(f);
                    await r.Expect<IOException>(()=>s.CloseAsync(),mode);
                    r.Need(!s.Closed && s.Faulted && f.Closes==1,mode+" unavailable and one disposal");
                }return;
            }
            if(id==21) {
                var f=new ControlledSink{Mode="close-pending"};var s=new BoundedFileSink(f);
                Task t=s.CloseAsync();r.Tasks.Add(t);await r.Wait(f.Entered.Task);
                bool available=s.Closed;long? length=available?f.Length:(long?)null;string hash=null;
                r.Need(!t.IsCompleted && !available && length==null && hash==null,"pending not empty");
                r.ExpectSync<InvalidOperationException>(()=>{var ignored=s.Length;},"busy length refused");
                f.Release.TrySetResult(true);await r.Wait(t);r.Need(s.Closed,"released close");return;
            }
            if(id==22) {
                var s=BoundedFileSink.OpenNew(a);s.Write(Array.Empty<byte>(),0,0);await s.WriteAsync(ReadOnlyMemory<byte>.Empty);
                r.Need(s.ReservedBytes==0 && s.ConfirmedBytes==0 && !s.Closed,"empty not yet available");
                await r.Close(s,a);await r.Close(s,a);s.Dispose();
                r.ExpectSync<ObjectDisposedException>(()=>s.WriteByte(1),"closed write");
                r.Need(File.ReadAllBytes(a).Length==0,"stable empty");return;
            }
            if(id==23) {
                var expected=Encoding.UTF8.GetBytes("PASS: ContractAPIAvailable\r\nPASS: BaselineCore\r\n");
                var s=BoundedFileSink.OpenNew(a);var e=BoundedFileSink.OpenNew(b);
                using(var source=new MemoryStream(expected,false))await r.Wait(source.CopyToAsync(s,8192,CancellationToken.None));
                using(var source=new MemoryStream(Array.Empty<byte>(),false))await r.Wait(source.CopyToAsync(e,8192,CancellationToken.None));
                await r.Close(s,a);await r.Close(e,b);
                r.Need(File.ReadAllBytes(a).SequenceEqual(expected) && new FileInfo(b).Length==0,"real48/0");return;
            }
            if(id==24) {
                using(var f=new FileStream(a,FileMode.CreateNew,FileAccess.Write,FileShare.Read)){f.Write(new byte[]{17,23,42});}
                r.Stable.Add(a);
                r.ExpectSync<IOException>(()=>BoundedFileSink.OpenNew(a),"exclusive create");
                r.Need(File.ReadAllBytes(a).SequenceEqual(new byte[]{17,23,42}),"sentinel unchanged");
                var s=BoundedFileSink.OpenNew(b);s.WriteByte(99);await r.Close(s,b);
                r.Need(File.ReadAllBytes(b).SequenceEqual(new byte[]{99}),"fresh path");return;
            }
            throw new InvalidOperationException("Unknown case");
        }
        public static async Task<VerificationReport> RunAsync(string casesRoot) {
            var r=new RunState(casesRoot);var report=new VerificationReport();
            if(Directory.Exists(r.Root))throw new IOException("Cases occupied");
            Directory.CreateDirectory(r.Root);r.Event("verification","begin","");
            for(int i=1;i<=24;i++){
                r.Current=new CaseResult{Id=i.ToString("00"),Name=Names[i-1]};
                report.CaseResults.Add(r.Current);r.Event("case","begin",r.Current.Id);
                try{
                    string dir=Path.Combine(r.Root,r.Current.Id);Directory.CreateDirectory(dir);
                    await Case(r,i,dir).ConfigureAwait(false);
                    if(r.Tasks.Any(t=>!t.IsCompleted))throw new InvalidOperationException("Pending case operations");
                    r.Current.Passed=true;r.Event("case","end",r.Current.Id);
                }catch(Exception ex){
                    string err=ex.GetType().FullName+": "+ex.Message+"\n"+ex.StackTrace;
                    r.Current.Truncated=err.Length>4096;r.Current.Error=err.Substring(0,Math.Min(err.Length,4096));
                    report.Error=r.Current.Error;report.Truncated=r.Current.Truncated;r.Event("case","error",r.Current.Id);
                    for(int j=i+1;j<=24;j++)report.Unexecuted.Add(j.ToString("00"));
                    break;
                }
            }
            report.PendingOperations=r.Tasks.Count(t=>!t.IsCompleted);
            var files=Directory.GetFiles(r.Root,"*",SearchOption.AllDirectories);
            if(files.Length>48)throw new IOException("File inventory bound");
            long total=0;
            foreach(string path in files) {
                string rel=Path.GetRelativePath(r.Root,path).Replace('\\','/');
                if(!r.Stable.Contains(path) || report.PendingOperations!=0){report.UnavailablePaths.Add(rel);continue;}
                byte[] bytes=File.ReadAllBytes(path);
                if(bytes.Length>65536)throw new IOException("Data file bound");total+=bytes.Length;
                report.Files.Add(new FileResult{RelativePath=rel,Length=bytes.Length,SHA256=Convert.ToHexString(SHA256.HashData(bytes))});
            }
            if(total>3145728)throw new IOException("Total data bound");
            if(report.CaseResults.Count==24 && report.CaseResults.All(c=>c.Passed) && report.PendingOperations==0 && report.UnavailablePaths.Count==0)report.Status="ADAPTER_VERIFIED";
            r.Event("verification",report.Status=="ADAPTER_VERIFIED"?"end":"error","");
            return report;
        }
    }
}

