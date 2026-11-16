using Talenro.Observation.CloseGuardian;
using Xunit;
namespace Talenro.Demo.Tests;
public class GuardianIsolationTests
{
    [Fact] public void AbsentOrPartialActivationCannotCreateObservationIntent()
    {
        foreach(var environment in new[]{new Dictionary<string,string?>(),new Dictionary<string,string?>{{"TALENRO_GUARDIAN_V1_VERSION","1"}}})
        {
            var activation=GuardianActivation.Parse(environment);
            Assert.NotEqual(ActivationStatus.Valid,activation.Status); Assert.Null(activation.Value);
        }
    }
    [Fact] public void SharedLayerHasNoNativeProcessFileOrAppDependencies()
    {
        var directory=new DirectoryInfo(AppContext.BaseDirectory);
        while(directory is not null && !File.Exists(Path.Combine(directory.FullName,"Talenro.Demo.Tests.csproj"))) directory=directory.Parent;
        Assert.NotNull(directory);
        var shared=Path.GetFullPath(Path.Combine(directory!.FullName,"../CloseGuardian.Shared"));
        foreach(var path in Directory.GetFiles(shared,"*.cs"))
        {
            var text=File.ReadAllText(path);
            foreach(var forbidden in new[]{"DllImport","LibraryImport","Process.Start","Process.Kill","NamedPipe","File.","DateTime.Now","Stopwatch.Start","Thread.Sleep","Microsoft.UI","Environment.Exit"})
                Assert.DoesNotContain(forbidden,text);
        }
        var project=System.Xml.Linq.XDocument.Load(Path.Combine(directory.FullName,"Talenro.Demo.Tests.csproj"));
        Assert.DoesNotContain(project.Descendants("ProjectReference"),e=>e.Attribute("Include")!.Value.Contains("App")||e.Attribute("Include")!.Value.Contains("CloseGuardian"));
    }
}
