using Microsoft.UI.Xaml;
using Talenro.Demo.App.Layout;
#if EMAIL_GEOMETRY_RECORDING
using Talenro.Demo.App.Layout.EmailGeometry;
#endif
#if CLOSE_STAGE_RECORDING
using Talenro.Demo.App.Layout.CloseStages;
#endif
namespace Talenro.Demo.App;
public partial class App : Application
{
    private Window? window;
#if EMAIL_GEOMETRY_RECORDING
    internal static GeometryReceiptSession? EmailGeometrySession { get; private set; }
#endif
#if CLOSE_STAGE_RECORDING
    private CloseStageSession? closeStageSession;
    internal static CloseStageBuffer? CloseStageBuffer { get; private set; }
#endif
#if LAYOUT_DIAGNOSTICS
    internal static LayoutPresentation? DiagnosticPresentation { get; private set; }
#endif
    public App() => InitializeComponent();
    protected override void OnLaunched(LaunchActivatedEventArgs args)
    {
        var arguments = Environment.GetCommandLineArgs().Skip(1).ToArray();
#if EMAIL_GEOMETRY_RECORDING
        var geometry = GeometryActivation.Parse(arguments);
        if (geometry.ErrorCode is not null) { RejectLayoutArguments(); return; }
        arguments = geometry.LayoutArguments.ToArray();
#endif
#if CLOSE_STAGE_RECORDING
        var recording = CloseStageActivation.Parse(arguments);
        if (recording.ErrorCode is not null) { RejectLayoutArguments(); return; }
        arguments = recording.LayoutArguments.ToArray();
#endif
#if LAYOUT_DIAGNOSTICS
        var parsed = LayoutOptions.Parse(arguments);
        if (parsed.ErrorCode is not null) { RejectLayoutArguments(); return; }
        if (parsed.Enabled)
        {
            try { DiagnosticPresentation = LayoutResourceCatalog.Prepare(parsed.Options!, typeof(App).Assembly); }
            catch (InvalidDataException) { RejectLayoutArguments(); return; }
        }
#else
        if (ReservedLayoutArguments.ContainsReserved(arguments)) { RejectLayoutArguments(); return; }
#endif
#if CLOSE_STAGE_RECORDING
        try
        {
            CloseStageBuffer = new(Environment.CurrentManagedThreadId);
            var stream = CloseStageActivation.OpenNew(Path.Combine(AppContext.BaseDirectory, "close-stage-records"), recording.RunId!.Value);
            closeStageSession = new(recording.RunId.Value, CloseStageBuffer, stream);
            closeStageSession.Start();
            if (!closeStageSession.WaitUntilReady(TimeSpan.FromSeconds(2)))
            { CloseStageBuffer.Disable(); RejectLayoutArguments(); return; }
        }
        catch (Exception error) when (error is IOException or UnauthorizedAccessException or ArgumentException or InvalidOperationException)
        { CloseStageBuffer?.Disable(); RejectLayoutArguments(); return; }
#endif
#if EMAIL_GEOMETRY_RECORDING
        try
        {
            var stream = GeometryReceiptSession.OpenNew(Path.Combine(AppContext.BaseDirectory, "email-geometry-records"), geometry.RunId!.Value);
            EmailGeometrySession = new(geometry.RunId.Value, stream);
            window = new MainWindow();
            window.Activate();
        }
        catch
        {
            EmailGeometrySession?.Invalidate();
            EmailGeometrySession?.Dispose();
            RejectLayoutArguments();
        }
#else
        window = new MainWindow();
        window.Activate();
#endif
    }
    private void RejectLayoutArguments()
    {
        Console.Error.WriteLine("Layout diagnostic startup rejected.");
        Environment.ExitCode = 2;
        Exit();
    }
}
