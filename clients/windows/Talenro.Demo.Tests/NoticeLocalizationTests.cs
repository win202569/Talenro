using Talenro.Demo.App.ViewModels;
using Talenro.Demo.Core;
using Xunit;

namespace Talenro.Demo.Tests;

public class NoticeLocalizationTests
{
    private static string Render(DemoState state, string language) =>
        NoticeFormatter.Format(state, key => ShellContractTests.Copy(language, key));

    [Theory]
    [InlineData(EntitlementStatus.Expired, "NoticeExpired", "expired")]
    [InlineData(EntitlementStatus.Exhausted, "NoticeExhausted", "exhausted")]
    [InlineData(EntitlementStatus.Unknown, "NoticeUnknown", "unknown")]
    [InlineData(EntitlementStatus.Stale, "NoticeStale", "stale")]
    [InlineData(EntitlementStatus.DeviceLimit, "NoticeDeviceLimit", "device limit")]
    public void RefusalCarriesSemanticKeyAndRendersBothLanguages(EntitlementStatus status, string key, string englishMeaning)
    {
        var state = DemoReducer.Reduce(new DemoState { HasSession = true, Entitlement = status }, new RequestConnection());
        Assert.Equal(key, state.Notice);
        Assert.Contains(englishMeaning, Render(state, "en-US"), StringComparison.OrdinalIgnoreCase);
        Assert.Contains("演示", Render(state, "zh-Hans"));
        Assert.Equal(ConnectionPhase.Disconnected, state.Connection);
    }

    [Fact]
    public void CrossCountryConfirmationFormatsStructuredCountriesWithoutCommittingThem()
    {
        var before = new DemoState { HasSession = true, Page = DemoPage.Nodes,
            Connection = ConnectionPhase.Connected, PreferredNode = "jp-1", LockedCountry = "JP", DraftNode = "de-1" };
        var state = DemoReducer.Reduce(before, new ApplyNodeSelection());
        Assert.Equal("NoticeConfirmCountry", state.Notice);
        Assert.Equal(new NoticeContext("JP", "DE"), state.NoticeContext);
        Assert.Equal("Switch from JP to DE and update country lock to DE? The simulated connection will restart.", Render(state, "en-US"));
        Assert.Equal("确认从 JP 切换至 DE 并将国家锁定更新为 DE？模拟连接将重新开始。", Render(state, "zh-Hans"));
        Assert.Equal("jp-1", state.PreferredNode);
        Assert.Equal("JP", state.LockedCountry);
        foreach (var action in new DemoEvent[] { new CancelNodeChange(), new ConfirmNodeChange(), new Logout() })
        {
            var after = DemoReducer.Reduce(state, action);
            Assert.Equal(new NoticeContext(), after.NoticeContext);
            Assert.Equal("", Render(after, "en-US"));
        }
    }

    [Fact]
    public void PlaceholderUsesSameSemanticNoticeAndDoesNotMutateDomainState()
    {
        var before = new DemoState { HasSession = true, PreferredNode = "de-1" };
        var state = DemoReducer.Reduce(before, new ShowPlaceholder("delete"));
        Assert.Equal("NoticePlaceholder", state.Notice);
        Assert.Equal("This feature is outside this demo.", Render(state, "en-US"));
        Assert.Equal("此功能不在本次演示范围内", Render(state, "zh-Hans"));
        Assert.Equal(before, state with { Notice = before.Notice });
    }

    [Theory]
    [InlineData("NoticeConnectionFailed")]
    [InlineData("NoticeLockedFailed")]
    [InlineData("NoticePermissionUnknown")]
    [InlineData("NoticeFinishOperation")]
    [InlineData("NoticeChooseCountry")]
    [InlineData("NoticeNoRecoveryNode")]
    [InlineData("NoticeCannotEdit")]
    [InlineData("NoticePendingChange")]
    [InlineData("NoticeNodeUnavailable")]
    [InlineData("NoticeCannotApply")]
    [InlineData("NoticeConfirmNode")]
    public void OtherNoticeKeysRenderResourceCopyNotRawIdentifiers(string key)
    {
        var state = new DemoState { Notice = key };
        Assert.NotEqual(key, Render(state, "en-US"));
        Assert.NotEqual(key, Render(state, "zh-Hans"));
    }

    [Fact]
    public void EmptyNoticeDoesNotLookUpAnEmptyResourceKey()
    {
        Assert.Equal("", NoticeFormatter.Format(new DemoState(), _ => throw new InvalidOperationException("Unexpected lookup")));
    }
}
