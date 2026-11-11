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

    private static (int Exit, string Text) Run(params string[] args)
    {
        var directory = new DirectoryInfo(AppContext.BaseDirectory);
        while (directory is not null && !File.Exists(Path.Combine(directory.FullName, "Talenro.Demo.App", "Talenro.Demo.App.csproj"))) directory = directory.Parent;
        Assert.NotNull(directory);
        var start = new ProcessStartInfo("dotnet") { UseShellExecute = false, CreateNoWindow = true, RedirectStandardOutput = true, RedirectStandardError = true, WorkingDirectory = directory.FullName };
        start.ArgumentList.Add("msbuild");
        start.ArgumentList.Add("Talenro.Demo.App/Talenro.Demo.App.csproj");
        start.ArgumentList.Add("-nologo");
        foreach (var arg in args) start.ArgumentList.Add(arg);
        using var process = Process.Start(start)!;
        var stdout = process.StandardOutput.ReadToEndAsync();
        var stderr = process.StandardError.ReadToEndAsync();
        Assert.True(process.WaitForExit(60000), "Evaluation did not finish; no build/restore was requested. Inspect remaining process before retrying.");
        return (process.ExitCode, stdout.GetAwaiter().GetResult() + stderr.GetAwaiter().GetResult());
    }
}
