namespace Talenro.Observation.CloseGuardian;

public enum CloseEffect { ClaimReceipt, SendClose, AcceptReceipt, EnqueueClose, InvokeOwnClose, Finish }
public enum GuardianPhase { AwaitReady, Ready, AwaitBound, Armed, Claiming, Sending, Observing, Finished }
public enum GuardianEventKind { Tick, Ready, Launch, Bound, Early, ParentLost, ClaimWritten, ClaimFailed, SendComplete, SendFailed, Ack, OriginalExit, Fault }
// Native adapters must supply evidence from retained handles and authenticated channels.
public sealed record GuardianEvent(GuardianEventKind Kind, IdentityEvidence? Identity=null,
    bool Verified=false, int? ExitCode=null, ProcessIdentity? LaunchedTarget=null);
public sealed record GuardianDecision(GuardianPhase State, CloseEffect? Effect, string Reason);
public enum CloseVerdict { Failed, NotExercised, Passed }
public sealed class GuardianPhoneSteps
{
    private readonly object gate=new();
    private int next;
    private long last=-1,started;
    private bool outstanding,closed;
    private bool Clock(long now)
    {
        if(now<0 || now<last) closed=true;
        last=now; return !closed;
    }
    public bool TryPrompt(int step,long elapsedMs,bool bound)
    {
        lock(gate)
        {
            if(!Clock(elapsedMs) || !bound || outstanding || step!=next || elapsedMs>210000) return false;
            started=elapsedMs; outstanding=true; return true;
        }
    }
    public bool Confirm(int step,long elapsedMs)
    {
        lock(gate)
        {
            if(!Clock(elapsedMs) || !outstanding) return false;
            if(elapsedMs-started>60000) { closed=true; outstanding=false; return false; }
            if(step!=next) return false;
            outstanding=false; next++; return true;
        }
    }
}
public sealed record CloseEvidence(long? ClaimMs,long? SendStartMs,long? SendCompleteMs,long? AcceptedMs,long? InvokedMs,
    long? ExitMs,bool OriginalExitVerified,bool Finalized,bool Correlated,bool HashesUnchanged,bool ManualOrEarlyExit);
