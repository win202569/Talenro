using System.Text;
using System.Text.Json;
using Talenro.Demo.App.Layout.EmailGeometry;
using Xunit;

namespace Talenro.Demo.Tests;

public class EmailGeometryReceiptTests
{
    private static readonly Guid Run = Guid.ParseExact("11111111111141118111111111111111", "N");
    private static Dictionary<string, GeometryStatus> ScaleStatuses() => new()
    {
        ["RenderingScale"] = GeometryStatus.Ok, ["TextScale"] = GeometryStatus.Ok
    };
    private static GeometrySnapshot Sample(int sequence = 1) => new(Run, sequence,
        sequence <= 2 ? "Login" : "Account", sequence, 1, 1,
        new(0, 0, 100, 100, new Dictionary<string, GeometryStatus>
        {
            ["HorizontalOffset"] = GeometryStatus.Ok, ["VerticalOffset"] = GeometryStatus.Ok,
            ["ViewportWidth"] = GeometryStatus.Ok, ["ViewportHeight"] = GeometryStatus.Ok
        }), [], [new("root", new(GeometryStatus.Ok, new("RootDIP", 0, 0, 100, 100)), true)], false, ScaleStatuses());

    [Fact]
    public void AddIsMemoryOnlyAndFinishWritesCorrelatedBoundedRecordsOnce()
    {
        var stream = new TrackingStream();
        var session = new GeometryReceiptSession(Run, stream);
        for (var i = 1; i <= 4; i++) session.Add(Sample(i));
        Assert.Equal(0, stream.Length);
        var result = session.Finish();
        Assert.True(result.Completed);
        Assert.Equal(4, result.Samples);
        Assert.Equal(1, stream.Disposals);
        Assert.True(stream.Flushes > 0);
        var bytes = stream.ToArray();
        Assert.False(bytes.Take(3).SequenceEqual(new byte[] { 239, 187, 191 }));
        var lines = Encoding.UTF8.GetString(bytes).Split('\n', StringSplitOptions.RemoveEmptyEntries);
        Assert.Equal(6, lines.Length);
        for (var i = 0; i < lines.Length; i++)
        {
            using var json = JsonDocument.Parse(lines[i]);
            Assert.Equal(1, json.RootElement.GetProperty("Schema").GetInt32());
            Assert.Equal(Run.ToString("N"), json.RootElement.GetProperty("Run").GetString());
            Assert.Equal(i == 0 ? "header" : i == 5 ? "health" : "snapshot", json.RootElement.GetProperty("Type").GetString());
            if (i is >= 1 and <= 4) Assert.Equal(i, json.RootElement.GetProperty("Data").GetProperty("Sequence").GetInt32());
        }
        Assert.Same(result, session.Finish());
        session.Dispose();
        Assert.Equal(bytes, stream.ToArray());
        Assert.Equal(1, stream.Disposals);
    }

    [Fact]
    public void ZeroSamplesAreExplicitNotDiagnosisSuccess()
    {
        var stream = new TrackingStream();
        var session = new GeometryReceiptSession(Run, stream);
        var result = session.Finish();
        Assert.True(result.Completed); // transport only, not diagnostic acceptance
        Assert.Equal(0, result.Samples);
        Assert.Contains("\"Samples\":0", Encoding.UTF8.GetString(stream.ToArray()));
        Assert.Contains("\"Provisional\":true", Encoding.UTF8.GetString(stream.ToArray()));
    }

    [Fact]
    public void ExplicitBoundaryRolesSurviveReceiptValidation()
    {
        var sample = Sample();
        var session = new GeometryReceiptSession(Run, new MemoryStream());
        session.Add(sample with { Boundaries = [sample.Boundaries[0], sample.Boundaries[0] with { Role = "root:clip" },
            sample.Boundaries[0] with { Role = "presenter:clip" }] });
        Assert.True(session.Finish().Completed);
    }

