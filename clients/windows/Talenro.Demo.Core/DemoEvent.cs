namespace Talenro.Demo.Core;

public abstract record DemoEvent;
public sealed record EnterDemo : DemoEvent;
public sealed record Logout : DemoEvent;
public sealed record Navigate(DemoPage Page) : DemoEvent;
public sealed record ShowPlaceholder(string Destination) : DemoEvent;
