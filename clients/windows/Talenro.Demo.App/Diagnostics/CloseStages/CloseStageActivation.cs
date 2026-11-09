namespace Talenro.Demo.App.Layout.CloseStages;
internal sealed record CloseStageActivationResult(Guid? RunId, IReadOnlyList<string> LayoutArguments, string? ErrorCode);
internal static class CloseStageActivation
{
    public static CloseStageActivationResult Parse(IReadOnlyList<string> args)
    {
        const string prefix = "--layout-close-run=";
        Guid? run = null;
        var remaining = new List<string>();
        foreach (string arg in args)
        {
            if (arg.StartsWith("--layout-close", StringComparison.OrdinalIgnoreCase))
            {
                if (run is not null || !arg.StartsWith(prefix, StringComparison.Ordinal)) return Invalid();
                string value = arg[prefix.Length..];
                if (value.Length != 32 || value.Any(c => !(c is >= '0' and <= '9' or >= 'a' and <= 'f')) ||
                    !Guid.TryParseExact(value, "N", out var parsed) || parsed == Guid.Empty) return Invalid();
                run = parsed;
            }
            else remaining.Add(arg);
        }
        if (run is null || !remaining.Contains("--layout-test")) return Invalid();
        return new(run, remaining.AsReadOnly(), null);
    }

    private static CloseStageActivationResult Invalid() => new(null, Array.Empty<string>(), "InvalidCloseStageArguments");

    public static Stream OpenNew(string trustedRoot, Guid runId)
    {
        if (runId == Guid.Empty || !Path.IsPathFullyQualified(trustedRoot)) throw new ArgumentException("Invalid recording destination.");
        string root = Path.GetFullPath(trustedRoot);
        RejectReparseAncestors(root);
        Directory.CreateDirectory(root);
        RejectReparseAncestors(root);
        string path = Path.Combine(root, runId.ToString("N") + ".jsonl");
        if (!string.Equals(Path.GetDirectoryName(Path.GetFullPath(path)), Path.TrimEndingDirectorySeparator(root), StringComparison.OrdinalIgnoreCase))
            throw new IOException("Recording path escaped root.");
        var stream = new FileStream(path, FileMode.CreateNew, FileAccess.Write, FileShare.Read);
        try { RejectReparseAncestors(path); return stream; }
        catch { stream.Dispose(); throw; } // Never remove a possibly concurrently substituted path.
    }

    private static void RejectReparseAncestors(string path)
    {
        for (string? current = path; current is not null; current = Path.GetDirectoryName(current))
        {
            try
            {
                if ((File.GetAttributes(current) & FileAttributes.ReparsePoint) != 0)
                    throw new IOException("Reparse recording path rejected.");
            }
            catch (FileNotFoundException) { }
            catch (DirectoryNotFoundException) { }
        }
    }
}
