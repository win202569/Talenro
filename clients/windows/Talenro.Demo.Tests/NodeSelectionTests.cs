using Talenro.Demo.Core;
using Xunit;

namespace Talenro.Demo.Tests;

public class NodeSelectionTests
{
    // Synthetic catalog contract: jp-1/JP and de-1/DE enabled; disabled-1 disabled.
    // null means Auto, not absence of a pending change.
    private static DemoState Editing(ConnectionPhase phase = ConnectionPhase.Disconnected) => new()
    {
        HasSession = true, Page = DemoPage.Nodes, Connection = phase,
        PreferredNode = "jp-1", DraftNode = "jp-1"
    };

    [Fact]
    public void OpeningNodesCopiesCommittedPreferenceIntoDraft()
    {
        var before = Editing() with { Page = DemoPage.Connection, DraftNode = "de-1" };
        var after = DemoReducer.Reduce(before, new Navigate(DemoPage.Nodes));
        Assert.Equal(DemoPage.Nodes, after.Page);
        Assert.Equal("jp-1", after.DraftNode);
        Assert.Equal("jp-1", after.PreferredNode);
    }

    [Fact]
    public void SelectingOnlyChangesDraft()
    {
        var before = Editing();
        Assert.Equal(before with { DraftNode = "de-1" },
            DemoReducer.Reduce(before, new SelectNode("de-1")));
    }

    [Fact]
    public void BackDiscardsDraftAndPendingConfirmation()
    {
        var before = Editing() with { DraftNode = "de-1", PendingNode = "de-1", HasPendingNodeChange = true };
        var after = DemoReducer.Reduce(before, new CancelNodeSelection());
        Assert.Equal(DemoPage.Connection, after.Page);
        Assert.Equal("jp-1", after.PreferredNode);
        Assert.Equal("jp-1", after.DraftNode);
        Assert.Null(after.PendingNode);
        Assert.False(after.HasPendingNodeChange);
    }

    [Theory]
    [InlineData("de-1")]
    [InlineData(null)]
    public void DisconnectedApplyCommitsWithoutConnecting(string? node)
    {
        var after = DemoReducer.Reduce(Editing() with { DraftNode = node }, new ApplyNodeSelection());
        Assert.Equal(node, after.PreferredNode);
        Assert.Equal(DemoPage.Connection, after.Page);
        Assert.Equal(ConnectionPhase.Disconnected, after.Connection);
        Assert.False(after.HasPendingNodeChange);
    }

    [Theory]
    [InlineData("disabled-1")]
    [InlineData("missing-node")]
    public void InvalidNodeCannotBeSelectedOrApplied(string node)
    {
        var before = Editing();
        var selected = DemoReducer.Reduce(before, new SelectNode(node));
        Assert.Equal("jp-1", selected.DraftNode);
        Assert.Equal("jp-1", selected.PreferredNode);
        Assert.False(string.IsNullOrWhiteSpace(selected.Notice));
        var applied = DemoReducer.Reduce(before with { DraftNode = node }, new ApplyNodeSelection());
        Assert.Equal("jp-1", applied.PreferredNode);
        Assert.Equal(ConnectionPhase.Disconnected, applied.Connection);
        Assert.False(applied.HasPendingNodeChange);
        Assert.False(string.IsNullOrWhiteSpace(applied.Notice));
    }

    [Fact]
    public void ConnectedApplyRequiresConfirmationBeforeAnyCommit()
    {
        var before = Editing(ConnectionPhase.Connected) with { DraftNode = "de-1" };
        var after = DemoReducer.Reduce(before, new ApplyNodeSelection());
        Assert.Equal("jp-1", after.PreferredNode);
        Assert.Equal(ConnectionPhase.Connected, after.Connection);
        Assert.True(after.HasPendingNodeChange);
        Assert.Equal("de-1", after.PendingNode);
        Assert.False(string.IsNullOrWhiteSpace(after.Notice));
    }

