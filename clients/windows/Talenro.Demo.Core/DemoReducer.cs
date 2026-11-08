namespace Talenro.Demo.Core;

public static class DemoReducer
{
    public static DemoState Reduce(DemoState state, DemoEvent action) => action switch
    {
        EnterDemo => state with { HasSession = true, Page = DemoPage.Connection },
        Logout => DemoFixtures.Initial(),
        Navigate { Page: DemoPage.Nodes } when state.HasSession => state with
        {
            Page = DemoPage.Nodes, DraftNode = state.PreferredNode,
            PendingNode = null, HasPendingNodeChange = false
        },
        Navigate navigation when state.HasSession => state with { Page = navigation.Page },
        ShowPlaceholder => state with { Notice = "NoticePlaceholder" },
        SetEntitlement change when state.HasSession => ChangeEntitlement(state, change.Status),
        RequestConnection when state.HasSession && state.Connection == ConnectionPhase.Disconnected
            && state.Entitlement != EntitlementStatus.Active
            => state with { Notice = EntitlementNotice(state.Entitlement), NoticeContext = new() },
        RequestConnection when state.HasSession && state.Connection == ConnectionPhase.Disconnected
            => state with { Connection = ConnectionPhase.Connecting, Notice = "", NoticeContext = new(),
                RecoveryChoices = LockRecoveryChoices.None, RecoveryCountry = null, CurrentCountry = null },
        CompleteConnection when state.HasSession && state.Connection == ConnectionPhase.Connecting
            && state.Entitlement != EntitlementStatus.Active
            => ChangeEntitlement(state, state.Entitlement),
        CompleteConnection when state.HasSession && state.Connection == ConnectionPhase.Connecting
            => state with { Connection = ConnectionPhase.Connected,
                CurrentCountry = state.RecoveryCountry ?? CountryOf(state.PreferredNode), RecoveryCountry = null },
        CancelConnection when state.HasSession && state.Connection == ConnectionPhase.Connecting
            => state with { Connection = ConnectionPhase.Disconnected, RecoveryCountry = null, CurrentCountry = null },
        RequestDisconnection when state.HasSession && state.Connection == ConnectionPhase.Connected
            => state with { Connection = ConnectionPhase.Disconnecting },
        CompleteDisconnection when state.HasSession && state.Connection == ConnectionPhase.Disconnecting
            => state with { Connection = ConnectionPhase.Disconnected, CurrentCountry = null },
        FailConnection when state.HasSession && state.Connection == ConnectionPhase.Connecting
            => state with { Connection = ConnectionPhase.Disconnected, RecoveryCountry = null, CurrentCountry = null,
                Notice = "NoticeConnectionFailed", NoticeContext = new() },
        SetCountryLock change when state.HasSession => SetLock(state, change.Enabled),
        FailLockedCountry when state.HasSession && state.LockedCountry is not null
            && state.Connection == ConnectionPhase.Connecting => state with
            {
                Connection = ConnectionPhase.Disconnected, RecoveryCountry = null, CurrentCountry = null,
                RecoveryChoices = LockRecoveryChoices.CrossCountryOnce | LockRecoveryChoices.RetryLockedCountry
                    | LockRecoveryChoices.StayDisconnected,
                Notice = "NoticeLockedFailed", NoticeContext = new()
            },
        RecoverAcrossCountriesOnce when CanRecover(state, LockRecoveryChoices.CrossCountryOnce)
            => StartRecovery(state, crossCountry: true),
        RetryLockedCountry when CanRecover(state, LockRecoveryChoices.RetryLockedCountry)
            => StartRecovery(state, crossCountry: false),
        StayDisconnected when CanRecover(state, LockRecoveryChoices.StayDisconnected)
            => state with { RecoveryChoices = LockRecoveryChoices.None, RecoveryCountry = null, Notice = "", NoticeContext = new() },
        SelectNode selection when state.HasSession && state.Page == DemoPage.Nodes
            => SelectDraft(state, selection.NodeId),
        ApplyNodeSelection when state.HasSession && state.Page == DemoPage.Nodes
            => ApplyDraft(state),
        CancelNodeSelection when state.HasSession && state.Page == DemoPage.Nodes
            => state with { Page = DemoPage.Connection, DraftNode = state.PreferredNode,
                PendingNode = null, HasPendingNodeChange = false, Notice = "", NoticeContext = new() },
        CancelNodeChange when state.HasSession && state.HasPendingNodeChange
            => state with { PendingNode = null, HasPendingNodeChange = false, Notice = "", NoticeContext = new() },
        ConfirmNodeChange when state.HasSession && state.HasPendingNodeChange
            && state.Connection == ConnectionPhase.Connected => ConfirmSelection(state),
        _ => state
    };

    private static bool IsTransitioning(DemoState state) =>
        state.Connection is ConnectionPhase.Connecting or ConnectionPhase.Disconnecting;

    private static string EntitlementNotice(EntitlementStatus status) => status switch
    {
        EntitlementStatus.Expired => "NoticeExpired",
        EntitlementStatus.Exhausted => "NoticeExhausted",
        EntitlementStatus.Unknown => "NoticeUnknown",
        EntitlementStatus.Stale => "NoticeStale",
        EntitlementStatus.DeviceLimit => "NoticeDeviceLimit",
        _ => "NoticePermissionUnknown"
    };

