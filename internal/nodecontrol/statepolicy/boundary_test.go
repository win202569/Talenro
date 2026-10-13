package statepolicy

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestNodeLeafBindingConsumerException(t *testing.T) {
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	dir := filepath.FromSlash("internal/nodecontrol/nodeleafbinding")
	target := "talenro.local/platform/internal/nodecontrol/statepolicy"
	for _, c := range []struct {
		path, pkg string
		want      bool
	}{
		{filepath.Join(dir, "a.go"), "nodeleafbinding", true},
		{filepath.Join(root, dir, "a.go"), "nodeleafbinding", true},
		{filepath.Join(dir, "a_test.go"), "nodeleafbinding", true},
		{filepath.Join(root, dir, "a_test.go"), "nodeleafbinding", true},
		{filepath.Join(dir, "nested", "a.go"), "nodeleafbinding", false},
		{filepath.FromSlash("internal/nodecontrol/nodeleafbinding_extra/a.go"), "nodeleafbinding", false},
		{filepath.Join("else", dir, "a.go"), "nodeleafbinding", false},
		{filepath.Join(root, "..", "external", dir, "a.go"), "nodeleafbinding", false},
		{filepath.Join(dir, "a.go"), "other", false},
		{filepath.Join(dir, "a_test.go"), "nodeleafbinding_test", false},
		{"internal/nodecontrol/brother/a_test.go", "nodeleafbinding", false},
	} {
		if got := policyConsumerAllowed(c.path, c.pkg, target); got != c.want {
			t.Errorf("%s %s: got %v want %v", c.path, c.pkg, got, c.want)
		}
	}
	if policyConsumerAllowed(filepath.Join(dir, "a.go"), "nodeleafbinding", target+"/child") {
		t.Error("subpath")
	}
}

func TestNodeCertTimeConsumerException(t *testing.T) {
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	dir := filepath.FromSlash("internal/nodecontrol/nodecerttime")
	target := "talenro.local/platform/internal/nodecontrol/statepolicy"
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
		if got := policyConsumerAllowed(c.path, c.pkg, target); got != c.want {
			t.Errorf("%s %s: got %v want %v", c.path, c.pkg, got, c.want)
		}
	}
	if policyConsumerAllowed(filepath.Join(dir, "a.go"), "nodecerttime", target+"/child") {
		t.Error("subpath")
	}

}

func policyConsumerAllowed(path, pkg, imported string) bool {
	// Only the exact selected-leaf projection consumer is added.
	leafBindingTarget := "talenro.local/platform/internal/nodecontrol/statepolicy"
	if imported == leafBindingTarget || strings.HasPrefix(imported, leafBindingTarget+"/") {
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
		exact := e == nil && filepath.Dir(rel) == filepath.FromSlash("internal/nodecontrol/nodeleafbinding")
		if exact || pkg == "nodeleafbinding" || pkg == "nodeleafbinding_test" {
			return imported == leafBindingTarget && exact && pkg == "nodeleafbinding"
		}
	}
	// Only the new exact consumer is added; all previous rules below remain.
	certTimeTarget := "talenro.local/platform/internal/nodecontrol/statepolicy"
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
	return imported != "talenro.local/platform/internal/nodecontrol/statepolicy" ||
		(pkg == "statepolicybinding" && filepath.Dir(filepath.Clean(path)) == filepath.FromSlash("internal/nodecontrol/statepolicybinding"))
}

func TestPolicyTimeBindingBoundary(t *testing.T) {
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
		if got := policyConsumerAllowed(filepath.FromSlash(tc.path), tc.pkg, "talenro.local/platform/internal/nodecontrol/statepolicy"); got != tc.want {
			t.Errorf("%s %s: %v", tc.path, tc.pkg, got)
		}
	}

}
func TestPolicyBindingConsumerBoundary(t *testing.T) {
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
		if got := policyConsumerAllowed(filepath.FromSlash(tc.path), tc.pkg, "talenro.local/platform/internal/nodecontrol/statepolicy"); got != tc.want {
			t.Errorf("%s: %v", tc.path, got)
		}
	}

}

func TestPolicyBoundary(t *testing.T) {
	allowed := map[string]bool{"context": true, "time": true, "bytes": true, "crypto/sha256": true, "sort": true, "math": true, "reflect": true, "github.com/google/uuid": true, "talenro.local/platform/internal/nodecontrol/contracts": true, "talenro.local/platform/internal/nodecontrol/stateformat": true, "talenro.local/platform/internal/nodecontrol/stateverify": true}
	files, e := filepath.Glob("*.go")
	if e != nil {
		t.Fatal(e)
	}
	count := 0
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		count++
		f, e := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if e != nil {
			t.Fatal(e)
		}
		timeName := "time"
		for _, i := range f.Imports {
			p, e := strconv.Unquote(i.Path.Value)
			if e != nil || !allowed[p] {
				t.Fatalf("unexpected production dependency %s", i.Path.Value)
			}
			if p == "time" && i.Name != nil {
				timeName = i.Name.Name
				if timeName == "." {
					t.Fatal("dot time import")
				}
			}
		}
		ast.Inspect(f, func(n ast.Node) bool {
			if s, ok := n.(*ast.SelectorExpr); ok {
				if x, ok := s.X.(*ast.Ident); ok && x.Name == timeName {
					switch s.Sel.Name {
					case "Now", "Since", "Until", "Sleep", "After", "AfterFunc", "NewTimer", "NewTicker", "Tick":
						t.Errorf("ambient clock %s", s.Sel.Name)
					}
				}
			}
			if fn, ok := n.(*ast.FuncDecl); ok && fn.Recv != nil {
				switch fn.Name.Name {
				case "ObserveDeployment", "ObserveTime", "ObserveState":
					t.Error("production observer implementation")
				}
			}
			return true
		})
	}
	if count == 0 {
		t.Fatal("missing implementation")
	}
	// Inspect import declarations, not arbitrary string literals in source.
	e = filepath.WalkDir("../../..", func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
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
			p, _ := strconv.Unquote(i.Path.Value)
			rel, relErr := filepath.Rel("../../..", path)
			if relErr != nil {
				return relErr
			}
			if !policyConsumerAllowed(rel, f.Name.Name, p) {
				t.Errorf("production caller %s", path)
			}
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
}
