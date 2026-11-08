using System.Globalization;
using Talenro.Demo.App.Layout;
using Xunit;

namespace Talenro.Demo.Tests;

public class LayoutMeasurementTests
{
    [Fact]
    public void ReadoutSeparatesActualRootAndPageAndIsCultureInvariant()
    {
        var previous = CultureInfo.CurrentCulture;
        try
        {
            CultureInfo.CurrentCulture = CultureInfo.GetCultureInfo("fr-FR");
            Assert.Equal("LAYOUT TEST en-US/long-status inactive | root=1024.125x768.000 DIP | page=804.125x650.500 DIP | render=1.250 text=1.000",
                LayoutTitle.Format(new("en-US", "long-status"), false, new(1024.125, 768, 804.125, 650.5, 1.25, 1)));
            Assert.Contains(" active |", LayoutTitle.Format(new("zh-Hans", "baseline"), true, new(1, 2, 3, 4, 1, 1)));
        }
        finally { CultureInfo.CurrentCulture = previous; }
    }

    [Theory]
    [InlineData(null)]
    [InlineData(0d)]
    [InlineData(-1d)]
    [InlineData(double.NaN)]
    [InlineData(double.PositiveInfinity)]
    [InlineData(double.NegativeInfinity)]
    public void InvalidReadingsRemainUnknown(double? invalid)
    {
        Assert.Equal("LAYOUT TEST zh-Hans/baseline active | root=unknownxunknown DIP | page=unknownxunknown DIP | render=unknown text=unknown",
            LayoutTitle.Format(new("zh-Hans", "baseline"), true, new(invalid, invalid, invalid, invalid, invalid, invalid)));
    }

    [Fact]
    public void TitleDoesNotDecideTargetAcceptance()
    {
        var title = LayoutTitle.Format(new("en-US", "baseline"), true, new(1023.9998, 768, 800, 600, 1, 1));
        Assert.Contains("root=1024.000x768.000 DIP", title);
        Assert.DoesNotContain("PASS", title);
        Assert.DoesNotContain("accepted", title);
    }

    [Fact]
    public void ObserverLifecycleHasNoTimerOrModelDispatch()
    {
        var path = Path.Combine(AppContext.BaseDirectory, "AppSources/Diagnostics/LayoutWindowObserver.cs");
        var code = File.Exists(path) ? File.ReadAllText(path) : "";
        foreach (var marker in new[] { "root.Loaded +=", "root.SizeChanged +=", "pageHost.SizeChanged +=", "Changed +=", "window.Closed +=",
            "root.Loaded -=", "root.SizeChanged -=", "pageHost.SizeChanged -=", "Changed -=", "window.Closed -=", "disposed", "TryEnqueue" })
            Assert.Contains(marker, code);
        foreach (var forbidden in new[] { "Timer", "UpdateLayout(", ".Dispatch(", "ActualWidth =", "ActualHeight =" })
            Assert.DoesNotContain(forbidden, code);
    }
}