    [Theory]
    [InlineData("write", "write-failed")]
    [InlineData("flush", "flush-failed")]
    [InlineData("dispose", "dispose-failed")]
    [InlineData("construct", "stream-invalid")]
    public void StreamFailuresNeverCompleteAndAlwaysAttemptOwnedCleanup(string fault, string code)
    {
        var stream = new TrackingStream { Fault = fault };
        var session = new GeometryReceiptSession(Run, stream);
        session.Add(Sample());
        var result = session.Finish();
        Assert.False(result.Completed);
        Assert.Contains(code, result.ErrorCodes);
        Assert.Equal(1, stream.Disposals);
        session.Dispose();
        Assert.Same(result, session.Finish());
        Assert.Equal(1, stream.Disposals);
        Assert.DoesNotContain("private failure", Encoding.UTF8.GetString(stream.ToArray()));
    }

    [Fact]
    public void PrimaryWriteAndCleanupErrorsAreBothRetainedInOrder()
    {
        var stream = new TrackingStream { Fault = "write+dispose" };
        var session = new GeometryReceiptSession(Run, stream);
        session.Add(Sample());
        var result = session.Finish();
        Assert.False(result.Completed);
        Assert.Equal(new[] { "write-failed", "dispose-failed" }, result.ErrorCodes);
        Assert.Equal(1, stream.Disposals);
    }

    [Fact]
    public void DisposeWithoutFinishWritesNothingAndCannotLaterClaimCompletion()
    {
        var stream = new TrackingStream();
        var session = new GeometryReceiptSession(Run, stream);
        session.Add(Sample());
        session.Dispose();
        session.Dispose();
        Assert.Empty(stream.ToArray());
        Assert.False(session.Finish().Completed);
        Assert.Equal(1, stream.Disposals);
    }

    [Fact]
    public void WrongThreadDisablesFurtherCapture()
    {
        var stream = new TrackingStream();
        var session = new GeometryReceiptSession(Run, stream);
        var thread = new Thread(() => session.Add(Sample()));
        thread.Start();
        Assert.True(thread.Join(TimeSpan.FromSeconds(5)));
        session.Add(Sample());
        var result = session.Finish();
        Assert.False(result.Completed);
        Assert.Equal(0, result.Samples);
        Assert.Contains("wrong-thread", result.ErrorCodes);
    }

    [Theory]
    [InlineData("run")]
    [InlineData("sequence")]
    [InlineData("page")]
    [InlineData("nonfinite")]
    [InlineData("missing-status")]
    [InlineData("extra-status")]
    [InlineData("null-ok")]
    [InlineData("elements-cap")]
    [InlineData("boundaries-cap")]
    [InlineData("private-role")]
    [InlineData("invalid-enum")]
    public void InvalidSnapshotsDisableCaptureWithoutPublishingPrivatePayload(string invalid)
    {
        var sample = Sample();
        sample = invalid switch
        {
            "run" => sample with { RunId = Guid.Empty },
            "sequence" => sample with { Sequence = 2 },
            "page" => sample with { Page = "private@example.com" },
            "nonfinite" => sample with { ElapsedMs = double.NaN },
            "missing-status" => sample with { FieldStatuses = new Dictionary<string, GeometryStatus>() },
            "extra-status" => sample with { FieldStatuses = new Dictionary<string, GeometryStatus>(ScaleStatuses()) { ["private@example.com"] = GeometryStatus.Ok } },
            "null-ok" => sample with { TextScale = null },
            "elements-cap" => sample with { Elements = new GeometryElement[17] },
            "boundaries-cap" => sample with { Boundaries = Enumerable.Repeat(sample.Boundaries[0], 17).ToArray() },
            "private-role" => sample with { Boundaries = [sample.Boundaries[0] with { Role = "private@example.com" }] },
            "invalid-enum" => sample with { FieldStatuses = new Dictionary<string, GeometryStatus>(ScaleStatuses()) { ["TextScale"] = (GeometryStatus)99 } },
            _ => throw new InvalidOperationException()
        };
        var stream = new TrackingStream();
        var session = new GeometryReceiptSession(Run, stream);
        session.Add(sample);
        session.Add(Sample());
        var result = session.Finish();
        Assert.False(result.Completed);
        Assert.Equal(0, result.Samples);
        Assert.DoesNotContain("private@example.com", Encoding.UTF8.GetString(stream.ToArray()));
    }

