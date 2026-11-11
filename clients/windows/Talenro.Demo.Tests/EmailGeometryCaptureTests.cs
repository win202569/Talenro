using Talenro.Demo.App.Layout.EmailGeometry;
using Xunit;

namespace Talenro.Demo.Tests;

public class EmailGeometryCaptureTests
{
    [Fact]
    public void AttemptsRemainConsumedAcrossFailuresAndReentry()
    {
        var budget = new GeometryCaptureBudget();
        Assert.False(budget.TryReserve("Other", true, out _));
        Assert.False(budget.TryReserve("Login", false, out _));
        Assert.True(budget.TryReserve("Login", true, out var first));
        Assert.Equal(1, first);
        // An acquisition failure has no rollback path.
        Assert.True(budget.TryReserve("Login", true, out var second));
        Assert.Equal(2, second);
        Assert.False(budget.TryReserve("Login", true, out _));
        Assert.True(budget.TryReserve("Account", true, out var third));
        Assert.Equal(3, third);
        Assert.False(budget.TryReserve("Login", true, out _));
        Assert.True(budget.TryReserve("Account", true, out var fourth));
        Assert.Equal(4, fourth);
        Assert.True(budget.Stopped);
        Assert.Equal(2, budget.Attempts("Login"));
        Assert.Equal(2, budget.Attempts("Account"));
        Assert.Equal(0, budget.Attempts("Other"));
        Assert.False(budget.TryReserve("Account", true, out _));
    }

    [Fact]
    public void StopIsPermanent()
    {
        var budget = new GeometryCaptureBudget();
        budget.Stop();
        budget.Stop();
        Assert.True(budget.Stopped);
        Assert.False(budget.TryReserve("Login", true, out var sequence));
        Assert.Equal(0, sequence);
    }

    [Fact]
    public void NavigationDrainsPhysicalSlotBeforeAllowingNewWork()
    {
        var budget = new GeometryCaptureBudget();
        var coordinator = new GeometryCaptureCoordinator(budget);
        coordinator.ChangePage("Login");
        Assert.True(coordinator.TryQueue(out var old));
        Assert.False(coordinator.TryQueue(out _));
        coordinator.ChangePage("Account");
        Assert.False(coordinator.TryQueue(out _));
        Assert.False(coordinator.TryBegin(old, true, out _));
        Assert.True(coordinator.TryQueue(out var current));
        Assert.NotEqual(old, current);
        coordinator.Cancel(old);
        Assert.False(coordinator.TryBegin(old, true, out _));
        Assert.False(coordinator.TryQueue(out _));
        Assert.True(coordinator.TryBegin(current, true, out var sequence));
        Assert.Equal(1, sequence);
        Assert.False(coordinator.TryBegin(current, true, out _));
        Assert.Equal(0, budget.Attempts("Login"));
        Assert.Equal(1, budget.Attempts("Account"));
    }

    [Fact]
    public void RepeatedNavigationCannotAccumulatePhysicalDispatcherCallbacks()
    {
        var budget = new GeometryCaptureBudget();
        var coordinator = new GeometryCaptureCoordinator(budget);
        var callbacks = new Queue<Action>();
        var captures = 0;
        for (var i = 0; i < 100; i++)
        {
            coordinator.ChangePage(i % 2 == 0 ? "Login" : "Account");
            if (coordinator.TryQueue(out var token)) callbacks.Enqueue(() =>
            {
                if (coordinator.TryBegin(token, true, out _)) captures++;
            });
            Assert.Single(callbacks);
        }
        callbacks.Dequeue()();
        Assert.Equal(0, captures);
        Assert.True(coordinator.TryQueue(out var current));
        Assert.True(coordinator.TryBegin(current, true, out _));
        Assert.Equal(1, budget.Attempts("Account"));
    }

    [Fact]
    public void RejectedEnqueueAndIneligibleCallbackDoNotSpendBudget()
    {
        var budget = new GeometryCaptureBudget();
        var coordinator = new GeometryCaptureCoordinator(budget);
        Assert.False(coordinator.TryQueue(out _));
        coordinator.ChangePage("Login");
        Assert.True(coordinator.TryQueue(out var token));
        coordinator.Cancel(token);
        Assert.True(coordinator.TryQueue(out token));
        Assert.False(coordinator.TryBegin(token, false, out _));
        Assert.Equal(0, budget.Attempts("Login"));
        Assert.True(coordinator.TryQueue(out token));
        Assert.True(coordinator.TryBegin(token, true, out _));
        Assert.Equal(1, budget.Attempts("Login"));
        // No second event means no second attempt.
        Assert.Equal(1, budget.Attempts("Login"));
    }

    [Fact]
    public void CloseInvalidatesQueuedWorkAndReentryCannotRestart()
    {
        var budget = new GeometryCaptureBudget();
        var coordinator = new GeometryCaptureCoordinator(budget);
        coordinator.ChangePage("Login");
        Assert.True(coordinator.TryQueue(out var token));
        coordinator.Stop();
        Assert.False(coordinator.TryBegin(token, true, out _));
        coordinator.ChangePage("Account");
        Assert.False(coordinator.TryQueue(out _));
        Assert.Equal(0, budget.Attempts("Login"));
    }

    [Fact]
    public void ReentrantAcquisitionCannotExceedPageBudget()
    {
        var budget = new GeometryCaptureBudget();
        var coordinator = new GeometryCaptureCoordinator(budget);
        coordinator.ChangePage("Login");
        for (var i = 1; i <= 2; i++)
        {
            Assert.True(coordinator.TryQueue(out var token));
            Assert.True(coordinator.TryBegin(token, true, out var sequence));
            Assert.Equal(i, sequence);
        }
        Assert.False(coordinator.TryQueue(out _));
        coordinator.ChangePage("Account");
        coordinator.ChangePage("Login");
        Assert.False(coordinator.TryQueue(out _));
    }
}
