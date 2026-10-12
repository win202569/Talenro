package statecontract

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

func TestContractPolicyBindingBoundary(t *testing.T) {
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
		if got := contractConsumerAllowed(filepath.FromSlash(tc.path), tc.pkg, "talenro.local/platform/internal/nodecontrol/statecontract"); got != tc.want {
			t.Errorf("%s: %v", tc.path, got)
		}
	}

}

func TestContractTimeBindingBoundary(t *testing.T) {
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
		if got := contractConsumerAllowed(filepath.FromSlash(tc.path), tc.pkg, "talenro.local/platform/internal/nodecontrol/statecontract"); got != tc.want {
			t.Errorf("%s %s: %v", tc.path, tc.pkg, got)
		}
	}

}

func TestContractBindingBoundary(t *testing.T) {
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
		if got := contractConsumerAllowed(filepath.FromSlash(tc.path), tc.pkg, "talenro.local/platform/internal/nodecontrol/statecontract"); got != tc.want {
			t.Errorf("%s: %v", tc.path, got)
		}
	}
}

func contractConsumerAllowed(path, pkg, imported string) bool {
	if pkg == "statetimebinding" && filepath.Dir(filepath.Clean(path)) == filepath.FromSlash("internal/nodecontrol/statetimebinding") {
		return true
	}
	if imported != "talenro.local/platform/internal/nodecontrol/statecontract" {
		return true
	}
	return (pkg == "statebinding" && filepath.Dir(filepath.Clean(path)) == filepath.FromSlash("internal/nodecontrol/statebinding")) ||
		(pkg == "statepolicybinding" && filepath.Dir(filepath.Clean(path)) == filepath.FromSlash("internal/nodecontrol/statepolicybinding"))
}
func contractImportsAllowed(src []byte) bool {
	allowed := map[string]map[string]bool{
		"context": {"Context": true, "Canceled": true, "DeadlineExceeded": true},
		"errors":  {"New": true}, "bytes": {"Compare": true, "Equal": true},
		"strings": {"Contains": true, "HasPrefix": true, "HasSuffix": true, "Split": true, "ToLower": true},
		"reflect": {"DeepEqual": true}, "github.com/google/uuid": {"Parse": true, "Nil": true, "RFC4122": true},
		"talenro.local/platform/internal/nodecontrol/contracts": {"Digest": true},
	}
	f, e := parser.ParseFile(token.NewFileSet(), "guard.go", src, 0)
	if e != nil {
		return false
	}
	aliases := map[string]map[string]bool{}
	for _, i := range f.Imports {
		p, e := strconv.Unquote(i.Path.Value)
		if e != nil {
			return false
		}
		members, ok := allowed[p]
		if !ok {
			return false
		}
		parts := strings.Split(p, "/")
		alias := parts[len(parts)-1]
		if i.Name != nil {
			alias = i.Name.Name
		}
		if alias == "." || alias == "_" {
			return false
		}
		if _, exists := aliases[alias]; exists {
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
func TestContractBoundary(t *testing.T) {
	for _, tc := range []struct {
		source string
		want   bool
	}{
		{`package p; import "context"; var _ context.Context`, true},
		{`package p; import u "github.com/google/uuid"; var _=u.Parse`, true},
		{`package p; import u "github.com/google/uuid"; var _=u.New`, false},
		{`package p; import (u "github.com/google/uuid"; v "github.com/google/uuid");var _=u.Parse;var _=v.New`, false},
		{`package p; import . "github.com/google/uuid"`, false},
		{`package p; import _ "os"`, false},
		{`package p; import "time"; var _=time.Now`, false},
		{`package p; import "context"; var _=context.WithCancel`, false},
		{`package p; import "talenro.local/platform/internal/nodecontrol/statepolicy"`, false},
		{`not Go`, false},
	} {
		if contractImportsAllowed([]byte(tc.source)) != tc.want {
			t.Fatal("import guard case", tc.source)
		}
	}
	if contractConsumerAllowed("internal/nodecontrol/serving/x.go", "serving", "talenro.local/platform/internal/nodecontrol/statecontract") || !contractConsumerAllowed("internal/nodecontrol/serving/x.go", "serving", "talenro.local/platform/internal/nodecontrol/contracts") {
		t.Fatal("consumer classification")
	}
	files, e := filepath.Glob("*.go")
	if e != nil {
		t.Fatal(e)
	}
	for _, p := range files {
		if strings.HasSuffix(p, "_test.go") {
			continue
		}
		b, e := os.ReadFile(p)
		if e != nil {
			t.Fatal(e)
		}
		if !contractImportsAllowed(b) {
			t.Fatal("product dependency", p)
		}
	}
	root := filepath.Clean("../../..")
	e = filepath.WalkDir(root, func(p string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", ".worktrees", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return nil
		}
		f, e := parser.ParseFile(token.NewFileSet(), p, nil, parser.ImportsOnly)
		if e != nil {
			return e
		}
		for _, i := range f.Imports {
			s, e := strconv.Unquote(i.Path.Value)
			if e != nil {
				return e
			}
			rel, relErr := filepath.Rel(root, p)
			if relErr != nil {
				return relErr
			}
			if !contractConsumerAllowed(rel, f.Name.Name, s) {
				t.Errorf("production consumer: %s", p)
			}
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
}
