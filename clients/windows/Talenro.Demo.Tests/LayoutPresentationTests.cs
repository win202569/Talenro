using System.Text.Json.Nodes;
using Talenro.Demo.App.Layout;
using Talenro.Demo.Core;
using Xunit;

namespace Talenro.Demo.Tests;

public class LayoutPresentationTests
{
    private static string Fixture => File.ReadAllText(Path.Combine(AppContext.BaseDirectory, "AppSources/ui-copy-cases.json"));
    private static string Resource(string language) => File.ReadAllText(Path.Combine(AppContext.BaseDirectory, "AppSources/Strings", language, "Resources.resw"));
    private static LayoutPresentation Create(string profile) => new(new("en-US", profile), LayoutFixtureCatalog.Parse(Fixture),
        LayoutResourceCatalog.Parse(Resource("en-US")));

    [Fact]
    public void CommittedFixtureRetainsExactText()
    {
        Assert.Equal("不可用 / Unavailable — no numeric usage is known", LayoutFixtureCatalog.Parse(Fixture).Text("unknown-usage"));
        Assert.Equal("\u202eRTL synthetic label\u202c \u2066demo@example.com\u2069", LayoutFixtureCatalog.Parse(Fixture).Text("rtl-mixed"));
    }

    [Theory]
    [InlineData("duplicate")]
    [InlineData("missing")]
    [InlineData("unknown")]
    [InlineData("empty")]
    [InlineData("long")]
    [InlineData("direction")]
    [InlineData("extra-direction")]
    public void InvalidFixturesCannotReachPresentation(string mutation)
    {
        var json = JsonNode.Parse(Fixture)!;
        var cases = json["cases"]!.AsArray();
        switch (mutation)
        {
            case "duplicate": cases.Add(cases[0]!.DeepClone()); break;
            case "missing": cases.RemoveAt(0); break;
            case "unknown": cases[0]!["id"] = "unknown"; break;
            case "empty": cases[0]!["text"] = ""; break;
            case "long": cases[0]!["text"] = new string('x', 4097); break;
            case "direction": cases[2]!["direction"] = "LeftToRight"; break;
            case "extra-direction": cases[0]!["direction"] = "RightToLeft"; break;
        }
        Assert.Throws<InvalidDataException>(() => LayoutFixtureCatalog.Parse(json.ToJsonString()));
    }

    [Fact]
    public void MalformedAndOversizedFixtureRejected()
    {
        Assert.Throws<InvalidDataException>(() => LayoutFixtureCatalog.Parse("{"));
        Assert.Throws<InvalidDataException>(() => LayoutFixtureCatalog.Parse(new string('x', 65537)));
    }

    [Theory]
    [InlineData("zh-Hans", "进入演示")]
    [InlineData("en-US", "Enter demo")]
    public void ExplicitLanguageHasExactResourceValues(string language, string expected)
    {
        var resources = LayoutResourceCatalog.Parse(Resource(language));
        LayoutResourceCatalog.Validate(resources);
        var presentation = new LayoutPresentation(new(language, "baseline"), LayoutFixtureCatalog.Parse(Fixture), resources);
        Assert.Equal(expected, presentation.Copy("EnterDemo.Content"));
        var missing = resources.ToDictionary(x => x.Key, x => x.Value);
        missing.Remove("EnterDemo.Content");
        Assert.Throws<InvalidDataException>(() => LayoutResourceCatalog.Validate(missing));
        Assert.Throws<InvalidDataException>(() => presentation.Copy("missing-resource"));
    }

    [Theory]
    [InlineData("<!DOCTYPE root [<!ENTITY x 'bad'>]><root><data name='x'><value>&x;</value></data></root>")]
    [InlineData("<root><data name='x'><value>a</value></data><data name='x'><value>b</value></data></root>")]
    public void UnsafeOrAmbiguousResourcesRejected(string xml) =>
        Assert.Throws<InvalidDataException>(() => LayoutResourceCatalog.Parse(xml));

    [Fact]
    public void ProfileMappingDoesNotMutateStateOrUnrelatedText()
    {
        const string ordinary = "original";
        foreach (var profile in new[] { "baseline", "long-status", "long-country", "rtl-mixed", "unknown-usage" })
        foreach (var phase in Enum.GetValues<ConnectionPhase>())
        foreach (var entitlement in Enum.GetValues<EntitlementStatus>())
        {
            var view = Create(profile);
            var state = new DemoState { Connection = phase, Entitlement = entitlement, PreferredNode = "jp-1", LockedCountry = "JP" };
            var before = state with { };
            Assert.Equal(profile == "long-status" && phase == ConnectionPhase.Connecting
                ? "Simulated connection is waiting for an explicit demonstration result; no VPN connection is established and no account information is submitted."
                : ordinary, view.Text("ConnectionStatus", state, ordinary));
            Assert.Equal(profile == "unknown-usage" && entitlement is EntitlementStatus.Unknown or EntitlementStatus.Stale
                ? "不可用 / Unavailable — no numeric usage is known" : ordinary, view.Text("UsageValue", state, ordinary));
            Assert.Equal(profile == "long-country" ? "Fictional demonstration location with a deliberately extended country and node label" : ordinary,
                view.Text("CountryJP", state, ordinary));
            Assert.Equal(profile == "long-country" ? "Fictional demonstration location with a deliberately extended country and node label · jp-1" : ordinary,
                view.Text("NodeJp1", state, ordinary));
            Assert.Equal(profile == "rtl-mixed" ? "\u202eRTL synthetic label\u202c \u2066demo@example.com\u2069" : ordinary,
                view.Text("LoginHeading", state, ordinary));
            Assert.Equal(ordinary, view.Text("OfflineNotice", state, ordinary));
            Assert.Equal(ordinary, view.Text("PrimaryAction", state, ordinary));
            Assert.Equal(before, state);
            Assert.Equal(profile == "rtl-mixed", view.IsRtl);
        }
    }

    [Fact]
    public void ActiveConditionDoesNotClaimOffPageCoverage()
    {
        Assert.False(Create("long-status").IsProfileActive(new() { Page = DemoPage.Account, Connection = ConnectionPhase.Connecting }));
        Assert.True(Create("long-status").IsProfileActive(new() { Page = DemoPage.Connection, Connection = ConnectionPhase.Connecting }));
        Assert.False(Create("long-country").IsProfileActive(new() { Page = DemoPage.Connection }));
        Assert.True(Create("long-country").IsProfileActive(new() { Page = DemoPage.Nodes }));
        Assert.False(Create("unknown-usage").IsProfileActive(new() { Page = DemoPage.Account, Entitlement = EntitlementStatus.Active }));
        Assert.True(Create("unknown-usage").IsProfileActive(new() { Page = DemoPage.Account, Entitlement = EntitlementStatus.Stale }));
        Assert.True(Create("baseline").IsProfileActive(new()));
        Assert.True(Create("rtl-mixed").IsProfileActive(new()));
    }
}
