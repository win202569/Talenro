using Talenro.Observation.CloseGuardian;
using Xunit;
namespace Talenro.Demo.Tests;
public class GuardianApplicationCloseTests
{
    internal static readonly byte[] Key=Enumerable.Repeat((byte)0x11,32).ToArray();
    internal static GuardianFrame Frame=>new(ChannelDomain.Application,FrameDirection.GuardianToApp,FrameKind.CloseNormal,
        new(Guid.Parse("11111111-1111-1111-1111-111111111111"),new string('2',64),1000,10000000),1,
        GuardianIdentityTests.Expected,GuardianIdentityTests.Expected,new string('3',64),new string('4',64),300000,CloseReason.Deadline);
    internal static AppEvent Request=>new(AppEventKind.Request,GuardianProtocol.Encode(Frame,Key),GuardianIdentityTests.Evidence);
    internal static ApplicationCloseState Queued()
    {
        var s=new ApplicationCloseState(Frame,7);
        Assert.Equal(CloseEffect.AcceptReceipt,s.Advance(Request,290000,Key).Effect);
        Assert.Equal(CloseEffect.EnqueueClose,s.Advance(new(AppEventKind.ReceiptWritten),290001).Effect);
        s.Advance(new(AppEventKind.Enqueued),290002); return s;
    }
    [Theory][InlineData(299999,7,true,true)][InlineData(300000,7,true,false)]
    [InlineData(299999,8,true,false)][InlineData(299999,7,false,false)]
    public void CallbackRechecksDeadlineAndOriginalObject(long now,long generation,bool alive,bool invokes)
    {
        var s=Queued(); var callback=new AppEvent(AppEventKind.Callback,WindowGeneration:generation,WindowAlive:alive);
        Assert.Equal(invokes,s.Advance(callback,now).Effect==CloseEffect.InvokeOwnClose);
        Assert.NotEqual(CloseEffect.InvokeOwnClose,s.Advance(callback,now).Effect);
    }
    [Fact] public void DuplicateRequestDoesNotScheduleTwice()
    {
        var s=Queued(); Assert.Null(s.Advance(Request,290003,Key).Effect);
        Assert.Equal(CloseEffect.InvokeOwnClose,s.Advance(new(AppEventKind.Callback,WindowGeneration:7),290004).Effect);
    }
    [Fact] public void BadMacPeerReceiptQueueAndGuardianLossNeverInvoke()
    {
        var bad=Request; bad.Frame![10]^=1;
        var s=new ApplicationCloseState(Frame,7);
        Assert.NotEqual(CloseEffect.AcceptReceipt,s.Advance(bad,290000,Key).Effect);
        var wrong=new ApplicationCloseState(Frame,7);
        Assert.NotEqual(CloseEffect.AcceptReceipt,wrong.Advance(Request with {Peer=GuardianIdentityTests.Evidence with {ConnectedPeerPid=1}},290000,Key).Effect);
        foreach(var failure in new[]{AppEventKind.ReceiptFailed,AppEventKind.EnqueueFailed,AppEventKind.GuardianLost,AppEventKind.Fault})
        {
            var a=new ApplicationCloseState(Frame,7); a.Advance(Request,290000,Key); a.Advance(new(failure),290001);
            Assert.NotEqual(CloseEffect.EnqueueClose,a.Advance(new(AppEventKind.ReceiptWritten),290002).Effect);
            Assert.NotEqual(CloseEffect.InvokeOwnClose,a.Advance(new(AppEventKind.Callback,WindowGeneration:7),290003).Effect);
        }
    }
    [Fact] public void CallbackMayRunBeforeEnqueueAcknowledgementButOnlyOnce()
    {
        var s=new ApplicationCloseState(Frame,7); s.Advance(Request,290000,Key);
        s.Advance(new(AppEventKind.ReceiptWritten),290001);
        Assert.Equal(CloseEffect.InvokeOwnClose,s.Advance(new(AppEventKind.Callback,WindowGeneration:7),290002).Effect);
        Assert.Null(s.Advance(new(AppEventKind.Enqueued),290003).Effect);
    }
    [Fact] public void BackwardClockAndHungReceiptExpireWithoutAction()
    {
        var s=Queued(); Assert.Equal(CloseEffect.Finish,s.Advance(new(AppEventKind.Tick),290001).Effect);
        var hung=new ApplicationCloseState(Frame,7); hung.Advance(Request,290000,Key);
        Assert.Equal(CloseEffect.Finish,hung.Advance(new(AppEventKind.Tick),360000).Effect);
        Assert.Null(hung.Advance(new(AppEventKind.ReceiptWritten),360001).Effect);
    }
}
