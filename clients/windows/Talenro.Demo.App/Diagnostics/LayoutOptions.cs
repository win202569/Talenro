namespace Talenro.Demo.App.Layout;

internal sealed record LayoutParseResult(bool Enabled, LayoutOptions? Options, string? ErrorCode);
internal sealed record LayoutOptions(string Language, string Profile)
{
    public static LayoutParseResult Parse(IReadOnlyList<string> args)
    {
        if (args.Count == 0) return new(false, null, null);
        var enabled = false;
        string? language = null, profile = null;
        foreach (var arg in args)
        {
            if (arg == "--layout-test" && !enabled) enabled = true;
            else if (arg.StartsWith("--layout-language=", StringComparison.Ordinal) && language is null)
                language = arg[18..];
            else if (arg.StartsWith("--layout-profile=", StringComparison.Ordinal) && profile is null)
                profile = arg[17..];
            else return Invalid();
        }
        language ??= "zh-Hans";
        profile ??= "baseline";
        if (!enabled || language is not ("zh-Hans" or "en-US") ||
            profile is not ("baseline" or "long-status" or "long-country" or "rtl-mixed" or "unknown-usage"))
            return Invalid();
        return new(true, new(language, profile), null);
    }

    private static LayoutParseResult Invalid() => new(false, null, "InvalidLayoutArguments");
}
