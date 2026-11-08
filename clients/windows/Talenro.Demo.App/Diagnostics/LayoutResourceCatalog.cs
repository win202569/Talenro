using System.Reflection;
using System.Xml;
using System.Xml.Linq;
namespace Talenro.Demo.App.Layout;

internal static class LayoutResourceCatalog
{
    // Required inventory includes static, computed-state, and dialog/notice keys.
    internal static readonly string[] RequiredKeys = """
        OfflineNotice.Text EnterDemo.Content DemoScenarios.Text NavConnection.Content NavNodes.Content NavTraffic.Content
        NavDevices.Content NavAccount.Content NavSettings.Content PendingView DemoTitle Close LoginHeading.Text
        EmailLabel.Text PasswordLabel.Text ResetPassword.Content RegisterAccount.Content LoginScenario.Header
        ScenarioNormal.Content ScenarioInvalid.Content ScenarioSubmitting.Content ScenarioRejected.Content ScenarioUnavailable.Content
        ScenarioMfa.Content LoginNormal LoginInvalid LoginSubmitting LoginRejected LoginUnavailable LoginMfa
        ConnectionHeading.Text ChooseLocation.Content ScenarioConnected.Content ScenarioFailed.Content ScenarioDisconnected.Content
        EntitlementScenario.Header StatusActive.Content StatusExpired.Content StatusExhausted.Content StatusUnknown.Content
        StatusStale.Content StatusDeviceLimit.Content ConnectionDisconnected ConnectionConnecting ConnectionConnected
        ConnectionDisconnecting ActionDisconnected ActionConnecting ActionConnected ActionDisconnecting NodesHeading.Text
        NodeSearch.Header NodeSearch.PlaceholderText ProtocolHeading.Text Back.Content ApplyNode.Content ListScenario.Header
        ScenarioLoading.Content ScenarioEmpty.Content NodesLoading NodesUnavailable NodesEmpty CountryJP CountryDE AutoNode
        ProtocolAuto NodeDisabled NodeTransitionBlocked NodeFixtures ConfirmNodeChange Confirm Cancel AccountHeading.Text
        MockPlan.Text UsageHeading.Text UsageUnavailable.Text DevicesHeading.Text DemoUsage DemoUsageExhausted DemoDevices
        DemoDevicesLimit Mfa.Content Devices.Content Language.Content Delete.Content ExitDemo.Content CountryLock.Header
        CountryLock.OnContent CountryLock.OffContent
        LockOff RecoverOnce.Content RetryCountry.Content StayDisconnected.Content ScenarioLockedFailed.Content NoticePlaceholder
        NoticeConnectionFailed NoticeLockedFailed NoticeExpired NoticeExhausted NoticeUnknown NoticeStale NoticeDeviceLimit
        NoticePermissionUnknown NoticeFinishOperation NoticeChooseCountry NoticeNoRecoveryNode NoticeCannotEdit NoticePendingChange
        NoticeNodeUnavailable NoticeCannotApply NoticeConfirmCountry NoticeConfirmNode
        """.Split((char[]?)null, StringSplitOptions.RemoveEmptyEntries);

    public static IReadOnlyDictionary<string, string> Parse(string xml)
    {
        try
        {
            using var reader = XmlReader.Create(new StringReader(xml), new XmlReaderSettings
            { DtdProcessing = DtdProcessing.Prohibit, XmlResolver = null, MaxCharactersInDocument = 262144 });
            var doc = XDocument.Load(reader);
            if (doc.Root?.Name != "root") throw new InvalidDataException("Invalid resource root.");
            var strings = new Dictionary<string, string>(StringComparer.Ordinal);
            foreach (var item in doc.Root.Elements("data"))
            {
                var key = (string?)item.Attribute("name");
                var values = item.Elements("value").ToArray();
                if (string.IsNullOrWhiteSpace(key) || values.Length != 1 || string.IsNullOrWhiteSpace(values[0].Value) ||
                    !strings.TryAdd(key, values[0].Value)) throw new InvalidDataException("Invalid resource entry.");
            }
            return strings;
        }
        catch (XmlException error) { throw new InvalidDataException("Invalid layout resources.", error); }
    }
    public static void Validate(IReadOnlyDictionary<string, string> strings)
    {
        if (RequiredKeys.Any(key => !strings.TryGetValue(key, out var value) || string.IsNullOrWhiteSpace(value)))
            throw new InvalidDataException("Incomplete layout resources.");
    }

    public static LayoutPresentation Prepare(LayoutOptions options, Assembly assembly)
    {
        var fixtures = LayoutFixtureCatalog.Parse(ReadEmbedded(assembly, "LayoutFixtures.json", 65536));
        var chinese = Parse(ReadEmbedded(assembly, "LayoutStrings.zh-Hans.xml", 262144));
        var english = Parse(ReadEmbedded(assembly, "LayoutStrings.en-US.xml", 262144));
        Validate(chinese);
        Validate(english);
        return new(options, fixtures, options.Language == "zh-Hans" ? chinese : english);
    }

    private static string ReadEmbedded(Assembly assembly, string name, int maximum)
    {
        using var stream = assembly.GetManifestResourceStream(name) ?? throw new InvalidDataException("Missing diagnostic resource.");
        if (stream.Length > maximum) throw new InvalidDataException("Oversized diagnostic resource.");
        using var reader = new StreamReader(stream);
        return reader.ReadToEnd();
    }
}
