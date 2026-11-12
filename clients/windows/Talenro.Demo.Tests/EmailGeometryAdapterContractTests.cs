using System.Security.Cryptography;
using System.Text;
using System.Xml.Linq;
using Talenro.Demo.App.Layout.EmailGeometry;
using Xunit;

namespace Talenro.Demo.Tests;

public class EmailGeometryAdapterContractTests
{
    [Theory]
    [InlineData("root")]
    [InlineData("presenter")]
    public void ExplicitClipsRemainDistinctFromLargerContainingBoxes(string role)
    {
        var box = new GeometryField<GeometryRect>(GeometryStatus.Ok, new("RootDIP", 0, 0, 100, 100));
        var email = new GeometryField<GeometryRect>(GeometryStatus.Ok, new("RootDIP", 10, 10, 30, 30));
        var clip = new GeometryField<GeometryRect>(GeometryStatus.Ok, new("RootDIP", 20, 0, 100, 100));
        var boundaries = new List<GeometryBoundary> { new(role, box, true) };
        GeometryBoundaryCollector.AddExplicitClip(boundaries, role, () => clip);
        Assert.Equal(2, boundaries.Count);
        Assert.Equal(role + ":clip", boundaries[1].Role);
        Assert.Equal(GeometryVerdict.Inside, GeometryEvaluator.Compare(email, boundaries[0]).Verdict);
        Assert.Equal(GeometryVerdict.Overflow, GeometryEvaluator.Compare(email, boundaries[1]).Verdict);
        Assert.Equal(-10, GeometryEvaluator.Compare(email, boundaries[1]).LeftInset);
    }

    [Fact]
    public void ClipErrorsAndUnsupportedMappingsAreUnknownAndCollectionIsBounded()
    {
        var boundaries = new List<GeometryBoundary>();
        GeometryBoundaryCollector.AddExplicitClip(boundaries, "root", () => throw new IOException());
        GeometryBoundaryCollector.AddExplicitClip(boundaries, "presenter", () => new(GeometryStatus.Unsupported, null));
        GeometryBoundaryCollector.AddExplicitClip(boundaries, "page", () => null);
        Assert.Equal(2, boundaries.Count);
        Assert.Equal(GeometryStatus.Error, boundaries[0].Bounds.Status);
        Assert.False(boundaries[0].Complete);
        Assert.False(boundaries[1].Complete);
        for (var i = 1; i <= 14; i++) GeometryBoundaryCollector.AddExplicitClip(boundaries, "ancestor-" + i, () => new(GeometryStatus.Missing, null));
        var called = false;
        GeometryBoundaryCollector.AddExplicitClip(boundaries, "email", () => { called = true; return null; });
        Assert.Equal(16, boundaries.Count);
        Assert.False(called);
    }
    [Theory]
    [InlineData("LoginPage", "A3C830FDC2F8E493B7FFDD5BC67F039D7511DA565161FB2989243424BA0C7FD0", 2)]
    [InlineData("AccountPage", "F01C2618893423F1002CBD3EDE467FF634D821BE7031D45492437ED088D6B57F", 3)]
    public void OnlyApprovedPageChangesArePresent(string page, string baseline, int expectedNames)
    {
        AssertApprovedPage(LoadPage(page), page, baseline, expectedNames);
    }

    private const string AccountBaseline = "F01C2618893423F1002CBD3EDE467FF634D821BE7031D45492437ED088D6B57F";

    private static XDocument LoadPage(string page) =>
        XDocument.Load(Path.Combine(AppContext.BaseDirectory, "AppSources", "Views", page + ".xaml"));

    private static XElement AccountPanel(XDocument doc)
    {
        XNamespace x = "http://schemas.microsoft.com/winfx/2006/xaml";
        var panel = Assert.Single(doc.Descendants(), e => (string?)e.Attribute(x + "Name") == "AccountGeometryPanel");
        Assert.Equal(XName.Get("StackPanel", "http://schemas.microsoft.com/winfx/2006/xaml/presentation"), panel.Name);
        return panel;
    }

    private static void AssertApprovedPage(XDocument source, string page, string baseline, int expectedNames)
    {
        var doc = new XDocument(source);
        if (page == "AccountPage")
        {
            var panel = AccountPanel(doc);
            Assert.Equal("Center", (string?)panel.Attribute("HorizontalAlignment"));
            panel.Attribute("HorizontalAlignment")!.Remove();
        }
        XNamespace x = "http://schemas.microsoft.com/winfx/2006/xaml";
        string[] names = ["LoginGeometryScroll", "LoginGeometryPanel", "AccountGeometryScroll", "AccountGeometryPanel", "AccountGeometryEmail"];
        var added = doc.Descendants().Attributes(x + "Name").Where(a => names.Contains(a.Value)).ToArray();
        Assert.Equal(expectedNames, added.Length);
        foreach (var attr in added) attr.Remove();
        Assert.Equal(baseline, Convert.ToHexString(SHA256.HashData(Encoding.UTF8.GetBytes(doc.ToString(SaveOptions.DisableFormatting)))));
    }