public static class GuardianVerdict
{
    public static CloseVerdict Evaluate(CloseEvidence e)
    {
        if(!e.OriginalExitVerified || !e.Finalized || !e.Correlated || !e.HashesUnchanged ||
            e.ExitMs is not (>=0 and <360000)) return CloseVerdict.Failed;
        if(e.ManualOrEarlyExit) return CloseVerdict.NotExercised;
        if(e.ClaimMs is not (>=0 and <300000) || e.SendStartMs is not (>=0 and <300000) ||
            e.SendCompleteMs is not (>=0 and <300000) || e.AcceptedMs is not (>=0 and <300000) ||
            e.InvokedMs is not (>=0 and <300000)) return CloseVerdict.Failed;
        // Receiver and sender completion receipts can race; no invented total ordering.
        return e.ClaimMs<=e.SendStartMs && e.SendStartMs<=e.SendCompleteMs && e.SendStartMs<=e.AcceptedMs &&
            e.AcceptedMs<=e.InvokedMs && e.InvokedMs<=e.ExitMs ? CloseVerdict.Passed : CloseVerdict.Failed;
    }
}
public sealed class GuardianState(ProcessIdentity? target=null)
{
    private readonly object gate=new();
    private long last=-1;
    public GuardianPhase State { get; private set; } = GuardianPhase.AwaitReady;
    public bool LaunchAllowed => State==GuardianPhase.Ready;
    public bool RequestTimely { get; private set; }
    public bool ExitVerified { get; private set; }
    public CloseReason RequestReason { get; private set; }
    public GuardianDecision Advance(GuardianEvent input,long elapsedMs)
    {
        lock(gate)
        {
            if(State==GuardianPhase.Finished) return Decision(null,"terminal");
            // Only the sole launch resets the prelaunch READY clock to immutable T0.
            if(input.Kind==GuardianEventKind.Launch && State==GuardianPhase.Ready && input.Verified && elapsedMs==0)
            {
                if(input.LaunchedTarget is not null)
                {
                    if(target is not null && !GuardianValidation.SameIdentity(target,input.LaunchedTarget)) return Finish("target-replacement");
                    target=input.LaunchedTarget;
                    if(!Valid(input.Identity)) return Finish("launch-identity-blocked");
                }
                if(!GuardianValidation.Identity(target)) return Finish("missing-target");
                last=0; State=GuardianPhase.AwaitBound; return Decision(null,"launched");
            }
            if(elapsedMs<0 || elapsedMs<last) return Finish("invalid-clock");
            last=elapsedMs;
            if(elapsedMs>=360000) return Finish("observation-cap");
            if(input.Kind==GuardianEventKind.Fault) return Finish("fault");
            if(input.Kind==GuardianEventKind.OriginalExit)
            {
                if(!input.Verified || input.ExitCode is null) return Finish("unverified-exit");
                ExitVerified=true; return Finish("original-exit");
            }
            if(State is GuardianPhase.AwaitReady or GuardianPhase.Ready)
            {
                if(elapsedMs>10000) return Finish("ready-timeout");
                if(input.Kind==GuardianEventKind.Ready && input.Verified)
                { State=GuardianPhase.Ready; return Decision(null,"ready"); }
                return input.Kind==GuardianEventKind.Tick ? Decision(null,"waiting-ready") : Finish("ready-required");
            }
            if(State==GuardianPhase.AwaitBound)
            {
                if(elapsedMs>10000) return Finish("binding-timeout");
                if(input.Kind==GuardianEventKind.Bound && input.Verified && Valid(input.Identity))
                { State=GuardianPhase.Armed; return Decision(null,"bound"); }
                return input.Kind==GuardianEventKind.Tick ? Decision(null,"waiting-bound") : Finish("binding-failed");
            }
            if(input.Kind is GuardianEventKind.ClaimFailed or GuardianEventKind.SendFailed) return Finish("effect-failed-no-retry");
            if(elapsedMs>=300000 && State is not GuardianPhase.Observing) return Finish("request-deadline");
            if(State==GuardianPhase.Armed && (input.Kind is GuardianEventKind.Early or GuardianEventKind.ParentLost ||
                input.Kind==GuardianEventKind.Tick && elapsedMs>=290000))
            {
                if(!input.Verified || !Valid(input.Identity)) return Finish("identity-blocked");
                RequestReason=input.Kind switch {GuardianEventKind.Early=>CloseReason.Early,
                    GuardianEventKind.ParentLost=>CloseReason.ParentLost,_=>CloseReason.Deadline};
                State=GuardianPhase.Claiming; return Decision(CloseEffect.ClaimReceipt,"claimed");
            }
            if(State==GuardianPhase.Claiming && input.Kind==GuardianEventKind.ClaimWritten)
            {
                if(!input.Verified || !Valid(input.Identity)) return Finish("identity-blocked");
                State=GuardianPhase.Sending; return Decision(CloseEffect.SendClose,"send-once");
            }
            if(State==GuardianPhase.Sending && input.Kind==GuardianEventKind.SendComplete)
            {
                if(!input.Verified) return Finish("send-unknown");
                RequestTimely=true; State=GuardianPhase.Observing;
            }
            return Decision(null,"no-effect");
        }
    }
    private bool Valid(IdentityEvidence? identity)=>target is not null && identity is not null && GuardianIdentity.Verify(target,identity);
    private GuardianDecision Decision(CloseEffect? effect,string reason)=>new(State,effect,reason);
    private GuardianDecision Finish(string reason) { State=GuardianPhase.Finished; return Decision(CloseEffect.Finish,reason); }
}
