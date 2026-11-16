using System.Security.Cryptography;
namespace Talenro.Observation.CloseGuardian;

public sealed record ProcessIdentity(int Pid, long CreationFileTimeUtc, string Path, string CommandLine);
public sealed record GuardianPolicy(Guid RunId, string Digest, long T0Ticks, long Frequency);
public sealed record IdentityEvidence(bool NativeSucceeded,bool CimSucceeded,int ConnectedPeerPid,
    int RetainedPid,long RetainedCreation,int CimPid,long CimCreationUtcTicks,string CimPath,string CimCommandLine,bool Alive);
public static class GuardianIdentity
{
    public static bool Verify(ProcessIdentity expected,IdentityEvidence evidence) =>
        GuardianValidation.Identity(expected) && evidence is not null && evidence.NativeSucceeded &&
        evidence.CimSucceeded && evidence.Alive && evidence.ConnectedPeerPid==expected.Pid &&
        evidence.RetainedPid==expected.Pid && evidence.RetainedCreation==expected.CreationFileTimeUtc &&
        evidence.CimPid==expected.Pid && evidence.CimCreationUtcTicks==
            (expected.CreationFileTimeUtc+504911232000000000L)/10*10 &&
        string.Equals(expected.Path,evidence.CimPath,StringComparison.OrdinalIgnoreCase) &&
        string.Equals(expected.CommandLine,evidence.CimCommandLine,StringComparison.Ordinal);
}
public enum ActivationStatus { Disabled, Valid, Rejected }
public sealed record ActivationResult(ActivationStatus Status, GuardianBootstrap? Value, string? ErrorCode);
public sealed class GuardianBootstrap : IDisposable
{
    private readonly byte[] key;
    private bool disposed;
    internal GuardianBootstrap(GuardianPolicy policy,string pipe,int pid,long creation,byte[] secret)
    { Policy=policy; Pipe=pipe; GuardianPid=pid; GuardianCreation=creation; key=secret; }
    public GuardianPolicy Policy { get; }
    public string Pipe { get; }
    public int GuardianPid { get; }
    public long GuardianCreation { get; }
    public ReadOnlySpan<byte> Key { get { ObjectDisposedException.ThrowIf(disposed,this); return key; } }
    public void Dispose() { CryptographicOperations.ZeroMemory(key); disposed=true; }
    public override string ToString() => "GuardianBootstrap [private activation data]";
}
public enum ChannelDomain : byte { Application = 1, Control = 2 }
public enum FrameDirection : byte { GuardianToApp = 1, AppToGuardian = 2, CoordinatorToGuardian = 3, GuardianToCoordinator = 4 }
public enum FrameKind : byte { Hello = 1, Challenge, Proof, Bound, CloseNormal, Ack, Reject, Ready, AssignTarget, CloseEarly, Abort }
public enum CloseReason : byte { None, Deadline, Early, ParentLost }
public sealed record GuardianFrame(ChannelDomain Domain, FrameDirection Direction, FrameKind Kind,
    GuardianPolicy Policy, ulong Sequence, ProcessIdentity Guardian, ProcessIdentity App,
    string GuardianChallenge, string AppChallenge, long DeadlineMs, CloseReason Reason);
public sealed class GuardianProtocolException : Exception
{
    public GuardianProtocolException() : base("Guardian frame rejected.") { }
}

internal static class GuardianValidation
{
    internal static bool Hex(string? value,int length) => value is not null && value.Length==length &&
        value.All(c=>c is >= '0' and <= '9' or >= 'a' and <= 'f');
    internal static bool Identity(ProcessIdentity? value) => value is not null && value.Pid>0 &&
        value.CreationFileTimeUtc>0 && value.CreationFileTimeUtc<=2650467743999999999L &&
        value.Path is {Length: >=3 and <=4096} && char.IsAsciiLetter(value.Path[0]) &&
        value.Path[1]==':' && value.Path[2]=='\\' && !value.Path.Contains('\0') &&
        value.CommandLine is {Length: >0 and <=4096} && !value.CommandLine.Contains('\0');
    internal static bool Policy(GuardianPolicy? value) => value is not null && value.RunId!=Guid.Empty &&
        Hex(value.Digest,64) && value.T0Ticks>=0 && value.Frequency>0;
    internal static bool SameIdentity(ProcessIdentity a,ProcessIdentity b) => a.Pid==b.Pid &&
        a.CreationFileTimeUtc==b.CreationFileTimeUtc && string.Equals(a.Path,b.Path,StringComparison.OrdinalIgnoreCase) &&
        string.Equals(a.CommandLine,b.CommandLine,StringComparison.Ordinal);
}
