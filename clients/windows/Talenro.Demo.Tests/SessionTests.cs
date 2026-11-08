using Talenro.Demo.Core;
using Xunit;

namespace Talenro.Demo.Tests;

public class SessionTests
{
    // A missing session transition or a partial logout reset must fail these tests.
    [Fact]
    public void EnterDemoOpensConnectionWithoutCredentials()
    {
        var state = DemoReducer.Reduce(DemoFixtures.Initial(), new EnterDemo());
        Assert.True(state.HasSession);
        Assert.Equal(DemoPage.Connection, state.Page);
        Assert.Equal(ConnectionPhase.Disconnected, state.Connection);
    }

    [Fact]
    public void LogoutClearsAllSessionState()
    {
        var dirty = new DemoState { HasSession = true, Page = DemoPage.Nodes,
            Connection = ConnectionPhase.Connected, PreferredNode = "jp-1", DraftNode = "de-1",
            LockedCountry = "JP", PendingNode = "de-1", Notice = "old",
            Entitlement = EntitlementStatus.Exhausted, Search = "abc" };
        var state = DemoReducer.Reduce(dirty, new Logout());
        Assert.Equal(new DemoState(), state);
    }

    [Theory]
    [InlineData("reset")]
    [InlineData("register")]
    [InlineData("mfa")]
    [InlineData("language")]
    [InlineData("devices")]
    [InlineData("delete")]
    [InlineData("traffic")]
    [InlineData("settings")]
    public void PlaceholderOnlyChangesNotice(string destination)
    {
        var before = new DemoState { HasSession = true, Page = DemoPage.Account,
            PreferredNode = "jp-1", LockedCountry = "JP" };
        var after = DemoReducer.Reduce(before, new ShowPlaceholder(destination));
        Assert.Equal("NoticePlaceholder", after.Notice);
        Assert.Equal(before, after with { Notice = before.Notice });
    }

    [Theory]
    [InlineData(DemoPage.Connection)]
    [InlineData(DemoPage.Nodes)]
    [InlineData(DemoPage.Account)]
    public void NavigationRequiresDemoSession(DemoPage page)
    {
        Assert.Equal(new DemoState(), DemoReducer.Reduce(new DemoState(), new Navigate(page)));
        var entered = new DemoState { HasSession = true, Page = DemoPage.Connection };
        Assert.Equal(page, DemoReducer.Reduce(entered, new Navigate(page)).Page);
    }
}
