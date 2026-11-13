using System.Diagnostics;
using System.Text.Json;
using Xunit;

namespace Talenro.Demo.Tests;

public class OrdinaryBuildIsolationTests
{
    private const string App = "Talenro.Demo.App";
    private const string Properties = "BaseOutputPath,BaseIntermediateOutputPath,MSBuildProjectExtensionsPath,MSBuildProjectDirectory,TargetPath,RuntimeIdentifier,DefineConstants";
    private static readonly string[] Ordinary = ["-p:Configuration=Debug", "-p:EnableLayoutDiagnostics=false",
        "-p:EnableEmailGeometryRecording=false", "-p:EnableCloseStageRecording=false",
        "-p:EmailGeometryBuildVariant=", "-p:CloseStageVerificationMode="];

    // A missing App-only override would overwrite the historical ordinary artifacts.
    [Fact]
    public void OrdinaryCenterPathsAndPayload()
    {
        using var json = Evaluate(App, [.. Ordinary, "-p:OrdinaryBuildVariant=account-center"]);
        var p = json.RootElement.GetProperty("Properties");
        string Value(string key) => p.GetProperty(key).GetString()!.Replace('\\', '/');
        Assert.Equal("bin/ordinary-account-center/", Value("BaseOutputPath"));
        Assert.Equal("obj/ordinary-account-center/", Value("BaseIntermediateOutputPath"));
        Assert.Equal(Value("MSBuildProjectDirectory") + "/obj/ordinary-account-center/", Value("MSBuildProjectExtensionsPath"));
        Assert.Equal(Value("MSBuildProjectDirectory") + "/bin/ordinary-account-center/Debug/net10.0-windows10.0.19041.0/win-x64/Talenro.Demo.App.dll", Value("TargetPath"));
        Assert.Equal("win-x64", Value("RuntimeIdentifier"));
        var defines = Value("DefineConstants").Split(';');
        foreach (var symbol in new[] { "LAYOUT_DIAGNOSTICS", "EMAIL_GEOMETRY_RECORDING", "CLOSE_STAGE_RECORDING" })
            Assert.DoesNotContain(symbol, defines);
        var items = json.RootElement.GetProperty("Items");
        var files = items.GetProperty("Compile").EnumerateArray()
            .Select(x => x.GetProperty("Identity").GetString()!.Replace('\\', '/')).ToArray();
        Assert.DoesNotContain(files, x => x.StartsWith("Diagnostics/", StringComparison.Ordinal));
        Assert.Contains("Layout/ReservedLayoutArguments.cs", files);
        Assert.DoesNotContain(items.GetProperty("EmbeddedResource").EnumerateArray(), x =>
            x.TryGetProperty("LogicalName", out var name) &&
            (name.GetString()!.StartsWith("LayoutFixtures", StringComparison.Ordinal) || name.GetString()!.StartsWith("LayoutStrings", StringComparison.Ordinal)));
    }

    // An empty selector must not alter either preexisting ordinary or geometry modes.
    [Theory]
    [InlineData(false)]
    [InlineData(true)]
    public void EmptyVariantPreservesDefaults(bool geometry)
    {
        string[] args = geometry
            ? ["-p:Configuration=Debug", "-p:EnableLayoutDiagnostics=true", "-p:EnableEmailGeometryRecording=true",
                "-p:EnableCloseStageRecording=false", "-p:EmailGeometryBuildVariant=account-center", "-p:CloseStageVerificationMode="]
            : Ordinary;
        using var omitted = Evaluate(App, args);
        using var empty = Evaluate(App, [.. args, "-p:OrdinaryBuildVariant="]);
        Assert.Equal(Normalize(omitted.RootElement), Normalize(empty.RootElement));
    }

    // An unguarded shared props group would redirect dependency outputs too.
    [Theory]
    [InlineData("Talenro.Demo.Core")]
    [InlineData("Talenro.Demo.Tests")]
    public void VariantLeavesDependenciesUnchanged(string project)
    {
        using var empty = Evaluate(project, [.. Ordinary, "-p:OrdinaryBuildVariant="]);
        using var selected = Evaluate(project, [.. Ordinary, "-p:OrdinaryBuildVariant=account-center"]);
        Assert.Equal(Normalize(empty.RootElement.GetProperty("Properties")), Normalize(selected.RootElement.GetProperty("Properties")));
        if (project == "Talenro.Demo.Core")
            Assert.Equal("", selected.RootElement.GetProperty("Properties").GetProperty("RuntimeIdentifier").GetString());
    }

