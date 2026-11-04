using System;
using System.IO;
using System.Text;
using System.Text.Json;
using System.Linq;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;

namespace Talenro.Diagnostics.IsolatedWriterV1
{
    public sealed class GateException : IOException
    {
        public string Reason { get; }
        public GateException(string reason,string detail):base(detail){Reason=reason;}
    }
    public sealed class Frame
    {
        public byte Type { get; }
        public byte Channel { get; }
        public uint Sequence { get; }
        public byte[] Payload { get; }
        public Frame(byte type,byte channel,uint sequence,byte[] payload)
        {Type=type;Channel=channel;Sequence=sequence;Payload=payload ?? throw new ArgumentNullException(nameof(payload));}
    }
    public sealed class ProtocolState
    {
        private int frames;
        private readonly long[] bytes=new long[2];
        private readonly int[] dataFrames=new int[2];
        private readonly bool[] ended=new bool[2];
        public long[] Bytes => (long[])bytes.Clone();
        public int Frames=>frames;
        public uint NextSequence { get; private set; } = 1;
        public bool CloseSeen { get; private set; }
        public bool InputEOF { get; private set; }
        public void Accept(Frame f)
        {
            FrameCodec.ValidateHeader(f.Type,f.Channel,f.Payload.Length);
            if(CloseSeen || InputEOF || f.Sequence!=NextSequence || frames>=35)
                throw new GateException("ProtocolViolation","frame order/count");
            int c=f.Channel-1;
            if(f.Type==3)
            {if(!ended[0] || !ended[1])throw new GateException("ProtocolViolation","early CLOSE");}
            else
            {
                if(ended[c])throw new GateException("ProtocolViolation","frame after END");
                if(f.Type==1)
                {
                    if(f.Payload.Length>65536-bytes[c])throw new GateException("CapacityExceeded","channel byte bound");
                    if(dataFrames[c]>=16)throw new GateException("ProtocolViolation","channel frame bound");
                }
            }
            // Validation is atomic with respect to all admission counters.
            if(f.Type==1){bytes[c]+=f.Payload.Length;dataFrames[c]++;}
            else if(f.Type==2)ended[c]=true;
            else CloseSeen=true;
            frames++;NextSequence++;
        }
        internal void ObserveEOF()
        {if(!CloseSeen || InputEOF)throw new GateException("ProtocolViolation","unexpected EOF");InputEOF=true;}
    }
    public static class FrameCodec
    {
        internal static void ValidateHeader(byte type,byte channel,int length)
        {
            bool valid=type==1?(channel==1||channel==2)&&length>=1&&length<=8192:
                type==2?(channel==1||channel==2)&&length==0:type==3&&channel==0&&length==0;
            if(!valid)throw new GateException("ProtocolViolation","header type/channel/length");
        }
        public static byte[] Encode(Frame f)
        {
            ValidateHeader(f.Type,f.Channel,f.Payload.Length);
            var result=new byte[16+f.Payload.Length];
            EncodeInto(result,f.Type,f.Channel,f.Sequence,f.Payload);return result;
        }
        internal static int EncodeInto(byte[] target,byte type,byte channel,uint seq,ReadOnlySpan<byte> payload)
        {
            ValidateHeader(type,channel,payload.Length);
            if(target.Length<16+payload.Length)throw new ArgumentException("send buffer");
            target[0]=84;target[1]=76;target[2]=87;target[3]=49;target[4]=1;target[5]=type;target[6]=channel;target[7]=0;
            for(int i=0;i<4;i++){target[8+i]=(byte)(seq>>(i*8));target[12+i]=(byte)((uint)payload.Length>>(i*8));}
            payload.CopyTo(target.AsSpan(16));return 16+payload.Length;
        }
        private static async Task Exact(Stream source,Memory<byte> buffer,CancellationToken ct)
        {
            int read=0;
            while(read<buffer.Length){int n=await source.ReadAsync(buffer.Slice(read),ct).ConfigureAwait(false);if(n==0)throw new GateException("ProtocolViolation","truncated frame");read+=n;}
        }
        public static async Task<Frame> ReadAsync(Stream source,ProtocolState state,CancellationToken ct)
        {
            byte[] header=new byte[16];
            int first=await source.ReadAsync(header.AsMemory(0,1),ct).ConfigureAwait(false);
            if(first==0){state.ObserveEOF();return null;}
            if(state.CloseSeen)throw new GateException("ProtocolViolation","extra bytes after CLOSE");
            await Exact(source,header.AsMemory(1),ct).ConfigureAwait(false);
            if(header[0]!=84||header[1]!=76||header[2]!=87||header[3]!=49||header[4]!=1||header[7]!=0)
                throw new GateException("ProtocolViolation","magic/version/reserved");
            uint seq=0,length=0;for(int i=0;i<4;i++){seq|=(uint)header[8+i]<<(8*i);length|=(uint)header[12+i]<<(8*i);}
            if(length>8192)throw new GateException("ProtocolViolation","payload length");
            ValidateHeader(header[5],header[6],(int)length);
            byte[] payload=new byte[(int)length];await Exact(source,payload,ct).ConfigureAwait(false);
            var frame=new Frame(header[5],header[6],seq,payload);state.Accept(frame);return frame;
        }
    }
    public sealed class ReceiptState
    {
        internal readonly string Token,InputSHA256;
        private int total;
        public bool Ready { get; internal set; }
        public bool Terminal { get; internal set; }
        public bool Complete { get; internal set; }
        public long StdoutLength { get; internal set; }
        public long StderrLength { get; internal set; }
        public string StdoutSHA256 { get; internal set; }
        public string StderrSHA256 { get; internal set; }
        public string Error { get; internal set; }
        public ReceiptState(string token,string inputSHA256){Token=token;InputSHA256=inputSHA256;}
        internal void Admit(int length)
        {if(length>8192-total)throw new GateException("ReceiptViolation","control total");total+=length;}
        public void ValidateSuccess(RelayResult relay,bool producerExited,int producerCode,bool writerExited,int writerCode)
        {
            if(!Complete)throw new GateException("MissingReceipt","COMPLETE absent");
            if(!producerExited||!writerExited||producerCode!=0||writerCode!=0||relay==null||!relay.StdoutEOF||!relay.StderrEOF||
                relay.Sent[0]!=StdoutLength||relay.Sent[1]!=StderrLength||relay.StdoutSHA256!=StdoutSHA256||relay.StderrSHA256!=StderrSHA256)
                throw new GateException("ReceiptViolation","COMPLETE disagrees with bytes or exits");
        }
    }
    public static class ReceiptValidator
    {
        private static readonly UTF8Encoding StrictUTF8=new UTF8Encoding(false,true);
        public static void Accept(ReadOnlySpan<byte> line,ReceiptState state)
        {
            if(line.Length==0||line.Length>4096||line[line.Length-1]!=10||line.IndexOf((byte)13)>=0||state.Terminal)
                throw new GateException("ReceiptViolation","line/order bound");
            if(line.Length>=3&&line[0]==239&&line[1]==187&&line[2]==191)throw new GateException("ReceiptViolation","BOM");
            state.Admit(line.Length);
            try
            {
                string json=StrictUTF8.GetString(line.Slice(0,line.Length-1));
                using var doc=JsonDocument.Parse(json);
                JsonElement root=doc.RootElement;
                if(root.ValueKind!=JsonValueKind.Object)throw new GateException("ReceiptViolation","object required");
                var keys=new HashSet<string>(StringComparer.Ordinal);
                foreach(var p in root.EnumerateObject())if(!keys.Add(p.Name))throw new GateException("ReceiptViolation","duplicate field");
                string status=root.GetProperty("State").GetString();
                string[] basic={"Schema","Token","InputSHA256","Role","State"};
                string[] more=status=="COMPLETE"?new[]{"StdoutLength","StderrLength","StdoutSHA256","StderrSHA256","InputEOF","CloseSucceeded","Faulted"}:
                    status=="ERROR"?new[]{"Reason"}:Array.Empty<string>();
                if(!keys.SetEquals(basic.Concat(more))||root.GetProperty("Schema").GetInt32()!=1||root.GetProperty("Token").GetString()!=state.Token||
                    root.GetProperty("InputSHA256").GetString()!=state.InputSHA256||root.GetProperty("Role").GetString()!="writer")
                    throw new GateException("ReceiptViolation","identity/schema");
                if(status=="READY")
                {if(state.Ready)throw new GateException("ReceiptViolation","duplicate READY");state.Ready=true;return;}
                if(!state.Ready)throw new GateException("ReceiptViolation","terminal before READY");
                if(status=="ERROR") {state.Error=root.GetProperty("Reason").GetString();if(string.IsNullOrEmpty(state.Error))throw new GateException("ReceiptViolation","empty reason");}
                else if(status=="COMPLETE")
                {
                    long a=root.GetProperty("StdoutLength").GetInt64(),b=root.GetProperty("StderrLength").GetInt64();
                    string ah=root.GetProperty("StdoutSHA256").GetString(),bh=root.GetProperty("StderrSHA256").GetString();
                    if(a<0||a>65536||b<0||b>65536||!Hash(ah)||!Hash(bh)||!root.GetProperty("InputEOF").GetBoolean()||!root.GetProperty("CloseSucceeded").GetBoolean()||root.GetProperty("Faulted").GetBoolean())
                        throw new GateException("ReceiptViolation","finalization fields");
                    state.StdoutLength=a;state.StderrLength=b;state.StdoutSHA256=ah;state.StderrSHA256=bh;state.Complete=true;
                }
                else throw new GateException("ReceiptViolation","unknown state");
                state.Terminal=true;
            }
            catch(GateException){throw;}
            catch(Exception e) when(e is JsonException||e is DecoderFallbackException||e is InvalidOperationException||e is KeyNotFoundException||e is FormatException||e is OverflowException)
            {throw new GateException("ReceiptViolation",e.Message);}
        }
        private static bool Hash(string s)=>s!=null&&s.Length==64&&s.All(c=>(c>='0'&&c<='9')||(c>='A'&&c<='F'));
    }
    public sealed class OutputEvidence
    {
        public bool Available { get; set; }
        public long? Length { get; set; }
        public string SHA256 { get; set; }
        public static OutputEvidence From(bool finalized,long length,string hash)
            =>new OutputEvidence{Available=finalized,Length=finalized?(long?)length:null,SHA256=finalized?hash:null};
    }
    public interface IProcessFacade
    {
        bool Associated { get; }
        bool HasExited { get; }
        int ExitCode { get; }
        bool Start();
        void Kill();
    }
    public sealed class Ownership
    {
        private readonly IProcessFacade original;
        public bool Associated { get; private set; }
        public Ownership(IProcessFacade process){original=process;}
        public void Start()
        {
            try{if(!original.Start())throw new IOException("Start returned false");}
            catch(Exception e){throw new GateException("ProcessStartFailed",e.Message);}
            finally{Associated=original.Associated;}
        }
        public void KillIfRunning(){if(Associated&&!original.HasExited)original.Kill();}
    }
    public sealed class State
    {
        public string Failure { get; internal set; }
        public long? DrainStart { get; internal set; }
        public long? ExitStart { get; internal set; }
        internal bool CancellationIssued;
    }
    public sealed class Observation
    {
        public bool Ready,WriterAssociated,ProducerAssociated,WriterExited,ProducerExited,TasksTerminal,ValidComplete,ControlFinished;
        public int WriterExitCode,ProducerExitCode;
        public string Error;
    }
    public sealed class Decision
    {
        public string Reason { get; set; }
        public bool Cancel,KillWriter,KillProducer,Done,Success,DisposeAllowed,NextAllowed,ExitUnconfirmed;
    }
    public static class SupervisorPolicy
    {
        public static Decision Step(State s,Observation o,long ticks)
        {
            if(ticks<0)throw new ArgumentOutOfRangeException(nameof(ticks));
            if(s.Failure==null)
            {
                string fault=o.Error;
                if(fault==null&&o.ProducerAssociated&&o.ProducerExited&&o.ProducerExitCode!=0)fault="ProducerFailed";
                if(fault==null&&o.WriterAssociated&&o.WriterExited&&o.WriterExitCode!=0)fault="WriterExited";
                if(fault==null&&o.WriterAssociated&&o.WriterExited&&o.ControlFinished&&!o.ValidComplete)fault="MissingReceipt";
                if(fault==null&&!o.Ready&&ticks>=5000)fault="ReadyTimeout";
                if(fault==null&&ticks>=120000)fault="ExecutionTimeout";
                if(o.ProducerAssociated&&o.ProducerExited&&o.ProducerExitCode==0&&s.DrainStart==null)s.DrainStart=ticks;
                if(fault==null&&s.DrainStart.HasValue&&ticks-s.DrainStart.Value>=2000)fault="DrainTimeout";
                if(fault!=null){s.Failure=fault;s.ExitStart=ticks;}
            }
            bool exits=(!o.WriterAssociated||o.WriterExited)&&(!o.ProducerAssociated||o.ProducerExited);
            var d=new Decision{Reason=s.Failure};
            if(s.Failure!=null)
            {
                d.Cancel=!s.CancellationIssued;s.CancellationIssued=true;
                d.KillWriter=d.Cancel&&o.WriterAssociated&&!o.WriterExited;
                d.KillProducer=d.Cancel&&o.ProducerAssociated&&!o.ProducerExited;
                d.Done=(exits&&o.TasksTerminal)||ticks-s.ExitStart.Value>=2000;
                d.ExitUnconfirmed=d.Done&&!exits;d.DisposeAllowed=d.Done&&exits&&o.TasksTerminal;
                d.NextAllowed=d.DisposeAllowed;return d;
            }
            d.Success=o.WriterAssociated&&o.ProducerAssociated&&exits&&o.WriterExitCode==0&&o.ProducerExitCode==0&&o.TasksTerminal&&o.ValidComplete;
            d.Done=d.Success;d.DisposeAllowed=d.Success;d.NextAllowed=d.Success;return d;
        }
    }
}
