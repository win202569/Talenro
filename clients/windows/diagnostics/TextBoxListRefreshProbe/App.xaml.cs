using Microsoft.UI.Xaml;

namespace TextBoxListRefreshProbe;

public partial class App : Application
{
    private Window? window;

    public App() => InitializeComponent();

    protected override void OnLaunched(LaunchActivatedEventArgs args)
    {
        var arguments = Environment.GetCommandLineArgs().Skip(1).ToArray();
        bool rebuildOnTextChange = arguments switch
        {
            [] => false,
            ["--mode=static"] => false,
            ["--mode=rebuild"] => true,
            _ => throw new ArgumentException("Expected --mode=static or --mode=rebuild.")
        };
        window = new MainWindow(rebuildOnTextChange);
        window.Activate();
    }
}
