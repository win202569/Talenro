namespace Talenro.Demo.App.Layout.EmailGeometry;

public sealed record GeometryActivationResult(Guid? RunId, IReadOnlyList<string> LayoutArguments, string? ErrorCode);

public static class GeometryActivation
{
    public static GeometryActivationResult Parse(IReadOnlyList<string> args)
    {
        const string prefix = "--layout-email-geometry-run=";
        string[] required = ["--layout-test", "--layout-language=en-US", "--layout-profile=rtl-mixed"];
        if (args is null || args.Count != 4 || required.Any(r => args.Count(a => a == r) != 1)) return Reject();
        var value = args.SingleOrDefault(a => a is not null && a.StartsWith(prefix, StringComparison.Ordinal));
        if (value is null) return Reject();
        var id = value[prefix.Length..];
        if (!Guid.TryParseExact(id, "N", out var run) || run == Guid.Empty || run.ToString("N") != id) return Reject();
        return new(run, Array.AsReadOnly(args.Where(a => a != value).ToArray()), null);
    }

    private static GeometryActivationResult Reject() => new(null, Array.Empty<string>(), "invalid-activation");
}
