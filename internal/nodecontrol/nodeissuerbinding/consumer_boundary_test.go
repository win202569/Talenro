package nodeissuerbinding

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestIssuerBindingConsumers(t *testing.T) {
	const imported = "talenro.local/platform/internal/nodecontrol/nodeissuerbinding"
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	dir := filepath.Join(root, "internal/nodecontrol/nodeissuerbinding")
	for _, c := range []struct {
		path, pkg, imp string
		want           bool
	}{
		{filepath.Join(dir, "a_test.go"), "nodeissuerbinding", imported, true},
		{"internal/nodecontrol/nodeissuerbinding/a_test.go", "nodeissuerbinding", imported, true},
		{filepath.Join(dir, "a.go"), "nodeissuerbinding", imported, false},
		{filepath.Join(dir, "a_test.go"), "nodeissuerbinding_test", imported, false},
		{filepath.Join(dir, "nested/a_test.go"), "nodeissuerbinding", imported, false},
		{filepath.Join(root, "internal/nodecontrol/nodeissuerbinding_sibling/a_test.go"), "nodeissuerbinding", imported, false},
		{filepath.Join(root, "else/internal/nodecontrol/nodeissuerbinding/a_test.go"), "nodeissuerbinding", imported, false},
		{filepath.Join(root, "../external/internal/nodecontrol/nodeissuerbinding/a_test.go"), "nodeissuerbinding", imported, false},
		{filepath.Join(dir, "a_test.go"), "nodeissuerbinding", imported + "/child", false},
		{filepath.Join(dir, "a_test.go"), "nodeissuerbinding", "other", false},
	} {
		if got := issuerBindingConsumerAllowed(c.path, c.pkg, c.imp); got != c.want {
			t.Fatal(c, got)
		}
	}
	e = filepath.WalkDir(root, func(path string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			switch d.Name() {
			case ".git", ".superpowers", ".worktrees", "vendor", "node_modules":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		f, e := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if e != nil {
			return e
		}
		for _, im := range f.Imports {
			p, e := strconv.Unquote(im.Path.Value)
			if e != nil {
				return e
			}
			if p == imported || strings.HasPrefix(p, imported+"/") {
				if !issuerBindingConsumerAllowed(path, f.Name.Name, p) {
					t.Errorf("unauthorized consumer %s %s", path, p)
				}
			}
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
}
func issuerBindingConsumerAllowed(path, pkg, imported string) bool {
	if imported != "talenro.local/platform/internal/nodecontrol/nodeissuerbinding" || pkg != "nodeissuerbinding" || !strings.HasSuffix(path, "_test.go") {
		return false
	}
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
	return e == nil && filepath.Dir(rel) == filepath.FromSlash("internal/nodecontrol/nodeissuerbinding")
}
