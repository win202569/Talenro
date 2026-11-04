using System;
using System.IO;
using System.Linq;
using System.Text;
using System.Text.Json;
using System.Diagnostics;
using System.Collections.Generic;
using System.Security.Cryptography;
using System.Threading;
using System.Threading.Tasks;
using Talenro.Diagnostics.BoundedDiskV1;

namespace Talenro.Diagnostics.IsolatedWriterV1
{
    public sealed class RelayResult
    {
        public long[] Read { get; set; }=new long[2];
        public long[] Admitted { get; set; }=new long[2];
        public long[] Sent { get; set; }=new long[2];
        public bool StdoutEOF { get; set; }
        public bool StderrEOF { get; set; }
        public string StdoutSHA256 { get; set; }
        public string StderrSHA256 { get; set; }
        public int Frames { get; set; }
        public string Error { get; set; }
    }
    public static class Relay
    {
        private sealed class Channel
        {
            internal readonly byte[] Buffer=new byte[8192];
            internal readonly Stream Source;
            internal int Count;
            internal bool EOF,Ended;
            internal Task Fill;
            internal readonly IncrementalHash Hash=IncrementalHash.CreateHash(HashAlgorithmName.SHA256);
            internal Channel(Stream source){Source=source;}
            internal async Task ReadChunk(RelayResult result,int index,CancellationToken ct)
            {
                Count=0;
                while(Count<8192)
                {
                    int n=await Source.ReadAsync(Buffer.AsMemory(Count,8192-Count),ct).ConfigureAwait(false);
                    result.Read[index]+=n;if(n==0){EOF=true;break;}Count+=n;
                }
            }
        }
        public static Task<RelayResult> RunAsync(Stream stdout,Stream stderr,Stream writerInput,CancellationToken ct)
            =>RunCore(stdout,stderr,writerInput,new RelayResult(),ct);
        internal static async Task<RelayResult> RunCore(Stream stdout,Stream stderr,Stream writerInput,RelayResult result,CancellationToken ct)
        {
            var channels=new[]{new Channel(stdout),new Channel(stderr)};
            byte[] send=new byte[8208];uint sequence=1;int preferred=0;
            foreach(int i in new[]{0,1})channels[i].Fill=channels[i].ReadChunk(result,i,ct);
            try
            {
                while(!channels[0].Ended||!channels[1].Ended)
                {
                    int selected=-1;
                    for(int offset=0;offset<2;offset++){int i=(preferred+offset)%2;if(!channels[i].Ended&&channels[i].Fill.IsCompleted){selected=i;break;}}
                    if(selected<0)
                    {await Task.WhenAny(channels.Where(c=>!c.Ended).Select(c=>c.Fill)).ConfigureAwait(false);continue;}
                    var channel=channels[selected];await channel.Fill.ConfigureAwait(false);preferred=1-selected;
                    if(channel.Count>65536-result.Admitted[selected])throw new GateException("CapacityExceeded","relay channel bound");
                    if(channel.Count>0)
                    {
                        result.Admitted[selected]+=channel.Count;
                        int length=FrameCodec.EncodeInto(send,1,(byte)(selected+1),sequence++,channel.Buffer.AsSpan(0,channel.Count));
                        await writerInput.WriteAsync(send.AsMemory(0,length),ct).ConfigureAwait(false);
                        channel.Hash.AppendData(channel.Buffer,0,channel.Count);result.Sent[selected]+=channel.Count;result.Frames++;
                    }
                    if(channel.EOF)
                    {
                        int length=FrameCodec.EncodeInto(send,2,(byte)(selected+1),sequence++,ReadOnlySpan<byte>.Empty);
                        await writerInput.WriteAsync(send.AsMemory(0,length),ct).ConfigureAwait(false);result.Frames++;channel.Ended=true;
                        string hash=Convert.ToHexString(channel.Hash.GetHashAndReset());
                        if(selected==0){result.StdoutEOF=true;result.StdoutSHA256=hash;}else{result.StderrEOF=true;result.StderrSHA256=hash;}
                    }
                    else channel.Fill=channel.ReadChunk(result,selected,ct);
                }
                int close=FrameCodec.EncodeInto(send,3,0,sequence,ReadOnlySpan<byte>.Empty);
                await writerInput.WriteAsync(send.AsMemory(0,close),ct).ConfigureAwait(false);result.Frames++;
                // No in-flight send remains; EOF is part of successful protocol delivery.
                await writerInput.DisposeAsync().ConfigureAwait(false);return result;
            }
            catch(Exception e)
            {
                result.Error=e is GateException g?g.Reason:"TransportFailure";
                throw;
            }
            finally
            {
                // Do not dispose a source or reuse a buffer while its read remains pending.
                // Cancellation belongs to the supervisor; waiting here keeps pending work visible.
                foreach(var channel in channels)
                {
                    if(!channel.Fill.IsCompleted){try{await channel.Fill.ConfigureAwait(false);}catch{}}
                    channel.Hash.Dispose();
                }
            }
        }
    }
    public sealed class WriterOptions
    {
        public string Scenario { get; set; }
        public string Token { get; set; }
        public string InputSHA256 { get; set; }
        public string Directory { get; set; }
    }
    public sealed class ProducerOptions
    {
        public string Scenario { get; set; }
    }
    public static class Writer
    {
        private static readonly Task Never=new TaskCompletionSource<bool>(TaskCreationOptions.RunContinuationsAsynchronously).Task;
        public static async Task<int> RunAsync(WriterOptions options)
        {
            Stream input=Console.OpenStandardInput(),output=Console.OpenStandardOutput();
            BoundedFileSink first=null,second=null;bool ready=false;
            async Task Emit(string status,Dictionary<string,object> extra=null)
            {
                var receipt=new Dictionary<string,object>{{"Schema",1},{"Token",options.Token},{"InputSHA256",options.InputSHA256},{"Role","writer"},{"State",status}};
                if(extra!=null)foreach(var p in extra)receipt.Add(p.Key,p.Value);
                byte[] bytes=Encoding.UTF8.GetBytes(JsonSerializer.Serialize(receipt)+"\n");
                if(bytes.Length>4096)throw new GateException("ReceiptViolation","emitted record bound");
                await output.WriteAsync(bytes).ConfigureAwait(false);await output.FlushAsync().ConfigureAwait(false);
            }
            try
            {
                if(options.Scenario=="S03")await Never.ConfigureAwait(false);
                string a=Path.Combine(options.Directory,"stdout.bin"),b=Path.Combine(options.Directory,"stderr.bin");
                first=BoundedFileSink.OpenNew(a);second=BoundedFileSink.OpenNew(b);
                await Emit("READY").ConfigureAwait(false);ready=true;var state=new ProtocolState();Frame frame;bool firstData=true;
                while((frame=await FrameCodec.ReadAsync(input,state,CancellationToken.None).ConfigureAwait(false))!=null)
                {
                    if(frame.Type!=1)continue;
                    if(firstData&&options.Scenario=="S06")return 23;
                    if(firstData&&options.Scenario=="S04")await Never.ConfigureAwait(false);
                    firstData=false;
                    await (frame.Channel==1?first:second).WriteAsync(frame.Payload,0,frame.Payload.Length,CancellationToken.None).ConfigureAwait(false);
                }
                if(options.Scenario=="S05")await Never.ConfigureAwait(false);
                await first.CloseAsync().ConfigureAwait(false);await second.CloseAsync().ConfigureAwait(false);
                if(!state.InputEOF||!first.Closed||!second.Closed||first.Faulted||second.Faulted)throw new GateException("ProtocolViolation","unfinalized writer");
                long al=new FileInfo(a).Length,bl=new FileInfo(b).Length;
                string ah,bh;using(var file=File.OpenRead(a))ah=Convert.ToHexString(SHA256.HashData(file));using(var file=File.OpenRead(b))bh=Convert.ToHexString(SHA256.HashData(file));
                if(options.Scenario!="S09")await Emit("COMPLETE",new Dictionary<string,object>{{"StdoutLength",al},{"StderrLength",bl},{"StdoutSHA256",ah},{"StderrSHA256",bh},{"InputEOF",true},{"CloseSucceeded",true},{"Faulted",false}}).ConfigureAwait(false);
                return 0;
            }
            catch(Exception e)
            {
                // Never manufacture finalization by retrying a failed close/write.
                if(ready)try{await Emit("ERROR",new Dictionary<string,object>{{"Reason",e is GateException g?g.Reason:e.GetType().Name}}).ConfigureAwait(false);}catch{}
                return 24;
            }
        }
    }
    public static class Producer
    {
        public static async Task<int> RunAsync(ProducerOptions options)
        {
            if(options.Scenario=="S07")return 23;
            Stream stdout=Console.OpenStandardOutput(),stderr=Console.OpenStandardError();
            if(options.Scenario=="S02")return 0;
            if(options.Scenario=="S01"||options.Scenario=="S04"||options.Scenario=="S08")
            {
                int length=options.Scenario=="S08"?65537:8193;
                await Pattern(stdout,length).ConfigureAwait(false);
                if(options.Scenario!="S08")await Pattern(stderr,48).ConfigureAwait(false);
            }
            else
            {
                byte[] fixture=Encoding.ASCII.GetBytes("PASS: ContractAPIAvailable\r\nPASS: BaselineCore\r\n");
                if(fixture.Length!=48)throw new InvalidOperationException("fixture length");
                await stdout.WriteAsync(fixture).ConfigureAwait(false);
            }
            await stdout.FlushAsync().ConfigureAwait(false);await stderr.FlushAsync().ConfigureAwait(false);return 0;
        }
        private static async Task Pattern(Stream output,int length)
        {
            byte[] buffer=new byte[8192];int sent=0;
            while(sent<length){int n=Math.Min(buffer.Length,length-sent);for(int i=0;i<n;i++)buffer[i]=(byte)((sent+i)%251);await output.WriteAsync(buffer.AsMemory(0,n)).ConfigureAwait(false);sent+=n;}
        }
    }
    public sealed class ScenarioOptions
    {
        public string Scenario { get; set; }
        public string Token { get; set; }
        public string InputSHA256 { get; set; }
        public string IntentSHA256 { get; set; }
        public string Engine { get; set; }
        public string RoleScript { get; set; }
        public string Root { get; set; }
    }
    internal sealed class NativeProcess : IProcessFacade
    {
        internal readonly Process Original;
        internal NativeProcess(Process p){Original=p;}
        public bool Associated{get{try{return Original.Id>0;}catch(InvalidOperationException){return false;}}}
        public bool HasExited=>Original.HasExited;
        public int ExitCode=>Original.ExitCode;
        public bool Start()=>Original.Start();
        public void Kill()=>Original.Kill();
    }
    public sealed class ProcessEvidence
    {
        public bool StartAttempted { get; set; }
        public bool Associated { get; set; }
        public int? PID { get; set; }
        public bool OwnedExit { get; set; }
        public int? ExitCode { get; set; }
        public bool KillRequested { get; set; }
        public string Error { get; set; }
    }
    public sealed class ScenarioResult
    {
        public string Scenario { get; set; }
        public string Token { get; set; }
        public string InputSHA256 { get; set; }
        public string SupervisionOutcome { get; set; }="INCOMPLETE";
        public bool EvidencePersisted { get; set; }
        public string Reason { get; set; }
        public bool NextAllowed { get; set; }
        public bool PendingResources { get; set; }
        public bool ExitUnconfirmed { get; set; }
        public long ElapsedMs { get; set; }
        public long? DrainStartMs { get; set; }
        public long? ExitStartMs { get; set; }
        public ProcessEvidence Writer { get; set; }=new ProcessEvidence();
        public ProcessEvidence Producer { get; set; }=new ProcessEvidence();
        public RelayResult Relay { get; set; }=new RelayResult();
        public string RelayTask { get; set; }="NotCreated";
        public string ControlTask { get; set; }="NotCreated";
        public string DiagnosticTask { get; set; }="NotCreated";
        public string ControlBase64 { get; set; }
        public string DiagnosticBase64 { get; set; }
        public bool ControlTruncated { get; set; }
        public bool DiagnosticTruncated { get; set; }
        public OutputEvidence Stdout { get; set; }=OutputEvidence.From(false,0,null);
        public OutputEvidence Stderr { get; set; }=OutputEvidence.From(false,0,null);
        public string Status { get; set; }="NOT_ACCEPTED";
        public bool SourceVerified { get; set; }
        public bool LiveActionsPerformed { get; set; }
        public bool RuntimeApproved { get; set; }
        public bool FormalAcceptance { get; set; }
        public bool RealGateImplemented { get; set; }
    }
    public static class Supervisor
    {
        // Retention is deliberately bounded by stop-on-first-pending-result.
        private static object retained;
        private sealed class Control
        {
            internal readonly object Gate=new object();
            internal readonly ReceiptState Receipt;
            internal readonly byte[] Raw=new byte[8192];internal int Count;internal bool Truncated;
            internal Control(string token,string hash){Receipt=new ReceiptState(token,hash);}
            internal async Task Read(Stream source,CancellationToken ct)
            {
                byte[] line=new byte[4096],one=new byte[1];int length=0;
                while(true)
                {
                    int n=await source.ReadAsync(one,ct).ConfigureAwait(false);
                    if(n==0){if(length!=0)throw new GateException("ReceiptViolation","unterminated control line");return;}
                    if(Count>=Raw.Length||length>=line.Length){Truncated=true;throw new GateException("ReceiptViolation","control bounds");}
                    Raw[Count++]=one[0];line[length++]=one[0];
                    if(one[0]==10){lock(Gate)ReceiptValidator.Accept(line.AsSpan(0,length),Receipt);length=0;}
                }
            }
        }
        private sealed class Diagnostics
        {
            internal readonly byte[] Raw=new byte[4096];internal int Count;internal bool Truncated;
            internal async Task Read(Stream stream,CancellationToken ct)
            {
                while(Count<Raw.Length){int n=await stream.ReadAsync(Raw.AsMemory(Count),ct).ConfigureAwait(false);if(n==0)return;Count+=n;}
                byte[] extra=new byte[1];if(await stream.ReadAsync(extra,ct).ConfigureAwait(false)>0){Truncated=true;throw new GateException("ReceiptViolation","writer stderr bound");}
            }
        }
        private static Process Make(ScenarioOptions options,string role)
        {
            var p=new Process();p.StartInfo=new ProcessStartInfo{FileName=options.Engine,WorkingDirectory=options.Root,UseShellExecute=false,CreateNoWindow=true,RedirectStandardOutput=true,RedirectStandardError=true,RedirectStandardInput=role=="Writer"};
            foreach(string arg in new[]{"-NoLogo","-NoProfile","-NonInteractive","-File",options.RoleScript,"-Role",role,"-Scenario",options.Scenario,"-InputSHA256",options.InputSHA256,"-IntentSHA256",options.IntentSHA256})p.StartInfo.ArgumentList.Add(arg);
            return p;
        }
        private static void Observe(Process p,Ownership owner,ProcessEvidence evidence)
        {
            evidence.Associated=owner.Associated;
            if(owner.Associated){evidence.PID=p.Id;evidence.OwnedExit=p.HasExited;if(evidence.OwnedExit)evidence.ExitCode=p.ExitCode;}
        }
        private static string Fault(Task task)
        {
            if(task==null||!task.IsCompleted||task.IsCompletedSuccessfully)return null;
            if(task.IsCanceled)return "TransportCancelled";
            Exception e=task.Exception.GetBaseException();return e is GateException gate?gate.Reason:"TransportFailure";
        }
        public static async Task<ScenarioResult> RunAsync(ScenarioOptions options)
        {
            if(retained!=null)throw new GateException("ExitUnconfirmed","previous resources retained");
            var result=new ScenarioResult{Scenario=options.Scenario,Token=options.Token,InputSHA256=options.InputSHA256};
            var clock=Stopwatch.StartNew();var state=new State();var cancel=new CancellationTokenSource();
            Process writer=Make(options,"Writer"),producer=Make(options,"Producer");
            var writerOwner=new Ownership(new NativeProcess(writer));var producerOwner=new Ownership(new NativeProcess(producer));
            var control=new Control(options.Token,options.InputSHA256);var diagnostic=new Diagnostics();
            Task controlTask=null,diagnosticTask=null;Task<RelayResult> relayTask=null;string observedError=null;Decision decision=null;
            try
            {
                result.Writer.StartAttempted=true;writerOwner.Start();
                controlTask=control.Read(writer.StandardOutput.BaseStream,cancel.Token);
                diagnosticTask=diagnostic.Read(writer.StandardError.BaseStream,cancel.Token);
            }
            catch(Exception e){observedError=e is GateException g?g.Reason:"ProcessStartFailed";result.Writer.Error=e.Message;}
            while(true)
            {
                try{Observe(writer,writerOwner,result.Writer);Observe(producer,producerOwner,result.Producer);}
                catch(Exception e){observedError=observedError??"ExitUnconfirmed";result.Writer.Error=result.Writer.Error??e.Message;}
                bool ready,complete;string receiptError;
                lock(control.Gate){ready=control.Receipt.Ready;complete=control.Receipt.Complete;receiptError=control.Receipt.Error;}
                string fault=observedError??result.Relay.Error??Fault(relayTask)??Fault(controlTask)??Fault(diagnosticTask)??receiptError;
                bool terminal=(relayTask==null||relayTask.IsCompleted)&&(controlTask==null||controlTask.IsCompleted)&&(diagnosticTask==null||diagnosticTask.IsCompleted);
                if(complete&&terminal&&result.Producer.OwnedExit&&result.Writer.OwnedExit&&result.Producer.ExitCode==0&&result.Writer.ExitCode==0&&fault==null)
                {
                    try{lock(control.Gate)control.Receipt.ValidateSuccess(result.Relay,true,result.Producer.ExitCode.Value,true,result.Writer.ExitCode.Value);}
                    catch(GateException e){fault=e.Reason;}
                }
                var observation=new Observation{Ready=ready,ValidComplete=complete,WriterAssociated=result.Writer.Associated,ProducerAssociated=result.Producer.Associated,
                    WriterExited=result.Writer.OwnedExit,ProducerExited=result.Producer.OwnedExit,WriterExitCode=result.Writer.ExitCode??0,ProducerExitCode=result.Producer.ExitCode??0,
                    TasksTerminal=terminal,ControlFinished=controlTask!=null&&controlTask.IsCompleted,Error=fault};
                decision=SupervisorPolicy.Step(state,observation,clock.ElapsedMilliseconds);
                if(decision.Cancel)cancel.Cancel();
                if(decision.KillWriter){result.Writer.KillRequested=true;try{writerOwner.KillIfRunning();}catch(Exception e){result.Writer.Error=e.Message;}}
                if(decision.KillProducer){result.Producer.KillRequested=true;try{producerOwner.KillIfRunning();}catch(Exception e){result.Producer.Error=e.Message;}}
                if(decision.Done)break;
                if(state.Failure==null&&ready&&!result.Producer.StartAttempted)
                {
                    try
                    {
                        result.Producer.StartAttempted=true;producerOwner.Start();
                        relayTask=Relay.RunCore(producer.StandardOutput.BaseStream,producer.StandardError.BaseStream,writer.StandardInput.BaseStream,result.Relay,cancel.Token);
                    }
                    catch(Exception e){observedError=e is GateException g?g.Reason:"ProcessStartFailed";result.Producer.Error=e.Message;}
                }
                await Task.Delay(1).ConfigureAwait(false);
            }
            result.ElapsedMs=clock.ElapsedMilliseconds;result.DrainStartMs=state.DrainStart;result.ExitStartMs=state.ExitStart;
            result.Reason=decision.Reason;result.NextAllowed=decision.NextAllowed;result.ExitUnconfirmed=decision.ExitUnconfirmed;
            result.PendingResources=!decision.DisposeAllowed;result.SupervisionOutcome=decision.Success?"SUCCESS":"FAILED";
            result.RelayTask=relayTask?.Status.ToString()??"NotCreated";result.ControlTask=controlTask?.Status.ToString()??"NotCreated";result.DiagnosticTask=diagnosticTask?.Status.ToString()??"NotCreated";
            if(controlTask==null||controlTask.IsCompleted){result.ControlBase64=Convert.ToBase64String(control.Raw,0,control.Count);result.ControlTruncated=control.Truncated;}
            if(diagnosticTask==null||diagnosticTask.IsCompleted){result.DiagnosticBase64=Convert.ToBase64String(diagnostic.Raw,0,diagnostic.Count);result.DiagnosticTruncated=diagnostic.Truncated;}
            if(decision.Success)
            {
                result.Stdout=OutputEvidence.From(true,control.Receipt.StdoutLength,control.Receipt.StdoutSHA256);
                result.Stderr=OutputEvidence.From(true,control.Receipt.StderrLength,control.Receipt.StderrSHA256);
            }
            if(decision.DisposeAllowed){writer.Dispose();producer.Dispose();cancel.Dispose();}
            else retained=new object[]{writer,producer,cancel,relayTask,controlTask,diagnosticTask,control,diagnostic};
            return result;
        }
    }
}
