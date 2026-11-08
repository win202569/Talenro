using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using Talenro.Demo.App.ViewModels;
using Talenro.Demo.Core;
#if LAYOUT_DIAGNOSTICS
using Talenro.Demo.App.Layout;
#endif
namespace Talenro.Demo.App.Views;
public sealed partial class LoginPage : Page
{
    private readonly DemoViewModel model;
    public LoginPage(DemoViewModel model)
    {
        this.model = model;
        InitializeComponent();
#if LAYOUT_DIAGNOSTICS
        ApplyDiagnosticPresentation();
#endif
    }
    public void Reset()
    {
        LoginScenario.SelectedIndex = 0;
        ScenarioNotice.Text = MainWindow.Copy("LoginNormal");
        EnterDemoButton.IsEnabled = true;
#if LAYOUT_DIAGNOSTICS
        ApplyDiagnosticPresentation();
#endif
    }
#if LAYOUT_DIAGNOSTICS
    private void ApplyDiagnosticPresentation()
    {
        if (App.DiagnosticPresentation is not { } presentation) return;
        LayoutStaticBindings.Apply("Login", this, presentation);
        LoginHeading.Text = presentation.Text("LoginHeading", model.State, LoginHeading.Text);
    }
#endif
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
            Content = NoticeFormatter.Format(model.State, MainWindow.Copy), CloseButtonText = MainWindow.Copy("Close") }.ShowAsync();
        (sender as Control)?.Focus(FocusState.Programmatic);
    }
}
