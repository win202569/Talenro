package statematerials

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestSubsetPolicyBindingBoundary(t *testing.T) {
	for _, tc := range []struct {
		path, pkg string
		want      bool
	}{
		{"internal/nodecontrol/statepolicybinding/x.go", "statepolicybinding", true},
		{"internal/nodecontrol/statepolicybinding/sub/x.go", "statepolicybinding", false},
		{"internal/nodecontrol/statepolicybinding-other/x.go", "statepolicybinding", false},
		{"internal/nodecontrol/statepolicybinding/x.go", "other", false},
		{"internal/nodecontrol/serving/x.go", "statepolicybinding", false},
		{"internal/nodecontrol/reader/x.go", "reader", false},
		{"internal/nodecontrol/authority/x.go", "authority", false},
		{"internal/store/x.go", "store", false},
	} {
		if got := materialsConsumerAllowed(filepath.FromSlash(tc.path), tc.pkg, "talenro.local/platform/internal/nodecontrol/statematerials"); got != tc.want {
			t.Errorf("%s: %v", tc.path, got)
		}
	}

}

func TestSubsetTimeBindingBoundary(t *testing.T) {
	for _, tc := range []struct {
		path, pkg string
		want      bool
	}{
		{"internal/nodecontrol/statetimebinding/x.go", "statetimebinding", true},
		{"internal/nodecontrol/statetimebinding/sub/x.go", "statetimebinding", false},
		{"internal/nodecontrol/statetimebinding-other/x.go", "statetimebinding", false},
		{"internal/nodecontrol/statetimebinding/x.go", "wrong", false},
		{"internal/nodecontrol/serving/x.go", "statetimebinding", false},
		{"internal/nodecontrol/reader/x.go", "reader", false},
		{"internal/nodecontrol/authority/x.go", "authority", false},
		{"internal/store/x.go", "store", false},
	} {
		if got := materialsConsumerAllowed(filepath.FromSlash(tc.path), tc.pkg, "talenro.local/platform/internal/nodecontrol/statematerials"); got != tc.want {
			t.Errorf("%s %s: %v", tc.path, tc.pkg, got)
		}
	}

}

func TestSubsetBindingBoundary(t *testing.T) {
	for _, tc := range []struct {
		path, pkg string
		want      bool
	}{
		{"internal/nodecontrol/statebinding/x.go", "statebinding", true},
		{"internal/nodecontrol/statebinding/sub/x.go", "statebinding", false},
		{"internal/nodecontrol/statebinding-other/x.go", "statebinding", false},
		{"internal/nodecontrol/statebinding/x.go", "other", false},
		{"internal/nodecontrol/serving/x.go", "serving", false},
		{"internal/nodecontrol/reader/x.go", "reader", false},
	} {
		if got := materialsConsumerAllowed(filepath.FromSlash(tc.path), tc.pkg, "talenro.local/platform/internal/nodecontrol/statematerials"); got != tc.want {
			t.Errorf("%s: %v", tc.path, got)
		}
	}
}

func TestSubsetCompositionBoundary(t *testing.T) {
	for _, tc := range []struct {
		path, pkg string
		want      bool
	}{
		{"internal/nodecontrol/statecomposition/evaluate.go", "statecomposition", true},
		{"internal/nodecontrol/statecomposition/sub/x.go", "statecomposition", false},
		{"internal/nodecontrol/statecomposition-other/x.go", "statecomposition", false},
		{"internal/nodecontrol/statecomposition/x.go", "other", false},
		{"internal/nodecontrol/serving/reader.go", "serving", false},
	} {
		if got := materialsConsumerAllowed(filepath.FromSlash(tc.path), tc.pkg, "talenro.local/platform/internal/nodecontrol/statematerials"); got != tc.want {
			t.Errorf("%s: %v", tc.path, got)
		}
	}
}

