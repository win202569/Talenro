using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using Talenro.Demo.App.ViewModels;
using Talenro.Demo.Core;
namespace Talenro.Demo.App.Views;
public sealed partial class NodesPage : Page
{
    private readonly DemoViewModel model;
    private bool rendering;
    public NodesPage(DemoViewModel model) { this.model = model; InitializeComponent(); Refresh(); }
    public void Reset()
    {
        SearchBox.Text = "";
        ListScenario.SelectedIndex = 0;
    }
    public void Refresh()
    {
        if (NodeList is null || ListStateNotice is null || ApplyNodeButton is null) return;
        rendering = true;
        try
        {
            NodeList.Items.Clear();
            var scenario = (ListScenario?.SelectedItem as ComboBoxItem)?.Tag as string ?? "Ready";
            var blocked = model.State.Connection is ConnectionPhase.Connecting or ConnectionPhase.Disconnecting;
            ApplyNodeButton.IsEnabled = !blocked && scenario == "Ready" && !model.State.HasPendingNodeChange;
            NodeList.IsEnabled = !blocked && !model.State.HasPendingNodeChange;
            SearchBox.IsEnabled = !blocked;
            if (scenario != "Ready")
            {
                ListStateNotice.Text = MainWindow.Copy("Nodes" + scenario);
                return;
            }
            var query = SearchBox.Text.Trim();
            var visible = DemoFixtures.Nodes.Where(node =>
                (node.Id + " " + node.Country + " " + MainWindow.Copy("Country" + node.Country))
                    .Contains(query, StringComparison.OrdinalIgnoreCase)).ToArray();
            if (query.Length == 0) AddNode(null, MainWindow.Copy("AutoNode"), true);
            foreach (var group in visible.GroupBy(node => node.Country))
            {
                NodeList.Items.Add(new ListViewItem { Content = MainWindow.Copy("Country" + group.Key), IsEnabled = false });
                foreach (var node in group)
                    AddNode(node.Id, node.Id + " · " + MainWindow.Copy("ProtocolAuto") +
                        (node.Enabled ? "" : " · " + MainWindow.Copy("NodeDisabled")), node.Enabled);
            }
            ListStateNotice.Text = blocked ? MainWindow.Copy("NodeTransitionBlocked")
                : visible.Length == 0 && query.Length != 0 ? MainWindow.Copy("NodesEmpty") : MainWindow.Copy("NodeFixtures");
        }
        finally { rendering = false; }
    }
    private void AddNode(string? id, string label, bool enabled)
    {
        var item = new ListViewItem { Content = label, Tag = id, IsEnabled = enabled };
        NodeList.Items.Add(item);
        if (enabled && id == model.State.DraftNode) NodeList.SelectedItem = item;
    }
    private void Search_Changed(object sender, TextChangedEventArgs e) => Refresh();
    private void ListScenario_Changed(object sender, SelectionChangedEventArgs e) => Refresh();
    private void Node_Selected(object sender, SelectionChangedEventArgs e)
    {
        if (!rendering && NodeList.SelectedItem is ListViewItem { IsEnabled: true } item)
            model.Dispatch(new SelectNode(item.Tag as string));
    }
    private void Back_Click(object sender, RoutedEventArgs e) => model.Dispatch(new CancelNodeSelection());
    private async void Apply_Click(object sender, RoutedEventArgs e)
    {
        if ((ListScenario.SelectedItem as ComboBoxItem)?.Tag as string != "Ready") return;
        model.Dispatch(new ApplyNodeSelection());
        if (!model.State.HasPendingNodeChange) return;
        var result = await new ContentDialog { XamlRoot = XamlRoot,
            Title = MainWindow.Copy("ConfirmNodeChange"), Content = NoticeFormatter.Format(model.State, MainWindow.Copy),
            PrimaryButtonText = MainWindow.Copy("Confirm"), CloseButtonText = MainWindow.Copy("Cancel"),
            DefaultButton = ContentDialogButton.Close }.ShowAsync();
        model.Dispatch(result == ContentDialogResult.Primary ? new ConfirmNodeChange() : new CancelNodeChange());
        if (model.State.Page == DemoPage.Nodes) ApplyNodeButton.Focus(FocusState.Programmatic);
    }
}
