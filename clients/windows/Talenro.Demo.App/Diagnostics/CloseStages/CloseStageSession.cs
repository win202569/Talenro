using System.Diagnostics;
using System.Text;
using System.Text.Json;
namespace Talenro.Demo.App.Layout.CloseStages;

internal sealed class CloseStageSession(Guid runId, CloseStageBuffer buffer, Stream ownedStream)
{
    private readonly TaskCompletionSource<bool> startup = new(TaskCreationOptions.RunContinuationsAsynchronously);
    private int started, ready, failed, completed;
    public bool Ready => Volatile.Read(ref ready) != 0;
    public bool Failed => Volatile.Read(ref failed) != 0;
    public bool WorkerCompleted => Volatile.Read(ref completed) != 0;
    private Thread? worker;
    public bool IsBackground => worker?.IsBackground == true;

    public void Start()
    {
        if (Interlocked.Exchange(ref started, 1) != 0) throw new InvalidOperationException("Already started.");
        worker = new Thread(WriteLoop) { IsBackground = true, Name = "Close stage writer" };
        try { worker.Start(); }
        catch { buffer.Disable(); ownedStream.Dispose(); throw; }
    }

    // Startup only. Never invoked from a close callback.
    public bool WaitUntilReady(TimeSpan timeout) => startup.Task.Wait(timeout) && startup.Task.Result;

    private void WriteLoop()
    {
        try
        {
            using var writer = new StreamWriter(ownedStream, new UTF8Encoding(false), 1024);
            string run = runId.ToString("N");
            writer.WriteLine(JsonSerializer.Serialize(new { Type = "ready", Schema = 1, Run = run, Frequency = Stopwatch.Frequency }));
            writer.Flush();
            Volatile.Write(ref ready, 1);
            startup.TrySetResult(true);
            bool terminal = false;
            while (true)
            {
                while (buffer.TryRead(out var record))
                {
                    writer.WriteLine(JsonSerializer.Serialize(new { Type = "stage", Schema = 1, Run = run,
                        record.Sequence, Stage = (int)record.Stage, record.Ticks }));
                    if (record.Stage == CloseStage.CloseReturn) terminal = true;
                }
                writer.Flush();
                if (terminal || buffer.Disabled || buffer.Dropped != 0)
                {
                    writer.WriteLine(JsonSerializer.Serialize(new { Type = "health", Schema = 1, Run = run,
                        buffer.Dropped, buffer.InvalidProducer, Failed = buffer.Disabled }));
                    writer.Flush();
                    break;
                }
                Thread.Sleep(10);
            }
        }
        catch (Exception)
        {
            // Recorder failures never escape onto the application thread, and carry no private payload.
            Volatile.Write(ref failed, 1);
            buffer.Disable();
        }
        finally
        {
            startup.TrySetResult(false);
            Volatile.Write(ref completed, 1);
        }
    }
}
