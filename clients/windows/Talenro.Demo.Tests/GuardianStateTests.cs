using Talenro.Observation.CloseGuardian;
using Xunit;
namespace Talenro.Demo.Tests;
public class GuardianStateTests
{
    internal static GuardianState Armed()
    {
        var s=new GuardianState(GuardianIdentityTests.Expected);
        s.Advance(new(GuardianEventKind.Ready,Verified:true),100);
        Assert.True(s.LaunchAllowed);
        s.Advance(new(GuardianEventKind.Launch,Verified:true),0);
        s.Advance(new(GuardianEventKind.Bound,GuardianIdentityTests.Evidence,true),10000);
        Assert.Equal(GuardianPhase.Armed,s.State); return s;
    }
    internal static GuardianEvent Valid(GuardianEventKind kind)=>new(kind,GuardianIdentityTests.Evidence,true);
    [Fact] public void TimerClaimsAndSendsOnceThenAckIsNotExit()
    {
        var s=Armed(); Assert.Null(s.Advance(Valid(GuardianEventKind.Tick),289999).Effect);
        Assert.Equal(CloseEffect.ClaimReceipt,s.Advance(Valid(GuardianEventKind.Tick),290000).Effect);
        Assert.Null(s.Advance(Valid(GuardianEventKind.Tick),290000).Effect);
        Assert.Equal(CloseEffect.SendClose,s.Advance(Valid(GuardianEventKind.ClaimWritten),299998).Effect);
        Assert.Null(s.Advance(Valid(GuardianEventKind.ClaimWritten),299998).Effect);
        s.Advance(Valid(GuardianEventKind.SendComplete),299999);
        Assert.True(s.RequestTimely); s.Advance(Valid(GuardianEventKind.Ack),299999);
        Assert.False(s.ExitVerified);
        Assert.Equal(CloseEffect.Finish,s.Advance(Valid(GuardianEventKind.Tick),360000).Effect);
        Assert.Null(s.Advance(Valid(GuardianEventKind.Tick),360001).Effect);
    }
    [Fact] public void HandshakeGatesAndLostBoundAckDoNotDisarm()
    {
        var missing=new GuardianState(GuardianIdentityTests.Expected);
        missing.Advance(new(GuardianEventKind.Launch,Verified:true),0);
        Assert.False(missing.LaunchAllowed); Assert.Equal(GuardianPhase.Finished,missing.State);
        var late=new GuardianState(GuardianIdentityTests.Expected);
        late.Advance(new(GuardianEventKind.Ready,Verified:true),10001);
        Assert.False(late.LaunchAllowed);
        var s=Armed(); s.Advance(Valid(GuardianEventKind.Bound),10001);
        Assert.Equal(GuardianPhase.Armed,s.State);
        Assert.Equal(CloseEffect.ClaimReceipt,s.Advance(Valid(GuardianEventKind.ParentLost),10002).Effect);
        var gap=new GuardianState(GuardianIdentityTests.Expected);
        gap.Advance(new(GuardianEventKind.Ready,Verified:true),0);
        gap.Advance(new(GuardianEventKind.Launch,Verified:true),0);
        gap.Advance(Valid(GuardianEventKind.Bound),10001);
        Assert.Equal(GuardianPhase.Finished,gap.State);
    }
    [Theory][InlineData(299999,true)][InlineData(300000,false)][InlineData(360000,false)]
    public void ReceiptCompletionRechecksDeadline(long time,bool sends)
    {
        var s=Armed(); s.Advance(Valid(GuardianEventKind.Tick),290000);
        Assert.Equal(sends,s.Advance(Valid(GuardianEventKind.ClaimWritten),time).Effect==CloseEffect.SendClose);
    }
    [Fact] public void AllEarlyTimerParentLossOrdersHaveOnlyOneClaimAndSend()
    {
        var kinds=new[]{GuardianEventKind.Early,GuardianEventKind.Tick,GuardianEventKind.ParentLost};
        foreach(var a in kinds) foreach(var b in kinds.Where(k=>k!=a))
        {
            var s=Armed(); var c=kinds.Single(k=>k!=a&&k!=b);
            var effects=new[]{a,b,c}.Select(k=>s.Advance(Valid(k),290000).Effect).ToList();
            effects.Add(s.Advance(Valid(GuardianEventKind.ClaimWritten),290000).Effect);
            effects.Add(s.Advance(Valid(GuardianEventKind.ClaimWritten),290000).Effect);
            Assert.Single(effects,e=>e==CloseEffect.ClaimReceipt);
            Assert.Single(effects,e=>e==CloseEffect.SendClose);
        }
    }
    [Fact] public void BadIdentityReceiptSendClockAndOriginalExitCannotRetry()
    {
        foreach(var bad in new[]{GuardianIdentityTests.Evidence with {RetainedCreation=1},GuardianIdentityTests.Evidence with {CimSucceeded=false}})
        { var s=Armed(); Assert.NotEqual(CloseEffect.ClaimReceipt,s.Advance(new(GuardianEventKind.Tick,bad,true),290000).Effect); }
        foreach(var failure in new[]{GuardianEventKind.ClaimFailed,GuardianEventKind.SendFailed,GuardianEventKind.Fault})
        {
            var s=Armed(); s.Advance(Valid(GuardianEventKind.Tick),290000);
            s.Advance(Valid(failure),290001);
            Assert.NotEqual(CloseEffect.SendClose,s.Advance(Valid(GuardianEventKind.ClaimWritten),290002).Effect);
        }
        var late=Armed(); late.Advance(Valid(GuardianEventKind.Tick),290000);
        late.Advance(Valid(GuardianEventKind.ClaimWritten),299999); late.Advance(Valid(GuardianEventKind.SendComplete),300000);
        Assert.False(late.RequestTimely);
        var back=Armed(); back.Advance(Valid(GuardianEventKind.Tick),9999); Assert.Equal(GuardianPhase.Finished,back.State);
        var exited=Armed(); exited.Advance(new(GuardianEventKind.OriginalExit,Verified:true,ExitCode:0),10001);
        Assert.True(exited.ExitVerified); Assert.False(exited.RequestTimely);
        Assert.Null(exited.Advance(Valid(GuardianEventKind.Tick),290000).Effect);
    }
    [Fact] public void ConcurrentClaimsStillEmitOneEffect()
    {
        var s=Armed(); var effects=new System.Collections.Concurrent.ConcurrentBag<CloseEffect?>();
        Parallel.For(0,50,_=>effects.Add(s.Advance(Valid(GuardianEventKind.Tick),290000).Effect));
        Assert.Single(effects,e=>e==CloseEffect.ClaimReceipt);
    }
    [Fact] public void ActualTargetCanOnlyBeAssignedBySoleVerifiedLaunch()
    {
        var s=new GuardianState();
        s.Advance(new(GuardianEventKind.Ready,Verified:true),100);
        Assert.True(s.LaunchAllowed);
        s.Advance(new(GuardianEventKind.Launch,GuardianIdentityTests.Evidence,true,LaunchedTarget:GuardianIdentityTests.Expected),0);
        s.Advance(Valid(GuardianEventKind.Bound),10);
        Assert.Equal(GuardianPhase.Armed,s.State);
        Assert.NotEqual(GuardianPhase.AwaitBound,s.Advance(new(GuardianEventKind.Launch,Verified:true,
            LaunchedTarget:GuardianIdentityTests.Expected with {Pid=789}),0).State);
    }
}
