namespace Talenro.Demo.Core;

public enum DemoPage { Login, Connection, Nodes, Account }
public enum ConnectionPhase { Disconnected, Connecting, Connected, Disconnecting }
public enum EntitlementStatus { Active, Expired, Exhausted, Unknown, Stale, DeviceLimit }

public sealed record DemoState
{
    public DemoPage Page { get; init; } = DemoPage.Login;
    public bool HasSession { get; init; }
    public ConnectionPhase Connection { get; init; } = ConnectionPhase.Disconnected;
    public string? PreferredNode { get; init; }
    public string? DraftNode { get; init; }
    public string? LockedCountry { get; init; }
    public string? PendingNode { get; init; }
    public bool HasPendingNodeChange { get; init; }
    public EntitlementStatus Entitlement { get; init; } = EntitlementStatus.Active;
    public string Notice { get; init; } = "";
    public string Search { get; init; } = "";
}
