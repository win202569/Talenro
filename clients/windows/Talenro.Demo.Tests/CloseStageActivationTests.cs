using Talenro.Demo.App.Layout.CloseStages;
using Xunit;

namespace Talenro.Demo.Tests;

public class CloseStageActivationTests
{
    private const string Id = "11111111111111111111111111111111";

    [Fact]
    public void ParseAcceptsOneRunAndPreservesLayoutArguments()
    {
        var result = CloseStageActivation.Parse(["--layout-test", "--layout-close-run=" + Id, "--layout-profile=baseline"]);
        Assert.Null(result.ErrorCode);
        Assert.Equal(Guid.ParseExact(Id, "N"), result.RunId);
        Assert.Equal(new[] { "--layout-test", "--layout-profile=baseline" }, result.LayoutArguments);
    }

    [Theory]
    [InlineData("missing")]
    [InlineData("duplicate")]
    [InlineData("empty")]
    [InlineData("upper-flag")]
    [InlineData("upper-value")]
    [InlineData("bad-guid")]
    [InlineData("no-layout")]
    public void RejectsDuplicateOrMalformedOrMissingLayout(string fault)
    {
        string[] args = fault switch
        {
            "missing" => ["--layout-test"],
            "duplicate" => ["--layout-test", "--layout-close-run=" + Id, "--layout-close-run=" + Id],
            "empty" => ["--layout-test", "--layout-close-run=" + new string('0', 32)],
            "upper-flag" => ["--layout-test", "--LAYOUT-CLOSE-RUN=" + Id],
            "upper-value" => ["--layout-test", "--layout-close-run=AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"],
            "bad-guid" => ["--layout-test", "--layout-close-run=../elsewhere"],
            _ => ["--layout-close-run=" + Id]
        };
        var result = CloseStageActivation.Parse(args);
        Assert.NotNull(result.ErrorCode);
        Assert.Null(result.RunId);
    }

    [Fact]
    public void CreatesNewOnlyUnderTrustedRoot()
    {
        string root = Path.Combine(Path.GetTempPath(), "talenro-close-test-" + Guid.NewGuid().ToString("N"));
        try
        {
            using (var stream = CloseStageActivation.OpenNew(root, Guid.ParseExact(Id, "N"))) stream.WriteByte(42);
            string file = Path.Combine(root, Id + ".jsonl");
            Assert.Equal(new byte[] { 42 }, File.ReadAllBytes(file));
            Assert.Throws<IOException>(() => CloseStageActivation.OpenNew(root, Guid.ParseExact(Id, "N")));
            Assert.Equal(new byte[] { 42 }, File.ReadAllBytes(file));
            Assert.Single(Directory.GetFiles(root));
        }
        finally
        {
            if (Directory.Exists(root))
            {
                foreach (var file in Directory.GetFiles(root)) File.Delete(file);
                Directory.Delete(root);
            }
        }
    }

    [Fact]
    public void RootContainmentRejectsRelativeAndEmptyRunBeforeCreatingAnything()
    {
        Assert.Throws<ArgumentException>(() => CloseStageActivation.OpenNew("../escape", Guid.ParseExact(Id, "N")));
        Assert.Throws<ArgumentException>(() => CloseStageActivation.OpenNew(Path.GetTempPath(), Guid.Empty));
    }

    [Fact]
    public void RejectsReparseAncestors()
    {
        // Existing Windows compatibility junction; never create or change it.
        string junction = Path.Combine(Path.GetPathRoot(Environment.SystemDirectory)!, "Users", "All Users");
        Assert.True((File.GetAttributes(junction) & FileAttributes.ReparsePoint) != 0);
        Assert.Throws<IOException>(() => CloseStageActivation.OpenNew(Path.Combine(junction, "TalenroNeverCreate"), Guid.ParseExact(Id, "N")));
    }

    [Theory]
    [InlineData(0)] [InlineData(1)] [InlineData(2)] [InlineData(3)] [InlineData(4)]
    [InlineData(5)] [InlineData(6)] [InlineData(7)] [InlineData(8)]
    public void EachFailingRemovalOmitsItsAfterMarker(int failureIndex)
    {
        var pairs = new[] {
            (CloseStage.RootLoadedBefore, CloseStage.RootLoadedAfter),
            (CloseStage.RootUnloadedBefore, CloseStage.RootUnloadedAfter),
            (CloseStage.RootSizeBefore, CloseStage.RootSizeAfter),
            (CloseStage.HostLoadedBefore, CloseStage.HostLoadedAfter),
            (CloseStage.HostUnloadedBefore, CloseStage.HostUnloadedAfter),
            (CloseStage.HostSizeBefore, CloseStage.HostSizeAfter),
            (CloseStage.WindowClosedBefore, CloseStage.WindowClosedAfter),
            (CloseStage.XamlRootBefore, CloseStage.XamlRootAfter),
            (CloseStage.UISettingsBefore, CloseStage.UISettingsAfter) };
        var buffer = new CloseStageBuffer(Environment.CurrentManagedThreadId);
        var scope = new CloseStageScope(buffer);
        var sentinel = new InvalidOperationException();
        Assert.Same(sentinel, Assert.Throws<InvalidOperationException>(() =>
        {
            for (int i = 0; i < pairs.Length; i++)
            {
                int current = i;
                scope.Run(pairs[i].Item1, pairs[i].Item2, () => { if (current == failureIndex) throw sentinel; });
            }
        }));
        var records = new List<CloseStageRecord>();
        while (buffer.TryRead(out var record)) records.Add(record);
        Assert.Equal(failureIndex * 2 + 1, records.Count);
        Assert.Equal(pairs[failureIndex].Item1, records[^1].Stage);
        Assert.DoesNotContain(records, r => r.Stage == pairs[failureIndex].Item2);
    }
}
