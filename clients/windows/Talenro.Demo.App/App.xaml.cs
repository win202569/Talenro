using Microsoft.UI.Xaml;
using Talenro.Demo.App.Layout;
namespace Talenro.Demo.App;
public partial class App : Application
{
    private Window? window;
#if LAYOUT_DIAGNOSTICS
    internal static LayoutPresentation? DiagnosticPresentation { get; private set; }
#endif
    public App() => InitializeComponent();
    protected override void OnLaunched(LaunchActivatedEventArgs args)
    {
        var arguments = Environment.GetCommandLineArgs().Skip(1).ToArray();
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
        window = new MainWindow();
        window.Activate();
    }
    private void RejectLayoutArguments()
    {
        Console.Error.WriteLine("Layout diagnostic startup rejected.");
        Environment.ExitCode = 2;
        Exit();
    }
}
