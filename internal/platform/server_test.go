package platform

import (
	"net/http"
	"testing"
	"time"
)

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
