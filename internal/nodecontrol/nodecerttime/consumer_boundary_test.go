package nodecerttime

import (
	"go/parser"
	"go/token"
	"os"
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
	target := "talenro.local/platform/internal/nodecontrol/nodecerttime"
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
		if got := certTimeConsumerAllowed(c.path, c.pkg, target); got != c.want {
			t.Errorf("%s %s: got %v want %v", c.path, c.pkg, got, c.want)
		}
	}
	if certTimeConsumerAllowed(filepath.Join(dir, "a.go"), "nodeleafbinding", target+"/child") {
		t.Error("subpath")
	}
}

func TestCertTimeConsumers(t *testing.T) {
	const imported = "talenro.local/platform/internal/nodecontrol/nodecerttime"
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	dir := filepath.Join(root, "internal/nodecontrol/nodecerttime")
	for _, tc := range []struct {
		path, pkg, imp string
		want           bool
	}{
		{filepath.Join(dir, "a_test.go"), "nodecerttime", imported, true},
		{"internal/nodecontrol/nodecerttime/a_test.go", "nodecerttime", imported, true},
		{filepath.Join(dir, "a.go"), "nodecerttime", imported, false},
		{filepath.Join(dir, "a_test.go"), "nodecerttime_test", imported, false},
		{filepath.Join(dir, "nested/a_test.go"), "nodecerttime", imported, false},
		{filepath.Join(root, "internal/nodecontrol/nodecerttime_sibling/a_test.go"), "nodecerttime", imported, false},
		{filepath.Join(root, "else/internal/nodecontrol/nodecerttime/a_test.go"), "nodecerttime", imported, false},
		{filepath.Join(root, "../external/internal/nodecontrol/nodecerttime/a_test.go"), "nodecerttime", imported, false},
		{filepath.Join(dir, "a_test.go"), "nodecerttime", imported + "/sub", false},
		{filepath.Join(dir, "a_test.go"), "nodecerttime", "other", false},
	} {
		if got := certTimeConsumerAllowed(tc.path, tc.pkg, tc.imp); got != tc.want {
			t.Fatal("exact consumer case", tc, got)
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
				if !certTimeConsumerAllowed(path, f.Name.Name, p) {
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

func certTimeConsumerAllowed(path, pkg, imported string) bool {
	// Only the exact selected-leaf projection consumer is added.
	leafBindingTarget := "talenro.local/platform/internal/nodecontrol/nodecerttime"
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
	if imported != "talenro.local/platform/internal/nodecontrol/nodecerttime" || pkg != "nodecerttime" || !strings.HasSuffix(path, "_test.go") {
		return false
	}
	root, e := filepath.Abs("../../..")
	if e != nil {
		return false
	}
	if !filepath.IsAbs(path) {
		path = filepath.Join(root, path)
	}
	resolved, e := filepath.Abs(path)
	if e != nil {
		return false
	}
	return filepath.Clean(filepath.Dir(resolved)) == filepath.Join(root, "internal", "nodecontrol", "nodecerttime")
}
