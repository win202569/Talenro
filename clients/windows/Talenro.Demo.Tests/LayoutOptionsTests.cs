using System.Xml.Linq;
using Talenro.Demo.App.Layout;
using Xunit;

namespace Talenro.Demo.Tests;

public class LayoutOptionsTests
{
    [Fact]
    public void EmptyArgumentsKeepDiagnosticBuildOrdinary()
    {
        var result = LayoutOptions.Parse([]);
        Assert.False(result.Enabled);
        Assert.Null(result.Options);
        Assert.Null(result.ErrorCode);
    }

    [Fact]
    public void ExplicitActivationSelectsSafeDefaults()
    {
        var result = LayoutOptions.Parse(["--layout-test"]);
        Assert.True(result.Enabled);
        Assert.Equal(new LayoutOptions("zh-Hans", "baseline"), result.Options);
        Assert.Null(result.ErrorCode);
    }

    [Theory]
    [InlineData("zh-Hans")]
    [InlineData("en-US")]
    public void AllProfilesRequireExplicitActivation(string language)
    {
        foreach (var profile in new[] { "baseline", "long-status", "long-country", "rtl-mixed", "unknown-usage" })
        {
            var result = LayoutOptions.Parse(["--layout-profile=" + profile, "--layout-test", "--layout-language=" + language]);
            Assert.True(result.Enabled);
            Assert.Equal(new LayoutOptions(language, profile), result.Options);
            Assert.Null(result.ErrorCode);
        }
    }

    [Theory]
    [InlineData("--layout-language=en-US")]
    [InlineData("--layout-profile=baseline")]
    [InlineData("--layout-test", "--layout-test")]
    [InlineData("--layout-test=true")]
    [InlineData("--layout-test", "--layout-unknown=secret")]
    [InlineData("--Layout-test")]
    [InlineData("--layout-test", "--layout-language=en-us")]
    [InlineData("--layout-test", "--layout-profile=Baseline")]
    [InlineData("--layout-test", "--layout-language=")]
    [InlineData("--layout-test", "--layout-language=en-US", "--layout-language=zh-Hans")]
    [InlineData("--layout-test", "--layout-profile=baseline", "--layout-profile=baseline")]
    [InlineData("--layout-test", "unrelated")]
    [InlineData("unrelated")]
    public void InvalidArgumentsFailWithoutEchoOrPartialOptions(params string[] args)
    {
        var result = LayoutOptions.Parse(args);
        Assert.False(result.Enabled);
        Assert.Null(result.Options);
        Assert.Equal("InvalidLayoutArguments", result.ErrorCode);
    }

    [Theory]
    [InlineData("--layout-test", true)]
    [InlineData("--LAYOUT-profile=x", true)]
    [InlineData("--layoutunknown", true)]
    [InlineData("unrelated", false)]
    public void OrdinaryBuildReservesPrefixButPreservesOtherArguments(string value, bool expected) =>
        Assert.Equal(expected, ReservedLayoutArguments.ContainsReserved([value]));

    [Fact]
    public void BuildDeclarationsSeparateDiagnosticCodeAndArtifacts()
    {
        var project = XDocument.Load(Path.Combine(AppContext.BaseDirectory, "AppSources", "Talenro.Demo.App.csproj"));
        var props = XDocument.Load(Path.Combine(AppContext.BaseDirectory, "AppSources", "Directory.Build.props"));
        Assert.Contains(project.Descendants("Compile"), x => (string?)x.Attribute("Remove") == "Diagnostics/**/*.cs");
        Assert.Contains(project.Descendants("Error"), x => (string?)x.Attribute("Code") == "LAYOUT001");
        Assert.Contains(props.Descendants("BaseIntermediateOutputPath"), x => x.Value == "obj/layout-diagnostics/");
        Assert.Contains(props.Descendants("BaseOutputPath"), x => x.Value == "bin/layout-diagnostics/");
        Assert.Contains(props.Descendants("EnableLayoutDiagnostics"), x => x.Value == "false");
    }
}
