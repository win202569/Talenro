using Talenro.Demo.Core;
using Xunit;

namespace Talenro.Demo.Tests;

public class EntitlementTests
{
    public static IEnumerable<object[]> DeniedStatuses => new[]
    {
        new object[] { EntitlementStatus.Expired },
        new object[] { EntitlementStatus.Exhausted },
        new object[] { EntitlementStatus.Unknown },
        new object[] { EntitlementStatus.Stale },
        new object[] { EntitlementStatus.DeviceLimit }
    };

    private static DemoState Session(EntitlementStatus status) => new()
    {
        HasSession = true, Page = DemoPage.Connection, Entitlement = status,
        PreferredNode = "jp-1", LockedCountry = "JP"
    };

    [Theory]
    [MemberData(nameof(DeniedStatuses))]
    public void IneligibleStateRefusesConnectWithoutChangingPreferences(EntitlementStatus status)
    {
        var before = Session(status);
        var after = DemoReducer.Reduce(before, new RequestConnection());
        Assert.Equal(ConnectionPhase.Disconnected, after.Connection);
        Assert.Equal(before, after with { Notice = before.Notice });
        Assert.False(string.IsNullOrWhiteSpace(after.Notice));
    }

    [Theory]
    [MemberData(nameof(DeniedStatuses))]
    public void IneligibleConfirmationCannotCommitNodeOrLock(EntitlementStatus status)
    {
        var before = Session(status) with { Page = DemoPage.Nodes, Connection = ConnectionPhase.Connected,
            DraftNode = "de-1", PendingNode = "de-1", HasPendingNodeChange = true };
        var after = DemoReducer.Reduce(before, new ConfirmNodeChange());
        Assert.NotEqual(ConnectionPhase.Connecting, after.Connection);
        Assert.Equal("jp-1", after.PreferredNode);
        Assert.Equal("JP", after.LockedCountry);
        Assert.False(after.HasPendingNodeChange);
        Assert.Null(after.PendingNode);
        Assert.False(string.IsNullOrWhiteSpace(after.Notice));
    }

    [Theory]
    [MemberData(nameof(DeniedStatuses))]
    public void NeitherRecoveryOptionBypassesEligibility(EntitlementStatus status)
    {
        var before = Session(status) with { RecoveryChoices = LockRecoveryChoices.CrossCountryOnce
            | LockRecoveryChoices.RetryLockedCountry | LockRecoveryChoices.StayDisconnected };
        foreach (var action in new DemoEvent[] { new RecoverAcrossCountriesOnce(), new RetryLockedCountry() })
        {
            var after = DemoReducer.Reduce(before, action);
            Assert.Equal(ConnectionPhase.Disconnected, after.Connection);
            Assert.Null(after.RecoveryCountry);
            Assert.Equal("JP", after.LockedCountry);
            Assert.Equal("jp-1", after.PreferredNode);
            Assert.False(string.IsNullOrWhiteSpace(after.Notice));
        }
    }

    [Theory]
    [MemberData(nameof(DeniedStatuses))]
    public void AccountRemainsAccessibleWhenConnectionIsRefused(EntitlementStatus status)
    {
        var after = DemoReducer.Reduce(Session(status), new Navigate(DemoPage.Account));
        Assert.Equal(DemoPage.Account, after.Page);
        Assert.Equal(status, after.Entitlement);
        Assert.True(after.HasSession);
    }

    [Theory]
    [MemberData(nameof(DeniedStatuses))]
    public void ScenarioRetainsExactEntitlementWithoutInventingActiveState(EntitlementStatus status)
    {
        var after = DemoReducer.Reduce(Session(EntitlementStatus.Active), new SetEntitlement(status));
        Assert.Equal(status, after.Entitlement);
        Assert.Equal(ConnectionPhase.Disconnected, after.Connection);
        Assert.True(after.HasSession);
    }

    [Theory]
    [InlineData(ConnectionPhase.Connected)]
    [InlineData(ConnectionPhase.Connecting)]
    public void QuotaExhaustionHardStopsAndClearsPendingOperations(ConnectionPhase phase)
    {
        var before = Session(EntitlementStatus.Active) with { Connection = phase, RecoveryCountry = "DE",
            PendingNode = "de-1", HasPendingNodeChange = true,
            RecoveryChoices = LockRecoveryChoices.CrossCountryOnce };
        var after = DemoReducer.Reduce(before, new SetEntitlement(EntitlementStatus.Exhausted));
        Assert.Equal(EntitlementStatus.Exhausted, after.Entitlement);
        Assert.Equal(ConnectionPhase.Disconnected, after.Connection);
        Assert.Null(after.RecoveryCountry);
        Assert.Null(after.PendingNode);
        Assert.False(after.HasPendingNodeChange);
        Assert.Equal(LockRecoveryChoices.None, after.RecoveryChoices);
        Assert.False(string.IsNullOrWhiteSpace(after.Notice));
        Assert.Equal("jp-1", after.PreferredNode);
        Assert.Equal("JP", after.LockedCountry);
        Assert.Equal(after, DemoReducer.Reduce(after, new CompleteConnection()));
    }

    [Theory]
    [MemberData(nameof(DeniedStatuses))]
    public void EntitlementInvalidatedAfterDialogCannotReconnect(EntitlementStatus status)
    {
        var before = Session(EntitlementStatus.Active) with { Page = DemoPage.Nodes,
            Connection = ConnectionPhase.Connected, DraftNode = "de-1" };
        var pending = DemoReducer.Reduce(before, new ApplyNodeSelection());
        Assert.True(pending.HasPendingNodeChange);
        var invalidated = DemoReducer.Reduce(pending, new SetEntitlement(status));
        var after = DemoReducer.Reduce(invalidated, new ConfirmNodeChange());
        Assert.Equal(status, after.Entitlement);
        Assert.NotEqual(ConnectionPhase.Connecting, after.Connection);
        Assert.Equal("jp-1", after.PreferredNode);
        Assert.Equal("JP", after.LockedCountry);
        Assert.False(after.HasPendingNodeChange);
    }

    [Theory]
    [MemberData(nameof(DeniedStatuses))]
    public void LateCompletionCannotAdmitAnIneligibleSession(EntitlementStatus status)
    {
        var before = Session(status) with { Connection = ConnectionPhase.Connecting, RecoveryCountry = "DE" };
        var after = DemoReducer.Reduce(before, new CompleteConnection());
        Assert.Equal(ConnectionPhase.Disconnected, after.Connection);
        Assert.Null(after.RecoveryCountry);
        Assert.Equal(status, after.Entitlement);
        Assert.False(string.IsNullOrWhiteSpace(after.Notice));
    }

    [Fact]
    public void RestoringEligibilityDoesNotAutomaticallyConnect()
    {
        var after = DemoReducer.Reduce(Session(EntitlementStatus.Exhausted), new SetEntitlement(EntitlementStatus.Active));
        Assert.Equal(EntitlementStatus.Active, after.Entitlement);
        Assert.Equal(ConnectionPhase.Disconnected, after.Connection);
    }

    [Fact]
    public void EntitlementEventDoesNotCreateUnauthenticatedSession()
    {
        Assert.Equal(new DemoState(), DemoReducer.Reduce(new DemoState(), new SetEntitlement(EntitlementStatus.Exhausted)));
    }
}
