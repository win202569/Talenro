using Talenro.Demo.App.Layout.EmailGeometry;
using Xunit;

namespace Talenro.Demo.Tests;

public class EmailGeometryEvidenceTests
{
    private static GeometryField<GeometryRect> Box(double l, double t, double r, double b, string space = "RootDIP") =>
        new(GeometryStatus.Ok, new(space, l, t, r, b));

    [Theory]
    [InlineData(10, 20, 30, 40, GeometryVerdict.Inside, 10, 20, 70, 60)]
    [InlineData(-2, 20, 30, 40, GeometryVerdict.Overflow, -2, 20, 70, 60)]
    [InlineData(10, -1, 30, 40, GeometryVerdict.Overflow, 10, -1, 70, 60)]
    [InlineData(10, 20, 101, 40, GeometryVerdict.Overflow, 10, 20, -1, 60)]
    [InlineData(10, 20, 30, 101, GeometryVerdict.Overflow, 10, 20, 70, -1)]
    [InlineData(0, 0, 100, 100, GeometryVerdict.Inside, 0, 0, 0, 0)]
    public void SignedEdgesDetectEveryOverflow(double l, double t, double r, double b,
        GeometryVerdict verdict, double left, double top, double right, double bottom)
    {
        var result = GeometryEvaluator.Compare(Box(l, t, r, b), new("root", Box(0, 0, 100, 100), true));
        Assert.Equal(verdict, result.Verdict);
        Assert.Equal(left, result.LeftInset);
        Assert.Equal(top, result.TopInset);
        Assert.Equal(right, result.RightInset);
        Assert.Equal(bottom, result.BottomInset);
    }

    [Fact]
    public void NegativeOriginIsNotAnOverflow()
    {
        var result = GeometryEvaluator.Compare(Box(-90, -80, -70, -60), new("root", Box(-100, -100, 0, 0), true));
        Assert.Equal(GeometryVerdict.Inside, result.Verdict);
        Assert.Equal(10, result.LeftInset);
        Assert.Equal(60, result.BottomInset);
    }

    [Fact]
    public void MirroredCornersNormalizeWithoutRounding()
    {
        var result = GeometryEvaluator.Normalize("RootDIP", [new(100, 0), new(80, 0), new(80, 20), new(100, 20)]);
        Assert.Equal(GeometryStatus.Ok, result.Status);
        Assert.Equal(new GeometryRect("RootDIP", 80, 0, 100, 20), result.Value);
        Assert.Equal(GeometryVerdict.Overflow, GeometryEvaluator.Compare(Box(-0.000001, 0, 20, 20),
            new("root", Box(0, 0, 100, 100), true)).Verdict);
    }

    [Theory]
    [InlineData(double.NaN, 0, 10, 10)]
    [InlineData(0, 0, double.PositiveInfinity, 10)]
    [InlineData(10, 0, 10, 10)]
    [InlineData(20, 0, 10, 10)]
    [InlineData(0, 10, 10, 10)]
    public void InvalidBoxesCannotPass(double l, double t, double r, double b)
    {
        Unknown(GeometryEvaluator.Compare(Box(l, t, r, b), new("root", Box(0, 0, 100, 100), true)));
        Unknown(GeometryEvaluator.Compare(Box(1, 1, 2, 2), new("root", Box(l, t, r, b), true)));
    }

    [Theory]
    [InlineData(GeometryStatus.Missing)]
    [InlineData(GeometryStatus.Unloaded)]
    [InlineData(GeometryStatus.Error)]
    [InlineData(GeometryStatus.Unsupported)]
    [InlineData(GeometryStatus.Invalid)]
    public void FailedAcquisitionCannotPass(GeometryStatus status) => Unknown(GeometryEvaluator.Compare(
        new(status, new("RootDIP", 1, 1, 2, 2)), new("root", Box(0, 0, 100, 100), true)));

    [Fact]
    public void MissingIncompleteAndMixedSpaceAreUnknown()
    {
        Unknown(GeometryEvaluator.Compare(new(GeometryStatus.Ok, null), new("root", Box(0, 0, 100, 100), true)));
        Unknown(GeometryEvaluator.Compare(Box(1, 1, 2, 2), new("root", Box(0, 0, 100, 100), false)));
        Unknown(GeometryEvaluator.Compare(Box(1, 1, 2, 2, "ParentDIP:panel"), new("root", Box(0, 0, 100, 100), true)));
    }

    [Fact]
    public void NormalizationRejectsMissingNonfiniteAndDegenerateCorners()
    {
        Assert.Equal(GeometryStatus.Invalid, GeometryEvaluator.Normalize("RootDIP", [new(0, 0), new(1, 0), new(1, 1)]).Status);
        Assert.Equal(GeometryStatus.Invalid, GeometryEvaluator.Normalize("RootDIP", [new(0, 0), new(double.NaN, 0), new(1, 1), new(0, 1)]).Status);
        Assert.Equal(GeometryStatus.Invalid, GeometryEvaluator.Normalize("RootDIP", [new(0, 0), new(0, 0), new(0, 1), new(0, 1)]).Status);
    }

    [Fact]
    public void SnapshotPreservesIncompleteClipChainAlongsideKnownBoundary()
    {
        var boundary = new GeometryBoundary("root", Box(0, 0, 100, 100), true);
        var snapshot = new GeometrySnapshot(Guid.Parse("11111111-1111-4111-8111-111111111111"), 1, "Login", 1,
            null, null, new(null, null, null, null, new Dictionary<string, GeometryStatus>()), [], [boundary], false,
            new Dictionary<string, GeometryStatus>());
        using var json = System.Text.Json.JsonDocument.Parse(System.Text.Json.JsonSerializer.Serialize(snapshot));
        Assert.False(json.RootElement.GetProperty("ClipChainComplete").GetBoolean());
        Assert.True(json.RootElement.GetProperty("Boundaries")[0].GetProperty("Complete").GetBoolean());
        Assert.Equal(GeometryVerdict.Inside, GeometryEvaluator.Compare(Box(1, 1, 2, 2), boundary).Verdict);
    }

    [Fact]
    public void FiniteCoordinatesWhoseDifferenceOverflowsRemainUnknown() => Unknown(GeometryEvaluator.Compare(
        Box(double.MaxValue / 2, 0, double.MaxValue, 10), new("root", Box(-double.MaxValue, 0, -1, 10), true)));

    private static void Unknown(GeometryComparison result)
    {
        Assert.Equal(GeometryVerdict.Unknown, result.Verdict);
        Assert.Null(result.LeftInset);
        Assert.Null(result.TopInset);
        Assert.Null(result.RightInset);
        Assert.Null(result.BottomInset);
    }
}
