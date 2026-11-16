using System.Buffers.Binary;
using System.Security.Cryptography;
using Talenro.Observation.CloseGuardian;
using Xunit;

namespace Talenro.Demo.Tests;
public class GuardianProtocolTests
{
    internal static byte[] Key => Enumerable.Range(0,32).Select(x=>(byte)x).ToArray();
    internal static GuardianFrame Frame() => new(ChannelDomain.Application,FrameDirection.GuardianToApp,
        FrameKind.CloseNormal,new(Guid.Parse("11111111-1111-1111-1111-111111111111"),new string('2',64),1000,10000000),1,
        new(123,133000000000000000,@"C:\g.exe",@"""C:\g.exe"" "),
        new(456,133000000000000010,@"C:\a.exe",@"""C:\a.exe"" "),new string('3',64),new string('4',64),300000,CloseReason.Deadline);
    [Fact] public void IndependentWireVectorMatchesCanonicalLayout()
    {
        // Independent PowerShell BitConverter/UTF8/HMAC construction, not Encode output.
        const string hex="330100001800000054616C656E726F2E436C6F7365477561726469616E2E76310101010520000000313131313131313131313131313131313131313131313131313131313131313101000000000000002222222222222222222222222222222222222222222222222222222222222222E80300000000000080969800000000007B0000000080209BCB82D80108000000433A5C672E6578650B00000022433A5C672E6578652220C80100000A80209BCB82D80108000000433A5C612E6578650B00000022433A5C612E657865222033333333333333333333333333333333333333333333333333333333333333334444444444444444444444444444444444444444444444444444444444444444E09304000000000001D75149F3103A200CBE19C2A1B704C0DBD894F38C109BA0A1C5AEB0EAD0841FC8";
        Assert.Equal(Convert.FromHexString(hex),GuardianProtocol.Encode(Frame(),Key));
        Assert.Equal(Frame(),GuardianProtocol.Decode(Convert.FromHexString(hex),Key));
    }
    [Fact] public void AuthenticatedInvalidUtf8AndUnknownVersionReject()
    {
        var bytes=GuardianProtocol.Encode(Frame(),Key);
        // First identity path begins at offset148 in this independently specified vector.
        bytes[148]=0xff; Resign(bytes);
        Assert.Throws<GuardianProtocolException>(()=>GuardianProtocol.Decode(bytes,Key));
        bytes=GuardianProtocol.Encode(Frame(),Key); bytes[32]=2; Resign(bytes);
        Assert.Throws<GuardianProtocolException>(()=>GuardianProtocol.Decode(bytes,Key));
    }
    private static void Resign(byte[] bytes) => HMACSHA256.HashData(Key,bytes.AsSpan(0,bytes.Length-32)).CopyTo(bytes,bytes.Length-32);
    [Fact] public void RoundTripPreservesExactTranscriptIncludingTrailingSpace()
    {
        var frame=Frame(); var bytes=GuardianProtocol.Encode(frame,Key);
        Assert.Equal(frame,GuardianProtocol.Decode(bytes,Key));
        Assert.Equal(bytes.Length-4,BinaryPrimitives.ReadInt32LittleEndian(bytes));
    }
    [Fact] public void AnyWireByteTamperOrWrongKeyRejects()
    {
        var original=GuardianProtocol.Encode(Frame(),Key);
        for(var i=0;i<original.Length;i++) { var bytes=original.ToArray(); bytes[i]^=1;
            Assert.Throws<GuardianProtocolException>(()=>GuardianProtocol.Decode(bytes,Key)); }
        Assert.Throws<GuardianProtocolException>(()=>GuardianProtocol.Decode(original,new byte[32]));
    }
    [Fact] public void EveryTruncatedFrameAndTrailingDataRejects()
    {
        var bytes=GuardianProtocol.Encode(Frame(),Key);
        for(var i=0;i<bytes.Length;i++) Assert.Throws<GuardianProtocolException>(()=>GuardianProtocol.Decode(bytes.AsSpan(0,i),Key));
        Assert.Throws<GuardianProtocolException>(()=>GuardianProtocol.Decode([..bytes,0],Key));
    }
    [Fact] public void SessionRejectsAuthenticatedWrongBindingAndReplay()
    {
        var f=Frame();
        GuardianFrame[] changed=[f with {Direction=FrameDirection.AppToGuardian,Kind=FrameKind.Ack,Reason=CloseReason.None},
            f with {Domain=ChannelDomain.Control,Direction=FrameDirection.CoordinatorToGuardian,Kind=FrameKind.CloseEarly,Reason=CloseReason.Early},
            f with {Policy=f.Policy with {RunId=Guid.NewGuid()}},f with {Policy=f.Policy with {Digest=new string('5',64)}},
            f with {Policy=f.Policy with {T0Ticks=1001}}, f with {Policy=f.Policy with {Frequency=1000}},
            f with {App=f.App with {CreationFileTimeUtc=f.App.CreationFileTimeUtc+1}},
            f with {App=f.App with {CommandLine=f.App.CommandLine.TrimEnd()}},
            f with {GuardianChallenge=new string('5',64)},f with {AppChallenge=new string('5',64)},f with {Sequence=2}];
        foreach(var bad in changed) Assert.Throws<GuardianProtocolException>(()=>new GuardianProtocolSession(f).Receive(GuardianProtocol.Encode(bad,Key),Key));
        var session=new GuardianProtocolSession(f); var bytes=GuardianProtocol.Encode(f,Key);
        Assert.Equal(f,session.Receive(bytes,Key));
        Assert.Throws<GuardianProtocolException>(()=>session.Receive(bytes,Key));
    }
    [Fact] public void MaximumFrameFitsAndOneExtraByteRejects()
    {
        var f=Frame(); var baseline=GuardianProtocol.Encode(f,Key).Length;
        var max=f with {App=f.App with {CommandLine=f.App.CommandLine+new string('x',4096-baseline)}};
        Assert.Equal(4096,GuardianProtocol.Encode(max,Key).Length);
        Assert.Equal(max,GuardianProtocol.Decode(GuardianProtocol.Encode(max,Key),Key));
        Assert.Throws<GuardianProtocolException>(()=>GuardianProtocol.Encode(max with {App=max.App with {CommandLine=max.App.CommandLine+"x"}},Key));
    }
    [Fact] public void InvalidSemanticFieldsDoNotGetAuthenticated()
    {
        var f=Frame();
        GuardianFrame[] invalid=[f with {Sequence=0},f with {DeadlineMs=300001},f with {Kind=(FrameKind)255},
            f with {App=f.App with {Pid=0}},f with {App=f.App with {Path="relative"}},
            f with {Reason=CloseReason.None},f with {GuardianChallenge="x"},f with {Policy=f.Policy with {Frequency=0}}];
        foreach(var bad in invalid) Assert.Throws<GuardianProtocolException>(()=>GuardianProtocol.Encode(bad,Key));
    }
}
