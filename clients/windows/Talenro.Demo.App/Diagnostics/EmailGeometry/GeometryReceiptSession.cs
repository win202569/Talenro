using System.Text;
using System.Text.Json;

namespace Talenro.Demo.App.Layout.EmailGeometry;

public sealed record GeometryFinishResult(bool Completed, int Samples, IReadOnlyList<string> ErrorCodes);

// Transport completion never means visual acceptance. The footer is provisional
// until an external observer also verifies normal process exit with code zero.
public sealed class GeometryReceiptSession : IDisposable
{
    private readonly Guid runId;
    private readonly Stream? stream;
    private readonly Func<object, string> serialize;
    private readonly int ownerThread = Environment.CurrentManagedThreadId;
    private readonly object sync = new();
    private readonly List<GeometrySnapshot> samples = [];
    private readonly List<string> errors = [];
    private readonly List<Exception> causes = [];
    private bool released;
    private GeometryFinishResult? result;
    internal Guid RunId => runId;

    internal void Invalidate()
    {
        lock (sync)
        {
            Fail("observer-failed");
            if (result is not null) result = new(false, samples.Count, Array.AsReadOnly(errors.ToArray()));
        }
    }

    public GeometryReceiptSession(Guid runId, Stream ownedStream)
        : this(runId, ownedStream, value => JsonSerializer.Serialize(value)) { }

    internal GeometryReceiptSession(Guid runId, Stream ownedStream, Func<object, string> serialize)
    {
        this.runId = runId;
        stream = ownedStream;
        this.serialize = serialize;
        try
        {
            if (runId == Guid.Empty || stream is null || !stream.CanWrite) Fail("stream-invalid");
        }
        catch (Exception ex) { Fail("stream-invalid", ex); }
        if (errors.Count != 0) Release();
    }

    public void Add(GeometrySnapshot snapshot)
    {
        lock (sync)
        {
            if (result is not null || released) return;
            if (!OnOwnerThread() || errors.Count != 0) return;
            try
            {
                if (!Valid(snapshot)) { Fail("invalid-snapshot"); return; }
                // Copy nested collections; callers must not be able to alter validated
                // payloads (or introduce private strings) after Add returns.
                samples.Add(snapshot with
                {
                    FieldStatuses = new Dictionary<string, GeometryStatus>(snapshot.FieldStatuses),
                    Scroll = snapshot.Scroll with { FieldStatuses = new Dictionary<string, GeometryStatus>(snapshot.Scroll.FieldStatuses) },
                    Elements = snapshot.Elements.Select(e => e with
                    {
                        FieldStatuses = new Dictionary<string, GeometryStatus>(e.FieldStatuses)
                    }).ToArray(),
                    Boundaries = snapshot.Boundaries.ToArray()
                });
            }
            catch (Exception ex) { Fail("snapshot-failed", ex); }
        }
    }

    public GeometryFinishResult Finish()
    {
        lock (sync)
        {
            if (result is not null) return result;
            OnOwnerThread();
            if (!released)
            {
                // Serialize completely before any IO, so serialization failure leaves
                // no misleading partial valid header/footer. At most four samples.
                byte[]? bytes = null;
                try
                {
                    var lines = new List<string>
                    {
                        serialize(new { Type = "header", Schema = 1, Run = runId.ToString("N") })
                    };
                    lines.AddRange(samples.Select(s => serialize(new
                    {
                        Type = "snapshot", Schema = 1, Run = runId.ToString("N"), Data = s
                    })));
                    lines.Add(serialize(new
                    {
                        Type = "health", Schema = 1, Run = runId.ToString("N"),
                        Samples = samples.Count, Provisional = true, Failed = errors.Count != 0,
                        ErrorCodes = errors.ToArray()
                    }));
                    bytes = new UTF8Encoding(false, true).GetBytes(string.Join('\n', lines) + "\n");
                }
                catch (Exception ex) { Fail("serialize-failed", ex); }
                if (bytes is not null)
                {
                    try { stream!.Write(bytes, 0, bytes.Length); }
                    catch (Exception ex) { Fail("write-failed", ex); }
                    try { stream!.Flush(); }
                    catch (Exception ex) { Fail("flush-failed", ex); }
                }
                Release();
            }
            result = new(errors.Count == 0, samples.Count, Array.AsReadOnly(errors.ToArray()));
            return result;
        }
    }

