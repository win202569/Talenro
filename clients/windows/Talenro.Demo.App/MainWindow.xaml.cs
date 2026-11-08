using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using Talenro.Demo.App.ViewModels;
using Talenro.Demo.App.Views;
using Talenro.Demo.Core;
namespace Talenro.Demo.App;
public sealed partial class MainWindow : Window
{
    private readonly DemoViewModel model = new();
    private readonly LoginPage login;
    private readonly ConnectionPage connection;
    private readonly NodesPage nodes;
    private readonly AccountPage account;
    private bool hadSession;
    public MainWindow()
    {
        InitializeComponent();
        login = new LoginPage(model);
        connection = new ConnectionPage(model);
        nodes = new NodesPage(model);
        account = new AccountPage(model);
        model.PropertyChanged += (_, _) => Refresh();
        Refresh();
    }
    private void Refresh()
    {
        if (hadSession && !model.State.HasSession) { login.Reset(); nodes.Reset(); }
        hadSession = model.State.HasSession;
        Sidebar.Visibility = model.State.HasSession ? Visibility.Visible : Visibility.Collapsed;
        Notice.Text = model.State.Notice;
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
    }
    internal static string Copy(string key) =>
        new Microsoft.Windows.ApplicationModel.Resources.ResourceLoader().GetString(key.Replace('.', '/'));
    private void Navigate_Click(object sender, RoutedEventArgs e)
    {
        if (sender is Button { Tag: string tag } && Enum.TryParse<DemoPage>(tag, out var page))
            model.Dispatch(new Navigate(page));
    }
    private async void Placeholder_Click(object sender, RoutedEventArgs e)
    {
        model.Dispatch(new ShowPlaceholder("sidebar"));
        var dialog = new ContentDialog { XamlRoot = Root.XamlRoot, Title = Copy("DemoTitle"),
            Content = model.State.Notice, CloseButtonText = Copy("Close") };
        await dialog.ShowAsync();
        (sender as Control)?.Focus(FocusState.Programmatic);
    }
}
