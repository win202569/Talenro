namespace Talenro.Demo.App.Layout.CloseStages;

internal enum CloseStage
{
    CloseEntry = 1, DisposeEntry, AlreadyDisposedReturn,
    RootLoadedBefore, RootLoadedAfter, RootUnloadedBefore, RootUnloadedAfter,
    RootSizeBefore, RootSizeAfter, HostLoadedBefore, HostLoadedAfter,
    HostUnloadedBefore, HostUnloadedAfter, HostSizeBefore, HostSizeAfter,
    WindowClosedBefore, WindowClosedAfter, XamlRootBefore, XamlRootAfter,
    XamlRootSkipped, UISettingsBefore, UISettingsAfter, UISettingsSkipped,
    RootClearBefore, RootClearAfter, DisposeReturn, CloseReturn
}

internal readonly record struct CloseStageRecord(int Sequence, CloseStage Stage, long Ticks);
