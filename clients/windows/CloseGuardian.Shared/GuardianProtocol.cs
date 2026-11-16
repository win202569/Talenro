using System.Buffers.Binary;
using System.Security.Cryptography;
using System.Text;
namespace Talenro.Observation.CloseGuardian;
public static class GuardianProtocol
{
    private static readonly UTF8Encoding Utf8=new(false,true);
    private const string DomainLabel="Talenro.CloseGuardian.v1";
    internal static void Validate(GuardianFrame f)
    {
        if(f is null || !GuardianValidation.Policy(f.Policy) || !GuardianValidation.Identity(f.Guardian) ||
            !GuardianValidation.Identity(f.App) || !GuardianValidation.Hex(f.GuardianChallenge,64) ||
            !GuardianValidation.Hex(f.AppChallenge,64) || f.Sequence==0 || f.DeadlineMs!=300000 ||
            !Enum.IsDefined(f.Kind) || !Enum.IsDefined(f.Reason)) throw new GuardianProtocolException();
        bool allowed=f.Domain switch
        {
            ChannelDomain.Application=>f.Direction switch
            {
                FrameDirection.GuardianToApp=>f.Kind is FrameKind.Challenge or FrameKind.Bound or FrameKind.CloseNormal or FrameKind.Ack or FrameKind.Reject,
                FrameDirection.AppToGuardian=>f.Kind is FrameKind.Hello or FrameKind.Proof or FrameKind.Bound or FrameKind.Ack or FrameKind.Reject,
                _=>false
            },
            ChannelDomain.Control=>f.Direction switch
            {
                FrameDirection.CoordinatorToGuardian=>f.Kind is FrameKind.AssignTarget or FrameKind.CloseEarly or FrameKind.Abort,
                FrameDirection.GuardianToCoordinator=>f.Kind is FrameKind.Ready or FrameKind.Ack or FrameKind.Reject,
                _=>false
            },
            _=>false
        };
        bool reason=f.Kind switch
        {
            FrameKind.CloseNormal=>f.Reason is CloseReason.Deadline or CloseReason.Early or CloseReason.ParentLost,
            FrameKind.CloseEarly=>f.Reason==CloseReason.Early,
            _=>f.Reason==CloseReason.None
        };
        if(!allowed || !reason) throw new GuardianProtocolException();
    }
    public static byte[] Encode(GuardianFrame frame, ReadOnlySpan<byte> key)
    {
        Validate(frame); if(key.Length!=32) throw new GuardianProtocolException();
        try
        {
            using var stream=new MemoryStream(); using var writer=new BinaryWriter(stream,Utf8,true);
            void Text(string text) { var bytes=Utf8.GetBytes(text); writer.Write((uint)bytes.Length); writer.Write(bytes); }
            void Identity(ProcessIdentity p) { writer.Write(p.Pid); writer.Write(p.CreationFileTimeUtc); Text(p.Path); Text(p.CommandLine); }
            writer.Write(0U); Text(DomainLabel); writer.Write((byte)1); writer.Write((byte)frame.Domain);
            writer.Write((byte)frame.Direction); writer.Write((byte)frame.Kind); Text(frame.Policy.RunId.ToString("N"));
            writer.Write(frame.Sequence); writer.Write(Convert.FromHexString(frame.Policy.Digest));
            writer.Write(frame.Policy.T0Ticks); writer.Write(frame.Policy.Frequency); Identity(frame.Guardian); Identity(frame.App);
            writer.Write(Convert.FromHexString(frame.GuardianChallenge)); writer.Write(Convert.FromHexString(frame.AppChallenge));
            writer.Write(frame.DeadlineMs); writer.Write((byte)frame.Reason);
            var body=stream.ToArray(); if(body.Length+32>4096) throw new GuardianProtocolException();
            BinaryPrimitives.WriteUInt32LittleEndian(body,(uint)(body.Length+32-4));
            return [..body,..HMACSHA256.HashData(key,body)];
        }
        catch(EncoderFallbackException) { throw new GuardianProtocolException(); }
    }
    public static GuardianFrame Decode(ReadOnlySpan<byte> bytes, ReadOnlySpan<byte> key)
    {
        if(key.Length!=32 || bytes.Length<36 || bytes.Length>4096 ||
            BinaryPrimitives.ReadUInt32LittleEndian(bytes)!=(uint)(bytes.Length-4)) throw new GuardianProtocolException();
        var data=bytes[..^32];
        if(!CryptographicOperations.FixedTimeEquals(HMACSHA256.HashData(key,data),bytes[^32..])) throw new GuardianProtocolException();
        try
        {
            using var stream=new MemoryStream(data[4..].ToArray(),false); using var reader=new BinaryReader(stream,Utf8);
            byte[] Exact(int n) { var value=reader.ReadBytes(n); if(value.Length!=n) throw new GuardianProtocolException(); return value; }
            string Text() { var n=reader.ReadUInt32(); if(n>4096 || n>stream.Length-stream.Position) throw new GuardianProtocolException(); return Utf8.GetString(Exact((int)n)); }
            ProcessIdentity Identity()=>new(reader.ReadInt32(),reader.ReadInt64(),Text(),Text());
            if(Text()!=DomainLabel || reader.ReadByte()!=1) throw new GuardianProtocolException();
            var domain=(ChannelDomain)reader.ReadByte(); var direction=(FrameDirection)reader.ReadByte(); var kind=(FrameKind)reader.ReadByte();
            var run=Text(); if(!GuardianValidation.Hex(run,32) || !Guid.TryParseExact(run,"N",out var runId)) throw new GuardianProtocolException();
            var sequence=reader.ReadUInt64(); var digest=Convert.ToHexString(Exact(32)).ToLowerInvariant();
            var policy=new GuardianPolicy(runId,digest,reader.ReadInt64(),reader.ReadInt64());
            var guardian=Identity(); var app=Identity();
            var guardianChallenge=Convert.ToHexString(Exact(32)).ToLowerInvariant(); var appChallenge=Convert.ToHexString(Exact(32)).ToLowerInvariant();
            var frame=new GuardianFrame(domain,direction,kind,policy,sequence,guardian,app,guardianChallenge,appChallenge,reader.ReadInt64(),(CloseReason)reader.ReadByte());
            if(stream.Position!=stream.Length) throw new GuardianProtocolException(); Validate(frame); return frame;
        }
        catch(Exception error) when(error is EndOfStreamException or DecoderFallbackException or OverflowException)
        { throw new GuardianProtocolException(); }
    }
}
public sealed class GuardianProtocolSession
{
    private readonly GuardianFrame expected;
    private ulong next;
    private bool failed;
    public GuardianProtocolSession(GuardianFrame expected) { GuardianProtocol.Validate(expected); this.expected=expected; next=expected.Sequence; }
    // Serialized receive owner. Authentication of a frame is NOT native peer verification.
    public GuardianFrame Receive(ReadOnlySpan<byte> bytes, ReadOnlySpan<byte> key)
    {
        if(failed) throw new GuardianProtocolException();
        try
        {
            var f=GuardianProtocol.Decode(bytes,key);
            if(f.Domain!=expected.Domain || f.Direction!=expected.Direction || f.Kind!=expected.Kind || f.Policy!=expected.Policy ||
                f.Sequence!=next || !GuardianValidation.SameIdentity(f.Guardian,expected.Guardian) || !GuardianValidation.SameIdentity(f.App,expected.App) ||
                f.GuardianChallenge!=expected.GuardianChallenge || f.AppChallenge!=expected.AppChallenge || f.DeadlineMs!=expected.DeadlineMs)
                throw new GuardianProtocolException();
            if(next==ulong.MaxValue) failed=true; else next++;
            return f;
        }
        catch(GuardianProtocolException) { failed=true; throw; }
    }
}
