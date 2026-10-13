package nodeleafbinding

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func leafBindingConsumerAllowed(path, pkg, imported string) bool {
	issuerBindingTarget := "talenro.local/platform/internal/nodecontrol/nodeleafbinding"
	if imported == issuerBindingTarget || strings.HasPrefix(imported, issuerBindingTarget+"/") {
		root, err := filepath.Abs("../../..")
		if err != nil {
			return false
		}
		normalized := filepath.Clean(path)
		if !filepath.IsAbs(normalized) {
			if strings.HasPrefix(filepath.ToSlash(normalized), "internal/") {
				normalized = filepath.Join(root, normalized)
			} else {
				normalized, err = filepath.Abs(normalized)
				if err != nil {
					return false
				}
			}
		}
		rel, err := filepath.Rel(root, normalized)
		exact := err == nil && filepath.Dir(rel) == filepath.FromSlash("internal/nodecontrol/nodeissuerbinding")
		if exact || pkg == "nodeissuerbinding" || pkg == "nodeissuerbinding_test" {
			return imported == issuerBindingTarget && exact && pkg == "nodeissuerbinding"
		}
	}
	if imported != "talenro.local/platform/internal/nodecontrol/nodeleafbinding" || pkg != "nodeleafbinding" || !strings.HasSuffix(path, "_test.go") {
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
	return e == nil && filepath.Dir(rel) == filepath.FromSlash("internal/nodecontrol/nodeleafbinding")
}
func TestLeafBindingConsumers(t *testing.T) {
	const imported = "talenro.local/platform/internal/nodecontrol/nodeleafbinding"
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	dir := filepath.Join(root, "internal/nodecontrol/nodeleafbinding")
	for _, c := range []struct {
		path, pkg, imp string
		want           bool
	}{
		{filepath.Join(dir, "a_test.go"), "nodeleafbinding", imported, true},
		{"internal/nodecontrol/nodeleafbinding/a_test.go", "nodeleafbinding", imported, true},
		{filepath.Join(dir, "a.go"), "nodeleafbinding", imported, false},
		{filepath.Join(dir, "a_test.go"), "nodeleafbinding_test", imported, false},
		{filepath.Join(dir, "nested/a_test.go"), "nodeleafbinding", imported, false},
		{filepath.Join(root, "internal/nodecontrol/nodeleafbinding_sibling/a_test.go"), "nodeleafbinding", imported, false},
		{filepath.Join(root, "else/internal/nodecontrol/nodeleafbinding/a_test.go"), "nodeleafbinding", imported, false},
		{filepath.Join(root, "../external/internal/nodecontrol/nodeleafbinding/a_test.go"), "nodeleafbinding", imported, false},
		{filepath.Join(dir, "a_test.go"), "nodeleafbinding", imported + "/child", false},
		{filepath.Join(dir, "a_test.go"), "nodeleafbinding", "other", false},
	} {
		if got := leafBindingConsumerAllowed(c.path, c.pkg, c.imp); got != c.want {
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
				if !leafBindingConsumerAllowed(path, f.Name.Name, p) {
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
func TestLeafBindingIssuerBindingBoundary(t *testing.T) {
	const target = "talenro.local/platform/internal/nodecontrol/nodeleafbinding"
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	dir := filepath.Join(root, "internal/nodecontrol/nodeissuerbinding")
	for _, c := range []struct {
		path, pkg, imp string
		want           bool
	}{
		{filepath.Join(dir, "x.go"), "nodeissuerbinding", target, true},
		{"internal/nodecontrol/nodeissuerbinding/x.go", "nodeissuerbinding", target, true},
		{filepath.Join(dir, "x_test.go"), "nodeissuerbinding", target, true},
		{filepath.Join(dir, "x_test.go"), "nodeissuerbinding_test", target, false},
		{filepath.Join(dir, "sub/x.go"), "nodeissuerbinding", target, false},
		{filepath.Join(root, "internal/nodecontrol/nodeissuerbinding_sibling/x.go"), "nodeissuerbinding", target, false},
		{filepath.Join(root, "else/internal/nodecontrol/nodeissuerbinding/x.go"), "nodeissuerbinding", target, false},
		{filepath.Join(root, "../external/internal/nodecontrol/nodeissuerbinding/x.go"), "nodeissuerbinding", target, false},
		{filepath.Join(dir, "x.go"), "wrong", target, false},
		{filepath.Join(dir, "x.go"), "nodeissuerbinding", target + "/child", false},
	} {
		if got := leafBindingConsumerAllowed(c.path, c.pkg, c.imp); got != c.want {
			t.Fatal("exact issuer binding consumer", c, got)
		}
	}
}