    [Theory]
    [InlineData("missing")]
    [InlineData("stretch")]
    [InlineData("left")]
    [InlineData("wrong-node")]
    [InlineData("duplicate-name")]
    [InlineData("max-width")]
    [InlineData("padding")]
    [InlineData("spacing")]
    [InlineData("email-direction")]
    [InlineData("email-alignment")]
    public void AccountCandidateRejectsUnapprovedMutations(string mutation)
    {
        var candidate = LoadPage("AccountPage");
        AccountPanel(candidate).SetAttributeValue("HorizontalAlignment", "Center");
        AssertApprovedPage(candidate, "AccountPage", AccountBaseline, 3);
        var changed = new XDocument(candidate);
        var panel = AccountPanel(changed);
        XNamespace x = "http://schemas.microsoft.com/winfx/2006/xaml";
        var email = changed.Descendants().Single(e => (string?)e.Attribute(x + "Name") == "AccountGeometryEmail");
        switch (mutation)
        {
            case "missing": panel.Attribute("HorizontalAlignment")!.Remove(); break;
            case "stretch": panel.SetAttributeValue("HorizontalAlignment", "Stretch"); break;
            case "left": panel.SetAttributeValue("HorizontalAlignment", "Left"); break;
            case "wrong-node":
                panel.Attribute("HorizontalAlignment")!.Remove();
                panel.Parent!.SetAttributeValue("HorizontalAlignment", "Center");
                break;
            case "duplicate-name": panel.AddAfterSelf(new XElement(panel)); break;
            case "max-width": panel.SetAttributeValue("MaxWidth", "821"); break;
            case "padding": panel.SetAttributeValue("Padding", "41"); break;
            case "spacing": panel.SetAttributeValue("Spacing", "21"); break;
            case "email-direction": email.SetAttributeValue("FlowDirection", "RightToLeft"); break;
            case "email-alignment": email.SetAttributeValue("HorizontalAlignment", "Center"); break;
            default: throw new ArgumentOutOfRangeException(nameof(mutation));
        }
        Assert.NotNull(Record.Exception(() => AssertApprovedPage(changed, "AccountPage", AccountBaseline, 3)));
    }

    [Fact]
    public void AccountCandidateValidationDoesNotMutateInput()
    {
        var candidate = LoadPage("AccountPage");
        AccountPanel(candidate).SetAttributeValue("HorizontalAlignment", "Center");
        var before = candidate.ToString(SaveOptions.DisableFormatting);
        AssertApprovedPage(candidate, "AccountPage", AccountBaseline, 3);
        Assert.Equal(before, candidate.ToString(SaveOptions.DisableFormatting));
    }

    [Fact]
    public void CleanupAttemptsAllRemovalsAndReportsFailure()
    {
        var calls = new List<int>();
        Assert.False(GeometryCleanup.Run([
            () => { calls.Add(1); throw new InvalidOperationException(); },
            () => calls.Add(2), () => { calls.Add(3); throw new IOException(); }
        ]));
        Assert.Equal(new[] { 1, 2, 3 }, calls);
        Assert.True(GeometryCleanup.Run([]));
    }

    [Fact]
    public void CloseStopsQueuedCaptureEvenWhenDetachmentFails()
    {
        var budget = new GeometryCaptureBudget();
        var coordinator = new GeometryCaptureCoordinator(budget);
        var dispatcher = new Queue<Action>();
        var captures = 0;
        coordinator.ChangePage("Login");
        Assert.True(coordinator.TryQueue(out var token));
        dispatcher.Enqueue(() => { if (coordinator.TryBegin(token, true, out _)) captures++; });
        coordinator.Stop();
        Assert.False(GeometryCleanup.Run([() => throw new IOException()]));
        dispatcher.Dequeue()();
        Assert.Equal(0, captures);
        Assert.True(budget.Stopped);
    }

    [Fact]
    public void ObserverFailureCannotProduceSuccessfulReceipt()
    {
        var session = new GeometryReceiptSession(Guid.NewGuid(), new MemoryStream());
        session.Invalidate();
        Assert.False(session.Finish().Completed);
        Assert.Contains("observer-failed", session.Finish().ErrorCodes);
    }
}
