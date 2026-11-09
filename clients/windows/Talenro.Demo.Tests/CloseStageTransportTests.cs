using System.Text;
using Talenro.Demo.App.Layout.CloseStages;
using Xunit;

namespace Talenro.Demo.Tests;

public class CloseStageTransportTests
{
    private static readonly Guid Run = Guid.Parse("11111111-1111-1111-1111-111111111111");
    private const string ReadyLine = "{\"Type\":\"ready\",\"Schema\":1,\"Run\":\"11111111111111111111111111111111\",\"Frequency\":1000}\n";
    private const string FirstLine = "{\"Type\":\"stage\",\"Schema\":1,\"Run\":\"11111111111111111111111111111111\",\"Sequence\":1,\"Stage\":1,\"Ticks\":10}\n";
    private static CloseStageReceipt Read(string text) => CloseStageReceiptReader.Read(new MemoryStream(Encoding.UTF8.GetBytes(text)), Run);

    [Fact]
    public void BeforeWithoutAfterIsOnlyAValidPrefix()
    {
        var result = Read(ReadyLine + FirstLine);
        Assert.Equal(CloseStageReceiptStatus.ValidPrefix, result.Status);
        Assert.Single(result.Records);
        Assert.False(result.CompletedClose);
    }

    [Theory]
    [InlineData("truncated")]
    [InlineData("wrong-run")]
    [InlineData("duplicate")]
    [InlineData("gap")]
    [InlineData("backward")]
    [InlineData("unknown")]
    [InlineData("frequency")]
    [InlineData("order")]
    [InlineData("drop")]
    [InlineData("failure")]
    [InlineData("missing-ready")]
    [InlineData("oversize")]
    [InlineData("too-many-lines")]
    public void RejectsUntrustworthyReceipts(string fault)
    {
        string second = "{\"Type\":\"stage\",\"Schema\":1,\"Run\":\"11111111111111111111111111111111\",\"Sequence\":2,\"Stage\":2,\"Ticks\":11}\n";
        string text = ReadyLine + FirstLine + second;
        text = fault switch
        {
            "truncated" => text[..^1],
            "wrong-run" => text.Replace("11111111111111111111111111111111", "22222222222222222222222222222222"),
            "duplicate" => ReadyLine + FirstLine + FirstLine,
            "gap" => text.Replace("\"Sequence\":2", "\"Sequence\":3"),
            "backward" => text.Replace("\"Ticks\":11", "\"Ticks\":9"),
            "unknown" => text.Replace("\"Stage\":2", "\"Stage\":99"),
            "frequency" => text.Replace("\"Frequency\":1000", "\"Frequency\":0"),
            "order" => text.Replace("\"Stage\":2", "\"Stage\":27"),
            "drop" => text + "{\"Type\":\"health\",\"Schema\":1,\"Run\":\"11111111111111111111111111111111\",\"Dropped\":1}\n",
            "failure" => text + "{\"Type\":\"health\",\"Schema\":1,\"Run\":\"11111111111111111111111111111111\",\"Failed\":true}\n",
            "missing-ready" => FirstLine,
            "oversize" => ReadyLine + new string(' ', 65536),
            "too-many-lines" => ReadyLine + string.Concat(Enumerable.Repeat(FirstLine, 128)),
            _ => text
        };
        Assert.Equal(CloseStageReceiptStatus.Inconclusive, Read(text).Status);
        Assert.False(Read(text).CompletedClose);
    }

    [Fact]
    public void MissingFinalHealthCannotClaimCompletedClose()
    {
        const string tail = "{\"Type\":\"stage\",\"Schema\":1,\"Run\":\"11111111111111111111111111111111\",\"Sequence\":2,\"Stage\":2,\"Ticks\":11}\n" +
            "{\"Type\":\"stage\",\"Schema\":1,\"Run\":\"11111111111111111111111111111111\",\"Sequence\":3,\"Stage\":3,\"Ticks\":12}\n" +
            "{\"Type\":\"stage\",\"Schema\":1,\"Run\":\"11111111111111111111111111111111\",\"Sequence\":4,\"Stage\":27,\"Ticks\":13}\n";
        Assert.Equal(CloseStageReceiptStatus.Inconclusive, Read(ReadyLine + FirstLine + tail).Status);
        Assert.False(Read(ReadyLine + FirstLine + tail).CompletedClose);
    }

