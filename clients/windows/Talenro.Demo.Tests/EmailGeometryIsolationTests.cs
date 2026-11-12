using System.Diagnostics;
using System.Text.Json;
using Xunit;

namespace Talenro.Demo.Tests;

public class EmailGeometryIsolationTests
{
    [Theory]
    [InlineData("false", "false", "Debug", false)]
    [InlineData("false", "true", "Debug", false)]
    [InlineData("true", "true", "Debug", true)]
    public void EvaluatedAppItemsAndPathsIsolateProbe(string geometry, string layout, string configuration, bool enabled)
    {
        var output = Run("-getProperty:BaseOutputPath,BaseIntermediateOutputPath,MSBuildProjectExtensionsPath,DefineConstants", "-getItem:Compile",
            "-p:EnableEmailGeometryRecording=" + geometry, "-p:EnableLayoutDiagnostics=" + layout, "-p:Configuration=" + configuration);
        Assert.Equal(0, output.Exit);
        using var json = JsonDocument.Parse(output.Text);
        var properties = json.RootElement.GetProperty("Properties");
        var files = json.RootElement.GetProperty("Items").GetProperty("Compile").EnumerateArray()
            .Select(x => x.GetProperty("Identity").GetString()!.Replace('\\', '/')).ToArray();
        Assert.Equal(enabled, files.Any(x => x.StartsWith("Diagnostics/EmailGeometry/", StringComparison.Ordinal)));
        Assert.Equal(enabled, properties.GetProperty("DefineConstants").GetString()!.Split(';').Contains("EMAIL_GEOMETRY_RECORDING"));
        Assert.Equal(enabled ? "bin/email-geometry/" : layout == "true" ? "bin/layout-diagnostics/" : "bin/", properties.GetProperty("BaseOutputPath").GetString()!.Replace('\\', '/'));
        if (enabled)
        {
            Assert.Equal("obj/email-geometry/", properties.GetProperty("BaseIntermediateOutputPath").GetString());
            Assert.Contains("obj/email-geometry", properties.GetProperty("MSBuildProjectExtensionsPath").GetString()!.Replace('\\', '/'));
        }
    }

    [Theory]
    [InlineData("true", "true", "Release", "false", "")]
    [InlineData("true", "false", "Debug", "false", "")]
    [InlineData("true", "true", "Debug", "true", "")]
    [InlineData("true", "true", "Debug", "false", "layout")]
    [InlineData("wrong", "true", "Debug", "false", "")]
    public void InvalidCombinationsFailBeforeBuild(string geometry, string layout, string configuration, string close, string verification)
    {
        var output = Run("-target:ValidateEmailGeometryConfiguration", "-p:EnableEmailGeometryRecording=" + geometry,
            "-p:EnableLayoutDiagnostics=" + layout, "-p:Configuration=" + configuration,
            "-p:EnableCloseStageRecording=" + close, "-p:CloseStageVerificationMode=" + verification);
        Assert.NotEqual(0, output.Exit);
        Assert.Contains("EMAILGEOMETRY001", output.Text);
    }

    [Fact]
    public void AccountCenterPathsStayInApp()
    {
        var output = Run("-getProperty:BaseOutputPath,BaseIntermediateOutputPath,MSBuildProjectExtensionsPath,MSBuildProjectDirectory,TargetPath,DefineConstants",
            "-getItem:Compile", "-p:EmailGeometryBuildVariant=account-center", "-p:Configuration=Debug",
            "-p:EnableEmailGeometryRecording=true", "-p:EnableLayoutDiagnostics=true",
            "-p:EnableCloseStageRecording=false", "-p:CloseStageVerificationMode=");
        Assert.Equal(0, output.Exit);
        using var json = JsonDocument.Parse(output.Text);
        var properties = json.RootElement.GetProperty("Properties");
        string Property(string name) => properties.GetProperty(name).GetString()!.Replace('\\', '/');
        Assert.Equal("bin/email-geometry-account-center/", Property("BaseOutputPath"));
        Assert.Equal("obj/email-geometry-account-center/", Property("BaseIntermediateOutputPath"));
        var projectDirectory = Property("MSBuildProjectDirectory");
        Assert.Equal(projectDirectory + "/obj/email-geometry-account-center/", Property("MSBuildProjectExtensionsPath"));
        Assert.Equal(projectDirectory + "/bin/email-geometry-account-center/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.dll", Property("TargetPath"));
        var defines = Property("DefineConstants").Split(';');
        Assert.Contains("EMAIL_GEOMETRY_RECORDING", defines);
        Assert.Contains("LAYOUT_DIAGNOSTICS", defines);
        Assert.DoesNotContain("CLOSE_STAGE_RECORDING", defines);
        var files = json.RootElement.GetProperty("Items").GetProperty("Compile").EnumerateArray()
            .Select(x => x.GetProperty("Identity").GetString()!.Replace('\\', '/')).ToArray();
        Assert.Contains(files, x => x.StartsWith("Diagnostics/EmailGeometry/", StringComparison.Ordinal));
        Assert.DoesNotContain(files, x => x.StartsWith("Diagnostics/CloseStages/", StringComparison.Ordinal));
    }

