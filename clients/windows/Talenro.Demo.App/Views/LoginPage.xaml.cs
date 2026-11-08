using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using Talenro.Demo.App.ViewModels;
using Talenro.Demo.Core;
namespace Talenro.Demo.App.Views;
public sealed partial class LoginPage : Page
{
    private readonly DemoViewModel model;
    public LoginPage(DemoViewModel model) { this.model = model; InitializeComponent(); }
    public void Reset() { LoginScenario.SelectedIndex = 0; ScenarioNotice.Text = MainWindow.Copy("LoginNormal"); EnterDemoButton.IsEnabled = true; }
    private void EnterDemo_Click(object sender, RoutedEventArgs e) => model.Dispatch(new EnterDemo());
    private void Scenario_Changed(object sender, SelectionChangedEventArgs e)
    {
        if (ScenarioNotice is null || EnterDemoButton is null) return;
        var tag = (LoginScenario.SelectedItem as ComboBoxItem)?.Tag as string ?? "Normal";
        ScenarioNotice.Text = MainWindow.Copy("Login" + tag);
        EnterDemoButton.IsEnabled = tag == "Normal";
    }
    private async void Placeholder_Click(object sender, RoutedEventArgs e)
    {
        model.Dispatch(new ShowPlaceholder("login"));
        await new ContentDialog { XamlRoot = XamlRoot, Title = MainWindow.Copy("DemoTitle"),
            Content = model.State.Notice, CloseButtonText = MainWindow.Copy("Close") }.ShowAsync();
        (sender as Control)?.Focus(FocusState.Programmatic);
    }
}
