using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using Talenro.Demo.App.ViewModels;
using Talenro.Demo.App.Views;
using Talenro.Demo.Core;
#if LAYOUT_DIAGNOSTICS
using Talenro.Demo.App.Layout;
#endif
namespace Talenro.Demo.App;
public sealed partial class MainWindow : Window
{
    private readonly DemoViewModel model = new();
    private readonly LoginPage login;
    private readonly ConnectionPage connection;
    private readonly NodesPage nodes;
    private readonly AccountPage account;
    private bool hadSession;
#if LAYOUT_DIAGNOSTICS
    private readonly LayoutWindowObserver? layoutObserver;
#endif
    public MainWindow()
    {
        InitializeComponent();
#if LAYOUT_DIAGNOSTICS
        if (App.DiagnosticPresentation is { } presentation)
        {
            LayoutStaticBindings.Apply("Shell", Root, presentation);
            Root.FlowDirection = presentation.IsRtl ? FlowDirection.RightToLeft : FlowDirection.LeftToRight;
        }
#endif
        login = new LoginPage(model);
        connection = new ConnectionPage(model);
        nodes = new NodesPage(model);
        account = new AccountPage(model);
#if LAYOUT_DIAGNOSTICS
        if (App.DiagnosticPresentation is { } diagnostics)
            layoutObserver = new(this, Root, PageHost, () => diagnostics.IsProfileActive(model.State), diagnostics.Options);
#endif
        model.PropertyChanged += (_, _) => Refresh();
        Refresh();
    }
    private void Refresh()
    {
        if (hadSession && !model.State.HasSession) { login.Reset(); nodes.Reset(); }
        hadSession = model.State.HasSession;
        Sidebar.Visibility = model.State.HasSession ? Visibility.Visible : Visibility.Collapsed;
        Notice.Text = NoticeFormatter.Format(model.State, Copy);
        PageHost.Content = model.State.Page switch
        {
            DemoPage.Login => login,
            DemoPage.Connection => connection,
            DemoPage.Nodes => nodes,
            DemoPage.Account => account,
            _ => login
        };
        connection.Refresh();
        nodes.Refresh();
        account.Refresh();
#if LAYOUT_DIAGNOSTICS
        layoutObserver?.Refresh();
#endif
    }
    internal static string Copy(string key)
    {
#if LAYOUT_DIAGNOSTICS
        if (App.DiagnosticPresentation is { } presentation) return presentation.Copy(key);
#endif
        return new Microsoft.Windows.ApplicationModel.Resources.ResourceLoader().GetString(key.Replace('.', '/'));
    }
    private void Navigate_Click(object sender, RoutedEventArgs e)
    {
        if (sender is Button { Tag: string tag } && Enum.TryParse<DemoPage>(tag, out var page))
            model.Dispatch(new Navigate(page));
    }
    private async void Placeholder_Click(object sender, RoutedEventArgs e)
    {
        model.Dispatch(new ShowPlaceholder("sidebar"));
        var dialog = new ContentDialog { XamlRoot = Root.XamlRoot, Title = Copy("DemoTitle"),
            Content = NoticeFormatter.Format(model.State, Copy), CloseButtonText = Copy("Close") };
        await dialog.ShowAsync();
        (sender as Control)?.Focus(FocusState.Programmatic);
    }
}
