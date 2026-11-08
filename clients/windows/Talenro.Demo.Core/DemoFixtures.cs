namespace Talenro.Demo.Core;

public sealed record DemoNode(string Id, string Country, bool Enabled);

public static class DemoFixtures
{
    public static DemoState Initial() => new();

    public static IReadOnlyList<DemoNode> Nodes { get; } = Array.AsReadOnly(new[]
    {
        new DemoNode("jp-1", "JP", true),
        new DemoNode("de-1", "DE", true),
        new DemoNode("disabled-1", "JP", false)
    });
}
