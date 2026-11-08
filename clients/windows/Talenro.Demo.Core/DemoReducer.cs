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
        ShowPlaceholder => state with { Notice = "此功能不在本次演示范围内" },
        RequestConnection when state.HasSession && state.Connection == ConnectionPhase.Disconnected
            => state with { Connection = ConnectionPhase.Connecting, Notice = "" },
        CompleteConnection when state.HasSession && state.Connection == ConnectionPhase.Connecting
            => state with { Connection = ConnectionPhase.Connected },
        CancelConnection when state.HasSession && state.Connection == ConnectionPhase.Connecting
            => state with { Connection = ConnectionPhase.Disconnected },
        RequestDisconnection when state.HasSession && state.Connection == ConnectionPhase.Connected
            => state with { Connection = ConnectionPhase.Disconnecting },
        CompleteDisconnection when state.HasSession && state.Connection == ConnectionPhase.Disconnecting
            => state with { Connection = ConnectionPhase.Disconnected },
        FailConnection when state.HasSession && state.Connection == ConnectionPhase.Connecting
            => state with { Connection = ConnectionPhase.Disconnected, Notice = "模拟连接失败，请重试或选择其他位置" },
        SelectNode selection when state.HasSession && state.Page == DemoPage.Nodes
            => SelectDraft(state, selection.NodeId),
        ApplyNodeSelection when state.HasSession && state.Page == DemoPage.Nodes
            => ApplyDraft(state),
        CancelNodeSelection when state.HasSession && state.Page == DemoPage.Nodes
            => state with { Page = DemoPage.Connection, DraftNode = state.PreferredNode,
                PendingNode = null, HasPendingNodeChange = false, Notice = "" },
        CancelNodeChange when state.HasSession && state.HasPendingNodeChange
            => state with { PendingNode = null, HasPendingNodeChange = false, Notice = "" },
        ConfirmNodeChange when state.HasSession && state.HasPendingNodeChange
            && state.Connection == ConnectionPhase.Connected => ConfirmSelection(state),
        _ => state
    };

    private static bool IsTransitioning(DemoState state) =>
        state.Connection is ConnectionPhase.Connecting or ConnectionPhase.Disconnecting;

    private static bool IsSelectable(string? nodeId) =>
        nodeId is null || DemoFixtures.Nodes.Any(node => node.Id == nodeId && node.Enabled);

    private static string? CountryOf(string? nodeId) =>
        DemoFixtures.Nodes.FirstOrDefault(node => node.Id == nodeId)?.Country;

    private static DemoState SelectDraft(DemoState state, string? nodeId)
    {
        if (IsTransitioning(state))
            return state with { Notice = "模拟连接正在转换，暂不能编辑节点" };
        if (state.HasPendingNodeChange)
            return state with { Notice = "请先确认或取消当前节点切换" };
        if (!IsSelectable(nodeId))
            return state with { Notice = "该演示节点不可用" };
        return state with { DraftNode = nodeId, Notice = "" };
    }

    private static DemoState ApplyDraft(DemoState state)
    {
        if (IsTransitioning(state))
            return state with { Notice = "模拟连接正在转换，暂不能应用节点" };
        if (state.HasPendingNodeChange)
            return state;
        if (!IsSelectable(state.DraftNode))
            return state with { Notice = "该演示节点不可用" };
        if (state.Connection == ConnectionPhase.Disconnected)
            return state with { PreferredNode = state.DraftNode, Page = DemoPage.Connection,
                PendingNode = null, HasPendingNodeChange = false, Notice = "" };

        var country = CountryOf(state.DraftNode);
        var notice = state.LockedCountry is not null && country is not null && country != state.LockedCountry
            ? $"确认从 {state.LockedCountry} 切换至 {country} 并将国家锁定更新为 {country}？模拟连接将重新开始。"
            : "确认切换节点？模拟连接将重新开始。";
        return state with { PendingNode = state.DraftNode, HasPendingNodeChange = true, Notice = notice };
    }

    private static DemoState ConfirmSelection(DemoState state)
    {
        if (!IsSelectable(state.PendingNode))
            return state with { PendingNode = null, HasPendingNodeChange = false, Notice = "该演示节点不可用" };
        var country = CountryOf(state.PendingNode);
        // Auto has no country identity and must not invent or clear a saved lock.
        return state with
        {
            PreferredNode = state.PendingNode, DraftNode = state.PendingNode,
            LockedCountry = state.LockedCountry is not null && country is not null ? country : state.LockedCountry,
            Connection = ConnectionPhase.Connecting, Page = DemoPage.Connection,
            PendingNode = null, HasPendingNodeChange = false, Notice = ""
        };
    }
}
