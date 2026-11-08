using System.ComponentModel;
using Talenro.Demo.Core;
namespace Talenro.Demo.App.ViewModels;
public sealed class DemoViewModel : INotifyPropertyChanged
{
    public DemoState State { get; private set; } = DemoFixtures.Initial();
    public event PropertyChangedEventHandler? PropertyChanged;
    public void Dispatch(DemoEvent action)
    {
        State = DemoReducer.Reduce(State, action);
        PropertyChanged?.Invoke(this, new PropertyChangedEventArgs(nameof(State)));
    }
}
