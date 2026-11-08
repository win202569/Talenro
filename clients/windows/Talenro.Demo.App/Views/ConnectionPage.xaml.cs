using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using Talenro.Demo.App.ViewModels;
using Talenro.Demo.Core;
#if LAYOUT_DIAGNOSTICS
using Talenro.Demo.App.Layout;
#endif
namespace Talenro.Demo.App.Views;
public sealed partial class ConnectionPage : Page
{
    private readonly DemoViewModel model;
    private bool refreshing;
    public ConnectionPage(DemoViewModel model) { this.model = model; InitializeComponent(); Refresh(); }
    public void Refresh()
    {
        refreshing = true;
        try
        {
        var phase = model.State.Connection;
        ConnectionStatus.Text = MainWindow.Copy("Connection" + phase);
        PrimaryConnectionButton.Content = MainWindow.Copy("Action" + phase);
        PrimaryConnectionButton.IsEnabled = phase != ConnectionPhase.Disconnecting;
        CountryLock.IsOn = model.State.LockedCountry is not null;
        CountryLock.IsEnabled = phase is not (ConnectionPhase.Connecting or ConnectionPhase.Disconnecting);
        LockedCountryValue.Text = model.State.LockedCountry ?? MainWindow.Copy("LockOff");
        RecoveryPanel.Visibility = model.State.RecoveryChoices == LockRecoveryChoices.None ? Visibility.Collapsed : Visibility.Visible;
        foreach (ComboBoxItem item in EntitlementScenario.Items)
            if ((string)item.Tag == model.State.Entitlement.ToString()) { EntitlementScenario.SelectedItem = item; break; }
#if LAYOUT_DIAGNOSTICS
        ApplyDiagnosticPresentation();
#endif
        }
        finally { refreshing = false; }
    }
#if LAYOUT_DIAGNOSTICS
    private void ApplyDiagnosticPresentation()
    {
        if (App.DiagnosticPresentation is not { } presentation) return;
        LayoutStaticBindings.Apply("Connection", this, presentation);
        ConnectionStatus.Text = presentation.Text("ConnectionStatus", model.State, ConnectionStatus.Text);
        if (presentation.IsRtl) LockedCountryValue.FlowDirection = FlowDirection.LeftToRight;
    }
#endif
    private void Lock_Toggled(object sender, RoutedEventArgs e) { if (!refreshing) model.Dispatch(new SetCountryLock(CountryLock.IsOn)); }
    private void Recover_Click(object sender, RoutedEventArgs e) => model.Dispatch(new RecoverAcrossCountriesOnce());
    private void Retry_Click(object sender, RoutedEventArgs e) => model.Dispatch(new RetryLockedCountry());
    private void Stay_Click(object sender, RoutedEventArgs e) => model.Dispatch(new StayDisconnected());
    private void LockedFailed_Click(object sender, RoutedEventArgs e) => model.Dispatch(new FailLockedCountry());
    private void PrimaryConnection_Click(object sender, RoutedEventArgs e)
    {
        DemoEvent? action = model.State.Connection switch
        {
            ConnectionPhase.Disconnected => new RequestConnection(),
            ConnectionPhase.Connecting => new CancelConnection(),
            ConnectionPhase.Connected => new RequestDisconnection(),
            _ => null
        };
        if (action is not null) model.Dispatch(action);
    }
    private void ChooseLocation_Click(object sender, RoutedEventArgs e) => model.Dispatch(new Navigate(DemoPage.Nodes));
    private void Complete_Click(object sender, RoutedEventArgs e) => model.Dispatch(new CompleteConnection());
    private void Fail_Click(object sender, RoutedEventArgs e) => model.Dispatch(new FailConnection());
    private void Disconnected_Click(object sender, RoutedEventArgs e) => model.Dispatch(new CompleteDisconnection());
    private void Entitlement_Changed(object sender, SelectionChangedEventArgs e)
    {
        if (refreshing) return;
        if (EntitlementScenario?.SelectedItem is ComboBoxItem { Tag: string tag } && Enum.TryParse<EntitlementStatus>(tag, out var status))
            model.Dispatch(new SetEntitlement(status));
    }
}
