using System.Text;
using System.Text.Json;
namespace Talenro.Demo.App.Layout;

internal sealed class LayoutFixtureCatalog
{
    private readonly Dictionary<string, string> texts;
    private LayoutFixtureCatalog(Dictionary<string, string> texts) => this.texts = texts;
    public static LayoutFixtureCatalog Parse(string json)
    {
        try
        {
            if (Encoding.UTF8.GetByteCount(json) > 65536) throw new InvalidDataException("Fixture too large.");
            using var document = JsonDocument.Parse(json, new JsonDocumentOptions { MaxDepth = 8 });
            var root = document.RootElement;
            UniqueProperties(root);
            var cases = root.GetProperty("cases");
            if (cases.ValueKind != JsonValueKind.Array) throw new InvalidDataException("Invalid fixture cases.");
            var values = new Dictionary<string, string>(StringComparer.Ordinal);
            foreach (var item in cases.EnumerateArray())
            {
                UniqueProperties(item);
                var id = item.GetProperty("id").GetString();
                var text = item.GetProperty("text").GetString();
                if (id is not ("long-status" or "long-country" or "rtl-mixed" or "unknown-usage") ||
                    string.IsNullOrWhiteSpace(text) || text.Length > 4096 || !values.TryAdd(id, text))
                    throw new InvalidDataException("Invalid fixture entry.");
                var hasDirection = item.TryGetProperty("direction", out var direction);
                if (id == "rtl-mixed" ? !hasDirection || direction.GetString() != "RightToLeft" : hasDirection)
                    throw new InvalidDataException("Invalid fixture direction.");
            }
            if (values.Count != 4) throw new InvalidDataException("Missing fixture.");
            return new(values);
        }
        catch (Exception error) when (error is JsonException or InvalidOperationException or KeyNotFoundException or ArgumentException)
        { throw new InvalidDataException("Invalid layout fixture.", error); }
    }
    public string Text(string id) => texts.TryGetValue(id, out var text) ? text : throw new InvalidDataException("Missing fixture.");
    private static void UniqueProperties(JsonElement element)
    {
        var names = new HashSet<string>(StringComparer.Ordinal);
        foreach (var property in element.EnumerateObject())
            if (!names.Add(property.Name)) throw new InvalidDataException("Duplicate fixture property.");
    }
}
