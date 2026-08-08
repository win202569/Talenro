package platform

import (
	"bytes"
	"log/slog"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestHTTPServerReportsSanitizedInternalErrors(t *testing.T) {
	var logs bytes.Buffer
	previous := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&logs, nil)))
	t.Cleanup(func() { slog.SetDefault(previous) })

	server := NewHTTPServer("unused", http.NewServeMux())
	if server.ErrorLog == nil {
		t.Fatal("ErrorLog is nil; net/http would use its raw default logger")
	}
	server.ErrorLog.Print("panic serving credential@private.example:8443 with stack trace")

	output := logs.String()
	if !strings.Contains(output, "msg=http_server_error") || !strings.Contains(output, "category=http_internal") {
		t.Fatalf("sanitized event/category missing from log: %q", output)
	}
	for _, private := range []string{"credential", "private.example", "8443", "stack trace"} {
		if strings.Contains(output, private) {
			t.Fatalf("private net/http detail %q leaked in log: %q", private, output)
		}
	}
}

func TestNewHTTPServerConfiguresResourceBounds(t *testing.T) {
	handler := http.NewServeMux()
	server := NewHTTPServer("127.0.0.1:8080", handler)

	if server.Addr != "127.0.0.1:8080" {
		t.Fatalf("address = %q, want %q", server.Addr, "127.0.0.1:8080")
	}
	if server.Handler != handler {
		t.Fatal("handler was not retained")
	}
	if server.ReadHeaderTimeout != 5*time.Second {
		t.Fatalf("read header timeout = %s, want 5s", server.ReadHeaderTimeout)
	}
	if server.ReadTimeout != 10*time.Second {
		t.Fatalf("read timeout = %s, want 10s", server.ReadTimeout)
	}
	if server.WriteTimeout != 10*time.Second {
		t.Fatalf("write timeout = %s, want 10s", server.WriteTimeout)
	}
	if server.IdleTimeout != 60*time.Second {
		t.Fatalf("idle timeout = %s, want 1m", server.IdleTimeout)
	}
	if server.MaxHeaderBytes != 16<<10 {
		t.Fatalf("max header bytes = %d, want %d", server.MaxHeaderBytes, 16<<10)
	}
}
