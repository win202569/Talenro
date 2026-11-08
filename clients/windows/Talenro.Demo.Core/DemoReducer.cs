namespace Talenro.Demo.Core;

public static class DemoReducer
{
    public static DemoState Reduce(DemoState state, DemoEvent action) => action switch
    {
        EnterDemo => state with { HasSession = true, Page = DemoPage.Connection },
        Logout => DemoFixtures.Initial(),
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
        _ => state
    };
}
