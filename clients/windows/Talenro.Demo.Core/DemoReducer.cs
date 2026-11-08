namespace Talenro.Demo.Core;

public static class DemoReducer
{
    public static DemoState Reduce(DemoState state, DemoEvent action) => action switch
    {
        EnterDemo => state with { HasSession = true, Page = DemoPage.Connection },
        Logout => DemoFixtures.Initial(),
        Navigate navigation when state.HasSession => state with { Page = navigation.Page },
        ShowPlaceholder => state with { Notice = "此功能不在本次演示范围内" },
        _ => state
    };
}
