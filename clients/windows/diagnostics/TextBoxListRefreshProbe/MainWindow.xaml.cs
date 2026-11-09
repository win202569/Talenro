using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;

namespace TextBoxListRefreshProbe;

// Experimental diagnostic spike, not a production search implementation.
public sealed partial class MainWindow : Window
{
    private readonly bool rebuildOnTextChange;
    private bool ready;

    public MainWindow(bool rebuildOnTextChange)
    {
        this.rebuildOnTextChange = rebuildOnTextChange;
        InitializeComponent();
        RebuildRows();
        ready = true;
    }

    private void Input_Changed(object sender, TextChangedEventArgs args)
    {
        if (ready && rebuildOnTextChange)
            RebuildRows();
    }

    private void RebuildRows()
    {
        ProbeRows.Items.Clear();
        foreach (var row in new[] { "jp-1", "us-1", "de-1" })
            ProbeRows.Items.Add(new TextBlock { Text = row });
    }
}