    [Fact]
    public void SnapshotsAreFrozenAgainstCallerMutation()
    {
        var statuses = ScaleStatuses();
        var boundaries = new List<GeometryBoundary>(Sample().Boundaries);
        var stream = new TrackingStream();
        var session = new GeometryReceiptSession(Run, stream);
        session.Add(Sample() with { FieldStatuses = statuses, Boundaries = boundaries });
        statuses["private@example.com"] = GeometryStatus.Ok;
        boundaries.Clear();
        Assert.True(session.Finish().Completed);
        var text = Encoding.UTF8.GetString(stream.ToArray());
        Assert.DoesNotContain("private@example.com", text);
        Assert.Contains("\"Role\":\"root\"", text);
    }

    [Fact]
    public void FifthSnapshotAndThirdPageSampleFailClosed()
    {
        foreach (var extraPage in new[] { false, true })
        {
            var session = new GeometryReceiptSession(Run, new TrackingStream());
            for (var i = 1; i <= (extraPage ? 2 : 4); i++) session.Add(Sample(i));
            session.Add(extraPage ? Sample(3) with { Page = "Login" } : Sample(5));
            Assert.False(session.Finish().Completed);
        }
    }

    [Fact]
    public void InvalidConstructorIdentityStillReleasesOwnedStream()
    {
        var stream = new TrackingStream();
        var session = new GeometryReceiptSession(Guid.Empty, stream);
        Assert.False(session.Finish().Completed);
        Assert.Equal(1, stream.Disposals);
    }

    [Fact]
    public void SerializationFailureStillDisposesAndDoesNotWritePartialRecords()
    {
        var stream = new TrackingStream { Fault = "dispose" };
        var session = new GeometryReceiptSession(Run, stream, _ => throw new InvalidOperationException("private failure"));
        session.Add(Sample());
        var result = session.Finish();
        Assert.False(result.Completed);
        Assert.Equal(new[] { "serialize-failed", "dispose-failed" }, result.ErrorCodes);
        Assert.Empty(stream.ToArray());
        Assert.Equal(1, stream.Disposals);
    }

    [Fact]
    public void FinishFromWrongThreadCannotClaimCompletion()
    {
        var stream = new TrackingStream();
        var session = new GeometryReceiptSession(Run, stream);
        session.Add(Sample());
        GeometryFinishResult? result = null;
        var thread = new Thread(() => result = session.Finish());
        thread.Start();
        Assert.True(thread.Join(TimeSpan.FromSeconds(5)));
        Assert.NotNull(result);
        Assert.False(result.Completed);
        Assert.Contains("wrong-thread", result.ErrorCodes);
        Assert.Equal(1, stream.Disposals);
    }

