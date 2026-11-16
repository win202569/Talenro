using Talenro.Observation.CloseGuardian;
using Xunit;
namespace Talenro.Demo.Tests;
public class GuardianIdentityTests
{
    internal static ProcessIdentity Expected=>new(456,133000000000000017,@"C:\a.exe",@"""C:\a.exe"" ");
    internal static IdentityEvidence Evidence=>new(true,true,456,456,133000000000000017,456,637911232000000010,@"c:\a.exe",@"""C:\a.exe"" ",true);
    [Fact] public void ExactNativeIdentityAndTruncatedCimMicrosecondMatch()
    { Assert.True(GuardianIdentity.Verify(Expected,Evidence)); }
    [Fact] public void EveryIdentityFailureRejectsWithoutTolerance()
    {
        var e=Evidence;
        IdentityEvidence[] bad=[e with {NativeSucceeded=false},e with {CimSucceeded=false},e with {ConnectedPeerPid=457},
            e with {RetainedPid=457},e with {RetainedCreation=e.RetainedCreation+1},e with {CimPid=457},
            e with {CimCreationUtcTicks=e.CimCreationUtcTicks+1},e with {CimCreationUtcTicks=e.CimCreationUtcTicks+10},
            e with {CimPath=@"C:\other.exe"},e with {CimCommandLine=e.CimCommandLine.TrimEnd()},e with {Alive=false}];
        foreach(var value in bad) Assert.False(GuardianIdentity.Verify(Expected,value));
        Assert.False(GuardianIdentity.Verify(Expected with {CreationFileTimeUtc=long.MaxValue},e));
    }
}
