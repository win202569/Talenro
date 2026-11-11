namespace Talenro.Demo.App.Layout.EmailGeometry;

public enum GeometryStatus { Ok, Missing, Unloaded, Invalid, Unsupported, Error }
public enum GeometryVerdict { Inside, Overflow, Unknown }
public sealed record GeometryPoint(double X, double Y);
public sealed record GeometryRect(string Space, double Left, double Top, double Right, double Bottom);
public sealed record GeometryField<T>(GeometryStatus Status, T? Value) where T : class;
public sealed record GeometryThickness(double Left, double Top, double Right, double Bottom);
public sealed record GeometryElement(string Role, GeometryStatus Status, bool Loaded, bool Visible,
    GeometryField<GeometryRect> RootBounds, GeometryField<GeometryRect> LayoutSlot,
    double? ActualWidth, double? ActualHeight, double? DesiredWidth, double? DesiredHeight,
    string? FlowDirection, string? HorizontalAlignment, string? TextAlignment,
    GeometryThickness? Margin, GeometryThickness? Padding, double? Width, double? MaxWidth,
    IReadOnlyDictionary<string, GeometryStatus> FieldStatuses);
public sealed record GeometryBoundary(string Role, GeometryField<GeometryRect> Bounds, bool Complete);
public sealed record GeometryScroll(double? HorizontalOffset, double? VerticalOffset,
    double? ViewportWidth, double? ViewportHeight, IReadOnlyDictionary<string, GeometryStatus> FieldStatuses);
public sealed record GeometrySnapshot(Guid RunId, int Sequence, string Page, double ElapsedMs,
    double? RenderingScale, double? TextScale, GeometryScroll Scroll,
    IReadOnlyList<GeometryElement> Elements, IReadOnlyList<GeometryBoundary> Boundaries,
    bool ClipChainComplete, IReadOnlyDictionary<string, GeometryStatus> FieldStatuses);
public sealed record GeometryComparison(GeometryVerdict Verdict, double? LeftInset,
    double? TopInset, double? RightInset, double? BottomInset, string Reason);

public static class GeometryEvaluator
{
    public static GeometryField<GeometryRect> Normalize(string space, IReadOnlyList<GeometryPoint> corners)
    {
        if (string.IsNullOrWhiteSpace(space) || corners is null || corners.Count != 4 ||
            corners.Any(p => p is null || !double.IsFinite(p.X) || !double.IsFinite(p.Y)))
            return new(GeometryStatus.Invalid, null);
        var box = new GeometryRect(space, corners.Min(p => p.X), corners.Min(p => p.Y),
            corners.Max(p => p.X), corners.Max(p => p.Y));
        return Valid(box) ? new(GeometryStatus.Ok, box) : new(GeometryStatus.Invalid, null);
    }

    public static GeometryComparison Compare(GeometryField<GeometryRect> element, GeometryBoundary boundary)
    {
        if (element is null || boundary is null || !boundary.Complete ||
            element.Status != GeometryStatus.Ok || boundary.Bounds is null || boundary.Bounds.Status != GeometryStatus.Ok ||
            !Valid(element.Value) || !Valid(boundary.Bounds.Value))
            return Unknown("unavailable-or-invalid");
        var e = element.Value!;
        var b = boundary.Bounds.Value!;
        if (!string.Equals(e.Space, b.Space, StringComparison.Ordinal)) return Unknown("space-mismatch");
        var left = e.Left - b.Left;
        var top = e.Top - b.Top;
        var right = b.Right - e.Right;
        var bottom = b.Bottom - e.Bottom;
        if (!double.IsFinite(left) || !double.IsFinite(top) || !double.IsFinite(right) || !double.IsFinite(bottom))
            return Unknown("nonfinite-distance");
        return new(left < 0 || top < 0 || right < 0 || bottom < 0 ? GeometryVerdict.Overflow : GeometryVerdict.Inside,
            left, top, right, bottom, "known-box-only");
    }

    private static GeometryComparison Unknown(string reason) => new(GeometryVerdict.Unknown, null, null, null, null, reason);
    private static bool Valid(GeometryRect? box) => box is not null && !string.IsNullOrWhiteSpace(box.Space) &&
        double.IsFinite(box.Left) && double.IsFinite(box.Top) && double.IsFinite(box.Right) && double.IsFinite(box.Bottom) &&
        box.Right > box.Left && box.Bottom > box.Top;
}
