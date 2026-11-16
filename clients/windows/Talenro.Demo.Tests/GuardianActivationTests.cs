using Talenro.Observation.CloseGuardian;
using Xunit;

namespace Talenro.Demo.Tests;
public class GuardianActivationTests
{
    internal const string Prefix = "TALENRO_GUARDIAN_V1_";
    internal static Dictionary<string,string?> Valid() => new()
    {
        [Prefix+"VERSION"]="1", [Prefix+"RUN_ID"]="11111111111111111111111111111111",
        [Prefix+"PIPE"]="talenro.guardian.22222222222222222222222222222222",
        [Prefix+"KEY"]=new string('a',64), [Prefix+"GUARDIAN_PID"]="123",
        [Prefix+"GUARDIAN_CREATION"]="133000000000000000", [Prefix+"POLICY"]=new string('b',64),
        [Prefix+"T0"]="1000", [Prefix+"FREQUENCY"]="10000000"
    };
    [Fact] public void AbsentBootstrapIsDormant()
    {
        var result=GuardianActivation.Parse(new Dictionary<string,string?> { ["PATH"]="unchanged" });
        Assert.Equal(ActivationStatus.Disabled,result.Status); Assert.Null(result.Value);
    }
    [Fact] public void CompleteBootstrapOwnsSecretAndDisposesIt()
    {
        var map=Valid(); var result=GuardianActivation.Parse(map);
        Assert.Equal(ActivationStatus.Valid,result.Status); using var value=result.Value!;
        Assert.Equal(123,value.GuardianPid); Assert.Equal(1000,value.Policy.T0Ticks);
        Assert.Equal(32,value.Key.Length); var borrowed=value.Key;
        Assert.All(value.Key.ToArray(),b=>Assert.Equal(0xaa,b));
        map[Prefix+"KEY"]="untrusted change";
        Assert.DoesNotContain(new string('a',64),value.ToString());
        value.Dispose(); Assert.All(borrowed.ToArray(),b=>Assert.Equal(0,b));
        Assert.Throws<ObjectDisposedException>(()=>value.Key.ToArray());
    }
    [Fact] public void EveryMissingKeyRejectsInsteadOfStartingUnguarded()
    {
        foreach(var key in Valid().Keys) { var map=Valid(); map.Remove(key);
            var r=GuardianActivation.Parse(map); Assert.Equal(ActivationStatus.Rejected,r.Status); Assert.Null(r.Value); }
    }
    [Theory]
    [InlineData("VERSION","2")][InlineData("GUARDIAN_PID","-1")][InlineData("GUARDIAN_PID"," 123")]
    [InlineData("GUARDIAN_CREATION","0")][InlineData("T0","-1")][InlineData("FREQUENCY","0")]
    [InlineData("KEY","private-secret")][InlineData("POLICY","zz")]
    [InlineData("RUN_ID","00000000000000000000000000000000")][InlineData("PIPE","\\\\remote\\pipe\\x")]
    public void InvalidFieldsRejectWithoutEchoingSecrets(string field,string input)
    {
        var map=Valid(); map[Prefix+field]=input; var result=GuardianActivation.Parse(map);
        Assert.Equal(ActivationStatus.Rejected,result.Status); Assert.Null(result.Value);
        Assert.Equal("InvalidBootstrap",result.ErrorCode);
    }
    [Fact] public void UnknownAndCaseCollidingKeysReject()
    {
        var map=Valid(); map[Prefix+"EXTRA"]="x";
        Assert.Equal(ActivationStatus.Rejected,GuardianActivation.Parse(map).Status);
        map=Valid(); map[(Prefix+"KEY").ToLowerInvariant()]=new string('a',64);
        Assert.Equal(ActivationStatus.Rejected,GuardianActivation.Parse(map).Status);
    }
}
