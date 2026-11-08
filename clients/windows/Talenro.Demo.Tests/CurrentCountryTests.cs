using Talenro.Demo.Core;
using Xunit;

namespace Talenro.Demo.Tests;

public class CurrentCountryTests
{
    private static DemoState Start() => new() { HasSession = true, PreferredNode = "jp-1", Page = DemoPage.Connection };
    private static DemoState Connected() => DemoReducer.Reduce(
        DemoReducer.Reduce(Start(), new RequestConnection()), new CompleteConnection());

    [Fact]
    public void RecoveryCompletionThenRelockUsesActualRecoveredCountry()
    {
        var state = Start();
        foreach (var action in new DemoEvent[] { new SetCountryLock(true), new RequestConnection(),
                     new FailLockedCountry(), new RecoverAcrossCountriesOnce(), new CompleteConnection() })
            state = DemoReducer.Reduce(state, action);
        Assert.Equal("DE", state.CurrentCountry);
        Assert.Equal("JP", state.LockedCountry);
        Assert.Equal("jp-1", state.PreferredNode);
        Assert.Null(state.RecoveryCountry);
        state = DemoReducer.Reduce(state, new SetCountryLock(false));
        state = DemoReducer.Reduce(state, new SetCountryLock(true));
        Assert.Equal("DE", state.LockedCountry);
    }

    [Theory]
    [InlineData("jp-1", "JP")]
    [InlineData("de-1", "DE")]
    [InlineData(null, null)]
    public void NormalCompletionRecordsOnlyKnownEnabledFixtureCountry(string? node, string? country)
    {
        var state = DemoReducer.Reduce(Start() with { PreferredNode = node }, new RequestConnection());
        state = DemoReducer.Reduce(state, new CompleteConnection());
        Assert.Equal(country, state.CurrentCountry);
    }

    [Fact]
    public void DisconnectCompletionClearsCurrentButNotSavedPreference()
    {
        var state = Connected();
        Assert.Equal("JP", state.CurrentCountry);
        state = DemoReducer.Reduce(state, new RequestDisconnection());
        Assert.Equal("JP", state.CurrentCountry);
        state = DemoReducer.Reduce(state, new CompleteDisconnection());
        Assert.Null(state.CurrentCountry);
        Assert.Equal("jp-1", state.PreferredNode);
    }

    [Fact]
    public void ExhaustionHardStopClearsCurrentCountry()
    {
        var state = DemoReducer.Reduce(Connected(), new SetEntitlement(EntitlementStatus.Exhausted));
        Assert.Equal(ConnectionPhase.Disconnected, state.Connection);
        Assert.Null(state.CurrentCountry);
    }

    [Fact]
    public void ConfirmedNodeChangeClearsOldCountryUntilExplicitCompletion()
    {
        var state = DemoReducer.Reduce(Connected(), new Navigate(DemoPage.Nodes));
        state = DemoReducer.Reduce(state, new SelectNode("de-1"));
        state = DemoReducer.Reduce(state, new ApplyNodeSelection());
        state = DemoReducer.Reduce(state, new ConfirmNodeChange());
        Assert.Null(state.CurrentCountry);
        state = DemoReducer.Reduce(state, new CompleteConnection());
        Assert.Equal("DE", state.CurrentCountry);
    }

    [Fact]
    public void CancelOrFailureNeverLeavesStaleCurrentCountry()
    {
        foreach (var action in new DemoEvent[] { new CancelConnection(), new FailConnection(), new FailLockedCountry() })
        {
            var before = Start() with { Connection = ConnectionPhase.Connecting, CurrentCountry = "DE", LockedCountry = "JP" };
            Assert.Null(DemoReducer.Reduce(before, action).CurrentCountry);
        }
    }
}
