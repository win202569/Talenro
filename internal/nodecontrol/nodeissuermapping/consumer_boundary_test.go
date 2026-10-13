package nodeissuermapping

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestIssuerMappingConsumers(t *testing.T) {
	const imported = "talenro.local/platform/internal/nodecontrol/nodeissuermapping"
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	dir := filepath.Join(root, "internal/nodecontrol/nodeissuermapping")
	for _, c := range []struct {
		path, pkg, imp string
		want           bool
	}{
		{filepath.Join(dir, "a_test.go"), "nodeissuermapping", imported, true},
		{"internal/nodecontrol/nodeissuermapping/a_test.go", "nodeissuermapping", imported, true},
		{filepath.Join(dir, "a.go"), "nodeissuermapping", imported, false},
		{filepath.Join(dir, "a_test.go"), "nodeissuermapping_test", imported, false},
		{filepath.Join(dir, "child/a_test.go"), "nodeissuermapping", imported, false},
		{filepath.Join(root, "internal/nodecontrol/nodeissuermapping_sibling/a_test.go"), "nodeissuermapping", imported, false},
		{filepath.Join(root, "other/internal/nodecontrol/nodeissuermapping/a_test.go"), "nodeissuermapping", imported, false},
		{filepath.Join(root, "../external/internal/nodecontrol/nodeissuermapping/a_test.go"), "nodeissuermapping", imported, false},
		{filepath.Join(dir, "a_test.go"), "wrong", imported, false},
		{filepath.Join(dir, "a_test.go"), "nodeissuermapping", imported + "/child", false},
	} {
		if got := issuerMappingConsumerAllowed(c.path, c.pkg, c.imp); got != c.want {
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
				if !issuerMappingConsumerAllowed(path, f.Name.Name, p) {
					t.Errorf("unauthorized %s %s", path, p)
				}
			}
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
}
func issuerMappingConsumerAllowed(path, pkg, imported string) bool {
	if imported != "talenro.local/platform/internal/nodecontrol/nodeissuermapping" || pkg != "nodeissuermapping" || !strings.HasSuffix(path, "_test.go") {
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
	return e == nil && filepath.Dir(rel) == filepath.FromSlash("internal/nodecontrol/nodeissuermapping")
}