    public void Dispose()
    {
        lock (sync)
        {
            if (result is not null) return;
            OnOwnerThread();
            Fail("unfinished");
            Release();
            result = new(false, samples.Count, Array.AsReadOnly(errors.ToArray()));
        }
    }

    private bool OnOwnerThread()
    {
        if (Environment.CurrentManagedThreadId == ownerThread) return true;
        Fail("wrong-thread");
        return false;
    }

    private void Fail(string code, Exception? cause = null)
    {
        if (!errors.Contains(code)) errors.Add(code);
        if (cause is not null && causes.Count < 16) causes.Add(cause);
    }

    private void Release()
    {
        if (released) return;
        released = true;
        try { stream?.Dispose(); }
        catch (Exception ex) { Fail("dispose-failed", ex); }
    }

    private bool Valid(GeometrySnapshot s)
    {
        if (s is null || s.RunId != runId || s.Sequence != samples.Count + 1 || samples.Count >= 4 ||
            s.Page is not ("Login" or "Account") || samples.Count(x => x.Page == s.Page) >= 2 ||
            !double.IsFinite(s.ElapsedMs) || s.ElapsedMs < 0 ||
            (samples.Count > 0 && s.ElapsedMs < samples[^1].ElapsedMs) ||
            s.Elements is null || s.Elements.Count > 16 || s.Boundaries is null || s.Boundaries.Count > 16 ||
            !Fields(s.FieldStatuses, ("RenderingScale", s.RenderingScale), ("TextScale", s.TextScale)) ||
            !PositiveIfPresent(s.RenderingScale) || !PositiveIfPresent(s.TextScale) || s.Scroll is null) return false;
        var scroll = s.Scroll;
        if (!Fields(scroll.FieldStatuses, ("HorizontalOffset", scroll.HorizontalOffset), ("VerticalOffset", scroll.VerticalOffset),
            ("ViewportWidth", scroll.ViewportWidth), ("ViewportHeight", scroll.ViewportHeight))) return false;
        if (!Nonnegative(scroll.HorizontalOffset) || !Nonnegative(scroll.VerticalOffset) ||
            !Nonnegative(scroll.ViewportWidth) || !Nonnegative(scroll.ViewportHeight)) return false;
        if (s.Elements.Any(e => !ValidElement(e)) || s.Boundaries.Any(b => b is null || !BoundaryRole(b.Role) || !Bounds(b.Bounds, false))) return false;
        if (s.Elements.Select(e => e.Role).Distinct().Count() != s.Elements.Count ||
            s.Boundaries.Select(b => b.Role).Distinct().Count() != s.Boundaries.Count) return false;
        // A claim of complete ancestry requires at least one complete known boundary.
        return !s.ClipChainComplete || (s.Boundaries.Count > 0 && s.Boundaries.All(b => b.Complete && b.Bounds.Status == GeometryStatus.Ok));
    }

    private static bool ValidElement(GeometryElement e)
    {
        if (e is null || !Role(e.Role) || !Enum.IsDefined(e.Status) || !Bounds(e.RootBounds, false) || !Bounds(e.LayoutSlot, true)) return false;
        if (!Fields(e.FieldStatuses, ("Loaded", e.Loaded), ("Visible", e.Visible),
            ("ActualWidth", e.ActualWidth), ("ActualHeight", e.ActualHeight),
            ("DesiredWidth", e.DesiredWidth), ("DesiredHeight", e.DesiredHeight),
            ("FlowDirection", e.FlowDirection), ("HorizontalAlignment", e.HorizontalAlignment),
            ("TextAlignment", e.TextAlignment), ("Margin", e.Margin), ("Padding", e.Padding),
            ("Width", e.Width), ("MaxWidth", e.MaxWidth))) return false;
        return Nonnegative(e.ActualWidth) && Nonnegative(e.ActualHeight) && Nonnegative(e.DesiredWidth) &&
            Nonnegative(e.DesiredHeight) && Nonnegative(e.Width) && Nonnegative(e.MaxWidth) &&
            (e.FlowDirection is null or "LeftToRight" or "RightToLeft") &&
            (e.HorizontalAlignment is null or "Left" or "Center" or "Right" or "Stretch") &&
            (e.TextAlignment is null or "Left" or "Center" or "Right" or "Justify" or "DetectFromContent");
    }

