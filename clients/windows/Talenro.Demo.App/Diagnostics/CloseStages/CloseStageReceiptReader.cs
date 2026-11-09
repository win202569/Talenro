using System.Text;
using System.Text.Json;
namespace Talenro.Demo.App.Layout.CloseStages;

internal enum CloseStageReceiptStatus { ValidPrefix, Inconclusive }
internal sealed record CloseStageReceipt(CloseStageReceiptStatus Status,
    IReadOnlyList<CloseStageRecord> Records, bool CompletedClose);
internal static class CloseStageReceiptReader
{
    private static CloseStageReceipt Invalid() => new(CloseStageReceiptStatus.Inconclusive, Array.Empty<CloseStageRecord>(), false);

    public static CloseStageReceipt Read(Stream input, Guid expectedRun)
    {
        try
        {
            if (expectedRun == Guid.Empty) return Invalid();
            // Fixed upper bound even for an unseekable input stream.
            var bytes = new byte[65537];
            int length = 0;
            while (length < bytes.Length)
            {
                int count = input.Read(bytes, length, bytes.Length - length);
                if (count == 0) break;
                length += count;
            }
            if (length == 0 || length > 65536 || bytes[length - 1] != (byte)'\n') return Invalid();
            var lines = new UTF8Encoding(false, true).GetString(bytes, 0, length).Split('\n');
            if (lines.Length - 1 > 128) return Invalid();
            var records = new List<CloseStageRecord>();
            bool ready = false, health = false;
            CloseStage? prior = null;
            foreach (var line in lines[..^1])
            {
                using var doc = JsonDocument.Parse(line);
                var root = doc.RootElement;
                if (root.ValueKind != JsonValueKind.Object) return Invalid();
                var properties = root.EnumerateObject().Select(p => p.Name).ToArray();
                if (properties.Distinct(StringComparer.Ordinal).Count() != properties.Length) return Invalid();
                if (root.GetProperty("Schema").GetInt32() != 1 || root.GetProperty("Run").GetString() != expectedRun.ToString("N")) return Invalid();
                string? type = root.GetProperty("Type").GetString();
                string[] allowed = type switch
                {
                    "ready" => ["Type", "Schema", "Run", "Frequency"],
                    "stage" => ["Type", "Schema", "Run", "Sequence", "Stage", "Ticks"],
                    "health" => ["Type", "Schema", "Run", "Dropped", "InvalidProducer", "Failed"],
                    _ => []
                };
                if (properties.Length != allowed.Length || properties.Any(p => !allowed.Contains(p))) return Invalid();
                if (type == "ready")
                {
                    if (ready || records.Count != 0 || root.GetProperty("Frequency").GetInt64() <= 0) return Invalid();
                    ready = true;
                }
                else if (type == "stage")
                {
                    if (!ready || health) return Invalid();
                    int seq = root.GetProperty("Sequence").GetInt32();
                    var stage = (CloseStage)root.GetProperty("Stage").GetInt32();
                    long ticks = root.GetProperty("Ticks").GetInt64();
                    if (seq != records.Count + 1 || ticks < 0 ||
                        (records.Count != 0 && ticks < records[^1].Ticks) || !Follows(prior, stage)) return Invalid();
                    records.Add(new(seq, stage, ticks));
                    prior = stage;
                }
                else if (type == "health")
                {
                    if (!ready || health || root.GetProperty("Dropped").GetInt32() != 0 ||
                        root.GetProperty("InvalidProducer").GetBoolean() || root.GetProperty("Failed").GetBoolean()) return Invalid();
                    health = true;
                }
                else return Invalid();
            }
            bool closed = prior == CloseStage.CloseReturn;
            if (!ready || (closed && !health)) return Invalid();
            return new(CloseStageReceiptStatus.ValidPrefix, records.AsReadOnly(), closed && health);
        }
        catch (Exception e) when (e is JsonException or InvalidOperationException or KeyNotFoundException or
            FormatException or OverflowException or IOException or DecoderFallbackException or ArgumentException)
        { return Invalid(); }
    }

    private static bool Follows(CloseStage? prior, CloseStage next) => prior switch
    {
        null => next == CloseStage.CloseEntry,
        CloseStage.CloseEntry => next == CloseStage.DisposeEntry,
        CloseStage.DisposeEntry => next is CloseStage.AlreadyDisposedReturn or CloseStage.RootLoadedBefore,
        CloseStage.AlreadyDisposedReturn => next == CloseStage.CloseReturn,
        CloseStage.WindowClosedAfter => next is CloseStage.XamlRootBefore or CloseStage.XamlRootSkipped,
        CloseStage.XamlRootAfter or CloseStage.XamlRootSkipped => next is CloseStage.UISettingsBefore or CloseStage.UISettingsSkipped,
        CloseStage.UISettingsAfter or CloseStage.UISettingsSkipped => next == CloseStage.RootClearBefore,
        CloseStage.CloseReturn => false,
        _ => next == prior + 1
    };
}
