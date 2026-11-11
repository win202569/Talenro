using Talenro.Demo.App.Layout.EmailGeometry;
using Xunit;

namespace Talenro.Demo.Tests;

public class EmailGeometryActivationTests
{
    private const string Run = "11111111111141118111111111111111";
    private static string[] Valid => ["--layout-test", "--layout-language=en-US", "--layout-profile=rtl-mixed", "--layout-email-geometry-run=" + Run];

    [Fact]
    public void FixedArgumentsPreserveIdentityAndLayoutParserInput()
    {
        var result = GeometryActivation.Parse(Valid.Reverse().ToArray());
        Assert.Null(result.ErrorCode);
        Assert.Equal(Guid.ParseExact(Run, "N"), result.RunId);
        Assert.Equal(new[] { "--layout-profile=rtl-mixed", "--layout-language=en-US", "--layout-test" }, result.LayoutArguments);
    }

    [Theory]
    [InlineData("")]
    [InlineData("00000000000000000000000000000000")]
    [InlineData("AAAAAAAAAAAA4AAA8AAAAAAAAAAAAAAA")]
    [InlineData("11111111-1111-4111-8111-111111111111")]
    [InlineData("11111111111141118111111111111111 ")]
    public void NoncanonicalRunRejectsWithoutLayoutArguments(string value)
    {
        var args = Valid;
        args[3] = "--layout-email-geometry-run=" + value;
        Reject(args);
    }

    [Fact]
    public void MissingExtraDuplicateWrongProfileAndCloseModeReject()
    {
        Reject([]);
        Reject(Valid[..3]);
        Reject([.. Valid, "--extra"]);
        Reject([.. Valid, "--close-stage-run=" + Run]);
        Reject([Valid[0], Valid[0], Valid[2], Valid[3]]);
        Reject([Valid[0], "--layout-language=zh-Hans", Valid[2], Valid[3]]);
        Reject([Valid[0], Valid[1], "--layout-profile=baseline", Valid[3]]);
        Reject([Valid[0], Valid[1], Valid[3], Valid[3]]);
    }

    private static void Reject(string[] args)
    {
        var result = GeometryActivation.Parse(args);
        Assert.NotNull(result.ErrorCode);
        Assert.Null(result.RunId);
        Assert.Empty(result.LayoutArguments);
    }
}
