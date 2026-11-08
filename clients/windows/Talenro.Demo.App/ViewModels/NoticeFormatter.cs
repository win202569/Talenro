using Talenro.Demo.Core;

namespace Talenro.Demo.App.ViewModels;

public static class NoticeFormatter
{
    public static string Format(DemoState state, Func<string, string> resource) =>
        string.IsNullOrEmpty(state.Notice) ? "" : string.Format(
            System.Globalization.CultureInfo.CurrentUICulture, resource(state.Notice),
            state.NoticeContext.FromCountry, state.NoticeContext.ToCountry);
}
