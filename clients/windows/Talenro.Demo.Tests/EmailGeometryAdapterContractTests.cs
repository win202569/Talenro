using System.Security.Cryptography;
using System.Text;
using System.Xml.Linq;
using Talenro.Demo.App.Layout.EmailGeometry;
using Xunit;

namespace Talenro.Demo.Tests;

public class EmailGeometryAdapterContractTests
{
    [Theory]
    [InlineData("LoginPage", "A3C830FDC2F8E493B7FFDD5BC67F039D7511DA565161FB2989243424BA0C7FD0", 2)]
    [InlineData("AccountPage", "F01C2618893423F1002CBD3EDE467FF634D821BE7031D45492437ED088D6B57F", 3)]
    public void OnlyApprovedNonvisualNamesAreAdded(string page, string baseline, int expectedNames)
    {
        var doc = XDocument.Load(Path.Combine(AppContext.BaseDirectory, "AppSources", "Views", page + ".xaml"));
        XNamespace x = "http://schemas.microsoft.com/winfx/2006/xaml";
        string[] names = ["LoginGeometryScroll", "LoginGeometryPanel", "AccountGeometryScroll", "AccountGeometryPanel", "AccountGeometryEmail"];
        var added = doc.Descendants().Attributes(x + "Name").Where(a => names.Contains(a.Value)).ToArray();
        Assert.Equal(expectedNames, added.Length);
        foreach (var attr in added) attr.Remove();
        Assert.Equal(baseline, Convert.ToHexString(SHA256.HashData(Encoding.UTF8.GetBytes(doc.ToString(SaveOptions.DisableFormatting)))));
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
