using Talenro.Demo.Core;
using Xunit;

namespace Talenro.Demo.Tests;

public class ConnectionTests
{
    private static DemoState At(ConnectionPhase phase) => new()
    {
        HasSession = true, Page = DemoPage.Connection, Connection = phase,
        PreferredNode = "jp-1", LockedCountry = "JP"
    };

    // Missing/wrong transition branches must fail, independently of UI or timers.
    [Fact]
    public void RequestOnlyStartsConnecting()
    {
        var before = At(ConnectionPhase.Disconnected);
        var after = DemoReducer.Reduce(before, new RequestConnection());
        Assert.Equal(before with { Connection = ConnectionPhase.Connecting }, after);
    }

    [Fact]
    public void ExplicitCompletionMarksSimulatedConnectionConnected()
    {
        var before = At(ConnectionPhase.Connecting);
        Assert.Equal(before with { Connection = ConnectionPhase.Connected, CurrentCountry = "JP" },
            DemoReducer.Reduce(before, new CompleteConnection()));
    }

    [Fact]
    public void CancelReturnsToDisconnectedWithoutLosingPreferences()
    {
        var before = At(ConnectionPhase.Connecting);
        Assert.Equal(before with { Connection = ConnectionPhase.Disconnected },
            DemoReducer.Reduce(before, new CancelConnection()));
    }

    [Fact]
    public void DisconnectRequiresExplicitCompletion()
    {
        var before = At(ConnectionPhase.Connected);
        var pending = DemoReducer.Reduce(before, new RequestDisconnection());
        Assert.Equal(before with { Connection = ConnectionPhase.Disconnecting }, pending);
        Assert.Equal(before with { Connection = ConnectionPhase.Disconnected },
            DemoReducer.Reduce(pending, new CompleteDisconnection()));
    }

    [Fact]
    public void CompletionAfterCancelCannotReviveConnection()
    {
        var canceled = DemoReducer.Reduce(At(ConnectionPhase.Connecting), new CancelConnection());
        var late = DemoReducer.Reduce(canceled, new CompleteConnection());
        Assert.Equal(ConnectionPhase.Disconnected, late.Connection);
        Assert.Equal(canceled, late);
    }

    [Fact]
    public void CompletionAfterLogoutPreservesEntireInitialState()
    {
        var loggedOut = DemoReducer.Reduce(At(ConnectionPhase.Connecting), new Logout());
        Assert.Equal(new DemoState(), DemoReducer.Reduce(loggedOut, new CompleteConnection()));
    }

    [Theory]
    [InlineData(ConnectionPhase.Disconnected)]
    [InlineData(ConnectionPhase.Connected)]
    [InlineData(ConnectionPhase.Disconnecting)]
    public void CompletionOutsideConnectingIsIgnored(ConnectionPhase phase)
    {
        var before = At(phase);
        Assert.Equal(before, DemoReducer.Reduce(before, new CompleteConnection()));
    }

    [Theory]
    [InlineData(ConnectionPhase.Connecting)]
    [InlineData(ConnectionPhase.Connected)]
    [InlineData(ConnectionPhase.Disconnecting)]
    public void ConnectRequestOutsideDisconnectedIsIgnored(ConnectionPhase phase)
    {
        var before = At(phase);
        Assert.Equal(before, DemoReducer.Reduce(before, new RequestConnection()));
    }

    [Theory]
    [InlineData(ConnectionPhase.Disconnected)]
    [InlineData(ConnectionPhase.Connecting)]
    [InlineData(ConnectionPhase.Disconnecting)]
    public void DisconnectRequestOutsideConnectedIsIgnored(ConnectionPhase phase)
    {
        var before = At(phase);
        Assert.Equal(before, DemoReducer.Reduce(before, new RequestDisconnection()));
    }

    [Theory]
    [InlineData(ConnectionPhase.Disconnected)]
    [InlineData(ConnectionPhase.Connecting)]
    [InlineData(ConnectionPhase.Connected)]
    public void DisconnectCompletionOutsideDisconnectingIsIgnored(ConnectionPhase phase)
    {
        var before = At(phase);
        Assert.Equal(before, DemoReducer.Reduce(before, new CompleteDisconnection()));
    }

    [Theory]
    [InlineData(ConnectionPhase.Disconnected)]
    [InlineData(ConnectionPhase.Connected)]
    [InlineData(ConnectionPhase.Disconnecting)]
    public void CancelOutsideConnectingIsIgnored(ConnectionPhase phase)
    {
        var before = At(phase);
        Assert.Equal(before, DemoReducer.Reduce(before, new CancelConnection()));
    }

    [Fact]
    public void UnauthenticatedConnectionEventsDoNothing()
    {
        DemoEvent[] events = [new RequestConnection(), new CompleteConnection(),
            new CancelConnection(), new RequestDisconnection(), new CompleteDisconnection(),
            new FailConnection()];
        foreach (var action in events)
            Assert.Equal(new DemoState(), DemoReducer.Reduce(new DemoState(), action));
    }

    [Fact]
    public void RepeatedRequestsNeverAdvancePastTransitionalPhase()
    {
        var connecting = DemoReducer.Reduce(At(ConnectionPhase.Disconnected), new RequestConnection());
        Assert.Equal(ConnectionPhase.Connecting, connecting.Connection);
        Assert.Equal(connecting, DemoReducer.Reduce(connecting, new RequestConnection()));
        var disconnecting = DemoReducer.Reduce(At(ConnectionPhase.Connected), new RequestDisconnection());
        Assert.Equal(ConnectionPhase.Disconnecting, disconnecting.Connection);
        Assert.Equal(disconnecting, DemoReducer.Reduce(disconnecting, new RequestDisconnection()));
    }

    [Fact]
    public void FailureWaitsForExplicitRetry()
    {
        var failed = DemoReducer.Reduce(At(ConnectionPhase.Connecting), new FailConnection());
        Assert.Equal(ConnectionPhase.Disconnected, failed.Connection);
        Assert.False(string.IsNullOrWhiteSpace(failed.Notice));
        Assert.Equal("jp-1", failed.PreferredNode);
        Assert.Equal("JP", failed.LockedCountry);
        Assert.Equal(failed, DemoReducer.Reduce(failed, new CompleteConnection()));
        var retry = DemoReducer.Reduce(failed, new RequestConnection());
        Assert.Equal(ConnectionPhase.Connecting, retry.Connection);
        Assert.Equal("", retry.Notice);
    }
}
