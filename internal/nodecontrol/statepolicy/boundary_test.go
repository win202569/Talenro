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

func policyConsumerAllowed(path, pkg, imported string) bool {
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
