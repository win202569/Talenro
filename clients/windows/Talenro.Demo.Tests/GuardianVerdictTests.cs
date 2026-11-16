using Talenro.Observation.CloseGuardian;
using Xunit;
namespace Talenro.Demo.Tests;
public class GuardianVerdictTests
{
    internal static CloseEvidence Good=>new(290000,290001,290002,290003,290004,290005,true,true,true,true,false);
    [Fact] public void CompleteTimelyCorrelatedEvidencePasses()
    { Assert.Equal(CloseVerdict.Passed,GuardianVerdict.Evaluate(Good)); }
    [Fact] public void AckOrRequestWithoutExitAndMissingFinalizationNeverPass()
    {
        var e=Good;
        CloseEvidence[] bad=[e with {OriginalExitVerified=false},e with {ExitMs=null},e with {Finalized=false},
            e with {Correlated=false},e with {HashesUnchanged=false},e with {ExitMs=360000},
            e with {SendCompleteMs=300000},e with {AcceptedMs=300000},e with {InvokedMs=300000},
            e with {ClaimMs=null},e with {SendStartMs=null},e with {AcceptedMs=null},e with {InvokedMs=null},
            e with {ClaimMs=-1},e with {InvokedMs=290002},e with {ExitMs=290003}];
        foreach(var v in bad) Assert.Equal(CloseVerdict.Failed,GuardianVerdict.Evaluate(v));
    }
    [Fact] public void VerifiedManualExitIsNotAutonomousCoverage()
    { Assert.Equal(CloseVerdict.NotExercised,GuardianVerdict.Evaluate(Good with {ManualOrEarlyExit=true,ClaimMs=null})); }
    [Fact] public void ReceiveCanPrecedeSendCompletionButNotSendStart()
    {
        Assert.Equal(CloseVerdict.Passed,GuardianVerdict.Evaluate(Good with {SendCompleteMs=290004}));
        Assert.Equal(CloseVerdict.Failed,GuardianVerdict.Evaluate(Good with {SendStartMs=290004}));
    }
}
