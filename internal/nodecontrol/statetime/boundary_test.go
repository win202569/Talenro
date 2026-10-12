package statetime

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

func timeImportsAllowed(src []byte) bool {
	f, e := parser.ParseFile(token.NewFileSet(), "", src, 0)
	if e != nil {
		return false
	}
	allowed := map[string]bool{"context": true, "crypto/ed25519": true, "errors": true, "time": true, "github.com/google/uuid": true, "talenro.local/platform/internal/nodecontrol/contracts": true, "talenro.local/platform/internal/nodecontrol/stateformat": true}
	timeAliases := map[string]bool{}
	for _, im := range f.Imports {
		p, e := strconv.Unquote(im.Path.Value)
		if e != nil || !allowed[p] {
			return false
		}
		if im.Name != nil && (im.Name.Name == "." || im.Name.Name == "_") {
			return false
		}
		if p == "time" {
			timeAlias := "time"
			if im.Name != nil {
				timeAlias = im.Name.Name
			}
			timeAliases[timeAlias] = true
		}
	}
	ok := true
	ast.Inspect(f, func(n ast.Node) bool {
		if s, yes := n.(*ast.SelectorExpr); yes {
			if x, yes := s.X.(*ast.Ident); yes && timeAliases[x.Name] {
				if s.Sel.Name != "Time" && s.Sel.Name != "Duration" && s.Sel.Name != "Minute" {
					ok = false
				}
			}
		}
		return true
	})
	return ok
}

func TestTimeBindingBoundary(t *testing.T) {
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
		if got := timeConsumerAllowed(filepath.FromSlash(tc.path), tc.pkg, "talenro.local/platform/internal/nodecontrol/statetime"); got != tc.want {
			t.Errorf("%s %s: %v", tc.path, tc.pkg, got)
		}
	}

}
func timeConsumerAllowed(path, packageName, importPath string) bool {
	if packageName == "statetimebinding" && filepath.Dir(filepath.Clean(path)) == filepath.FromSlash("internal/nodecontrol/statetimebinding") {
		return true
	}
	return importPath != "talenro.local/platform/internal/nodecontrol/statetime"
}
func TestTimeBoundary(t *testing.T) {
	for _, src := range []string{`package p; import "net"`, `package p; import "os"`, `package p; import "database/sql"`, `package p; import "talenro.local/platform/internal/nodecontrol/statepolicy"`, `package p; import "talenro.local/platform/internal/nodecontrol/stateverify"`, `package p; import "time"; var x=time.Now`, `package p; import clock "time"; var x=clock.Now`, `package p; import . "time"; var x=Now`, `package p; import "time"; var x=time.NewTimer`} {
		if timeImportsAllowed([]byte(src)) {
			t.Fatal("allowed", src)
		}
	}
	if !timeImportsAllowed([]byte(`package p; import "time"; var x time.Time`)) {
		t.Fatal("pure type")
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
		if e != nil || !timeImportsAllowed(b) {
			t.Fatal(p, e)
		}
	}
	for _, p := range []string{"cmd/main.go", "internal/nodecontrol/statetime_extra/x.go", "internal/nodecontrol/statepolicy/x.go"} {
		if timeConsumerAllowed(p, "example", "talenro.local/platform/internal/nodecontrol/statetime") {
			t.Fatal("consumer", p)
		}
	}
	if !timeConsumerAllowed("cmd/main.go", "main", "context") {
		t.Fatal("other import")
	}
	e = filepath.WalkDir("../../..", func(p string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == ".worktrees" || d.Name() == "vendor" {
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
		rel, e := filepath.Rel("../../..", p)
		if e != nil {
			return e
		}
		for _, im := range f.Imports {
			s, _ := strconv.Unquote(im.Path.Value)
			if !timeConsumerAllowed(rel, f.Name.Name, s) {
				t.Errorf("production consumer %s", p)
			}
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
}

// Every alias must be checked, not just the last import of the package.
func TestTimeBoundaryMultipleAliases(t *testing.T) {
	for _, src := range []string{
		`package p; import(a "time"; b "time"); var x=a.Now; var y b.Time`,
		`package p; import(a "time"; b "time"); var x=a.Now(); var y b.Time`,
		`package p; import(a "time"; b "time"); var x=b.Now; var y a.Time`,
	} {
		if timeImportsAllowed([]byte(src)) {
			t.Fatal("ambient clock accepted through duplicate aliases")
		}
	}
	if !timeImportsAllowed([]byte(`package p; import(a "time"; b "time"); var x a.Time; var y b.Time`)) {
		t.Fatal("pure types rejected")
	}
}