    [Fact]
    public void FactoryCreatesOnlyNewRunAndNeverOverwrites()
    {
        var root = Path.Combine(Path.GetTempPath(), "email-geometry-test-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(root);
        try
        {
            using (var stream = GeometryReceiptSession.OpenNew(root, Run)) stream.WriteByte(42);
            var file = Path.Combine(root, Run.ToString("N"), "geometry.jsonl");
            Assert.ThrowsAny<IOException>(() => GeometryReceiptSession.OpenNew(root, Run));
            Assert.Equal(new byte[] { 42 }, File.ReadAllBytes(file));
            Assert.ThrowsAny<ArgumentException>(() => GeometryReceiptSession.OpenNew(root, Guid.Empty));
            Assert.False(Directory.Exists(Path.Combine(root, Guid.Empty.ToString("N"))));
            Assert.ThrowsAny<ArgumentException>(() => GeometryReceiptSession.OpenNew("relative-root", Run));
        }
        finally { Directory.Delete(root, true); }
    }

    private static GeometryElement Element() => new("email", GeometryStatus.Ok, true, true,
        new(GeometryStatus.Ok, new("RootDIP", 10, 10, 20, 20)),
        new(GeometryStatus.Ok, new("ParentDIP:panel", 0, 0, 10, 10)),
        10, 10, 10, 10, "LeftToRight", "Stretch", "Left", new(0, 0, 0, 0), null, null, null,
        new Dictionary<string, GeometryStatus>
        {
            ["Loaded"] = GeometryStatus.Ok, ["Visible"] = GeometryStatus.Ok,
            ["ActualWidth"] = GeometryStatus.Ok, ["ActualHeight"] = GeometryStatus.Ok,
            ["DesiredWidth"] = GeometryStatus.Ok, ["DesiredHeight"] = GeometryStatus.Ok,
            ["FlowDirection"] = GeometryStatus.Ok, ["HorizontalAlignment"] = GeometryStatus.Ok,
            ["TextAlignment"] = GeometryStatus.Ok, ["Margin"] = GeometryStatus.Ok,
            ["Padding"] = GeometryStatus.Unsupported, ["Width"] = GeometryStatus.Unsupported,
            ["MaxWidth"] = GeometryStatus.Unsupported
        });

    [Fact]
    public void ElementPropertiesAndUnsupportedAutoConstraintsAreRetained()
    {
        var stream = new TrackingStream();
        var session = new GeometryReceiptSession(Run, stream);
        session.Add(Sample() with { Elements = [Element()] });
        Assert.True(session.Finish().Completed);
        Assert.Contains("ParentDIP:panel", Encoding.UTF8.GetString(stream.ToArray()));
    }

    [Theory]
    [InlineData("root-space")]
    [InlineData("slot-space")]
    [InlineData("flow")]
    [InlineData("margin")]
    [InlineData("negative-size")]
    public void InvalidElementFieldsCannotEnterReceipt(string fault)
    {
        var e = Element();
        e = fault switch
        {
            "root-space" => e with { RootBounds = e.LayoutSlot },
            "slot-space" => e with { LayoutSlot = e.RootBounds },
            "flow" => e with { FlowDirection = "private@example.com" },
            "margin" => e with { Margin = new(double.NaN, 0, 0, 0) },
            "negative-size" => e with { ActualWidth = -1 },
            _ => throw new InvalidOperationException()
        };
        var stream = new TrackingStream();
        var session = new GeometryReceiptSession(Run, stream);
        session.Add(Sample() with { Elements = [e] });
        Assert.False(session.Finish().Completed);
        Assert.DoesNotContain("private@example.com", Encoding.UTF8.GetString(stream.ToArray()));
    }

    [Fact]
    public void FactoryRejectsExistingDirectoryAndReparseMetadata()
    {
        var root = Path.Combine(Path.GetTempPath(), "email-geometry-test-" + Guid.NewGuid().ToString("N"));
        Directory.CreateDirectory(Path.Combine(root, Run.ToString("N")));
        try
        {
            Assert.ThrowsAny<IOException>(() => GeometryReceiptSession.OpenNew(root, Run));
            Assert.ThrowsAny<IOException>(() => GeometryReceiptSession.OpenNew(root, Guid.NewGuid(),
                _ => FileAttributes.Directory | FileAttributes.ReparsePoint));
            Assert.Empty(Directory.GetFiles(root, "*", SearchOption.AllDirectories));
        }
        finally { Directory.Delete(root, true); }
    }

    private sealed class TrackingStream : MemoryStream
    {
        public string Fault { get; init; } = "";
        public int Disposals { get; private set; }
        public int Flushes { get; private set; }
        public override bool CanWrite => Fault == "construct" ? throw new IOException("private failure") : base.CanWrite;
        public override void Write(byte[] buffer, int offset, int count)
        {
            if (Fault.StartsWith("write", StringComparison.Ordinal)) throw new IOException("private failure");
            base.Write(buffer, offset, count);
        }
        public override void Flush()
        {
            Flushes++;
            if (Fault == "flush") throw new IOException("private failure");
            base.Flush();
        }
        protected override void Dispose(bool disposing)
        {
            Disposals++;
            base.Dispose(disposing);
            if (Fault.Contains("dispose", StringComparison.Ordinal)) throw new IOException("private failure");
        }
    }
}
