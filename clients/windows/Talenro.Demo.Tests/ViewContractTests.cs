using System.Xml.Linq;
using Xunit;

namespace Talenro.Demo.Tests;

public class ViewContractTests
{
    private static readonly XNamespace X = "http://schemas.microsoft.com/winfx/2006/xaml";

    [Theory]
    [InlineData("Login")]
    [InlineData("Connection")]
    [InlineData("Nodes")]
    [InlineData("Account")]
    public void FourNativePagesHaveDeclaredDestinations(string page)
    {
        var document = ShellContractTests.Source($"Views/{page}Page.xaml");
        Assert.Equal("Page", document.Root!.Name.LocalName);
        Assert.Equal($"Talenro.Demo.App.Views.{page}Page", (string?)document.Root.Attribute(X + "Class"));
        if (page != "Login")
            Assert.Single(ShellContractTests.Source("MainWindow.xaml").Descendants(),
                e => (string?)e.Attribute("Tag") == page);
    }

    [Fact]
    public void NodesExposeProtocolSearchResultsAndExplicitApplyOrBack()
    {
        var nodes = ShellContractTests.Source("Views/NodesPage.xaml");
        Assert.Equal("TextBox", ShellContractTests.Named(nodes, "SearchBox").Name.LocalName);
        Assert.Equal("ProtocolHeading", (string?)ShellContractTests.Named(nodes, "ProtocolHeading").Attribute(X + "Uid"));
        Assert.Equal("协议", ShellContractTests.Copy("zh-Hans", "ProtocolHeading.Text"));
        Assert.Equal("Protocol", ShellContractTests.Copy("en-US", "ProtocolHeading.Text"));
        Assert.Equal("ListView", ShellContractTests.Named(nodes, "NodeList").Name.LocalName);
        Assert.Equal("ListStateNotice", (string?)ShellContractTests.Named(nodes, "ListStateNotice").Attribute(X + "Name"));
        Assert.Equal("Apply_Click", (string?)ShellContractTests.Named(nodes, "ApplyNodeButton").Attribute("Click"));
        Assert.Equal("Back_Click", (string?)ShellContractTests.Named(nodes, "BackButton").Attribute("Click"));
    }

    [Fact]
    public void AccountUsageStartsUnavailableRatherThanInventingZero()
    {
        var account = ShellContractTests.Source("Views/AccountPage.xaml");
        Assert.Equal("UsageUnavailable", (string?)ShellContractTests.Named(account, "UsageValue").Attribute(X + "Uid"));
        Assert.Equal("不可用", ShellContractTests.Copy("zh-Hans", "UsageUnavailable.Text"));
        Assert.Equal("Unavailable", ShellContractTests.Copy("en-US", "UsageUnavailable.Text"));
        Assert.Equal("Logout_Click", (string?)ShellContractTests.Named(account, "ExitDemoButton").Attribute("Click"));
        Assert.DoesNotContain(account.Descendants(), e => e.Name.LocalName == "Hyperlink");
    }

    [Theory]
    [InlineData("Mfa")]
    [InlineData("Devices")]
    [InlineData("Language")]
    [InlineData("Delete")]
    public void ExcludedAccountActionsUseExplanationNotSuccessOrPayment(string action)
    {
        var account = ShellContractTests.Source("Views/AccountPage.xaml");
        var button = ShellContractTests.Named(account, action + "Button");
        Assert.Equal("Placeholder_Click", (string?)button.Attribute("Click"));
        Assert.Contains("未实现", ShellContractTests.Copy("zh-Hans", action + ".Content"));
        Assert.Contains("not implemented", ShellContractTests.Copy("en-US", action + ".Content"));
    }

    [Theory]
    [InlineData("NodesLoading")]
    [InlineData("NodesUnavailable")]
    [InlineData("NodesEmpty")]
    [InlineData("DemoUsage")]
    [InlineData("DemoDevices")]
    [InlineData("ConfirmNodeChange")]
    [InlineData("Cancel")]
    public void AlternateStatesAndConfirmationHaveBothLanguageResources(string key)
    {
        Assert.False(string.IsNullOrWhiteSpace(ShellContractTests.Copy("zh-Hans", key)));
        Assert.False(string.IsNullOrWhiteSpace(ShellContractTests.Copy("en-US", key)));
    }
}
