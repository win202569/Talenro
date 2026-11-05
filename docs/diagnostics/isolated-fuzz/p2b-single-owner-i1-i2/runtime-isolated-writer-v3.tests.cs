// Authored before implementation. Single-run approval omits an executable RED build.
// Mutation targets: skipped bounds/order checks, reused pending buffers, reset
// deadlines, late success after failure, and inferred empty/unclosed output.
using System;
using System.IO;
using System.Linq;
using System.Text;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;

namespace Talenro.Diagnostics.IsolatedWriterV1
{
    public sealed class UnitCase
    {
        public string Id { get; set; }
        public bool Passed { get; set; }
        public int Assertions { get; set; }
        public string Error { get; set; }
    }
    public sealed class UnitReport
    {
        public List<UnitCase> Cases { get; set; } = new List<UnitCase>();
        public bool Passed { get; set; }
    }
    public static class ProtocolTests
    {
        public static readonly string[] Names = { "U01", "U02", "U03", "U04", "U05", "U06", "U07", "U08", "U09", "U10", "U11", "U12", "U13", "U14", "U15", "U16", "U17", "U18" };
        private const string EmptyHash = "E3B0C44298FC1C149AFBF4C8996FB92427AE41E4649B934CA495991B7852B855";
        private static UnitCase current;
        private static void Check(bool value, string why)
        {
            current.Assertions++;
            if (!value) throw new Exception(why);
        }
        private static void Reject(Action action, string reason)
        {
            try { action(); } catch (GateException e) { Check(e.Reason == reason, "wrong rejection: " + e.Reason); return; }
            throw new Exception("accepted invalid input: " + reason);
        }
        private static async Task RejectAsync(Func<Task> action, string reason)
        {
            try { await action(); } catch (GateException e) { Check(e.Reason == reason, "wrong rejection: " + e.Reason); return; }
            throw new Exception("accepted invalid input: " + reason);
        }
        private static Frame F(byte type, byte channel, uint sequence, int length = 0)
            => new Frame(type, channel, sequence, new byte[length]);
        private static Task<Frame> Read(byte[] bytes, ProtocolState state = null)
            => FrameCodec.ReadAsync(new MemoryStream(bytes), state ?? new ProtocolState(), CancellationToken.None);
        private static byte[] Header(byte type, byte channel, uint seq, uint len)
        {
            byte[] b = { 84, 76, 87, 49, 1, type, channel, 0, 0, 0, 0, 0, 0, 0, 0, 0 };
            for (int i = 0; i < 4; i++) { b[8+i]=(byte)(seq>>(8*i)); b[12+i]=(byte)(len>>(8*i)); }
            return b;
        }
        private static byte[] Line(string state = "READY", string extra = "") => Encoding.UTF8.GetBytes(
            "{\"Schema\":1,\"Token\":\"token\",\"InputSHA256\":\"hash\",\"Role\":\"writer\",\"State\":\"" + state + "\"" + extra + "}\n");
        private static byte[] Complete(long length = 0, string hash = EmptyHash) => Line("COMPLETE",
            ",\"StdoutLength\":" + length + ",\"StderrLength\":0,\"StdoutSHA256\":\"" + hash +
            "\",\"StderrSHA256\":\"" + EmptyHash + "\",\"InputEOF\":true,\"CloseSucceeded\":true,\"Faulted\":false");
        private static ReceiptState R() => new ReceiptState("token", "hash");
        public static async Task<UnitReport> RunAsync()
        {
            var report = new UnitReport();
            for (int i = 1; i <= 18; i++)
            {
                current = new UnitCase { Id = Names[i-1] }; report.Cases.Add(current);
                try { await Case(i); current.Passed = true; }
                catch (Exception e) { current.Error = e.ToString(); break; }
            }
            report.Passed = report.Cases.Count == 18 && report.Cases.All(c => c.Passed);
            return report;
        }
        private static async Task Case(int n)
        {
            if (n == 1)
            {
                byte[] literal = {84,76,87,49,1,1,1,0,1,0,0,0,1,0,0,0,42};
                Check(FrameCodec.Encode(new Frame(1,1,1,new byte[]{42})).SequenceEqual(literal), "golden encoding");
                Frame f = await Read(literal); Check(f.Payload.SequenceEqual(new byte[]{42}) && f.Sequence==1 && f.Channel==1, "golden decoding");
            }
            else if (n == 2)
            {
                foreach (int index in new[]{0,4,7,6,5})
                {
                    byte[] b=Header(1,1,1,1); b[index]=99;
                    var source=new HeaderOnlyStream(b);
                    await RejectAsync(async()=>{await FrameCodec.ReadAsync(source,new ProtocolState(),CancellationToken.None);},"ProtocolViolation");
                    Check(!source.PayloadRead,"invalid header reached payload");
                }
            }
            else if (n == 3)
            {
                foreach(uint len in new uint[]{0,8193}) await RejectAsync(async()=>{await Read(Header(1,1,1,len));},"ProtocolViolation");
                Check((await Read(Header(1,1,1,8192).Concat(new byte[8192]).ToArray())).Payload.Length==8192,"8192 accepted");
                foreach(byte type in new byte[]{2,3}) await RejectAsync(async()=>{await Read(Header(type,type==3?(byte)0:(byte)1,1,1));},"ProtocolViolation");
            }
            else if (n == 4)
            {
                await RejectAsync(async()=>{await Read(Header(1,1,1,1).Take(15).ToArray());},"ProtocolViolation");
                await RejectAsync(async()=>{await Read(Header(1,1,1,2).Concat(new byte[]{42}).ToArray());},"ProtocolViolation");
            }
            else if (n == 5)
            {
                foreach(uint seq in new uint[]{0,1,3}) { var s=new ProtocolState(); s.Accept(F(1,1,1,1)); Reject(()=>s.Accept(F(1,1,seq,1)),"ProtocolViolation"); }
            }
            else if (n == 6)
            {
                var s=new ProtocolState(); s.Accept(F(2,1,1));
                Reject(()=>s.Accept(F(1,1,2,1)),"ProtocolViolation"); Reject(()=>s.Accept(F(2,1,2)),"ProtocolViolation");
                Reject(()=>s.Accept(F(3,0,2)),"ProtocolViolation"); s.Accept(F(2,2,2)); s.Accept(F(3,0,3));
                Check(s.CloseSeen && !s.InputEOF,"close is not EOF"); Reject(()=>s.Accept(F(3,0,4)),"ProtocolViolation");
            }
            else if (n == 7)
            {
                var s=new ProtocolState(); for(uint i=1;i<=8;i++) s.Accept(F(1,1,i,8192));
                Check(s.Bytes[0]==65536,"capacity accepted"); Reject(()=>s.Accept(F(1,1,9,1)),"CapacityExceeded");
                Check(s.Bytes[0]==65536 && s.Frames==8 && s.NextSequence==9,"rejected frame changed admission");
            }
            else if (n == 8)
            {
                var s=new ProtocolState(); for(uint i=1;i<=16;i++) s.Accept(F(1,1,i,1));
                Reject(()=>s.Accept(F(1,1,17,1)),"ProtocolViolation");
                var global=new ProtocolState();
                // Unreachable otherwise-valid 36th frame: reflection is test-only.
                typeof(ProtocolState).GetField("frames",System.Reflection.BindingFlags.Instance|System.Reflection.BindingFlags.NonPublic).SetValue(global,35);
                Reject(()=>global.Accept(F(1,1,1,1)),"ProtocolViolation");
            }
            else if (n == 9)
            {
                foreach(bool extra in new[]{true,false})
                {
                    var s=new ProtocolState();s.Accept(F(2,1,1));s.Accept(F(2,2,2));s.Accept(F(3,0,3));
                    Check(!s.InputEOF,"premature finalization");
                    if(extra) await RejectAsync(async()=>{await FrameCodec.ReadAsync(new MemoryStream(new byte[]{1}),s,CancellationToken.None);},"ProtocolViolation");
                    else { Check(await FrameCodec.ReadAsync(new MemoryStream(),s,CancellationToken.None)==null && s.InputEOF,"actual EOF required"); }
                }
            }
            else if (n == 10)
            {
                string good=Encoding.UTF8.GetString(Line());
                foreach(string bad in new[]{good.Replace("token","wrong"),good.Replace("hash","wrong"),good.Replace("writer","producer"),good.Replace("\"Schema\":1","\"Schema\":2"),good.Replace("}\n",",\"Unknown\":1}\n")})
                    Reject(()=>ReceiptValidator.Accept(Encoding.UTF8.GetBytes(bad),R()),"ReceiptViolation");
            }
            else if (n == 11)
            {
                var r=R();ReceiptValidator.Accept(Line(),r);Reject(()=>ReceiptValidator.Accept(Line(),r),"ReceiptViolation");
                Reject(()=>ReceiptValidator.Accept(Complete(),R()),"ReceiptViolation");ReceiptValidator.Accept(Complete(),r);
                Reject(()=>ReceiptValidator.Accept(Complete(),r),"ReceiptViolation");Reject(()=>ReceiptValidator.Accept(Line(),r),"ReceiptViolation");
            }
            else if (n == 12)
            {
                Reject(()=>ReceiptValidator.Accept(new byte[4097],R()),"ReceiptViolation");
                var r=R(); typeof(ReceiptState).GetField("total",System.Reflection.BindingFlags.Instance|System.Reflection.BindingFlags.NonPublic).SetValue(r,8192);
                Reject(()=>ReceiptValidator.Accept(new byte[]{10},r),"ReceiptViolation");
                Reject(()=>ReceiptValidator.Accept(Line().Take(Line().Length-1).ToArray(),R()),"ReceiptViolation");
                Reject(()=>ReceiptValidator.Accept(new byte[]{255,10},R()),"ReceiptViolation");
                Reject(()=>ReceiptValidator.Accept(new byte[]{239,187,191}.Concat(Line()).ToArray(),R()),"ReceiptViolation");
            }
            else if (n == 13)
            {
                var sent=new RelayResult { StdoutSHA256=EmptyHash,StderrSHA256=EmptyHash,StdoutEOF=true,StderrEOF=true };
                foreach(byte[] terminal in new[]{Complete(1),Complete(0,new string('0',64))})
                {var r=R();ReceiptValidator.Accept(Line(),r);ReceiptValidator.Accept(terminal,r);Reject(()=>r.ValidateSuccess(sent,true,0,true,0),"ReceiptViolation");}
                var missing=R();ReceiptValidator.Accept(Line(),missing);Reject(()=>missing.ValidateSuccess(sent,true,0,true,0),"MissingReceipt");
                var good=R();ReceiptValidator.Accept(Line(),good);ReceiptValidator.Accept(Complete(),good);
                Reject(()=>good.ValidateSuccess(sent,true,0,false,0),"ReceiptViolation");Reject(()=>good.ValidateSuccess(sent,true,23,true,0),"ReceiptViolation");
                good.ValidateSuccess(sent,true,0,true,0);Check(true,"good complete/exits");
            }
            else if (n == 14)
            {
                var stdout=new ChunkStream(8193);var stderr=new ChunkStream(8193);var sink=new HeldWriteStream();
                Task<RelayResult> run=Relay.RunAsync(stdout,stderr,sink,CancellationToken.None);
                await sink.Entered.Task; byte[] frozen=sink.Pending.ToArray();
                Check(!run.IsCompleted && sink.Writes==1,"send must remain serialized while held");
                Check(sink.Pending.Span.SequenceEqual(frozen),"pending send buffer mutated");
                sink.Release.TrySetResult(true); var result=await run;
                Check(sink.MaxInFlight==1,"concurrent writes"); Check(result.Sent[0]==8193 && result.Sent[1]==8193 && result.StdoutEOF && result.StderrEOF,"relay counts/EOF");
                var wire=new MemoryStream(sink.Bytes.ToArray());var state=new ProtocolState();var frames=new List<Frame>();Frame frame;
                while((frame=await FrameCodec.ReadAsync(wire,state,CancellationToken.None))!=null) frames.Add(frame);
                Check(frames.Where(f=>f.Type==1).Select(f=>f.Channel).SequenceEqual(new byte[]{1,2,1,2}),"ready channel fairness");
                foreach(byte ch in new byte[]{1,2})
                {
                    var data=frames.Where(f=>f.Type==1 && f.Channel==ch).ToArray();
                    Check(data.Select(f=>f.Payload.Length).SequenceEqual(new[]{8192,1}),"partial reads not coalesced");
                    byte[] expected=Enumerable.Range(0,8193).Select(i=>(byte)(i%251)).ToArray();
                    Check(data.SelectMany(f=>f.Payload).SequenceEqual(expected),"tail bytes changed");
                }
            }
            else if (n == 15)
            {
                var s=new State();Check(SupervisorPolicy.Step(s,new Observation(),4999).Reason==null,"early readiness timeout");
                Check(SupervisorPolicy.Step(s,new Observation(),5000).Reason=="ReadyTimeout","readiness origin");
                s=new State();Check(SupervisorPolicy.Step(s,new Observation{Ready=true},119999).Reason==null,"early execution timeout");
                Check(SupervisorPolicy.Step(s,new Observation{Ready=true},120000).Reason=="ExecutionTimeout","execution origin reset");
                s=new State();var o=new Observation{Ready=true,ProducerAssociated=true,ProducerExited=true,ProducerExitCode=0};
                SupervisorPolicy.Step(s,o,10);SupervisorPolicy.Step(s,o,1000);
                Check(SupervisorPolicy.Step(s,o,2010).Reason=="DrainTimeout","drain window reset");
                Check(!SupervisorPolicy.Step(s,o,4009).Done,"exit window too short");
                Check(SupervisorPolicy.Step(s,o,4010).Done,"exit window reset");
            }
            else if (n == 16)
            {
                var s=new State();var pending=new TaskCompletionSource<bool>();
                var o=new Observation{Error="TransportFailure",WriterAssociated=true,TasksTerminal=pending.Task.IsCompleted};
                var d=SupervisorPolicy.Step(s,o,1);Check(d.Cancel && d.KillWriter && !d.Done,"failure actions");
                o.Error=null;o.ValidComplete=true;o.Ready=true;o.WriterExited=true;o.ProducerExited=true;o.ProducerAssociated=true;
                d=SupervisorPolicy.Step(s,o,2);Check(d.Reason=="TransportFailure" && !d.Success && !d.DisposeAllowed,"late success/pending dispose");
                pending.SetResult(true);o.TasksTerminal=true;d=SupervisorPolicy.Step(s,o,3);
                Check(d.Done && !d.Success && d.DisposeAllowed,"failure lost at terminal");
            }
            else if (n == 17)
            {
                var unavailable=OutputEvidence.From(false,0,EmptyHash);Check(!unavailable.Available && unavailable.Length==null && unavailable.SHA256==null,"unclosed empty inferred");
                var available=OutputEvidence.From(true,0,EmptyHash);Check(available.Available && available.Length==0 && available.SHA256==EmptyHash,"finalized empty rejected");
            }
            else
            {
                var fake=new AssociatedThenThrow();var ownership=new Ownership(fake);
                Reject(()=>ownership.Start(),"ProcessStartFailed");Check(ownership.Associated,"association lost after throw");
                ownership.KillIfRunning();Check(fake.Kills==1,"original association not targeted");
                var s=new State();var o=new Observation{WriterAssociated=ownership.Associated,Error="ProcessStartFailed",TasksTerminal=true};
                SupervisorPolicy.Step(s,o,0);var d=SupervisorPolicy.Step(s,o,2000);
                Check(d.Done && !d.Success && !d.NextAllowed && !d.DisposeAllowed && d.ExitUnconfirmed,"unknown exit allowed handoff");
            }
        }
        private sealed class HeaderOnlyStream : MemoryStream
        {
            public bool PayloadRead; public HeaderOnlyStream(byte[] b):base(b){}
            public override ValueTask<int> ReadAsync(Memory<byte> b,CancellationToken c=default)
            {if(Position>=16){PayloadRead=true;throw new Exception("payload read");}return base.ReadAsync(b,c);}
        }
        private sealed class ChunkStream : Stream
        {
            private readonly int size;private int offset;public ChunkStream(int n){size=n;}
            public override ValueTask<int> ReadAsync(Memory<byte> b,CancellationToken c=default)
            {int n=Math.Min(Math.Min(b.Length,137),size-offset);for(int i=0;i<n;i++)b.Span[i]=(byte)((offset+i)%251);offset+=n;return ValueTask.FromResult(n);}
            public override bool CanRead=>true;public override bool CanWrite=>false;public override bool CanSeek=>false;
            public override long Length=>throw new NotSupportedException();public override long Position{get=>throw new NotSupportedException();set=>throw new NotSupportedException();}
            public override int Read(byte[] b,int o,int n)=>throw new NotSupportedException();public override void Write(byte[] b,int o,int n)=>throw new NotSupportedException();
            public override void Flush(){}public override long Seek(long o,SeekOrigin s)=>throw new NotSupportedException();public override void SetLength(long n)=>throw new NotSupportedException();
        }
        private sealed class HeldWriteStream : MemoryStream
        {
            public readonly TaskCompletionSource<bool> Entered=new TaskCompletionSource<bool>(TaskCreationOptions.RunContinuationsAsynchronously);
            public readonly TaskCompletionSource<bool> Release=new TaskCompletionSource<bool>(TaskCreationOptions.RunContinuationsAsynchronously);
            public ReadOnlyMemory<byte> Pending { get; private set; } public int Writes,MaxInFlight;private int inFlight;
            public readonly List<byte> Bytes=new List<byte>();
            public override async ValueTask WriteAsync(ReadOnlyMemory<byte> b,CancellationToken c=default)
            {
                Writes++;inFlight++;MaxInFlight=Math.Max(MaxInFlight,inFlight);
                if(Writes==1){Pending=b;Entered.SetResult(true);await Release.Task;}
                Bytes.AddRange(b.ToArray());inFlight--;
            }
        }
        private sealed class AssociatedThenThrow : IProcessFacade
        {
            public bool Associated{get;private set;}public bool HasExited=>false;public int ExitCode=>throw new InvalidOperationException();public int Kills;
            public bool Start(){Associated=true;throw new IOException("associated then threw");}public void Kill(){Kills++;}
        }
    }
}
