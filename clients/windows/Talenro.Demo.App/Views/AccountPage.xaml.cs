using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using Talenro.Demo.App.ViewModels;
using Talenro.Demo.Core;
#if LAYOUT_DIAGNOSTICS
using Talenro.Demo.App.Layout;
#endif
namespace Talenro.Demo.App.Views;
public sealed partial class AccountPage : Page
{
    private readonly DemoViewModel model;
    public AccountPage(DemoViewModel model) { this.model = model; InitializeComponent(); Refresh(); }
    public void Refresh()
    {
#if LAYOUT_DIAGNOSTICS
        ApplyDiagnosticPresentation();
#endif
        var status = model.State.Entitlement;
        EntitlementValue.Text = MainWindow.Copy("Status" + status + ".Content");
        var unknown = status is EntitlementStatus.Unknown or EntitlementStatus.Stale;
        UsageValue.Text = MainWindow.Copy(unknown ? "UsageUnavailable.Text"
            : status == EntitlementStatus.Exhausted ? "DemoUsageExhausted" : "DemoUsage");
        DevicesValue.Text = MainWindow.Copy(unknown ? "UsageUnavailable.Text"
            : status == EntitlementStatus.DeviceLimit ? "DemoDevicesLimit" : "DemoDevices");
#if LAYOUT_DIAGNOSTICS
        if (App.DiagnosticPresentation is { } presentation)
            UsageValue.Text = presentation.Text("UsageValue", model.State, UsageValue.Text);
#endif
    }
#if LAYOUT_DIAGNOSTICS
    private void ApplyDiagnosticPresentation()
    {
        if (App.DiagnosticPresentation is { } presentation)
            LayoutStaticBindings.Apply("Account", this, presentation);
    }
#endif
    private void Logout_Click(object sender, RoutedEventArgs e) => model.Dispatch(new Logout());
    private async void Placeholder_Click(object sender, RoutedEventArgs e)
    {
        model.Dispatch(new ShowPlaceholder("account"));
        await new ContentDialog { XamlRoot = XamlRoot, Title = MainWindow.Copy("DemoTitle"),
            Content = NoticeFormatter.Format(model.State, MainWindow.Copy), CloseButtonText = MainWindow.Copy("Close") }.ShowAsync();
        (sender as Control)?.Focus(FocusState.Programmatic);
    }
}