func TestSubsetBoundary(t *testing.T) {
	for _, src := range []string{
		`package p; import "os"`, `package p; import "net"`, `package p; import "time"`, `package p; import "crypto/sha256"`,
		`package p; import "talenro.local/platform/internal/nodecontrol/statecontract"`, `package p; import "talenro.local/platform/internal/nodecontrol/stateverify"`, `package p; import "talenro.local/platform/internal/nodecontrol/statepolicy"`,
		`package p; import u "github.com/google/uuid"; var _=u.New`, `package p; import "context"; var _=context.WithCancel`,
		`package p; import . "bytes"`, `package p; import _ "bytes"`, `package p; import (a "bytes"; b "bytes");var _=a.Equal;var _=b.Clone`, `broken`,
	} {
		if materialsImportsAllowed([]byte(src)) {
			t.Fatal("forbidden source accepted", src)
		}
	}
	if !materialsImportsAllowed([]byte(`package p;import "bytes";var _=bytes.Equal`)) {
		t.Fatal("allowed source rejected")
	}
	files, e := filepath.Glob("*.go")
	if e != nil {
		t.Fatal(e)
	}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		src, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		if !materialsImportsAllowed(src) {
			t.Fatal("production dependency", path)
		}
	}
	e = filepath.WalkDir("../../..", func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == ".worktrees" || d.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		f, e := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if e != nil {
			return e
		}
		for _, i := range f.Imports {
			p, e := strconv.Unquote(i.Path.Value)
			if e != nil {
				return e
			}
			rel, relErr := filepath.Rel("../../..", path)
			if relErr != nil {
				return relErr
			}
			if !materialsConsumerAllowed(rel, f.Name.Name, p) {
				t.Error("production consumer", path)
			}
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
}
func TestNodeCertTimeConsumerException(t *testing.T) {
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	dir := filepath.FromSlash("internal/nodecontrol/nodecerttime")
	target := "talenro.local/platform/internal/nodecontrol/statematerials"
	cases := []struct {
		path, pkg string
		want      bool
	}{
		{filepath.Join(dir, "a.go"), "nodecerttime", true},
		{filepath.Join(root, dir, "a.go"), "nodecerttime", true},
		{filepath.Join(dir, "a_test.go"), "nodecerttime", true},
		{filepath.Join(root, dir, "a_test.go"), "nodecerttime", true},
		{filepath.Join(dir, "nested", "a.go"), "nodecerttime", false},
		{filepath.FromSlash("internal/nodecontrol/nodecerttime_extra/a.go"), "nodecerttime", false},
		{filepath.Join("else", dir, "a.go"), "nodecerttime", false},
		{filepath.Join(root, "..", "external", dir, "a.go"), "nodecerttime", false},
		{filepath.Join(dir, "a.go"), "other", false},
		{filepath.Join(dir, "a_test.go"), "nodecerttime_test", false},
		{"internal/nodecontrol/brother/a_test.go", "nodecerttime", false},
	}
	for _, c := range cases {
		if got := materialsConsumerAllowed(c.path, c.pkg, target); got != c.want {
			t.Errorf("%s %s: got %v want %v", c.path, c.pkg, got, c.want)
		}
	}
	if materialsConsumerAllowed(filepath.Join(dir, "a.go"), "nodecerttime", target+"/child") {
		t.Error("subpath")
	}

}

func materialsConsumerAllowed(path, pkg, imported string) bool {
	// Only the new exact consumer is added; all previous rules below remain.
	certTimeTarget := "talenro.local/platform/internal/nodecontrol/statematerials"
	if imported == certTimeTarget || strings.HasPrefix(imported, certTimeTarget+"/") {
		root, e := filepath.Abs("../../..")
		if e != nil {
			return false
		}
		normalized := filepath.Clean(path)
		if !filepath.IsAbs(normalized) {
			if strings.HasPrefix(filepath.ToSlash(normalized), "internal/") {
				normalized = filepath.Join(root, normalized)
			} else {
				normalized, e = filepath.Abs(normalized)
				if e != nil {
					return false
				}
			}
		}
		rel, e := filepath.Rel(root, normalized)
		exact := e == nil && filepath.Dir(rel) == filepath.FromSlash("internal/nodecontrol/nodecerttime")
		if exact || pkg == "nodecerttime" || pkg == "nodecerttime_test" {
			return imported == certTimeTarget && exact && pkg == "nodecerttime"
		}
	}

	if pkg == "statetimebinding" && filepath.Dir(filepath.Clean(path)) == filepath.FromSlash("internal/nodecontrol/statetimebinding") {
		return true
	}
	if imported != "talenro.local/platform/internal/nodecontrol/statematerials" {
		return true
	}
	dir := filepath.Dir(filepath.Clean(path))
	return (pkg == "statecomposition" && dir == filepath.FromSlash("internal/nodecontrol/statecomposition")) ||
		(pkg == "statebinding" && dir == filepath.FromSlash("internal/nodecontrol/statebinding")) ||
		(pkg == "statepolicybinding" && dir == filepath.FromSlash("internal/nodecontrol/statepolicybinding"))
}
func materialsImportsAllowed(src []byte) bool {
	allowed := map[string]map[string]bool{
		"context": {"Context": true, "Canceled": true, "DeadlineExceeded": true},
		"errors":  {"New": true}, "bytes": {"Compare": true, "Equal": true, "Clone": true}, "reflect": {"DeepEqual": true}, "sort": {"Slice": true},
		"github.com/google/uuid":                                  {"Parse": true, "Nil": true, "RFC4122": true},
		"talenro.local/platform/internal/nodecontrol/contracts":   {"Digest": true},
		"talenro.local/platform/internal/nodecontrol/stateformat": {"ParsedRootSet": true, "ParsedArtifact": true, "ParseRootSet": true, "ParseEnvelope": true, "PublicKeyMaterial": true, "Kind": true, "RootRotation": true, "Metadata": true, "Desired": true, "ErrInvalidFormat": true},
	}
	f, e := parser.ParseFile(token.NewFileSet(), "guard.go", src, 0)
	if e != nil {
		return false
	}
	aliases := map[string]map[string]bool{}
	seen := map[string]bool{}
	for _, i := range f.Imports {
		path, e := strconv.Unquote(i.Path.Value)
		if e != nil {
			return false
		}
		members, ok := allowed[path]
		if !ok || seen[path] {
			return false
		}
		seen[path] = true
		parts := strings.Split(path, "/")
		alias := parts[len(parts)-1]
		if i.Name != nil {
			alias = i.Name.Name
		}
		if alias == "." || alias == "_" || aliases[alias] != nil {
			return false
		}
		aliases[alias] = members
	}
	valid := true
	ast.Inspect(f, func(n ast.Node) bool {
		if s, ok := n.(*ast.SelectorExpr); ok {
			if id, ok := s.X.(*ast.Ident); ok {
				if members, exists := aliases[id.Name]; exists && !members[s.Sel.Name] {
					valid = false
				}
			}
		}
		return true
	})
	return valid
}
