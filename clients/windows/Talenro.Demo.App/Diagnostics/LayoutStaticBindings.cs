using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
namespace Talenro.Demo.App.Layout;

internal static class LayoutStaticBindings
{
    public static void Apply(string viewId, FrameworkElement root, LayoutPresentation presentation)
    {
        Bind<TextBlock>("Shell", viewId, root, "OfflineNotice", "OfflineNotice.Text", (control, value) => control.Text = value, presentation);
        Bind<Button>("Shell", viewId, root, "NavConnection", "NavConnection.Content", (control, value) => control.Content = value, presentation);
        Bind<Button>("Shell", viewId, root, "NavNodes", "NavNodes.Content", (control, value) => control.Content = value, presentation);
        Bind<Button>("Shell", viewId, root, "NavTraffic", "NavTraffic.Content", (control, value) => control.Content = value, presentation);
        Bind<Button>("Shell", viewId, root, "NavDevices", "NavDevices.Content", (control, value) => control.Content = value, presentation);
        Bind<Button>("Shell", viewId, root, "NavAccount", "NavAccount.Content", (control, value) => control.Content = value, presentation);
        Bind<Button>("Shell", viewId, root, "NavSettings", "NavSettings.Content", (control, value) => control.Content = value, presentation);
        Bind<TextBlock>("Login", viewId, root, "LoginHeading", "LoginHeading.Text", (control, value) => control.Text = value, presentation);
        Bind<TextBlock>("Login", viewId, root, "EmailLabel", "EmailLabel.Text", (control, value) => control.Text = value, presentation);
        Bind<TextBlock>("Login", viewId, root, "PasswordLabel", "PasswordLabel.Text", (control, value) => control.Text = value, presentation);
        Bind<Button>("Login", viewId, root, "EnterDemoButton", "EnterDemo.Content", (control, value) => control.Content = value, presentation);
        Bind<Button>("Login", viewId, root, "ResetPassword", "ResetPassword.Content", (control, value) => control.Content = value, presentation);
        Bind<Button>("Login", viewId, root, "RegisterAccount", "RegisterAccount.Content", (control, value) => control.Content = value, presentation);
        Bind<TextBlock>("Login", viewId, root, "DemoScenarios", "DemoScenarios.Text", (control, value) => control.Text = value, presentation);
        Bind<ComboBox>("Login", viewId, root, "LoginScenario", "LoginScenario.Header", (control, value) => control.Header = value, presentation);
        Bind<ComboBoxItem>("Login", viewId, root, "ScenarioNormal", "ScenarioNormal.Content", (control, value) => control.Content = value, presentation);
        Bind<ComboBoxItem>("Login", viewId, root, "ScenarioInvalid", "ScenarioInvalid.Content", (control, value) => control.Content = value, presentation);
        Bind<ComboBoxItem>("Login", viewId, root, "ScenarioSubmitting", "ScenarioSubmitting.Content", (control, value) => control.Content = value, presentation);
        Bind<ComboBoxItem>("Login", viewId, root, "ScenarioRejected", "ScenarioRejected.Content", (control, value) => control.Content = value, presentation);
        Bind<ComboBoxItem>("Login", viewId, root, "ScenarioUnavailable", "ScenarioUnavailable.Content", (control, value) => control.Content = value, presentation);
        Bind<ComboBoxItem>("Login", viewId, root, "ScenarioMfa", "ScenarioMfa.Content", (control, value) => control.Content = value, presentation);
        Bind<TextBlock>("Connection", viewId, root, "ConnectionHeading", "ConnectionHeading.Text", (control, value) => control.Text = value, presentation);
        Bind<Button>("Connection", viewId, root, "ChooseLocation", "ChooseLocation.Content", (control, value) => control.Content = value, presentation);
        Bind<ToggleSwitch>("Connection", viewId, root, "CountryLock", "CountryLock.Header", (control, value) => control.Header = value, presentation);
        Bind<Button>("Connection", viewId, root, "RecoverOnce", "RecoverOnce.Content", (control, value) => control.Content = value, presentation);
        Bind<Button>("Connection", viewId, root, "RetryCountry", "RetryCountry.Content", (control, value) => control.Content = value, presentation);
        Bind<Button>("Connection", viewId, root, "StayDisconnected", "StayDisconnected.Content", (control, value) => control.Content = value, presentation);
        Bind<TextBlock>("Connection", viewId, root, "DemoScenarios", "DemoScenarios.Text", (control, value) => control.Text = value, presentation);
        Bind<Button>("Connection", viewId, root, "ScenarioConnected", "ScenarioConnected.Content", (control, value) => control.Content = value, presentation);
        Bind<Button>("Connection", viewId, root, "ScenarioFailed", "ScenarioFailed.Content", (control, value) => control.Content = value, presentation);
        Bind<Button>("Connection", viewId, root, "ScenarioLockedFailed", "ScenarioLockedFailed.Content", (control, value) => control.Content = value, presentation);
        Bind<Button>("Connection", viewId, root, "ScenarioDisconnected", "ScenarioDisconnected.Content", (control, value) => control.Content = value, presentation);
        Bind<ComboBox>("Connection", viewId, root, "EntitlementScenario", "EntitlementScenario.Header", (control, value) => control.Header = value, presentation);
        Bind<ComboBoxItem>("Connection", viewId, root, "StatusActive", "StatusActive.Content", (control, value) => control.Content = value, presentation);
        Bind<ComboBoxItem>("Connection", viewId, root, "StatusExpired", "StatusExpired.Content", (control, value) => control.Content = value, presentation);
        Bind<ComboBoxItem>("Connection", viewId, root, "StatusExhausted", "StatusExhausted.Content", (control, value) => control.Content = value, presentation);
        Bind<ComboBoxItem>("Connection", viewId, root, "StatusUnknown", "StatusUnknown.Content", (control, value) => control.Content = value, presentation);
        Bind<ComboBoxItem>("Connection", viewId, root, "StatusStale", "StatusStale.Content", (control, value) => control.Content = value, presentation);
        Bind<ComboBoxItem>("Connection", viewId, root, "StatusDeviceLimit", "StatusDeviceLimit.Content", (control, value) => control.Content = value, presentation);
        Bind<TextBlock>("Nodes", viewId, root, "NodesHeading", "NodesHeading.Text", (control, value) => control.Text = value, presentation);
        Bind<TextBox>("Nodes", viewId, root, "SearchBox", "NodeSearch.Header", (control, value) => control.Header = value, presentation);
        Bind<TextBox>("Nodes", viewId, root, "SearchBox", "NodeSearch.PlaceholderText", (control, value) => control.PlaceholderText = value, presentation);
        Bind<TextBlock>("Nodes", viewId, root, "ProtocolHeading", "ProtocolHeading.Text", (control, value) => control.Text = value, presentation);
        Bind<Button>("Nodes", viewId, root, "BackButton", "Back.Content", (control, value) => control.Content = value, presentation);
        Bind<Button>("Nodes", viewId, root, "ApplyNodeButton", "ApplyNode.Content", (control, value) => control.Content = value, presentation);
        Bind<TextBlock>("Nodes", viewId, root, "DemoScenarios", "DemoScenarios.Text", (control, value) => control.Text = value, presentation);
        Bind<ComboBox>("Nodes", viewId, root, "ListScenario", "ListScenario.Header", (control, value) => control.Header = value, presentation);
        Bind<ComboBoxItem>("Nodes", viewId, root, "ScenarioNormal", "ScenarioNormal.Content", (control, value) => control.Content = value, presentation);
        Bind<ComboBoxItem>("Nodes", viewId, root, "ScenarioLoading", "ScenarioLoading.Content", (control, value) => control.Content = value, presentation);
        Bind<ComboBoxItem>("Nodes", viewId, root, "ScenarioUnavailable", "ScenarioUnavailable.Content", (control, value) => control.Content = value, presentation);
        Bind<ComboBoxItem>("Nodes", viewId, root, "ScenarioEmpty", "ScenarioEmpty.Content", (control, value) => control.Content = value, presentation);
        Bind<TextBlock>("Account", viewId, root, "AccountHeading", "AccountHeading.Text", (control, value) => control.Text = value, presentation);
        Bind<TextBlock>("Account", viewId, root, "MockPlan", "MockPlan.Text", (control, value) => control.Text = value, presentation);
        Bind<TextBlock>("Account", viewId, root, "UsageHeading", "UsageHeading.Text", (control, value) => control.Text = value, presentation);
        Bind<TextBlock>("Account", viewId, root, "UsageValue", "UsageUnavailable.Text", (control, value) => control.Text = value, presentation);
        Bind<TextBlock>("Account", viewId, root, "DevicesHeading", "DevicesHeading.Text", (control, value) => control.Text = value, presentation);
        Bind<Button>("Account", viewId, root, "MfaButton", "Mfa.Content", (control, value) => control.Content = value, presentation);
        Bind<Button>("Account", viewId, root, "DevicesButton", "Devices.Content", (control, value) => control.Content = value, presentation);
        Bind<Button>("Account", viewId, root, "LanguageButton", "Language.Content", (control, value) => control.Content = value, presentation);
        Bind<Button>("Account", viewId, root, "DeleteButton", "Delete.Content", (control, value) => control.Content = value, presentation);
        Bind<Button>("Account", viewId, root, "ExitDemoButton", "ExitDemo.Content", (control, value) => control.Content = value, presentation);
    }

    private static void Bind<T>(string owner, string viewId, FrameworkElement root, string name,
        string key, Action<T, string> assign, LayoutPresentation presentation) where T : FrameworkElement
    {
        if (owner != viewId) return;
        if (root.FindName(name) is not T control)
            throw new InvalidDataException("Diagnostic binding target unavailable.");
        assign(control, presentation.Copy(key));
    }
}
