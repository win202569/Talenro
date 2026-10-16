using System;

public interface ICompileStageRootProbe {CompileStageRootCheck Check();}
// Owns one delayed registry; no native calls and no second sampling cohort.
public sealed class CompileStageRetainedSession:IDisposable {
    readonly CompileStageTrigger trigger;readonly ICompileStageRootProbe rootProbe;
    readonly Func<RetainedLifetimeRegistry> factory;RetainedLifetimeRegistry registry;bool disposed;
    public CompileStageTriggerReport TriggerReport {get{return trigger.Report;}}
    public LifetimeReport LifetimeReport {get{return registry==null?null:registry.Report;}}
    public CompileStageRetainedSession(CompileStageTrigger trigger,ICompileStageRootProbe rootProbe,Func<RetainedLifetimeRegistry> factory){
        if(trigger==null || rootProbe==null || factory==null)throw new ArgumentException("Trigger, exact root probe and registry factory required");
        this.trigger=trigger;this.rootProbe=rootProbe;this.factory=factory;
    }
    public void Tick(long runtimeDeadlineTicks){
        if(disposed)return;
        if(registry==null){
            if(TriggerReport.State!="Waiting")return;
            trigger.Observe(runtimeDeadlineTicks);
            if(TriggerReport.State!="Waiting")return;
            try{
                var check=rootProbe.Check();
                if(trigger.HasCandidate){if(!trigger.ResolveCandidate(check,runtimeDeadlineTicks))return;registry=factory();if(registry==null){trigger.Fail("missing-registry");return;}}
                else {if(check==null || !check.Succeeded || check.Error!=0){trigger.Fail("root-probe-failed");return;}if(check.WaitResult==0)RootExited();return;}
            }catch(Exception e){trigger.Fail("stage-activation:"+e.GetType().Name);return;}
        }
        registry.Poll(runtimeDeadlineTicks);
        var first=registry.Report.FirstQuery;
        if(first!=null && TriggerReport.FirstRegistryQuery==null)TriggerReport.FirstRegistryQuery=new LifetimeQueryEvidence{Succeeded=first.Succeeded,Error=first.Error,StartTimestamp=first.StartTimestamp,EndTimestamp=first.EndTimestamp,Frequency=first.Frequency,StartUTC=first.StartUTC,EndUTC=first.EndUTC,Assigned=first.Assigned,Returned=first.Returned,Capacity=first.Capacity,BufferBytes=first.BufferBytes};
    }
    public void RootExited(){trigger.RootExited();}
    public void Snapshot(string phase,long deadlineTicks,bool rootExited,bool timedOut,bool terminationRequested){
        if(disposed)return;if(rootExited)RootExited();if(registry!=null)registry.Snapshot(phase,deadlineTicks,rootExited,timedOut,terminationRequested);
    }
    public void Dispose(){
        if(disposed)return;disposed=true;
        try{trigger.Dispose();}catch(Exception e){trigger.Fail("stage-reader-release:"+e.GetType().Name);}finally{if(registry!=null)registry.Dispose();}
    }
}
public static class CompileStageRetainedTeardown {
    public static void Run(Action coreCleanup,CompileStageRetainedSession session){try{coreCleanup();}finally{if(session!=null)session.Dispose();}}
}
