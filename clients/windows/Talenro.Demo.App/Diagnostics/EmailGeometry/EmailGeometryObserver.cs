using System.Diagnostics;
using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using Microsoft.UI.Xaml.Controls.Primitives;
using Microsoft.UI.Xaml.Media;
using Windows.Foundation;
using Windows.UI.ViewManagement;

namespace Talenro.Demo.App.Layout.EmailGeometry;

internal sealed record EmailGeometryTargets(Page Page, ScrollViewer Scroll, StackPanel Panel, TextBlock Email);

// Read-only adapter. No timer, model actions, layout forcing or disk IO here.
internal sealed class EmailGeometryObserver : IDisposable
{
    private readonly FrameworkElement root;
    private readonly FrameworkElement pageHost;
    private readonly Func<string?> currentPage;
    private readonly Func<EmailGeometryTargets?> currentTargets;
    private readonly GeometryReceiptSession session;
    private readonly GeometryCaptureBudget budget = new();
    private readonly GeometryCaptureCoordinator coordinator;
    private readonly Stopwatch elapsed = Stopwatch.StartNew();
    private EmailGeometryTargets? subscribed;
    private string? page;
    private bool disposed;

    internal EmailGeometryObserver(FrameworkElement root, FrameworkElement pageHost,
        Func<string?> currentPage, Func<EmailGeometryTargets?> currentTargets, GeometryReceiptSession session)
    {
        this.root = root;
        this.pageHost = pageHost;
        this.currentPage = currentPage;
        this.currentTargets = currentTargets;
        this.session = session;
        coordinator = new(budget);
    }

    internal void RefreshPage()
    {
        if (disposed) return;
        try
        {
            var next = currentTargets();
            var name = currentPage();
            if (ReferenceEquals(subscribed?.Page, next?.Page) && page == name) return;
            coordinator.ChangePage(name);
            Detach();
            page = name;
            if (budget.Stopped || name is not ("Login" or "Account") || next is null || budget.Attempts(name) >= 2) return;
            subscribed = next;
            next.Page.Loaded += Loaded;
            next.Page.Unloaded += Unloaded;
            next.Page.LayoutUpdated += LayoutUpdated;
        }
        catch { Fail(); }
    }

    private void Loaded(object sender, RoutedEventArgs args)
    {
        // Loading does not itself count as a layout sample.
        coordinator.ChangePage(page);
    }

    private void Unloaded(object sender, RoutedEventArgs args) => coordinator.ChangePage(null);

    private void LayoutUpdated(object? sender, object args)
    {
        if (disposed || subscribed is null) return;
        try
        {
            var targets = subscribed;
            var name = page;
            if (!Eligible(targets, name) || !coordinator.TryQueue(out var token)) return;
            if (!root.DispatcherQueue.TryEnqueue(() => Capture(token, targets, name!))) coordinator.Cancel(token);
        }
        catch { Fail(); }
    }

    private bool Eligible(EmailGeometryTargets targets, string? name) =>
        !disposed && name == currentPage() && ReferenceEquals(targets.Page, currentTargets()?.Page) &&
        pageHost is ContentControl host && ReferenceEquals(host.Content, targets.Page) &&
        root.IsLoaded && targets.Page.IsLoaded && targets.Page.Visibility == Visibility.Visible &&
        root.XamlRoot is not null && ReferenceEquals(root.XamlRoot, targets.Page.XamlRoot);

    private void Capture(long token, EmailGeometryTargets targets, string name)
    {
        if (disposed) return;
        try
        {
            if (!coordinator.TryBegin(token, Eligible(targets, name), out var sequence)) return;
            session.Add(Measure(targets, name, sequence));
            if (budget.Stopped || budget.Attempts(name) >= 2) Detach();
        }
        catch { Fail(); }
    }

