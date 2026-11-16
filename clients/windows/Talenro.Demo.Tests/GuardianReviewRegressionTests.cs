using Talenro.Observation.CloseGuardian;
using Xunit;
namespace Talenro.Demo.Tests;
public class GuardianReviewRegressionTests
{
    [Fact] public void ClosePreemptsPhoneAndCancellationIsIrreversible()
    {
        var p=new GuardianPhoneSteps(); Assert.True(p.TryPrompt(0,10000,true));
        var g=GuardianStateTests.Armed();
        var decision=g.Advance(GuardianStateTests.Valid(GuardianEventKind.Early),10001);
        Assert.Equal(CloseEffect.ClaimReceipt,decision.Effect);
        if(decision.Effect is CloseEffect.ClaimReceipt or CloseEffect.Finish) p.Cancel();
        p.Cancel(); Assert.False(p.Confirm(0,10002)); Assert.False(p.Confirm(0,70001));
        Assert.False(p.TryPrompt(0,70002,true));
        var lost=new GuardianPhoneSteps(); Assert.True(lost.TryPrompt(0,100,true));
        Assert.False(lost.TryPrompt(1,101,false)); Assert.False(lost.Confirm(0,102));
        Assert.False(lost.TryPrompt(0,103,true));
    }
    [Fact] public void LaunchHandoffPreservesActualElapsedSinceT0WithoutRenewal()
    {
        GuardianState Ready()
        { var s=new GuardianState(); s.Advance(new(GuardianEventKind.Ready,Verified:true),9000); return s; }
        var launch=new GuardianEvent(GuardianEventKind.Launch,GuardianIdentityTests.Evidence,true,LaunchedTarget:GuardianIdentityTests.Expected);
        var good=Ready(); Assert.Equal(GuardianPhase.AwaitBound,good.Advance(launch,500).State);
        Assert.Equal(GuardianPhase.Armed,good.Advance(GuardianStateTests.Valid(GuardianEventKind.Bound),10000).State);
        var late=Ready(); Assert.Equal(GuardianPhase.Finished,late.Advance(launch,10001).State);
        var backwards=Ready(); backwards.Advance(launch,500);
        Assert.Equal(GuardianPhase.Finished,backwards.Advance(GuardianStateTests.Valid(GuardianEventKind.Bound),499).State);
        var duplicate=Ready(); duplicate.Advance(launch,500);
        Assert.Equal(GuardianPhase.Finished,duplicate.Advance(launch,500).State);
        var delayedBinding=Ready(); delayedBinding.Advance(launch,9999);
        Assert.Equal(GuardianPhase.Finished,delayedBinding.Advance(GuardianStateTests.Valid(GuardianEventKind.Bound),10001).State);
    }
    private static string Absent=>new('0',64);
    [Fact] public void WireLifecycleDoesNotInventUnlaunchedIdentityOrUndisclosedChallenge()
    {
        var full=GuardianProtocolTests.Frame();
        GuardianFrame Wire(GuardianFrame f)=>new GuardianProtocolSession(f).Receive(GuardianProtocol.Encode(f,GuardianProtocolTests.Key),GuardianProtocolTests.Key);
        var ready=full with {Domain=ChannelDomain.Control,Direction=FrameDirection.GuardianToCoordinator,
            Kind=FrameKind.Ready,Policy=full.Policy with {T0Ticks=0},App=ProcessIdentity.Unassigned,
            GuardianChallenge=Absent,AppChallenge=Absent,Reason=CloseReason.None};
        Assert.Equal(ProcessIdentity.Unassigned,Wire(ready).App);
        var assigned=ready with {Direction=FrameDirection.CoordinatorToGuardian,Kind=FrameKind.AssignTarget,
            Policy=full.Policy,App=full.App};
        Assert.Equal(full.App,Wire(assigned).App);
        var hello=full with {Direction=FrameDirection.AppToGuardian,Kind=FrameKind.Hello,
            GuardianChallenge=Absent,Reason=CloseReason.None};
        Assert.Equal(Absent,Wire(hello).GuardianChallenge);
        var challenge=full with {Kind=FrameKind.Challenge,Reason=CloseReason.None};
        Assert.Equal(full.GuardianChallenge,Wire(challenge).GuardianChallenge);
        Assert.Equal(full.AppChallenge,Wire(challenge).AppChallenge);
        Wire(full with {Direction=FrameDirection.AppToGuardian,Kind=FrameKind.Proof,Reason=CloseReason.None,Sequence=2});
        Wire(full with {Kind=FrameKind.Bound,Reason=CloseReason.None,Sequence=2});
        Wire(full with {Direction=FrameDirection.AppToGuardian,Kind=FrameKind.Bound,Reason=CloseReason.None,Sequence=3});
        Wire(full with {Sequence=3});
        GuardianFrame[] invalid=[ready with {App=full.App},ready with {Policy=full.Policy},
            assigned with {App=ProcessIdentity.Unassigned},hello with {GuardianChallenge=full.GuardianChallenge},
            full with {App=ProcessIdentity.Unassigned},full with {GuardianChallenge=Absent},full with {AppChallenge=Absent}];
        foreach(var frame in invalid) Assert.Throws<GuardianProtocolException>(()=>GuardianProtocol.Encode(frame,GuardianProtocolTests.Key));
    }
}