    [Theory]
    [InlineData("Talenro.Demo.Core")]
    [InlineData("Talenro.Demo.Tests")]
    public void VariantLeavesCoreAndTestsPathsUnchanged(string project)
    {
        string Evaluate(string variant)
        {
            var output = RunProject(project + "/" + project + ".csproj",
                "-getProperty:BaseOutputPath,BaseIntermediateOutputPath,MSBuildProjectExtensionsPath,TargetPath",
                "-p:Configuration=Debug", "-p:EmailGeometryBuildVariant=" + variant,
                "-p:EnableEmailGeometryRecording=true", "-p:EnableLayoutDiagnostics=true",
                "-p:EnableCloseStageRecording=false", "-p:CloseStageVerificationMode=");
            Assert.Equal(0, output.Exit);
            return output.Text;
        }
        Assert.Equal(Evaluate(""), Evaluate("account-center"));
    }

    [Theory]
    [InlineData("unknown", "Debug", "true", "true", "false", "")]
    [InlineData("../account-center", "Debug", "true", "true", "false", "")]
    [InlineData("account-center", "Release", "true", "true", "false", "")]
    [InlineData("account-center", "Debug", "false", "true", "false", "")]
    [InlineData("account-center", "Debug", "true", "false", "false", "")]
    [InlineData("account-center", "Debug", "true", "true", "true", "")]
    [InlineData("account-center", "Debug", "true", "true", "false", "layout")]
    public void InvalidAccountCenterVariantFailsEarly(string variant, string configuration, string layout, string geometry, string close, string verification)
    {
        var output = Run("-target:ValidateEmailGeometryBuildVariant", "-p:EmailGeometryBuildVariant=" + variant,
            "-p:Configuration=" + configuration, "-p:EnableLayoutDiagnostics=" + layout,
            "-p:EnableEmailGeometryRecording=" + geometry, "-p:EnableCloseStageRecording=" + close,
            "-p:CloseStageVerificationMode=" + verification);
        Assert.NotEqual(0, output.Exit);
        Assert.Contains("EMAILGEOMETRY002", output.Text);
    }

    private static (int Exit, string Text) Run(params string[] args) =>
        RunProject("Talenro.Demo.App/Talenro.Demo.App.csproj", args);

    private static (int Exit, string Text) RunProject(string project, params string[] args)
    {
        var directory = new DirectoryInfo(AppContext.BaseDirectory);
        while (directory is not null && !File.Exists(Path.Combine(directory.FullName, "Talenro.Demo.App", "Talenro.Demo.App.csproj"))) directory = directory.Parent;
        Assert.NotNull(directory);
        var start = new ProcessStartInfo("dotnet") { UseShellExecute = false, CreateNoWindow = true, RedirectStandardOutput = true, RedirectStandardError = true, WorkingDirectory = directory.FullName };
        start.ArgumentList.Add("msbuild");
        start.ArgumentList.Add(project);
        start.ArgumentList.Add("-nologo");
        foreach (var arg in args) start.ArgumentList.Add(arg);
        using var process = Process.Start(start)!;
        var stdout = process.StandardOutput.ReadToEndAsync();
        var stderr = process.StandardError.ReadToEndAsync();
        Assert.True(process.WaitForExit(60000), "Evaluation did not finish; no build/restore was requested. Inspect remaining process before retrying.");
        return (process.ExitCode, stdout.GetAwaiter().GetResult() + stderr.GetAwaiter().GetResult());
    }
}