    [Fact]
    public void CancelConfirmationPreservesConnectionPreferenceAndCountryLock()
    {
        var before = Editing(ConnectionPhase.Connected) with { DraftNode = "de-1",
            LockedCountry = "JP", PendingNode = "de-1", HasPendingNodeChange = true };
        var after = DemoReducer.Reduce(before, new CancelNodeChange());
        Assert.Equal("jp-1", after.PreferredNode);
        Assert.Equal("JP", after.LockedCountry);
        Assert.Equal(ConnectionPhase.Connected, after.Connection);
        Assert.False(after.HasPendingNodeChange);
        Assert.Null(after.PendingNode);
    }

    [Theory]
    [InlineData("de-1")]
    [InlineData(null)]
    public void ConfirmCommitsCapturedSelectionAndStartsConnecting(string? node)
    {
        var before = Editing(ConnectionPhase.Connected) with {
            DraftNode = node, PendingNode = node, HasPendingNodeChange = true };
        var after = DemoReducer.Reduce(before, new ConfirmNodeChange());
        Assert.Equal(node, after.PreferredNode);
        Assert.Equal(ConnectionPhase.Connecting, after.Connection);
        Assert.Equal(DemoPage.Connection, after.Page);
        Assert.False(after.HasPendingNodeChange);
        Assert.Null(after.PendingNode);
    }

    [Fact]
    public void CrossCountryConfirmationNamesBothCountriesBeforeAtomicCommit()
    {
        var before = Editing(ConnectionPhase.Connected) with { DraftNode = "de-1", LockedCountry = "JP" };
        var pending = DemoReducer.Reduce(before, new ApplyNodeSelection());
        Assert.Equal("NoticeConfirmCountry", pending.Notice);
        Assert.Equal(new NoticeContext("JP", "DE"), pending.NoticeContext);
        Assert.Equal("JP", pending.LockedCountry);
        Assert.Equal("jp-1", pending.PreferredNode);
        var after = DemoReducer.Reduce(pending, new ConfirmNodeChange());
        Assert.Equal("DE", after.LockedCountry);
        Assert.Equal("de-1", after.PreferredNode);
        Assert.Equal(ConnectionPhase.Connecting, after.Connection);
    }

    [Theory]
    [InlineData(ConnectionPhase.Connecting)]
    [InlineData(ConnectionPhase.Disconnecting)]
    public void TransitionalPhasesRefuseEditingAndApply(ConnectionPhase phase)
    {
        var before = Editing(phase) with { DraftNode = "de-1" };
        foreach (var action in new DemoEvent[] { new SelectNode(null), new ApplyNodeSelection() })
        {
            var after = DemoReducer.Reduce(before, action);
            Assert.Equal(before, after with { Notice = before.Notice });
            Assert.False(string.IsNullOrWhiteSpace(after.Notice));
        }
    }

    [Fact]
    public void PendingDialogFreezesDraftUntilResolved()
    {
        var before = Editing(ConnectionPhase.Connected) with { DraftNode = "de-1",
            PendingNode = "de-1", HasPendingNodeChange = true };
        var after = DemoReducer.Reduce(before, new SelectNode(null));
        Assert.Equal(before, after with { Notice = before.Notice });
    }

    [Fact]
    public void ConfirmationWithoutPendingChangeDoesNothing()
    {
        var before = Editing(ConnectionPhase.Connected);
        Assert.Equal(before, DemoReducer.Reduce(before, new ConfirmNodeChange()));
    }

    [Fact]
    public void UnauthenticatedSelectionEventsCannotChangeState()
    {
        foreach (var action in new DemoEvent[] { new SelectNode("de-1"), new ApplyNodeSelection(),
                     new CancelNodeSelection(), new ConfirmNodeChange(), new CancelNodeChange() })
            Assert.Equal(new DemoState(), DemoReducer.Reduce(new DemoState(), action));
    }
}