    [Fact]
    public void ReadyPrecedesStagesAndStageLinesRoundTrip()
    {
        var buffer = new CloseStageBuffer(Environment.CurrentManagedThreadId);
        var stream = new CapturingStream();
        var session = new CloseStageSession(Run, buffer, stream);
        session.Start();
        Assert.True(session.WaitUntilReady(TimeSpan.FromSeconds(3)));
        Assert.True(session.IsBackground);
        foreach (var stage in new[] { CloseStage.CloseEntry, CloseStage.DisposeEntry, CloseStage.AlreadyDisposedReturn, CloseStage.CloseReturn }) buffer.Mark(stage);
        Assert.True(SpinWait.SpinUntil(() => session.WorkerCompleted, TimeSpan.FromSeconds(3)));
        Assert.False(session.Failed);
        var result = CloseStageReceiptReader.Read(new MemoryStream(stream.Bytes), Run);
        Assert.Equal(CloseStageReceiptStatus.ValidPrefix, result.Status);
        Assert.True(result.CompletedClose);
        Assert.Equal(4, result.Records.Count);
        Assert.StartsWith("{\"Type\":\"ready\"", Encoding.UTF8.GetString(stream.Bytes));
    }

    [Fact]
    public void BlockedWriterDoesNotBlockMark()
    {
        var buffer = new CloseStageBuffer(Environment.CurrentManagedThreadId);
        using var stream = new BlockingStream();
        var session = new CloseStageSession(Run, buffer, stream);
        session.Start();
        try
        {
            Assert.True(stream.Entered.Wait(TimeSpan.FromSeconds(3)));
            buffer.Mark(CloseStage.CloseEntry);
            buffer.Mark(CloseStage.DisposeEntry);
            buffer.Mark(CloseStage.AlreadyDisposedReturn);
            buffer.Mark(CloseStage.CloseReturn);
            Assert.False(session.WorkerCompleted);
            Assert.Equal(0, buffer.Dropped);
        }
        finally { stream.Release.Set(); SpinWait.SpinUntil(() => session.WorkerCompleted, TimeSpan.FromSeconds(3)); }
    }

    [Theory]
    [InlineData(false, false)] [InlineData(false, true)] [InlineData(true, false)] [InlineData(true, true)]
    public void NormalAndNullConditionalRemovalTracesComplete(bool noRoot, bool noSettings)
    {
        // Literal stage fixture, not a claim that live WinUI event accessors ran.
        var stages = new List<int> { 1, 2, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17 };
        stages.AddRange(noRoot ? [20] : [18, 19]);
        stages.AddRange(noSettings ? [23] : [21, 22]);
        stages.AddRange([24, 25, 26, 27]);
        var text = new StringBuilder(ReadyLine);
        for (int i = 0; i < stages.Count; i++)
            text.Append($"{{\"Type\":\"stage\",\"Schema\":1,\"Run\":\"11111111111111111111111111111111\",\"Sequence\":{i + 1},\"Stage\":{stages[i]},\"Ticks\":{i + 10}}}\n");
        text.Append("{\"Type\":\"health\",\"Schema\":1,\"Run\":\"11111111111111111111111111111111\",\"Dropped\":0,\"InvalidProducer\":false,\"Failed\":false}\n");
        Assert.True(Read(text.ToString()).CompletedClose);
        Assert.False(Read(text.ToString().Replace("\"Dropped\":0", "\"Dropped\":1")).CompletedClose);
        Assert.False(Read(text.ToString().Replace("\"Failed\":false", "\"Failed\":true")).CompletedClose);
    }

    [Fact]
    public void FailureStreamDisablesRecordingWithoutThrowingOntoProducer()
    {
        var buffer = new CloseStageBuffer(Environment.CurrentManagedThreadId);
        var session = new CloseStageSession(Run, buffer, new FailureStream());
        session.Start();
        Assert.True(SpinWait.SpinUntil(() => session.WorkerCompleted, TimeSpan.FromSeconds(3)));
        Assert.True(session.Failed);
        Assert.True(buffer.Disabled);
        buffer.Mark(CloseStage.CloseEntry);
        Assert.False(buffer.TryRead(out _));
    }

    private sealed class CapturingStream : MemoryStream
    {
        public byte[] Bytes { get; private set; } = [];
        protected override void Dispose(bool disposing) { if (disposing) Bytes = ToArray(); base.Dispose(disposing); }
    }
    private sealed class FailureStream : MemoryStream
    {
        public override void Write(byte[] buffer, int offset, int count) => throw new IOException("test failure");
    }
    private sealed class BlockingStream : MemoryStream
    {
        public readonly ManualResetEventSlim Entered = new();
        public readonly ManualResetEventSlim Release = new();
        public override void Write(byte[] buffer, int offset, int count)
        {
            Entered.Set();
            if (!Release.Wait(TimeSpan.FromSeconds(5))) throw new IOException("test timeout");
            base.Write(buffer, offset, count);
        }
    }
}
