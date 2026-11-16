using Talenro.Observation.CloseGuardian;
using Xunit;
namespace Talenro.Demo.Tests;
public class GuardianSessionTests
{
    [Fact] public void InMemorySessionFromReadyThroughVerifiedExit()
    {
        var guardian=new GuardianState(); var app=new ApplicationCloseState(GuardianApplicationCloseTests.Frame,7);
        var effects=new List<(CloseEffect Effect,long At)>();
        guardian.Advance(new(GuardianEventKind.Ready,Verified:true),1);
        Assert.True(guardian.LaunchAllowed);
        guardian.Advance(new(GuardianEventKind.Launch,GuardianIdentityTests.Evidence,true,LaunchedTarget:GuardianIdentityTests.Expected),0);
        Assert.False(guardian.LaunchAllowed);
        guardian.Advance(GuardianStateTests.Valid(GuardianEventKind.Bound),1);
        void G(GuardianEventKind kind,long at)
        { var d=guardian.Advance(GuardianStateTests.Valid(kind),at); if(d.Effect is { } e) effects.Add((e,at)); }
        void A(AppEvent input,long at)
        { var d=app.Advance(input,at,GuardianApplicationCloseTests.Key); if(d.Effect is { } e) effects.Add((e,at)); }
        G(GuardianEventKind.Tick,289999); G(GuardianEventKind.Tick,290000);
        G(GuardianEventKind.ClaimWritten,290001);
        var transport=new Queue<byte[]>(); transport.Enqueue(GuardianProtocol.Encode(GuardianApplicationCloseTests.Frame,GuardianApplicationCloseTests.Key));
        G(GuardianEventKind.SendComplete,290002);
        A(GuardianApplicationCloseTests.Request with {Frame=transport.Dequeue()},290003);
        A(new(AppEventKind.ReceiptWritten),290004);
        A(new(AppEventKind.Enqueued),290004);
        A(new(AppEventKind.Callback,WindowGeneration:7),290005);
        guardian.Advance(new(GuardianEventKind.OriginalExit,Verified:true,ExitCode:0),290006);
        Assert.True(guardian.ExitVerified); Assert.True(guardian.RequestTimely);
        Assert.Equal(new[]{CloseEffect.ClaimReceipt,CloseEffect.SendClose,CloseEffect.AcceptReceipt,CloseEffect.EnqueueClose,CloseEffect.InvokeOwnClose},effects.Select(e=>e.Effect));
        Assert.All(effects,e=>Assert.InRange(e.At,290000,299999));
        Assert.Equal(CloseVerdict.Passed,GuardianVerdict.Evaluate(new(290000,290001,290002,290003,290005,290006,true,true,true,true,false)));
    }
    [Theory][InlineData("observer-loss")][InlineData("pipe-replacement")][InlineData("truncated")]
    [InlineData("delayed-ui")][InlineData("hung-receipt")][InlineData("suspend")]
    public void FaultsCannotCreateReplayOrLateInvocation(string fault)
    {
        var g=GuardianStateTests.Armed(); var a=new ApplicationCloseState(GuardianApplicationCloseTests.Frame,7);
        var first=fault=="observer-loss"?GuardianEventKind.ParentLost:GuardianEventKind.Tick;
        Assert.Equal(CloseEffect.ClaimReceipt,g.Advance(GuardianStateTests.Valid(first),290000).Effect);
        if(fault=="hung-receipt")
        { Assert.Equal(CloseEffect.Finish,g.Advance(GuardianStateTests.Valid(GuardianEventKind.Tick),360000).Effect); return; }
        Assert.Equal(CloseEffect.SendClose,g.Advance(GuardianStateTests.Valid(GuardianEventKind.ClaimWritten),290001).Effect);
        var request=GuardianApplicationCloseTests.Request;
        if(fault=="pipe-replacement") request=request with {Peer=GuardianIdentityTests.Evidence with {RetainedCreation=1}};
        if(fault=="truncated") request=request with {Frame=request.Frame![..^1]};
        var result=a.Advance(request,fault=="suspend"?360000:290002,GuardianApplicationCloseTests.Key);
        if(fault is "pipe-replacement" or "truncated" or "suspend") Assert.Equal(CloseEffect.Finish,result.Effect);
        else { a.Advance(new(AppEventKind.ReceiptWritten),290003); a.Advance(new(AppEventKind.Enqueued),290004); }
        Assert.NotEqual(CloseEffect.InvokeOwnClose,a.Advance(new(AppEventKind.Callback,WindowGeneration:7),300000).Effect);
        Assert.Null(g.Advance(GuardianStateTests.Valid(GuardianEventKind.ClaimWritten),290002).Effect);
    }
    [Fact] public void PhoneReserveAllows210SecondsButNoLateOrStaleAck()
    {
        var p=new GuardianPhoneSteps(); Assert.True(p.TryPrompt(0,210000,true));
        Assert.False(p.TryPrompt(1,210001,true)); Assert.False(p.Confirm(1,210002));
        Assert.True(p.Confirm(0,270000)); Assert.False(p.Confirm(0,270000));
        Assert.False(new GuardianPhoneSteps().TryPrompt(0,210001,true));
        Assert.False(new GuardianPhoneSteps().TryPrompt(0,1,false));
        var late=new GuardianPhoneSteps(); Assert.True(late.TryPrompt(0,100,true));
        Assert.False(late.Confirm(0,60101)); Assert.False(late.Confirm(0,60102));
    }
}
