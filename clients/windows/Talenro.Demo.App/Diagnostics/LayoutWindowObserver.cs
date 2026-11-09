using System.Runtime.InteropServices;
using Microsoft.UI.Xaml;
using Windows.UI.ViewManagement;
#if CLOSE_STAGE_RECORDING
using Talenro.Demo.App.Layout.CloseStages;
#endif
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
#if CLOSE_STAGE_RECORDING
    private readonly CloseStageBuffer stages;
    private readonly CloseStageScope closeScope;
    private readonly Action[] removals;
#endif

    public LayoutWindowObserver(MainWindow window, FrameworkElement root, FrameworkElement pageHost,
        Func<bool> active, LayoutOptions options
#if CLOSE_STAGE_RECORDING
        , CloseStageBuffer stages
#endif
        )
    {
        this.window = window;
        this.root = root;
        this.pageHost = pageHost;
        this.active = active;
        this.options = options;
#if CLOSE_STAGE_RECORDING
        this.stages = stages;
        closeScope = new(stages);
        removals = [() => root.Loaded -= Loaded, () => root.Unloaded -= Unloaded,
            () => root.SizeChanged -= SizeChanged, () => pageHost.Loaded -= Loaded,
            () => pageHost.Unloaded -= Unloaded, () => pageHost.SizeChanged -= SizeChanged,
            () => window.Closed -= Closed, () => observedRoot!.Changed -= RootChanged,
            () => settings!.TextScaleFactorChanged -= TextScaleChanged];
#endif
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
    private void Closed(object sender, WindowEventArgs args)
    {
#if CLOSE_STAGE_RECORDING
        stages.Mark(CloseStage.CloseEntry);
#endif
        Dispose();
#if CLOSE_STAGE_RECORDING
        stages.Mark(CloseStage.CloseReturn);
#endif
    }

    public void Dispose()
    {
#if CLOSE_STAGE_RECORDING
        stages.Mark(CloseStage.DisposeEntry);
        if (disposed) { stages.Mark(CloseStage.AlreadyDisposedReturn); return; }
#else
        if (disposed) return;
#endif
        disposed = true;
#if CLOSE_STAGE_RECORDING
        closeScope.Run(CloseStage.RootLoadedBefore, CloseStage.RootLoadedAfter, removals[0]);
        closeScope.Run(CloseStage.RootUnloadedBefore, CloseStage.RootUnloadedAfter, removals[1]);
        closeScope.Run(CloseStage.RootSizeBefore, CloseStage.RootSizeAfter, removals[2]);
        closeScope.Run(CloseStage.HostLoadedBefore, CloseStage.HostLoadedAfter, removals[3]);
        closeScope.Run(CloseStage.HostUnloadedBefore, CloseStage.HostUnloadedAfter, removals[4]);
        closeScope.Run(CloseStage.HostSizeBefore, CloseStage.HostSizeAfter, removals[5]);
        closeScope.Run(CloseStage.WindowClosedBefore, CloseStage.WindowClosedAfter, removals[6]);
        if (observedRoot is not null) closeScope.Run(CloseStage.XamlRootBefore, CloseStage.XamlRootAfter, removals[7]);
        else stages.Mark(CloseStage.XamlRootSkipped);
        if (settings is not null) closeScope.Run(CloseStage.UISettingsBefore, CloseStage.UISettingsAfter, removals[8]);
        else stages.Mark(CloseStage.UISettingsSkipped);
        stages.Mark(CloseStage.RootClearBefore);
#else
        root.Loaded -= Loaded;
        root.Unloaded -= Unloaded;
        root.SizeChanged -= SizeChanged;
        pageHost.Loaded -= Loaded;
        pageHost.Unloaded -= Unloaded;
        pageHost.SizeChanged -= SizeChanged;
        window.Closed -= Closed;
        if (observedRoot is not null) observedRoot.Changed -= RootChanged;
        if (settings is not null) settings.TextScaleFactorChanged -= TextScaleChanged;
#endif
        observedRoot = null;
#if CLOSE_STAGE_RECORDING
        stages.Mark(CloseStage.RootClearAfter);
        stages.Mark(CloseStage.DisposeReturn);
#endif
    }
}
