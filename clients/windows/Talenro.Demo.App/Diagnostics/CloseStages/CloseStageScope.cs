namespace Talenro.Demo.App.Layout.CloseStages;

internal sealed class CloseStageScope(CloseStageBuffer buffer)
{
    public void Run(CloseStage before, CloseStage after, Action operation)
    {
        buffer.Mark(before);
        operation();
        buffer.Mark(after);
    }
}
