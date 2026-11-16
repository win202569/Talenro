namespace Talenro.Observation.CloseGuardian;
public enum AppClosePhase { Bound, Accepting, Enqueuing, Queued, Invoked, Finished }
public enum AppEventKind { Request, ReceiptWritten, ReceiptFailed, Enqueued, EnqueueFailed, Callback, Tick, GuardianLost, Fault }
public sealed record AppEvent(AppEventKind Kind, byte[]? Frame=null, IdentityEvidence? Peer=null,
    long WindowGeneration=0, bool WindowAlive=true);
public sealed record AppDecision(AppClosePhase State, CloseEffect? Effect,string Reason);
public sealed class ApplicationCloseState(GuardianFrame expected, long originalWindowGeneration)
{
    private readonly object gate=new();
    private readonly GuardianProtocolSession session=new(expected);
    private long last=-1;
    public AppClosePhase State { get; private set; }=AppClosePhase.Bound;
    public AppDecision Advance(AppEvent input,long elapsedMs,ReadOnlySpan<byte> key=default)
    {
        lock(gate)
        {
            if(State is AppClosePhase.Finished or AppClosePhase.Invoked) return Decision(null,"terminal");
            if(elapsedMs<0 || elapsedMs<last) return Finish("invalid-clock");
            last=elapsedMs;
            if(elapsedMs>=300000) return Finish("close-deadline");
            if(input.Kind is AppEventKind.ReceiptFailed or AppEventKind.EnqueueFailed or AppEventKind.GuardianLost or AppEventKind.Fault)
                return Finish("effect-or-peer-failed");
            if(input.Kind==AppEventKind.Request && State==AppClosePhase.Bound)
            {
                if(expected.Domain!=ChannelDomain.Application || expected.Direction!=FrameDirection.GuardianToApp ||
                    expected.Kind!=FrameKind.CloseNormal || originalWindowGeneration<=0 || input.Frame is null ||
                    input.Peer is null || !GuardianIdentity.Verify(expected.Guardian,input.Peer)) return Finish("identity-blocked");
                try { session.Receive(input.Frame,key); }
                catch(GuardianProtocolException) { return Finish("frame-rejected"); }
                State=AppClosePhase.Accepting; return Decision(CloseEffect.AcceptReceipt,"accepted-once");
            }
            if(State==AppClosePhase.Accepting && input.Kind==AppEventKind.ReceiptWritten)
            { State=AppClosePhase.Enqueuing; return Decision(CloseEffect.EnqueueClose,"enqueue-once"); }
            if(State==AppClosePhase.Enqueuing && input.Kind==AppEventKind.Enqueued)
                State=AppClosePhase.Queued;
            // A dispatcher callback can run before TryEnqueue's caller reports success.
            if(State is AppClosePhase.Enqueuing or AppClosePhase.Queued && input.Kind==AppEventKind.Callback)
            {
                if(!input.WindowAlive || input.WindowGeneration!=originalWindowGeneration) return Finish("original-window-unavailable");
                State=AppClosePhase.Invoked; return Decision(CloseEffect.InvokeOwnClose,"invoke-once");
            }
            return Decision(null,"no-effect");
        }
    }
    private AppDecision Decision(CloseEffect? effect,string reason)=>new(State,effect,reason);
    private AppDecision Finish(string reason) { State=AppClosePhase.Finished; return Decision(CloseEffect.Finish,reason); }
}
