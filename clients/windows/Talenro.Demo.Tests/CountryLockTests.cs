using Talenro.Demo.Core;
using Xunit;

namespace Talenro.Demo.Tests;

public class CountryLockTests
{
    private const LockRecoveryChoices ThreeChoices = LockRecoveryChoices.CrossCountryOnce
        | LockRecoveryChoices.RetryLockedCountry | LockRecoveryChoices.StayDisconnected;

    private static DemoState Failed() => new()
    {
        HasSession = true, Page = DemoPage.Connection, LockedCountry = "JP",
        PreferredNode = "jp-1", RecoveryChoices = ThreeChoices
    };

    [Fact]
    public void InitialLockIsOffAndUnknownCountryCannotEnableIt()
    {
        Assert.Null(DemoFixtures.Initial().LockedCountry);
        var after = DemoReducer.Reduce(new DemoState { HasSession = true }, new SetCountryLock(true));
        Assert.Null(after.LockedCountry);
        Assert.False(string.IsNullOrWhiteSpace(after.Notice));
    }

    [Theory]
    [InlineData("jp-1", "JP")]
    [InlineData("de-1", "DE")]
    public void EnableUsesExplicitSelectedFixtureCountry(string node, string country)
    {
        var after = DemoReducer.Reduce(new DemoState { HasSession = true, PreferredNode = node }, new SetCountryLock(true));
        Assert.Equal(country, after.LockedCountry);
        Assert.Equal(ConnectionPhase.Disconnected, after.Connection);
    }

    [Fact]
    public void EnableCanUseKnownCurrentFixtureCountry()
    {
        var after = DemoReducer.Reduce(new DemoState { HasSession = true,
            Connection = ConnectionPhase.Connected, CurrentCountry = "DE" }, new SetCountryLock(true));
        Assert.Equal("DE", after.LockedCountry);
    }

    [Theory]
    [InlineData("missing-node")]
    [InlineData("disabled-1")]
    public void InvalidSelectionDoesNotInventLockCountry(string node)
    {
        var after = DemoReducer.Reduce(new DemoState { HasSession = true, PreferredNode = node }, new SetCountryLock(true));
        Assert.Null(after.LockedCountry);
        Assert.False(string.IsNullOrWhiteSpace(after.Notice));
    }

    [Fact]
    public void DisableClearsLockAndPendingRecoveryWithoutConnecting()
    {
        var after = DemoReducer.Reduce(Failed(), new SetCountryLock(false));
        Assert.Null(after.LockedCountry);
        Assert.Equal(LockRecoveryChoices.None, after.RecoveryChoices);
        Assert.Null(after.RecoveryCountry);
        Assert.Equal(ConnectionPhase.Disconnected, after.Connection);
        Assert.Equal("jp-1", after.PreferredNode);
    }

    [Fact]
    public void LockedFailureOffersExactlyThreeChoicesWithoutAutoFallback()
    {
        var before = Failed() with { Connection = ConnectionPhase.Connecting, RecoveryChoices = LockRecoveryChoices.None };
        var after = DemoReducer.Reduce(before, new FailLockedCountry());
        Assert.Equal(ConnectionPhase.Disconnected, after.Connection);
        Assert.Equal(ThreeChoices, after.RecoveryChoices);
        Assert.Equal("JP", after.LockedCountry);
        Assert.Null(after.RecoveryCountry);
        Assert.False(string.IsNullOrWhiteSpace(after.Notice));
    }

    [Fact]
    public void CrossCountryRecoveryIsAnImmediateSingleAttemptNotASavedPreference()
    {
        var after = DemoReducer.Reduce(Failed(), new RecoverAcrossCountriesOnce());
        Assert.Equal(ConnectionPhase.Connecting, after.Connection);
        Assert.Equal("DE", after.RecoveryCountry);
        Assert.Equal("JP", after.LockedCountry);
        Assert.Equal("jp-1", after.PreferredNode);
        Assert.Equal(LockRecoveryChoices.None, after.RecoveryChoices);
        Assert.Equal(after, DemoReducer.Reduce(after, new RecoverAcrossCountriesOnce()));
    }