    private static DemoState ChangeEntitlement(DemoState state, EntitlementStatus status)
    {
        if (status == EntitlementStatus.Active)
            return state with { Entitlement = status, Notice = "", NoticeContext = new() };

        return state with
        {
            Entitlement = status,
            Connection = status == EntitlementStatus.Exhausted || state.Connection == ConnectionPhase.Connecting
                ? ConnectionPhase.Disconnected : state.Connection,
            CurrentCountry = status == EntitlementStatus.Exhausted || state.Connection == ConnectionPhase.Connecting
                ? null : state.CurrentCountry,
            PendingNode = null, HasPendingNodeChange = false,
            RecoveryCountry = null, RecoveryChoices = LockRecoveryChoices.None,
            Notice = EntitlementNotice(status), NoticeContext = new()
        };
    }

    private static bool CanRecover(DemoState state, LockRecoveryChoices choice) =>
        state.HasSession && state.LockedCountry is not null
        && state.Connection == ConnectionPhase.Disconnected && (state.RecoveryChoices & choice) != 0;

    private static DemoState SetLock(DemoState state, bool enabled)
    {
        if (IsTransitioning(state) || state.HasPendingNodeChange)
            return state with { Notice = "NoticeFinishOperation", NoticeContext = new() };
        if (!enabled)
            return state with { LockedCountry = null, RecoveryChoices = LockRecoveryChoices.None,
                RecoveryCountry = null, Notice = "", NoticeContext = new() };

        var selected = DemoFixtures.Nodes.FirstOrDefault(node => node.Id == state.PreferredNode && node.Enabled);
        var country = state.Connection == ConnectionPhase.Connected && state.CurrentCountry is not null
            ? state.CurrentCountry : selected?.Country;
        if (country is null || !DemoFixtures.Nodes.Any(node => node.Enabled && node.Country == country))
            return state with { Notice = "NoticeChooseCountry", NoticeContext = new() };
        return state with { LockedCountry = country, RecoveryChoices = LockRecoveryChoices.None,
            RecoveryCountry = null, Notice = "", NoticeContext = new() };
    }

    private static DemoState StartRecovery(DemoState state, bool crossCountry)
    {
        if (state.Entitlement != EntitlementStatus.Active)
            return state with { Notice = EntitlementNotice(state.Entitlement), NoticeContext = new() };
        var candidate = DemoFixtures.Nodes.FirstOrDefault(node => node.Enabled &&
            (crossCountry ? node.Country != state.LockedCountry : node.Country == state.LockedCountry));
        if (candidate is null)
            return state with { Notice = "NoticeNoRecoveryNode", NoticeContext = new() };
        // Consume the offered choice immediately; no reusable cross-country permission.
        return state with { Connection = ConnectionPhase.Connecting, RecoveryCountry = candidate.Country, CurrentCountry = null,
            RecoveryChoices = LockRecoveryChoices.None, Notice = "", NoticeContext = new() };
    }

    private static bool IsSelectable(string? nodeId) =>
        nodeId is null || DemoFixtures.Nodes.Any(node => node.Id == nodeId && node.Enabled);

    private static string? CountryOf(string? nodeId) =>
        DemoFixtures.Nodes.FirstOrDefault(node => node.Id == nodeId && node.Enabled)?.Country;

    private static DemoState SelectDraft(DemoState state, string? nodeId)
    {
        if (IsTransitioning(state))
            return state with { Notice = "NoticeCannotEdit", NoticeContext = new() };
        if (state.HasPendingNodeChange)
            return state with { Notice = "NoticePendingChange", NoticeContext = new() };
        if (!IsSelectable(nodeId))
            return state with { Notice = "NoticeNodeUnavailable", NoticeContext = new() };
        return state with { DraftNode = nodeId, Notice = "", NoticeContext = new() };
    }

    private static DemoState ApplyDraft(DemoState state)
    {
        if (IsTransitioning(state))
            return state with { Notice = "NoticeCannotApply", NoticeContext = new() };
        if (state.HasPendingNodeChange)
            return state;
        if (!IsSelectable(state.DraftNode))
            return state with { Notice = "NoticeNodeUnavailable", NoticeContext = new() };
        if (state.Connection == ConnectionPhase.Disconnected)
            return state with { PreferredNode = state.DraftNode, Page = DemoPage.Connection,
                PendingNode = null, HasPendingNodeChange = false, Notice = "", NoticeContext = new() };

        var country = CountryOf(state.DraftNode);
        var changesCountry = state.LockedCountry is not null && country is not null && country != state.LockedCountry;
        return state with { PendingNode = state.DraftNode, HasPendingNodeChange = true,
            Notice = changesCountry ? "NoticeConfirmCountry" : "NoticeConfirmNode",
            NoticeContext = changesCountry ? new(state.LockedCountry, country) : new() };
    }

    private static DemoState ConfirmSelection(DemoState state)
    {
        if (state.Entitlement != EntitlementStatus.Active)
            return state with { PendingNode = null, HasPendingNodeChange = false,
                Notice = EntitlementNotice(state.Entitlement), NoticeContext = new() };
        if (!IsSelectable(state.PendingNode))
            return state with { PendingNode = null, HasPendingNodeChange = false, Notice = "NoticeNodeUnavailable", NoticeContext = new() };
        var country = CountryOf(state.PendingNode);
        // Auto has no country identity and must not invent or clear a saved lock.
        return state with
        {
            PreferredNode = state.PendingNode, DraftNode = state.PendingNode,
            LockedCountry = state.LockedCountry is not null && country is not null ? country : state.LockedCountry,
            Connection = ConnectionPhase.Connecting, Page = DemoPage.Connection, CurrentCountry = null,
            PendingNode = null, HasPendingNodeChange = false, Notice = "", NoticeContext = new()
        };
    }
}
