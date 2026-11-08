using System.Globalization;
namespace Talenro.Demo.App.Layout;

internal sealed record LayoutMeasurement(double? RootWidth, double? RootHeight, double? PageWidth,
    double? PageHeight, double? RenderingScale, double? TextScale);
internal static class LayoutTitle
{
    public static string Format(LayoutOptions options, bool active, LayoutMeasurement values) =>
        $"LAYOUT TEST {options.Language}/{options.Profile} {(active ? "active" : "inactive")} | " +
        $"root={Value(values.RootWidth)}x{Value(values.RootHeight)} DIP | " +
        $"page={Value(values.PageWidth)}x{Value(values.PageHeight)} DIP | " +
        $"render={Value(values.RenderingScale)} text={Value(values.TextScale)}";

    private static string Value(double? value) => value is > 0 && double.IsFinite(value.Value)
        ? value.Value.ToString("F3", CultureInfo.InvariantCulture) : "unknown";
}
