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
    public class Result {public long ExitCode;public bool TimedOut,CleanupVerified;public uint RootId,DescendantsRemaining,ActiveAfterCleanup,TotalProcesses;public string StartUTC,EndUTC;}
    public static Result Run(string executable,string[] args,IDictionary overrides,string cwd,string stdout,string stderr,int limitSeconds) {
        IntPtr job=IntPtr.Zero,output=IntPtr.Zero,error=IntPtr.Zero,input=IntPtr.Zero,attributes=IntPtr.Zero,handles=IntPtr.Zero,environment=IntPtr.Zero;
        ProcessInfo pi=new ProcessInfo();bool attributesReady=false,assigned=false;
        var result=new Result{StartUTC=DateTime.UtcNow.ToString("o")};
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
            Check(CreateProcess(executable,command,IntPtr.Zero,IntPtr.Zero,true,0x08080404,environment,cwd,ref startup,out pi));
            result.RootId=pi.Id;
            Check(AssignProcessToJobObject(job,pi.Process));assigned=true;
            Check(ResumeThread(pi.Thread)!=UInt32.MaxValue);
            bool rootExited=false;Accounting accounting;
            while(true) {
                double left=limitSeconds*1000-watch.Elapsed.TotalMilliseconds;
                if(left<=0){result.TimedOut=true;break;}
                uint status=WaitForSingleObject(pi.Process,(uint)Math.Min(100,Math.Ceiling(left)));
                Check(status!=UInt32.MaxValue);
                Check(QueryInformationJobObject(job,1,out accounting,Marshal.SizeOf(typeof(Accounting)),IntPtr.Zero));
                if(status==0){rootExited=true;break;}
            }
            Check(QueryInformationJobObject(job,1,out accounting,Marshal.SizeOf(typeof(Accounting)),IntPtr.Zero));
            result.TotalProcesses=accounting.Total;
            result.DescendantsRemaining=rootExited?accounting.Active:(accounting.Active>0?accounting.Active-1:0);
            if(result.TimedOut || accounting.Active>0) {
                Check(TerminateJobObject(job,0xE001));
                var cleanup=Stopwatch.StartNew();
                do {Thread.Sleep(10);Check(QueryInformationJobObject(job,1,out accounting,Marshal.SizeOf(typeof(Accounting)),IntPtr.Zero));}while(accounting.Active>0 && cleanup.ElapsedMilliseconds<5000);
            }
            result.ActiveAfterCleanup=accounting.Active;result.CleanupVerified=accounting.Active==0;
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
