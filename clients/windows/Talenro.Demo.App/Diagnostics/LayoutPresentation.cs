using Talenro.Demo.Core;
namespace Talenro.Demo.App.Layout;

internal sealed class LayoutPresentation(LayoutOptions options, LayoutFixtureCatalog fixtures, IReadOnlyDictionary<string, string> strings)
{
    public LayoutOptions Options { get; } = options;
    public string Copy(string key) => strings.TryGetValue(key, out var value) ? value : throw new InvalidDataException("Missing layout resource.");
    public string Text(string target, DemoState state, string ordinaryText) => (Options.Profile, target) switch
    {
        ("long-status", "ConnectionStatus") when state.Connection == ConnectionPhase.Connecting => fixtures.Text("long-status"),
        ("long-country", "CountryJP") => fixtures.Text("long-country"),
        ("long-country", "NodeJp1") => fixtures.Text("long-country") + " · jp-1",
        ("rtl-mixed", "LoginHeading") => fixtures.Text("rtl-mixed"),
        ("unknown-usage", "UsageValue") when state.Entitlement is EntitlementStatus.Unknown or EntitlementStatus.Stale => fixtures.Text("unknown-usage"),
        _ => ordinaryText
    };
    public bool IsProfileActive(DemoState state) => Options.Profile switch
    {
        "baseline" or "rtl-mixed" => true,
        "long-status" => state.Page == DemoPage.Connection && state.Connection == ConnectionPhase.Connecting,
        "long-country" => state.Page == DemoPage.Nodes,
        "unknown-usage" => state.Page == DemoPage.Account && state.Entitlement is EntitlementStatus.Unknown or EntitlementStatus.Stale,
        _ => false
    };
    public bool IsRtl => Options.Profile == "rtl-mixed";
}