    private GeometrySnapshot Measure(EmailGeometryTargets targets, string name, int sequence)
    {
        var nodes = new List<(string Role, FrameworkElement Node)>
        {
            ("root", root), ("PageHost", pageHost), ("page", targets.Page),
            ("scroll", targets.Scroll), ("panel", targets.Panel), ("email", targets.Email)
        };
        var chain = new List<FrameworkElement>();
        try
        {
            DependencyObject? node = targets.Email;
            for (var i = 0; i < 16 && node is not null; i++)
            {
                if (node is FrameworkElement element) chain.Add(element);
                if (ReferenceEquals(node, root)) break;
                node = VisualTreeHelper.GetParent(node);
            }
        }
        catch { /* Unknown chain is explicit below; never infer clipping from parents. */ }
        var presenters = chain.OfType<ScrollContentPresenter>().ToArray();
        if (presenters.Length == 1) nodes.Add(("presenter", presenters[0]));
        foreach (var element in chain)
        {
            if (nodes.Count == 16) break;
            if (!nodes.Any(n => ReferenceEquals(n.Node, element))) nodes.Add(("ancestor-" + (nodes.Count + 1), element));
        }
        var elements = nodes.Select(n => ReadElement(n.Role, n.Node, nodes)).ToArray();
        var boundaries = new List<GeometryBoundary>
        {
            new("root", RootBounds(root), true),
            new("presenter", presenters.Length == 1 ? RootBounds(presenters[0]) : new(GeometryStatus.Missing, null), presenters.Length == 1)
        };
        // Explicit rectangular clips can be mapped when no extra geometry transform
        // is present. Implicit template/compositor clipping remains unknown.
        foreach (var (role, node) in nodes)
        {
            if (role is "root" or "presenter" || boundaries.Count == 16) continue;
            try
            {
                var clip = node.Clip;
                if (clip is null) continue;
                var bounds = clip.Transform is null ? Transform(node, clip.Rect) : new GeometryField<GeometryRect>(GeometryStatus.Unsupported, null);
                boundaries.Add(new(role, bounds, bounds.Status == GeometryStatus.Ok));
            }
            catch { boundaries.Add(new(role, new(GeometryStatus.Error, null), false)); }
        }
        var scrollFields = new Dictionary<string, GeometryStatus>();
        var scroll = new GeometryScroll(
            Number("HorizontalOffset", () => targets.Scroll.HorizontalOffset, scrollFields),
            Number("VerticalOffset", () => targets.Scroll.VerticalOffset, scrollFields),
            Number("ViewportWidth", () => targets.Scroll.ViewportWidth, scrollFields),
            Number("ViewportHeight", () => targets.Scroll.ViewportHeight, scrollFields), scrollFields);
        var scaleFields = new Dictionary<string, GeometryStatus>();
        var render = Number("RenderingScale", () => root.XamlRoot.RasterizationScale, scaleFields, true);
        var text = Number("TextScale", () => new UISettings().TextScaleFactor, scaleFields, true);
        return new(session.RunId, sequence, name, elapsed.Elapsed.TotalMilliseconds, render, text, scroll,
            elements, boundaries, false, scaleFields);
    }

    private GeometryElement ReadElement(string role, FrameworkElement node, List<(string Role, FrameworkElement Node)> nodes)
    {
        var fields = new Dictionary<string, GeometryStatus>();
        var loaded = Value("Loaded", () => node.IsLoaded, fields);
        var visible = Value("Visible", () => node.Visibility == Visibility.Visible, fields);
        GeometryField<GeometryRect> slot;
        try
        {
            var parent = nodes.FirstOrDefault(n => ReferenceEquals(n.Node, node.Parent));
            var rect = LayoutInformation.GetLayoutSlot(node);
            slot = parent.Node is null ? new(GeometryStatus.Unsupported, null) : GeometryEvaluator.Normalize("ParentDIP:" + parent.Role,
                [new(rect.Left, rect.Top), new(rect.Right, rect.Top), new(rect.Right, rect.Bottom), new(rect.Left, rect.Bottom)]);
        }
        catch { slot = new(GeometryStatus.Error, null); }
        var actualWidth = Number("ActualWidth", () => node.ActualWidth, fields);
        var actualHeight = Number("ActualHeight", () => node.ActualHeight, fields);
        var desiredWidth = Number("DesiredWidth", () => node.DesiredSize.Width, fields);
        var desiredHeight = Number("DesiredHeight", () => node.DesiredSize.Height, fields);
        var flow = Text("FlowDirection", () => node.FlowDirection.ToString(), fields);
        var alignment = Text("HorizontalAlignment", () => node.HorizontalAlignment.ToString(), fields);
        string? textAlignment = null;
        fields["TextAlignment"] = GeometryStatus.Unsupported;
        if (node is TextBlock block) textAlignment = Text("TextAlignment", () => block.TextAlignment.ToString(), fields);
        var margin = ThicknessValue("Margin", () => node.Margin, fields);
        GeometryThickness? padding = null;
        fields["Padding"] = GeometryStatus.Unsupported;
        if (node is StackPanel panel) padding = ThicknessValue("Padding", () => panel.Padding, fields);
        else if (node is Control control) padding = ThicknessValue("Padding", () => control.Padding, fields);
        else if (node is TextBlock textBlock) padding = ThicknessValue("Padding", () => textBlock.Padding, fields);
        var width = Number("Width", () => node.Width, fields, constraint: true);
        var maxWidth = Number("MaxWidth", () => node.MaxWidth, fields, constraint: true);
        return new(role, loaded == true ? GeometryStatus.Ok : GeometryStatus.Unloaded, loaded ?? false, visible ?? false,
            RootBounds(node), slot, actualWidth, actualHeight, desiredWidth, desiredHeight,
            flow, alignment, textAlignment, margin, padding, width, maxWidth, fields);
    }