    [Fact]
    public void RetryStaysInSavedCountry()
    {
        var after = DemoReducer.Reduce(Failed(), new RetryLockedCountry());
        Assert.Equal(ConnectionPhase.Connecting, after.Connection);
        Assert.Equal("JP", after.RecoveryCountry);
        Assert.Equal("JP", after.LockedCountry);
        Assert.Equal(LockRecoveryChoices.None, after.RecoveryChoices);
    }

    [Fact]
    public void StayDisconnectedClosesChoicesWithoutRequestingConnection()
    {
        var after = DemoReducer.Reduce(Failed(), new StayDisconnected());
        Assert.Equal(ConnectionPhase.Disconnected, after.Connection);
        Assert.Equal(LockRecoveryChoices.None, after.RecoveryChoices);
        Assert.Equal("JP", after.LockedCountry);
        Assert.Null(after.RecoveryCountry);
    }

    [Fact]
    public void CancelConsumesOverrideAndNextRequestHasNoCrossCountryPermission()
    {
        var attempt = Failed() with { Connection = ConnectionPhase.Connecting,
            RecoveryChoices = LockRecoveryChoices.None, RecoveryCountry = "DE" };
        var canceled = DemoReducer.Reduce(attempt, new CancelConnection());
        Assert.Equal(ConnectionPhase.Disconnected, canceled.Connection);
        Assert.Null(canceled.RecoveryCountry);
        var next = DemoReducer.Reduce(canceled, new RequestConnection());
        Assert.Equal("JP", next.LockedCountry);
        Assert.Null(next.RecoveryCountry);
    }

    [Fact]
    public void CompletionConsumesOverrideWithoutChangingSavedLock()
    {
        var attempt = Failed() with { Connection = ConnectionPhase.Connecting,
            RecoveryChoices = LockRecoveryChoices.None, RecoveryCountry = "DE" };
        var after = DemoReducer.Reduce(attempt, new CompleteConnection());
        Assert.Equal(ConnectionPhase.Connected, after.Connection);
        Assert.Equal("JP", after.LockedCountry);
        Assert.Null(after.RecoveryCountry);
    }

    [Fact]
    public void LaterFailureRequiresANewExplicitRecoveryChoice()
    {
        var attempt = Failed() with { Connection = ConnectionPhase.Connecting,
            RecoveryChoices = LockRecoveryChoices.None, RecoveryCountry = "DE" };
        var failed = DemoReducer.Reduce(attempt, new FailLockedCountry());
        Assert.Null(failed.RecoveryCountry);
        Assert.Equal(ConnectionPhase.Disconnected, failed.Connection);
        Assert.Equal("JP", failed.LockedCountry);
        Assert.Equal(ThreeChoices, failed.RecoveryChoices);
        Assert.Equal(failed, DemoReducer.Reduce(failed, new CompleteConnection()));
    }

    [Fact]
    public void RecoveryActionsWithoutAnOfferCannotStartConnection()
    {
        var before = Failed() with { RecoveryChoices = LockRecoveryChoices.None };
        foreach (var action in new DemoEvent[] { new RecoverAcrossCountriesOnce(), new RetryLockedCountry(), new StayDisconnected() })
            Assert.Equal(before, DemoReducer.Reduce(before, action));
    }

    [Fact]
    public void OrdinaryFailureAlsoClearsSingleAttemptOverride()
    {
        var attempt = Failed() with { Connection = ConnectionPhase.Connecting,
            RecoveryChoices = LockRecoveryChoices.None, RecoveryCountry = "DE" };
        var after = DemoReducer.Reduce(attempt, new FailConnection());
        Assert.Equal(ConnectionPhase.Disconnected, after.Connection);
        Assert.Null(after.RecoveryCountry);
        Assert.Equal("JP", after.LockedCountry);
    }

    [Fact]
    public void UnauthenticatedLockEventsCannotChangeState()
    {
        foreach (var action in new DemoEvent[] { new SetCountryLock(true), new SetCountryLock(false),
                     new FailLockedCountry(), new RecoverAcrossCountriesOnce(), new RetryLockedCountry(), new StayDisconnected() })
            Assert.Equal(new DemoState(), DemoReducer.Reduce(new DemoState(), action));
    }
}
