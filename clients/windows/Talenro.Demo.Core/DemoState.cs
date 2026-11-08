namespace Talenro.Demo.Core;

public enum DemoPage { Login, Connection, Nodes, Account }
public enum ConnectionPhase { Disconnected, Connecting, Connected, Disconnecting }
public enum EntitlementStatus { Active, Expired, Exhausted, Unknown, Stale, DeviceLimit }
public sealed record NoticeContext(string? FromCountry = null, string? ToCountry = null);
[Flags]
public enum LockRecoveryChoices { None = 0, CrossCountryOnce = 1, RetryLockedCountry = 2, StayDisconnected = 4 }

public sealed record DemoState
{
    public DemoPage Page { get; init; } = DemoPage.Login;
    public bool HasSession { get; init; }
    public ConnectionPhase Connection { get; init; } = ConnectionPhase.Disconnected;
    public string? PreferredNode { get; init; }
    public string? DraftNode { get; init; }
    public string? LockedCountry { get; init; }
    public string? CurrentCountry { get; init; }
    public LockRecoveryChoices RecoveryChoices { get; init; }
    public string? RecoveryCountry { get; init; }
    public string? PendingNode { get; init; }
    public bool HasPendingNodeChange { get; init; }
    public EntitlementStatus Entitlement { get; init; } = EntitlementStatus.Active;
    public string Notice { get; init; } = "";
    public NoticeContext NoticeContext { get; init; } = new();
    public string Search { get; init; } = "";
}
