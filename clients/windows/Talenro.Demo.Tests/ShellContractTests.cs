using System.Xml.Linq;
using Xunit;

namespace Talenro.Demo.Tests;

// Approved static declaration checks; not evidence of native click/focus behavior.
public class ShellContractTests
{
    private static readonly XNamespace X = "http://schemas.microsoft.com/winfx/2006/xaml";

    internal static XDocument Source(string relative)
    {
        var path = Path.Combine(AppContext.BaseDirectory, "AppSources", relative);
        Assert.True(File.Exists(path), $"Missing native UI declaration: {relative}");
        return XDocument.Load(path);
    }

    internal static XElement Named(XDocument document, string name) =>
        Assert.Single(document.Descendants(), e => (string?)e.Attribute(X + "Name") == name);

    internal static string Copy(string language, string key)
    {
        var entry = Assert.Single(Source($"Strings/{language}/Resources.resw").Descendants("data"),
            e => (string?)e.Attribute("name") == key);
        var value = entry.Element("value")?.Value;
        Assert.False(string.IsNullOrWhiteSpace(value));
        return value!;
    }

    [Fact]
    public void OfflineNoticeLivesOutsideReplaceablePageContent()
    {
        var shell = Source("MainWindow.xaml");
        var notice = Named(shell, "OfflineNotice");
        var host = Named(shell, "PageHost");
        Assert.Equal("OfflineNotice", (string?)notice.Attribute(X + "Uid"));
        Assert.DoesNotContain(host, notice.Ancestors());
        Assert.DoesNotContain(notice.AncestorsAndSelf(), e => e.Attribute("Visibility") is not null);
        Assert.Equal("离线演示，不会连接 VPN 或提交账号信息", Copy("zh-Hans", "OfflineNotice.Text"));
        Assert.Contains("offline", Copy("en-US", "OfflineNotice.Text"), StringComparison.OrdinalIgnoreCase);
    }

    [Fact]
    public void LoginNeverRequestsEditableCredentials()
    {
        var login = Source("Views/LoginPage.xaml");
        Assert.Equal("TextBlock", Named(login, "DemoEmail").Name.LocalName);
        Assert.Equal("demo@example.com", (string?)Named(login, "DemoEmail").Attribute("Text"));
        Assert.Equal("LeftToRight", (string?)Named(login, "DemoEmail").Attribute("FlowDirection"));
        Assert.Equal("TextBlock", Named(login, "DemoPassword").Name.LocalName);
        Assert.Equal("••••••••", (string?)Named(login, "DemoPassword").Attribute("Text"));
        Assert.DoesNotContain(login.Descendants(), e => e.Name.LocalName is "PasswordBox" or "TextBox" or "RichEditBox");
    }

    [Fact]
    public void EntryButtonExplicitlyEntersDemoInsteadOfClaimingAuthentication()
    {
        var entry = Named(Source("Views/LoginPage.xaml"), "EnterDemoButton");
        Assert.Equal("Button", entry.Name.LocalName);
        Assert.Equal("EnterDemo", (string?)entry.Attribute(X + "Uid"));
        Assert.Equal("EnterDemo_Click", (string?)entry.Attribute("Click"));
        Assert.Equal("进入演示", Copy("zh-Hans", "EnterDemo.Content"));
        Assert.Equal("Enter demo", Copy("en-US", "EnterDemo.Content"));
    }

    [Fact]
    public void PowerGlyphIsDecorativeNotASecondFocusableConnectionAction()
    {
        var connection = Source("Views/ConnectionPage.xaml");
        var glyph = Named(connection, "PowerGlyph");
        Assert.Equal("TextBlock", glyph.Name.LocalName);
        Assert.Equal("False", (string?)glyph.Attribute("IsTextSelectionEnabled"));
        Assert.Equal("Raw", (string?)glyph.Attribute("AutomationProperties.AccessibilityView"));
        Assert.DoesNotContain(glyph.Ancestors(), e => e.Name.LocalName is "Button" or "HyperlinkButton");
        var primary = Assert.Single(connection.Descendants(), e => (string?)e.Attribute("Click") == "PrimaryConnection_Click");
        Assert.Equal("Button", primary.Name.LocalName);
        Assert.Equal("PrimaryConnectionButton", (string?)primary.Attribute(X + "Name"));
    }

    [Fact]
    public void ScenarioControlsAreVisiblySeparateFromPrimaryConnectionAction()
    {
        var connection = Source("Views/ConnectionPage.xaml");
        var scenarios = Named(connection, "ScenarioPanel");
        Assert.Contains(scenarios.Descendants(), e => (string?)e.Attribute(X + "Uid") == "DemoScenarios");
        Assert.DoesNotContain(Named(connection, "PrimaryConnectionButton"), scenarios.Descendants());
        Assert.Equal("演示场景", Copy("zh-Hans", "DemoScenarios.Text"));
        Assert.Equal("Demo scenarios", Copy("en-US", "DemoScenarios.Text"));
    }
}
