using System.Diagnostics;
using Talenro.Demo.App.Layout.CloseStages;
using Xunit;

namespace Talenro.Demo.Tests;

public class CloseStageBufferTests
{
    [Fact]
    public void BufferPreserves64ThenDrops65th()
    {
        var buffer = new CloseStageBuffer(Environment.CurrentManagedThreadId);
        var start = Stopwatch.GetTimestamp();
        for (int i = 0; i < 65; i++) buffer.Mark(CloseStage.CloseEntry);
        Assert.Equal(1, buffer.Dropped);
        for (int i = 1; i <= 64; i++)
        {
            Assert.True(buffer.TryRead(out var value));
            Assert.Equal(i, value.Sequence);
            Assert.Equal(CloseStage.CloseEntry, value.Stage);
            Assert.InRange(value.Ticks, start, Stopwatch.GetTimestamp());
        }
        Assert.False(buffer.TryRead(out _));
        buffer.Mark(CloseStage.DisposeEntry);
        Assert.True(buffer.TryRead(out var next));
        Assert.Equal(66, next.Sequence); // Dropped marker leaves an observable sequence gap.
    }

    [Fact]
    public void WrongThreadInvalidatesWithoutThrowing()
    {
        var buffer = new CloseStageBuffer(Environment.CurrentManagedThreadId);
        Exception? failure = null;
        var thread = new Thread(() => { try { buffer.Mark(CloseStage.CloseEntry); } catch (Exception e) { failure = e; } });
        thread.Start();
        Assert.True(thread.Join(TimeSpan.FromSeconds(5)));
        Assert.Null(failure);
        Assert.True(buffer.InvalidProducer);
        Assert.True(buffer.Disabled);
        Assert.False(buffer.TryRead(out _));
    }

    [Fact]
    public void DisabledProducerReturnsWithoutEnqueue()
    {
        var buffer = new CloseStageBuffer(Environment.CurrentManagedThreadId);
        buffer.Disable();
        buffer.Mark(CloseStage.CloseEntry);
        Assert.True(buffer.Disabled);
        Assert.False(buffer.TryRead(out _));
    }

    [Fact]
    public void AllocationFreeMark()
    {
        var buffer = new CloseStageBuffer(Environment.CurrentManagedThreadId);
        buffer.Mark(CloseStage.CloseEntry);
        buffer.TryRead(out _);
        long before = GC.GetAllocatedBytesForCurrentThread();
        for (int i = 0; i < 64; i++) buffer.Mark(CloseStage.DisposeEntry);
        long allocated = GC.GetAllocatedBytesForCurrentThread() - before;
        Assert.Equal(0, allocated);
        Assert.True(buffer.TryRead(out _));
    }

    [Fact]
    public void ScopePreservesOriginalExceptionAndOmitsAfter()
    {
        var buffer = new CloseStageBuffer(Environment.CurrentManagedThreadId);
        var scope = new CloseStageScope(buffer);
        var expected = new InvalidOperationException();
        Assert.Same(expected, Assert.Throws<InvalidOperationException>(() =>
            scope.Run(CloseStage.RootLoadedBefore, CloseStage.RootLoadedAfter, () => throw expected)));
        Assert.True(buffer.TryRead(out var value));
        Assert.Equal(CloseStage.RootLoadedBefore, value.Stage);
        Assert.False(buffer.TryRead(out _));
    }

    [Fact]
    public void ReentrantMarkersStayOrdered()
    {
        var buffer = new CloseStageBuffer(Environment.CurrentManagedThreadId);
        var scope = new CloseStageScope(buffer);
        scope.Run(CloseStage.RootLoadedBefore, CloseStage.RootLoadedAfter,
            () => buffer.Mark(CloseStage.DisposeEntry));
        foreach (var stage in new[] { CloseStage.RootLoadedBefore, CloseStage.DisposeEntry, CloseStage.RootLoadedAfter })
        {
            Assert.True(buffer.TryRead(out var value));
            Assert.Equal(stage, value.Stage);
        }
    }

    [Fact]
    public void ConcurrentPublicationPreservesRecordsAcrossSlotReuse()
    {
        var buffer = new CloseStageBuffer(Environment.CurrentManagedThreadId);
        using var barrier = new Barrier(2);
        Exception? failure = null;
        var consumer = new Thread(() =>
        {
            try
            {
                long prior = 0;
                for (int batch = 0; batch < 20; batch++)
                {
                    Assert.True(barrier.SignalAndWait(TimeSpan.FromSeconds(5)));
                    for (int i = 0; i < 32; i++)
                    {
                        CloseStageRecord value = default;
                        Assert.True(SpinWait.SpinUntil(() => buffer.TryRead(out value), TimeSpan.FromSeconds(5)));
                        Assert.Equal(batch * 32 + i + 1, value.Sequence);
                        Assert.Equal(i % 2 == 0 ? CloseStage.CloseEntry : CloseStage.DisposeEntry, value.Stage);
                        Assert.True(value.Ticks >= prior);
                        prior = value.Ticks;
                    }
                    Assert.True(barrier.SignalAndWait(TimeSpan.FromSeconds(5)));
                }
            }
            catch (Exception e) { failure = e; }
        }) { IsBackground = true };
        consumer.Start();
        try
        {
            for (int batch = 0; batch < 20; batch++)
            {
                Assert.True(barrier.SignalAndWait(TimeSpan.FromSeconds(5)));
                for (int i = 0; i < 32; i++) buffer.Mark(i % 2 == 0 ? CloseStage.CloseEntry : CloseStage.DisposeEntry);
                Assert.True(barrier.SignalAndWait(TimeSpan.FromSeconds(5)));
            }
        }
        finally { consumer.Join(TimeSpan.FromSeconds(6)); }
        Assert.Null(failure);
        Assert.Equal(0, buffer.Dropped);
    }
}
