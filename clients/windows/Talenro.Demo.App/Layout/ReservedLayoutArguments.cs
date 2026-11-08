namespace Talenro.Demo.App.Layout;

internal static class ReservedLayoutArguments
{
    public static bool ContainsReserved(IReadOnlyList<string> args) =>
        args.Any(value => value.StartsWith("--layout", StringComparison.OrdinalIgnoreCase));
}
