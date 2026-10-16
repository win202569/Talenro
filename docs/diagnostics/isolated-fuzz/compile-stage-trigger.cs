using System;
using System.Collections.Generic;
using System.IO;
using System.Security.Cryptography;
using System.Text;

public interface ICompileStageByteSource:IDisposable {
    long Length {get;}
    int Read(byte[] buffer,int offset,int count);
}
public sealed class FileCompileStageByteSource:ICompileStageByteSource {
    readonly FileStream stream;
    public FileCompileStageByteSource(string path){stream=new FileStream(path,FileMode.Open,FileAccess.Read,FileShare.ReadWrite);}
    public long Length {get{return stream.Length;}}
    public int Read(byte[] buffer,int offset,int count){return stream.Read(buffer,offset,count);}
    public void Dispose(){stream.Dispose();}
}
public sealed class CompileStageRootCheck:LifetimeApiEvidence {public uint WaitResult;}
public sealed class CompileStageTriggerReport {
    public int Schema=1,IterationByteLimit=65536,LineByteLimit=65536,PollLimit=60000,PollCount;
    public long TotalByteLimit=16777216,BytesConsumed;
    public string Policy="leaf-test-compile-log-v1",State="Waiting";
    public long? LineStart,LineEnd,LineNumber;
    public string RawLineBase64,LineSHA256,PrefixSHA256;
    public LifetimeApiEvidence CandidateObserved,Activation;
    public CompileStageRootCheck RootCheck;
    public LifetimeQueryEvidence FirstRegistryQuery;
    public List<string> Errors=new List<string>();
}
// Bounded, run-local text observer. Text never supplies process identity.
public sealed class CompileStageTrigger:IDisposable {
    const int Chunk=65536,MaxPolls=60000;const long MaxBytes=16777216;
    const string Package="talenro.local/platform/internal/nodecontrol/nodeleafbinding";
    readonly ICompileStageByteSource source;readonly IRetainedLifetimeClock clock;
    readonly string compiler,directory;readonly long frequency;
    readonly byte[] buffer=new byte[Chunk],line=new byte[Chunk];
    readonly SHA256 prefix=SHA256.Create();
    int lineLength;long lineStart,lineNumber=1,lastTick;bool candidate,disposed;
    public CompileStageTriggerReport Report {get;private set;}
    public bool HasCandidate {get{return candidate && Report.State=="Waiting" && !disposed;}}
    static string PathKey(string value){return value.Replace('\\','/').TrimEnd('/');}
    public CompileStageTrigger(ICompileStageByteSource source,IRetainedLifetimeClock clock,string expectedCompiler,string sourceDirectory){
        if(source==null || clock==null || clock.Frequency<=0 || String.IsNullOrEmpty(expectedCompiler) || String.IsNullOrEmpty(sourceDirectory) || !Path.IsPathRooted(expectedCompiler) || !Path.IsPathRooted(sourceDirectory))throw new ArgumentException("Fixed paths, source and clock required");
        this.source=source;this.clock=clock;frequency=clock.Frequency;compiler=PathKey(expectedCompiler);directory=PathKey(sourceDirectory);lastTick=clock.Ticks;
        Report=new CompileStageTriggerReport();
    }
    public void Fail(string error){
        Report.State="Failed";
        if(Report.Errors.Count<16){string text=String.IsNullOrEmpty(error)?"unspecified-trigger-error":error;Report.Errors.Add(text.Length>512?text.Substring(0,512):text);}
    }
    bool Budget(long deadline){
        long now=clock.Ticks;if(now<0 || now<lastTick || clock.Frequency!=frequency){Fail("trigger-clock");return false;}lastTick=now;
        if(now>=deadline){Fail("trigger-deadline");return false;}return true;
    }
    LifetimeApiEvidence Begin(){return new LifetimeApiEvidence{StartTimestamp=clock.Ticks,StartUTC=clock.UtcNow,Frequency=frequency};}
    void End(LifetimeApiEvidence value){value.EndTimestamp=clock.Ticks;value.EndUTC=clock.UtcNow;value.Succeeded=true;}
    static string Digest(byte[] bytes,int count){using(var hash=SHA256.Create())return Hex(hash.ComputeHash(bytes,0,count));}
    static string Hex(byte[] bytes){return BitConverter.ToString(bytes).Replace("-","");}
    static bool Space(char c){return c==' ' || c=='\t' || c=='\r' || c=='\n' || c=='\v' || c=='\f';}
    static List<string> Tokens(string text){
        var result=new List<string>();int i=0;
        while(i<text.Length){
            while(i<text.Length && Space(text[i]))i++;if(i==text.Length)break;
            var token=new StringBuilder();bool quoted=false,closed=false;
            if(text[i]=='"'){quoted=true;i++;}
            while(i<text.Length){
                char c=text[i++];
                if(quoted){
                    if(c=='"'){closed=true;break;}
                    if(c=='\\'){if(i==text.Length || (text[i]!='\\' && text[i]!='"'))throw new FormatException("unsupported-escape");c=text[i++];}
                    token.Append(c);
                }else{if(Space(c))break;if(c=='"')throw new FormatException("embedded-quote");token.Append(c);}
            }
            if(quoted && (!closed || (i<text.Length && !Space(text[i]))))throw new FormatException("unclosed-or-joined-quote");
            result.Add(token.ToString());
        }
        return result;
    }
    static bool LooksLikeCompiler(string text,string compiler){
        // Only a classification for fail-closed malformed candidates, never a match.
        string normalized=PathKey(text.TrimStart(' ','\t','\r','\v','\f').Replace("\\\\","\\"));
        string quoted="\""+compiler;
        if(normalized.StartsWith(quoted,StringComparison.OrdinalIgnoreCase)){
            int at=quoted.Length;return at==normalized.Length || normalized[at]=='"' || Space(normalized[at]);
        }
        if(normalized.StartsWith(compiler,StringComparison.OrdinalIgnoreCase)){
            int at=compiler.Length;return at==normalized.Length || Space(normalized[at]);
        }
        return false;
    }
    bool MatchLine(){
        int disposition=Disposition(line,lineLength,compiler,directory);
        if(disposition<0)Fail("candidate-grammar-or-utf8");return disposition==1;
    }
    static int Disposition(byte[] raw,int count,string compiler,string directory){
        string text;
        try{text=new UTF8Encoding(false,true).GetString(raw,0,count);}
        catch(DecoderFallbackException){return count==Chunk || LooksLikeCompiler(Encoding.UTF8.GetString(raw,0,count),compiler)?-1:0;}
        List<string> tokens;
        try{tokens=Tokens(text);}catch(FormatException){return count==Chunk || LooksLikeCompiler(text,compiler)?-1:0;}
        if((tokens.Count==0 || !String.Equals(PathKey(tokens[0]),compiler,StringComparison.OrdinalIgnoreCase)) && LooksLikeCompiler(text,compiler))return -1;
        return Eligible(tokens,compiler,directory)?1:(count==Chunk?-1:0);
    }
    // The saved-evidence consumer reuses the exact production grammar.
    public static bool IsEligibleCommand(byte[] raw,string expectedCompiler,string sourceDirectory){
        return GetCommandDisposition(raw,expectedCompiler,sourceDirectory)==1;
    }
    public static int GetCommandDisposition(byte[] raw,string expectedCompiler,string sourceDirectory){
        if(raw==null || raw.Length==0 || raw.Length>Chunk || raw[raw.Length-1]!=10)return -1;
        return Disposition(raw,raw.Length,PathKey(expectedCompiler),PathKey(sourceDirectory));
    }
    static bool Eligible(List<string> tokens,string compiler,string directory){
        if(tokens.Count<1 || !String.Equals(PathKey(tokens[0]),compiler,StringComparison.OrdinalIgnoreCase))return false;
        int p=-1,pack=-1;
        for(int i=1;i<tokens.Count;i++){
            if(tokens[i]=="-p"){if(p>=0)return false;p=i;}
            if(tokens[i]=="-pack"){if(pack>=0)return false;pack=i;}
        }
        if(p<1 || p+1>=tokens.Count || tokens[p+1]!=Package || pack<=p+1 || pack+1>=tokens.Count)return false;
        bool test=false;
        for(int i=pack+1;i<tokens.Count;i++){
            string path=PathKey(tokens[i]);int slash=path.LastIndexOf('/');
            if(!Path.IsPathRooted(path) || slash<0 || !String.Equals(path.Substring(0,slash),directory,StringComparison.OrdinalIgnoreCase) || !path.EndsWith(".go",StringComparison.OrdinalIgnoreCase))return false;
            if(path.EndsWith("_test.go",StringComparison.OrdinalIgnoreCase))test=true;
        }
        return test;
    }
    public void Observe(long runtimeDeadlineTicks){
        if(disposed || Report.State!="Waiting" || candidate)return;
        try{
            if(!Budget(runtimeDeadlineTicks))return;
            if(Report.PollCount>=MaxPolls || Report.BytesConsumed>=MaxBytes){Fail("scan-limit");return;}
            Report.PollCount++;var observed=Begin();
            long available=source.Length-Report.BytesConsumed;
            if(available<0){Fail("log-length-regression");return;}
            int want=(int)Math.Min(Math.Min((long)Chunk,available),MaxBytes-Report.BytesConsumed);
            int count=want==0?0:source.Read(buffer,0,want);
            if(count<0 || count>want){Fail("invalid-read-count");return;}
            if(!Budget(runtimeDeadlineTicks))return;
            for(int i=0;i<count;i++){
                line[lineLength++]=buffer[i];Report.BytesConsumed++;
                if(buffer[i]==10){
                    bool match=MatchLine();prefix.TransformBlock(line,0,lineLength,line,0);
                    if(Report.State=="Failed")return;
                    if(match){
                        End(observed);if(!Budget(runtimeDeadlineTicks))return;
                        prefix.TransformFinalBlock(new byte[0],0,0);
                        Report.LineStart=lineStart;Report.LineEnd=Report.BytesConsumed;Report.LineNumber=lineNumber;
                        Report.RawLineBase64=Convert.ToBase64String(line,0,lineLength);Report.LineSHA256=Digest(line,lineLength);Report.PrefixSHA256=Hex(prefix.Hash);
                        Report.CandidateObserved=observed;candidate=true;return;
                    }
                    lineLength=0;lineStart=Report.BytesConsumed;lineNumber++;
                }else if(lineLength==Chunk){Fail("line-limit");return;}
            }
            if(Report.BytesConsumed>=MaxBytes || Report.PollCount>=MaxPolls)Fail("scan-limit-without-candidate");
        }catch(Exception e){Fail("trigger-read:"+e.GetType().Name);}
    }
    public bool ResolveCandidate(CompileStageRootCheck rootCheck,long runtimeDeadlineTicks){
        if(!HasCandidate)return false;
        if(!Budget(runtimeDeadlineTicks))return false;
        if(rootCheck!=null)Report.RootCheck=new CompileStageRootCheck{Succeeded=rootCheck.Succeeded,Error=rootCheck.Error,WaitResult=rootCheck.WaitResult,StartTimestamp=rootCheck.StartTimestamp,EndTimestamp=rootCheck.EndTimestamp,Frequency=rootCheck.Frequency,StartUTC=rootCheck.StartUTC,EndUTC=rootCheck.EndUTC};
        if(rootCheck==null || !rootCheck.Succeeded || rootCheck.Error!=0 || rootCheck.Frequency!=frequency || rootCheck.StartTimestamp<Report.CandidateObserved.EndTimestamp || rootCheck.EndTimestamp<rootCheck.StartTimestamp || rootCheck.EndTimestamp>clock.Ticks){Fail("root-check-incomplete");return false;}
        if(rootCheck.WaitResult==0){RootExited();return false;}
        if(rootCheck.WaitResult!=258){Fail("root-check-state");return false;}
        var activation=Begin();End(activation);if(!Budget(runtimeDeadlineTicks))return false;
        Report.Activation=activation;Report.State="Activated";return true;
    }
    public void RootExited(){if(Report.State=="Waiting")Report.State="RootExitedWithoutActivation";}
    public void Dispose(){
        if(disposed)return;disposed=true;
        try{source.Dispose();}catch(Exception e){Fail("reader-dispose:"+e.GetType().Name);}finally{prefix.Dispose();}
    }
}