    private static bool Fields(IReadOnlyDictionary<string, GeometryStatus> statuses, params (string Name, object? Value)[] fields)
    {
        if (statuses is null || statuses.Count != fields.Length) return false;
        foreach (var (name, value) in fields)
        {
            if (!statuses.TryGetValue(name, out var status) || !Enum.IsDefined(status)) return false;
            if (status != GeometryStatus.Ok) { if (value is not null && value is not bool) return false; continue; }
            if (value is null || value is double d && !double.IsFinite(d)) return false;
            if (value is GeometryThickness t && (!double.IsFinite(t.Left) || !double.IsFinite(t.Top) ||
                !double.IsFinite(t.Right) || !double.IsFinite(t.Bottom))) return false;
        }
        return true;
    }

    private static bool PositiveIfPresent(double? value) => value is null || value > 0;
    private static bool Nonnegative(double? value) => value is null || value >= 0;
    private static bool Role(string role) => role is "root" or "PageHost" or "page" or "scroll" or "panel" or "email" or "presenter" ||
        Enumerable.Range(1, 16).Any(i => role == "ancestor-" + i);
    private static bool BoundaryRole(string role) => Role(role) ||
        role is not null && role.EndsWith(":clip", StringComparison.Ordinal) && Role(role[..^5]);
    private static bool Bounds(GeometryField<GeometryRect> field, bool local)
    {
        if (field is null || !Enum.IsDefined(field.Status)) return false;
        if (field.Status != GeometryStatus.Ok) return field.Value is null;
        var b = field.Value;
        return b is not null && (local ? b.Space is not null && b.Space.StartsWith("ParentDIP:", StringComparison.Ordinal) && Role(b.Space[10..]) : b.Space == "RootDIP") &&
            double.IsFinite(b.Left) && double.IsFinite(b.Top) && double.IsFinite(b.Right) && double.IsFinite(b.Bottom) &&
            b.Right > b.Left && b.Bottom > b.Top;
    }

    // Internal trusted-root API, not a user-selectable output path. A persistent
    // create-new claim serializes cooperating writers; receipts are never removed.
    internal static Stream OpenNew(string trustedAbsoluteRoot, Guid runId, Func<string, FileAttributes>? attributes = null)
    {
        if (runId == Guid.Empty || !Path.IsPathFullyQualified(trustedAbsoluteRoot)) throw new ArgumentException("Invalid recording location.");
        attributes ??= File.GetAttributes;
        var root = Path.GetFullPath(trustedAbsoluteRoot);
        CheckAncestors(root, attributes);
        Directory.CreateDirectory(root);
        CheckAncestors(root, attributes);
        var directory = Path.Combine(root, runId.ToString("N"));
        if (Directory.Exists(directory) || File.Exists(directory)) throw new IOException("Run already exists.");
        using var claim = new FileStream(Path.Combine(root, runId.ToString("N") + ".claim"), FileMode.CreateNew, FileAccess.Write, FileShare.None);
        CheckAncestors(directory, attributes);
        if (Directory.Exists(directory) || File.Exists(directory)) throw new IOException("Run already exists.");
        Directory.CreateDirectory(directory);
        CheckAncestors(directory, attributes);
        var path = Path.Combine(directory, "geometry.jsonl");
        var output = new FileStream(path, FileMode.CreateNew, FileAccess.Write, FileShare.Read);
        try { CheckAncestors(path, attributes); return output; }
        catch { output.Dispose(); throw; }
    }

    private static void CheckAncestors(string path, Func<string, FileAttributes> attributes)
    {
        for (string? current = path; current is not null; current = Path.GetDirectoryName(current))
        {
            try
            {
                if ((attributes(current) & FileAttributes.ReparsePoint) != 0) throw new IOException("Reparse path rejected.");
            }
            catch (FileNotFoundException) { }
            catch (DirectoryNotFoundException) { }
        }
    }
}
