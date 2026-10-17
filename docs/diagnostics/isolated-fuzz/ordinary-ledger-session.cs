using System;
using System.Collections.Generic;

public sealed class OrdinaryFinalOutcome {
    public bool EvidenceComplete,NaturalShutdownVerified,CountReconciled;
    public OrdinaryAccounting FinalAccounting;public ulong[] FinalMembers;
    public OrdinaryAccounting InitialAccounting;public ulong[] InitialMembers=new ulong[0];
    public LifetimeQueryEvidence InitialMemberQuery;
    public LifetimeQueryEvidence FinalMemberQuery;
    public long NaturalObservationMs,ShutdownStartTimestamp,EndTimestamp,Frequency;
    public List<string> Errors=new List<string>();
}
public sealed class OrdinaryLedgerSession {
    readonly OrdinaryMemberLedger ledger;readonly IntPtr job;
    readonly IOrdinaryLedgerApi api;readonly IRetainedLifetimeClock clock;
    bool finished;
    public OrdinaryLedgerSession(OrdinaryMemberLedger ledger,IntPtr job,IOrdinaryLedgerApi api,IRetainedLifetimeClock clock){
        if(ledger==null || job==IntPtr.Zero || job==new IntPtr(-1) || api==null || clock==null || clock.Frequency<=0)throw new ArgumentException("Ledger, job, API and clock required");
        this.ledger=ledger;this.job=job;this.api=api;this.clock=clock;
    }
    void Reject(OrdinaryFinalOutcome r,string error){ledger.Fail(error,-1);if(r.Errors.Count<96)r.Errors.Add(error);}
    public OrdinaryFinalOutcome Finish(bool rootExited,bool timedOut,long shutdownStartTicks){
        var r=new OrdinaryFinalOutcome{ShutdownStartTimestamp=shutdownStartTicks,Frequency=clock.Frequency};
        try{
            if(finished){Reject(r,"duplicate-finish");return r;}finished=true;
            long deadline=checked(shutdownStartTicks+checked(clock.Frequency*500)/1000);
            if(shutdownStartTicks<0 || clock.Ticks<shutdownStartTicks){Reject(r,"shutdown-clock");return r;}
            if(!rootExited || timedOut){Reject(r,"root-unfinished");return r;}
            if(!ledger.Budget(deadline) || ledger.Report.Incomplete)return r;
            r.InitialAccounting=api.QueryAccounting(job,deadline);
            if(!ledger.Budget(deadline) || !OrdinaryMemberLedger.ValidApi(r.InitialAccounting)){Reject(r,"initial-accounting");return r;}
            ledger.Poll(deadline);
            r.InitialMemberQuery=ledger.CurrentPollQuery;
            r.InitialMembers=(ulong[])ledger.LastMembers.Clone();
            while(!ledger.Report.Incomplete){
                ledger.ObserveFinal(deadline,false);
                if(ledger.Report.Incomplete)return r;
                int admitted=ledger.Report.Admitted;
                // Recheck candidates after member waits, before releasing identities.
                ledger.Poll(deadline);
                if(ledger.Report.Incomplete)return r;
                if(ledger.Report.Admitted==admitted)break;
            }
            if(ledger.Report.Incomplete)return r;
            foreach(var entry in ledger.Report.Entries){
                var final=entry.Final;
                if(final==null || !final.Signaled || !final.ExitCodeRead || final.TerminationRequested){Reject(r,"missing-final");return r;}
            }
            ledger.SealAndRelease();
            if(ledger.Report.Incomplete || !ledger.Budget(deadline))return r;
            r.FinalAccounting=api.QueryAccounting(job,deadline);
            if(!ledger.Budget(deadline) || !OrdinaryMemberLedger.ValidApi(r.FinalAccounting)){Reject(r,"final-accounting");return r;}
            var list=api.QueryMembers(job,deadline);r.FinalMemberQuery=OrdinaryMemberLedger.CopyQuery(list);
            if(!ledger.Budget(deadline) || !OrdinaryMemberLedger.ValidList(list)){Reject(r,"final-list");return r;}
            r.FinalMembers=(ulong[])list.Ids.Clone();
            r.CountReconciled=r.FinalAccounting.TotalProcesses==1u+(uint)ledger.Report.Admitted;
            if(!r.CountReconciled || r.FinalAccounting.ActiveProcesses!=0 || r.FinalAccounting.TotalTerminatedProcesses!=0 || r.FinalMembers.Length!=0){Reject(r,"final-reconciliation");return r;}
            r.EvidenceComplete=true;r.NaturalShutdownVerified=true;
        }catch(Exception){Reject(r,"final-exception");}
        finally{
            r.EndTimestamp=clock.Ticks;
            if(r.EndTimestamp>=shutdownStartTicks && shutdownStartTicks>=0 && clock.Frequency>0){
                r.NaturalObservationMs=(long)((r.EndTimestamp-shutdownStartTicks)*1000.0/clock.Frequency);
            }else{Reject(r,"elapsed-clock");}
            if(r.NaturalObservationMs>=500 || ledger.Report.Incomplete){r.EvidenceComplete=false;r.NaturalShutdownVerified=false;}
        }
        return r;
    }
}
public static class OrdinaryLedgerTeardown {
    public static void Run(Action coreCleanup,OrdinaryMemberLedger ledger){
        try{coreCleanup();}finally{if(ledger!=null)ledger.Dispose();}
    }
}
