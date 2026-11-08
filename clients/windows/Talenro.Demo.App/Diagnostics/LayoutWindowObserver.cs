using System.Runtime.InteropServices;
using Microsoft.UI.Xaml;
using Windows.UI.ViewManagement;
namespace Talenro.Demo.App.Layout;

internal sealed class LayoutWindowObserver : IDisposable
{
    private readonly MainWindow window;
    private readonly FrameworkElement root;
    private readonly FrameworkElement pageHost;
    private readonly Func<bool> active;
    private readonly LayoutOptions options;
    private readonly UISettings? settings;
    private XamlRoot? observedRoot;
    private bool disposed;
    private string? lastTitle;

    public LayoutWindowObserver(MainWindow window, FrameworkElement root, FrameworkElement pageHost,
        Func<bool> active, LayoutOptions options)
    {
        this.window = window;
        this.root = root;
        this.pageHost = pageHost;
        this.active = active;
        this.options = options;
        try
        {
            settings = new UISettings();
            settings.TextScaleFactorChanged += TextScaleChanged;
        }
        catch (Exception error) when (error is COMException or NotSupportedException)
        { settings = null; }
        root.Loaded += Loaded;
        root.Unloaded += Unloaded;
        root.SizeChanged += SizeChanged;
        pageHost.Loaded += Loaded;
        pageHost.Unloaded += Unloaded;
        pageHost.SizeChanged += SizeChanged;
        window.Closed += Closed;
        Refresh();
    }

    public void Refresh()
    {
        if (disposed) return;
        if (!root.DispatcherQueue.HasThreadAccess)
        {
            root.DispatcherQueue.TryEnqueue(Refresh);
            return;
        }
        var current = root.IsLoaded ? root.XamlRoot : null;
        if (!ReferenceEquals(current, observedRoot))
        {
            if (observedRoot is not null) observedRoot.Changed -= RootChanged;
            observedRoot = current;
            if (observedRoot is not null) observedRoot.Changed += RootChanged;
        }
        double? textScale = null;
        try { textScale = settings?.TextScaleFactor; }
        catch (Exception error) when (error is COMException or NotSupportedException) { }
        var values = new LayoutMeasurement(
            current is null ? null : root.ActualWidth, current is null ? null : root.ActualHeight,
            current is null || !pageHost.IsLoaded ? null : pageHost.ActualWidth,
            current is null || !pageHost.IsLoaded ? null : pageHost.ActualHeight,
            current?.RasterizationScale, textScale);
        var title = LayoutTitle.Format(options, active(), values);
        if (title == lastTitle) return;
        window.Title = title;
        lastTitle = title;
    }

    private void Loaded(object sender, RoutedEventArgs args) => Refresh();
    private void Unloaded(object sender, RoutedEventArgs args) => Refresh();
    private void SizeChanged(object sender, SizeChangedEventArgs args) => Refresh();
    private void RootChanged(XamlRoot sender, XamlRootChangedEventArgs args) => Refresh();
    private void TextScaleChanged(UISettings sender, object args) => root.DispatcherQueue.TryEnqueue(Refresh);
    private void Closed(object sender, WindowEventArgs args) => Dispose();

    public void Dispose()
    {
        if (disposed) return;
        disposed = true;
        root.Loaded -= Loaded;
        root.Unloaded -= Unloaded;
        root.SizeChanged -= SizeChanged;
        pageHost.Loaded -= Loaded;
        pageHost.Unloaded -= Unloaded;
        pageHost.SizeChanged -= SizeChanged;
        window.Closed -= Closed;
        if (observedRoot is not null) observedRoot.Changed -= RootChanged;
        if (settings is not null) settings.TextScaleFactorChanged -= TextScaleChanged;
        observedRoot = null;
    }
}
