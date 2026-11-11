namespace Talenro.Demo.App.Layout.EmailGeometry;

internal static class GeometryCleanup
{
    internal static bool Run(IEnumerable<Action> removals)
    {
        var success = true;
        foreach (var remove in removals)
        {
            try { remove(); }
            catch { success = false; }
        }
        return success;
    }
}

// UI-thread state. Reservations are attempts, not successful snapshots.
public sealed class GeometryCaptureBudget
{
    private int login;
    private int account;
    public bool Stopped { get; private set; }
    public int Attempts(string? page) => page == "Login" ? login : page == "Account" ? account : 0;

    public bool TryReserve(string page, bool eligible, out int sequence)
    {
        sequence = 0;
        if (Stopped || !eligible || page is not ("Login" or "Account") || Attempts(page) >= 2) return false;
        if (page == "Login") login++; else account++;
        sequence = login + account;
        if (sequence == 4) Stop();
        return true;
    }

    public void Stop() => Stopped = true;
}

public sealed class GeometryCaptureCoordinator(GeometryCaptureBudget budget)
{
    private string? page;
    private long nextToken;
    private long? pending;

    public void ChangePage(string? value)
    {
        pending = null;
        page = value;
    }

    public bool TryQueue(out long token)
    {
        token = 0;
        if (budget.Stopped || pending.HasValue || page is not ("Login" or "Account") || budget.Attempts(page) >= 2)
            return false;
        if (nextToken == long.MaxValue) { Stop(); return false; }
        token = ++nextToken;
        pending = token;
        return true;
    }

    public bool TryBegin(long token, bool eligible, out int sequence)
    {
        sequence = 0;
        if (pending != token) return false;
        pending = null;
        return page is not null && budget.TryReserve(page, eligible, out sequence);
    }

    public void Cancel(long token)
    {
        if (pending == token) pending = null;
    }

    public void Stop()
    {
        pending = null;
        budget.Stop();
    }
}
