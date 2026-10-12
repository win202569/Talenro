package stateformat

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func formatImportsAllowed(source []byte) bool {
	f, e := parser.ParseFile(token.NewFileSet(), "", source, parser.ImportsOnly)
	if e != nil {
		return false
	}
	// Exact, deliberately narrow allowlist. Additional pure standard imports
	// require review here; subpackages never inherit permission from a prefix.
	allowed := map[string]bool{}
	for _, path := range []string{"bytes", "crypto/sha256", "encoding/base64", "encoding/hex", "encoding/json", "errors", "io", "sort", "strconv", "strings", "time", "unicode/utf8", "github.com/google/uuid", "github.com/gowebpki/jcs", "talenro.local/platform/internal/nodecontrol/contracts"} {
		allowed[path] = true
	}
	for _, i := range f.Imports {
		path, e := strconv.Unquote(i.Path.Value)
		if e != nil || !allowed[path] {
			return false
		}
	}
	return true
}

func TestFormatDependencyBoundary(t *testing.T) {
	for _, path := range []string{"net", "net/http", "database/sql", "crypto/ed25519", "crypto/rsa", "talenro.local/platform/internal/nodecontrol/state", "talenro.local/platform/internal/nodecontrol/serving", "talenro.local/platform/internal/nodecontrol/authority", "talenro.local/platform/internal/store", "talenro.local/platform/internal/trust", "talenro.local/platform/internal/nodecontrol/contracts/unsafe", "github.com/google/uuid/other"} {
		for _, alias := range []string{"", "x ", "_ "} {
			source := []byte("package example\nimport " + alias + "\"" + path + "\"\n")
			if formatImportsAllowed(source) {
				t.Errorf("allowed forbidden import %s", path)
			}
		}
	}
	if !formatImportsAllowed([]byte("package example\nimport(\"bytes\";\"crypto/sha256\";\"github.com/google/uuid\")")) {
		t.Fatal("approved pure dependencies rejected")
	}
	if formatImportsAllowed([]byte("not go source")) {
		t.Fatal("malformed source accepted")
	}
	files, e := filepath.Glob("*.go")
	if e != nil {
		t.Fatal(e)
	}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		b, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		if !formatImportsAllowed(b) {
			t.Errorf("production import boundary: %s", path)
		}
	}
}