    private GeometryField<GeometryRect> RootBounds(FrameworkElement node)
    {
        try
        {
            if (!node.IsLoaded || !ReferenceEquals(node.XamlRoot, root.XamlRoot)) return new(GeometryStatus.Unloaded, null);
            return Transform(node, new Rect(0, 0, node.ActualWidth, node.ActualHeight));
        }
        catch { return new(GeometryStatus.Error, null); }
    }

    private GeometryField<GeometryRect> Transform(FrameworkElement node, Rect rect)
    {
        try
        {
            var transform = node.TransformToVisual(root);
            var points = new[] { new Point(rect.Left, rect.Top), new Point(rect.Right, rect.Top), new Point(rect.Right, rect.Bottom), new Point(rect.Left, rect.Bottom) };
            return GeometryEvaluator.Normalize("RootDIP", points.Select(p => transform.TransformPoint(p)).Select(p => new GeometryPoint(p.X, p.Y)).ToArray());
        }
        catch { return new(GeometryStatus.Error, null); }
    }

    private static T? Value<T>(string name, Func<T> read, Dictionary<string, GeometryStatus> fields) where T : struct
    {
        try { var value = read(); fields[name] = GeometryStatus.Ok; return value; }
        catch { fields[name] = GeometryStatus.Error; return null; }
    }

    private static double? Number(string name, Func<double> read, Dictionary<string, GeometryStatus> fields, bool positive = false, bool constraint = false)
    {
        var value = Value(name, read, fields);
        if (value is null) return null;
        if (!double.IsFinite(value.Value)) { fields[name] = constraint ? GeometryStatus.Unsupported : GeometryStatus.Invalid; return null; }
        if (value < 0 || positive && value <= 0) { fields[name] = GeometryStatus.Invalid; return null; }
        return value;
    }

    private static string? Text(string name, Func<string> read, Dictionary<string, GeometryStatus> fields)
    {
        try { var value = read(); fields[name] = GeometryStatus.Ok; return value; }
        catch { fields[name] = GeometryStatus.Error; return null; }
    }

    private static GeometryThickness? ThicknessValue(string name, Func<Thickness> read, Dictionary<string, GeometryStatus> fields)
    {
        var value = Value(name, read, fields);
        if (value is not { } t) return null;
        if (!double.IsFinite(t.Left) || !double.IsFinite(t.Top) || !double.IsFinite(t.Right) || !double.IsFinite(t.Bottom))
        { fields[name] = GeometryStatus.Invalid; return null; }
        return new(t.Left, t.Top, t.Right, t.Bottom);
    }

    private void Detach()
    {
        var previous = subscribed;
        subscribed = null;
        if (previous is null) return;
        if (!GeometryCleanup.Run([() => previous.Page.LayoutUpdated -= LayoutUpdated,
            () => previous.Page.Loaded -= Loaded, () => previous.Page.Unloaded -= Unloaded]))
        { session.Invalidate(); coordinator.Stop(); }
    }

    private void Fail()
    {
        session.Invalidate();
        Dispose();
    }

    public void Dispose()
    {
        if (disposed) return;
        disposed = true;
        coordinator.Stop();
        Detach();
    }
}
