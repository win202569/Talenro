using System.Text.RegularExpressions;
using System.Xml.Linq;
using Talenro.Demo.App.Layout;
using Xunit;

namespace Talenro.Demo.Tests;

// Declaration coverage only: live handler/focus behavior remains a native check.
public class LayoutViewContractTests
{
    // Deliberately source-contract checks, not a C# parser or native bidi/rendering proof.
    // The approved scope excludes launching WinUI or refactoring production views.
    public static TheoryData<string, string> DirectionContracts => new()
    {
        { "MainWindow.xaml.cs", "Root.FlowDirection = presentation.IsRtl ? FlowDirection.RightToLeft : FlowDirection.LeftToRight;" },
        { "Views/NodesPage.xaml.cs", "if (id is not null && App.DiagnosticPresentation?.IsRtl == true) item.FlowDirection = FlowDirection.LeftToRight;" },
        { "Views/ConnectionPage.xaml.cs", "if (presentation.IsRtl) LockedCountryValue.FlowDirection = FlowDirection.LeftToRight;" }
    };

    private static void AssertDirectionContract(string source, string declaration) =>
        Assert.Contains(Regex.Replace(declaration, @"\s+", ""), Regex.Replace(source, @"\s+", ""));

    [Theory]
    [MemberData(nameof(DirectionContracts))]
    public void RtlAndTechnicalIdentifierDirectionDeclarationsArePresent(string file, string declaration) =>
        AssertDirectionContract(Read(file), declaration);

    [Theory]
    [MemberData(nameof(DirectionContracts))]
    public void DirectionContractRejectsRemovedOrChangedDeclaration(string file, string declaration)
    {
        var source = Regex.Replace(Read(file), @"\s+", "");
        var expected = Regex.Replace(declaration, @"\s+", "");
        AssertDirectionContract(source, declaration);
        // Mutations stay in memory. Each must be real and fail the same positive assertion.
        foreach (var replacement in new[] { "", expected.Replace("FlowDirection.LeftToRight", "FlowDirection.RightToLeft") })
        {
            var mutated = source.Replace(expected, replacement, StringComparison.Ordinal);
            Assert.NotEqual(source, mutated);
            Assert.ThrowsAny<Xunit.Sdk.XunitException>(() => AssertDirectionContract(mutated, declaration));
        }
    }

    private static readonly XNamespace X = "http://schemas.microsoft.com/winfx/2006/xaml";
    private static string Read(string file) => File.Exists(Path.Combine(AppContext.BaseDirectory, "AppSources", file))
        ? File.ReadAllText(Path.Combine(AppContext.BaseDirectory, "AppSources", file)) : "";
    private static string FileFor(string view) => view == "Shell" ? "MainWindow.xaml" : $"Views/{view}Page.xaml";
    private static (string Name, string Key, string Property, string Type)[] Bindings(string view) =>
        Regex.Matches(Read("Diagnostics/LayoutStaticBindings.cs"), "Bind<(?<type>\\w+)>\\(\"(?<view>\\w+)\", viewId, root, \"(?<name>\\w+)\", \"(?<key>[^\"]+)\", \\(control, value\\) => control\\.(?<property>\\w+) = value, presentation\\);")
            .Where(m => m.Groups["view"].Value == view)
            .Select(m => (m.Groups["name"].Value, m.Groups["key"].Value, m.Groups["property"].Value, m.Groups["type"].Value)).ToArray();

    [Theory]
    [InlineData("Shell")]
    [InlineData("Login")]
    [InlineData("Connection")]
    [InlineData("Nodes")]
    [InlineData("Account")]
    public void EveryLocalizedPropertyHasTypedExistingControlBinding(string view)
    {
        var doc = XDocument.Parse(Read(FileFor(view)));
        var bindings = Bindings(view);
        var resources = LayoutResourceCatalog.Parse(Read("Strings/en-US/Resources.resw"));
        var chinese = LayoutResourceCatalog.Parse(Read("Strings/zh-Hans/Resources.resw"));
        AssertLocalizationCoverage(doc, bindings, resources, chinese);
    }

