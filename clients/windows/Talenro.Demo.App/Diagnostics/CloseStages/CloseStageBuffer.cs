using System.Diagnostics;
namespace Talenro.Demo.App.Layout.CloseStages;

internal sealed class CloseStageBuffer(int producerThreadId)
{
    private readonly CloseStageRecord[] slots = new CloseStageRecord[64];
    private int published, consumed, sequence, dropped, invalidProducer, disabled;
    public int Dropped => Volatile.Read(ref dropped);
    public bool InvalidProducer => Volatile.Read(ref invalidProducer) != 0;
    public bool Disabled => Volatile.Read(ref disabled) != 0;
    public void Disable() => Volatile.Write(ref disabled, 1);

    public void Mark(CloseStage stage)
    {
        if (Environment.CurrentManagedThreadId != producerThreadId)
        {
            Volatile.Write(ref invalidProducer, 1);
            Disable();
            return;
        }
        if (Disabled) return;
        if (sequence == int.MaxValue || stage < CloseStage.CloseEntry || stage > CloseStage.CloseReturn)
        {
            Disable();
            return;
        }
        int next = ++sequence;
        if (published - Volatile.Read(ref consumed) == 64)
        {
            Volatile.Write(ref dropped, dropped + 1);
            return;
        }
        slots[published % 64] = new(next, stage, Stopwatch.GetTimestamp());
        // Release publishes the whole slot; the reader acquires before copying it.
        Volatile.Write(ref published, published + 1);
    }

    // Exactly one consumer. Acquire/release also protects a reused slot from overwrite.
    public bool TryRead(out CloseStageRecord record)
    {
        if (consumed == Volatile.Read(ref published)) { record = default; return false; }
        record = slots[consumed % 64];
        Volatile.Write(ref consumed, consumed + 1);
        return true;
    }
}