    // Every incompatible mode must fail at this guard, not at a later target.
    [Theory]
    [InlineData("unknown", "Debug", "false", "false", "false", "", "")]
    [InlineData("../account-center", "Debug", "false", "false", "false", "", "")]
    [InlineData("account-center", "Release", "false", "false", "false", "", "")]
    [InlineData("account-center", "Debug", "true", "false", "false", "", "")]
    [InlineData("account-center", "Debug", "false", "true", "false", "", "")]
    [InlineData("account-center", "Debug", "false", "false", "true", "", "")]
    [InlineData("account-center", "Debug", "false", "false", "false", "account-center", "")]
    [InlineData("account-center", "Debug", "false", "false", "false", "", "ordinary")]
    [InlineData("account-center", "Debug", "wrong", "false", "false", "", "")]
    [InlineData("account-center", "Debug", "false", "wrong", "false", "", "")]
    [InlineData("account-center", "Debug", "false", "false", "wrong", "", "")]
    public void InvalidOrdinaryVariantFailsEarly(string variant, string configuration, string layout,
        string geometry, string close, string emailVariant, string verification)
    {
        var result = RunProject(App, "-target:ValidateOrdinaryBuildVariant", "-p:OrdinaryBuildVariant=" + variant,
            "-p:Configuration=" + configuration, "-p:EnableLayoutDiagnostics=" + layout,
            "-p:EnableEmailGeometryRecording=" + geometry, "-p:EnableCloseStageRecording=" + close,
            "-p:EmailGeometryBuildVariant=" + emailVariant, "-p:CloseStageVerificationMode=" + verification);
        Assert.NotEqual(0, result.Exit);
        Assert.Contains("ORDINARYBUILD001", result.Text);
    }

    private static JsonDocument Evaluate(string project, string[] args)
    {
        var result = RunProject(project, ["-getProperty:" + Properties, "-getItem:Compile,EmbeddedResource", .. args]);
        Assert.True(result.Exit == 0, result.Text);
        return JsonDocument.Parse(result.Text);
    }

    private static string Normalize(JsonElement value) => value.ValueKind switch
    {
        JsonValueKind.Object => "{" + string.Join(",", value.EnumerateObject().OrderBy(x => x.Name, StringComparer.Ordinal)
            .Select(x => JsonSerializer.Serialize(x.Name) + ":" + Normalize(x.Value))) + "}",
        JsonValueKind.Array => "[" + string.Join(",", value.EnumerateArray().Select(Normalize).OrderBy(x => x, StringComparer.Ordinal)) + "]",
        JsonValueKind.String => JsonSerializer.Serialize(value.GetString()!.Replace('\\', '/')),
        _ => value.GetRawText()
    };

    private static (int Exit, string Text) RunProject(string project, params string[] args)
    {
        var directory = new DirectoryInfo(AppContext.BaseDirectory);
        while (directory is not null && !File.Exists(Path.Combine(directory.FullName, App, App + ".csproj"))) directory = directory.Parent;
        Assert.NotNull(directory);
        var start = new ProcessStartInfo("dotnet") { UseShellExecute = false, CreateNoWindow = true,
            RedirectStandardOutput = true, RedirectStandardError = true, WorkingDirectory = directory.FullName };
        start.ArgumentList.Add("msbuild");
        start.ArgumentList.Add(project + "/" + project + ".csproj");
        start.ArgumentList.Add("-nologo");
        foreach (var arg in args) start.ArgumentList.Add(arg);
        using var process = Process.Start(start)!;
        var stdout = process.StandardOutput.ReadToEndAsync();
        var stderr = process.StandardError.ReadToEndAsync();
        Assert.True(process.WaitForExit(60000), "Evaluation did not finish; inspect remaining process before any retry. No build/restore requested.");
        return (process.ExitCode, stdout.GetAwaiter().GetResult() + stderr.GetAwaiter().GetResult());
    }
}