    [Theory]
    [InlineData("Shell")]
    [InlineData("Login")]
    [InlineData("Connection")]
    [InlineData("Nodes")]
    [InlineData("Account")]
    public void NewLocalizedControlWithoutResourcesIsRejected(string view)
    {
        var doc = XDocument.Parse(Read(FileFor(view)));
        // Mutate a real declaration in memory, not the checked-in application.
        doc.Root!.Add(new XElement(doc.Root.Name.Namespace + "TextBlock",
            new XAttribute(X + "Name", "NewUnmappedControl"),
            new XAttribute(X + "Uid", "NewUnmappedUid")));
        var error = Assert.Throws<InvalidDataException>(() => AssertLocalizationCoverage(doc, Bindings(view),
            LayoutResourceCatalog.Parse(Read("Strings/en-US/Resources.resw")),
            LayoutResourceCatalog.Parse(Read("Strings/zh-Hans/Resources.resw"))));
        Assert.Contains("NewUnmappedUid", error.Message);
    }

    private static void AssertLocalizationCoverage(XDocument doc,
        (string Name, string Key, string Property, string Type)[] bindings,
        IReadOnlyDictionary<string, string> resources, IReadOnlyDictionary<string, string> chinese)
    {
        var expectedCount = 0;
        foreach (var node in doc.Descendants().Where(x => x.Attribute(X + "Uid") is not null))
        {
            var uid = (string)node.Attribute(X + "Uid")!;
            var name = (string?)node.Attribute(X + "Name");
            var keys = resources.Keys.Where(key => key.StartsWith(uid + ".", StringComparison.Ordinal)).ToArray();
            if (keys.Length == 0)
                throw new InvalidDataException($"Localized UID '{uid}' has no resource properties.");
            foreach (var key in keys)
            {
                expectedCount++;
                Assert.Contains(bindings, entry => entry.Name == name && entry.Key == key &&
                    entry.Property == key[(uid.Length + 1)..] && entry.Type == node.Name.LocalName);
                Assert.Contains(key, chinese.Keys);
            }
        }
        Assert.Equal(expectedCount, bindings.Length);
        Assert.True(expectedCount > 0);
    }

    [Fact]
    public void AllCurrentKeysAreValidatedBeforeAnyViewAndLanguagesAgree()
    {
        var english = LayoutResourceCatalog.Parse(Read("Strings/en-US/Resources.resw"));
        var chinese = LayoutResourceCatalog.Parse(Read("Strings/zh-Hans/Resources.resw"));
        Assert.Equal(english.Keys.Order(), chinese.Keys.Order());
        Assert.Equal(english.Keys.Order(), LayoutResourceCatalog.RequiredKeys.Order());
        var startup = Read("App.xaml.cs");
        Assert.True(startup.IndexOf("LayoutResourceCatalog.Prepare", StringComparison.Ordinal) < startup.IndexOf("new MainWindow", StringComparison.Ordinal));
        Assert.Contains("DiagnosticPresentation", Read("MainWindow.xaml.cs"));
    }

    [Theory]
    [InlineData("Login")]
    [InlineData("Connection")]
    [InlineData("Nodes")]
    [InlineData("Account")]
    public void SharedViewsKeepHandlersAndHaveGuardedPresentationHooks(string view)
    {
        var code = Read($"Views/{view}Page.xaml.cs");
        Assert.Contains("#if LAYOUT_DIAGNOSTICS", code);
        Assert.Contains($"LayoutStaticBindings.Apply(\"{view}\"", code);
        Assert.Contains("ApplyDiagnosticPresentation", code);
        Assert.Contains("InitializeComponent()", code);
        Assert.DoesNotContain("PrimaryLanguageOverride", code);
        if (view == "Nodes")
        {
            Assert.Contains("Tag = id", code);
            Assert.Contains("model.Dispatch(new SelectNode(item.Tag as string))", code);
            Assert.Contains("MainWindow.Copy(\"Country\" + node.Country)", code);
            Assert.Contains("new ConfirmNodeChange()", code);
        }
        if (view == "Login") Assert.Contains("ApplyDiagnosticPresentation();", code[code.IndexOf("public void Reset", StringComparison.Ordinal)..]);
    }

    [Fact]
    public void TestPresentationCannotRemovePermanentWarningOrEmailIsolation()
    {
        foreach (var view in new[] { "Login", "Account" })
        {
            var doc = XDocument.Parse(Read(FileFor(view)));
            Assert.Contains(doc.Descendants("{http://schemas.microsoft.com/winfx/2006/xaml/presentation}TextBlock"),
                node => (string?)node.Attribute("Text") == "demo@example.com" && (string?)node.Attribute("FlowDirection") == "LeftToRight");
        }
        Assert.Contains("Root.FlowDirection", Read("MainWindow.xaml.cs"));
        Assert.Contains("IsTabStop=\"False\"", Read("MainWindow.xaml"));
        Assert.DoesNotContain("PropertyInfo", Read("Diagnostics/LayoutStaticBindings.cs"));
    }
}
