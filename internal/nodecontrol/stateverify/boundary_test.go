package stateverify

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

func verifySourceAllowed(source []byte) bool {
	f, e := parser.ParseFile(token.NewFileSet(), "", source, 0)
	if e != nil {
		return false
	}
	allowed := map[string]bool{}
	for _, s := range []string{"bytes", "context", "crypto/ed25519", "errors", "time", "github.com/google/uuid", "talenro.local/platform/internal/nodecontrol/contracts", "talenro.local/platform/internal/nodecontrol/stateformat"} {
		allowed[s] = true
	}
	aliases := map[string]string{}
	for _, i := range f.Imports {
		p, e := strconv.Unquote(i.Path.Value)
		if e != nil || !allowed[p] {
			return false
		}
		name := filepath.Base(p)
		if i.Name != nil {
			name = i.Name.Name
		}
		if name == "." || name == "_" {
			return false
		}
		aliases[name] = p
	}
	ok := true
	ast.Inspect(f, func(n ast.Node) bool {
		if _, yes := n.(*ast.GoStmt); yes {
			ok = false
		}
		if s, yes := n.(*ast.SelectorExpr); yes {
			if x, yes := s.X.(*ast.Ident); yes {
				switch aliases[x.Name] {
				case "time":
					if s.Sel.Name != "Time" {
						ok = false
					}
				case "crypto/ed25519":
					if s.Sel.Name != "Verify" {
						ok = false
					}
				}
			}
		}
		return true
	})
	return ok
}
func ownerCleanupBeforeParse(source []byte) bool {
	f, e := parser.ParseFile(token.NewFileSet(), "", source, 0)
	if e != nil {
		return false
	}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "Verify" || fn.Body == nil {
			continue
		}
		var cleanup, parse token.Pos
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if d, ok := n.(*ast.DeferStmt); ok {
				if s, ok := d.Call.Fun.(*ast.SelectorExpr); ok {
					if id, ok := s.X.(*ast.Ident); ok && id.Name == "o" && s.Sel.Name == "Clear" {
						cleanup = d.Pos()
					}
				}
			}
			if c, ok := n.(*ast.CallExpr); ok {
				if s, ok := c.Fun.(*ast.SelectorExpr); ok {
					if id, ok := s.X.(*ast.Ident); ok && id.Name == "o" && s.Sel.Name == "parse" && (parse == 0 || c.Pos() < parse) {
						parse = c.Pos()
					}
				}
			}
			return true
		})
		return cleanup != 0 && parse != 0 && cleanup < parse
	}
	return false
}

// Only the explicitly isolated policy and composition packages may use the verifier.
// No non-test package may consume the policy package itself.
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
		if got := isolatedImportAllowed(c.path, c.pkg, target); got != c.want {
			t.Errorf("%s %s: got %v want %v", c.path, c.pkg, got, c.want)
		}
	}
	if isolatedImportAllowed(filepath.Join(dir, "a.go"), "nodeleafbinding", target+"/child") {
		t.Error("subpath")
	}
	if isolatedImportAllowed(filepath.Join(dir, "a.go"), "nodeleafbinding", "talenro.local/platform/internal/nodecontrol/stateverify") {
		t.Error("verifier import")
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
		if got := isolatedImportAllowed(c.path, c.pkg, target); got != c.want {
			t.Errorf("%s %s: got %v want %v", c.path, c.pkg, got, c.want)
		}
	}
	if isolatedImportAllowed(filepath.Join(dir, "a.go"), "nodecerttime", target+"/child") {
		t.Error("subpath")
	}
	if isolatedImportAllowed(filepath.Join(dir, "a.go"), "nodecerttime", "talenro.local/platform/internal/nodecontrol/stateverify") {
		t.Error("verifier import")
	}
}

func isolatedImportAllowed(relativePath, packageName, imported string) bool {
	// Only the exact selected-leaf projection consumer is added.
	leafBindingTarget := "talenro.local/platform/internal/nodecontrol/statepolicy"
	if imported == leafBindingTarget || strings.HasPrefix(imported, leafBindingTarget+"/") {
		root, e := filepath.Abs("../../..")
		if e != nil {
			return false
		}
		normalized := filepath.Clean(relativePath)
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
		if exact || packageName == "nodeleafbinding" || packageName == "nodeleafbinding_test" {
			return imported == leafBindingTarget && exact && packageName == "nodeleafbinding"
		}
	}
	// Only the new exact consumer is added; all previous rules below remain.
	certTimeTarget := "talenro.local/platform/internal/nodecontrol/statepolicy"
	if imported == certTimeTarget || strings.HasPrefix(imported, certTimeTarget+"/") {
		root, e := filepath.Abs("../../..")
		if e != nil {
			return false
		}
		normalized := filepath.Clean(relativePath)
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
		if exact || packageName == "nodecerttime" || packageName == "nodecerttime_test" {
			return imported == certTimeTarget && exact && packageName == "nodecerttime"
		}
	}

	if imported == "talenro.local/platform/internal/nodecontrol/statepolicy" {
		return (packageName == "statepolicybinding" && filepath.Dir(filepath.Clean(relativePath)) == filepath.FromSlash("internal/nodecontrol/statepolicybinding")) ||
			(packageName == "statetimebinding" && filepath.Dir(filepath.Clean(relativePath)) == filepath.FromSlash("internal/nodecontrol/statetimebinding"))
	}
	if imported != "talenro.local/platform/internal/nodecontrol/stateverify" {
		return true
	}
	dir := filepath.Dir(filepath.Clean(relativePath))
	return (packageName == "statepolicy" && dir == filepath.FromSlash("internal/nodecontrol/statepolicy")) ||
		(packageName == "statecomposition" && dir == filepath.FromSlash("internal/nodecontrol/statecomposition"))
}
func TestStateverifyPolicyBindingBoundary(t *testing.T) {
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
		if got := isolatedImportAllowed(filepath.FromSlash(tc.path), tc.pkg, "talenro.local/platform/internal/nodecontrol/statepolicy"); got != tc.want {
			t.Errorf("%s: %v", tc.path, got)
		}
	}
	if isolatedImportAllowed(filepath.FromSlash("internal/nodecontrol/statepolicybinding/x.go"), "statepolicybinding", "talenro.local/platform/internal/nodecontrol/stateverify") {
		t.Fatal("direct verifier consumer")
	}
}

func TestStateverifyPolicyBoundary(t *testing.T) {
	for _, c := range []struct {
		path, pkg, imported string
		want                bool
	}{
		{"internal/nodecontrol/statepolicy/facts.go", "statepolicy", "stateverify", true},
		{"internal/nodecontrol/serving/reader.go", "serving", "stateverify", false},
		{"internal/nodecontrol/serving/reader.go", "serving", "statepolicy", false},
		{"internal/nodecontrol/statepolicy/sub/facts.go", "statepolicy", "stateverify", false},
		{"internal/nodecontrol/statepolicy-other/facts.go", "statepolicy", "stateverify", false},
		{"internal/nodecontrol/statepolicy/facts.go", "other", "stateverify", false},
	} {
		if got := isolatedImportAllowed(filepath.FromSlash(c.path), c.pkg, "talenro.local/platform/internal/nodecontrol/"+c.imported); got != c.want {
			t.Errorf("%s %s: allowed=%v", c.path, c.imported, got)
		}
	}
}
func TestStateverifyCompositionBoundary(t *testing.T) {
	for _, tc := range []struct {
		path, pkg, imported string
		want                bool
	}{
		{"internal/nodecontrol/statecomposition/evaluate.go", "statecomposition", "stateverify", true},
		{"internal/nodecontrol/statecomposition/sub/x.go", "statecomposition", "stateverify", false},
		{"internal/nodecontrol/statecomposition-other/x.go", "statecomposition", "stateverify", false},
		{"internal/nodecontrol/statecomposition/x.go", "other", "stateverify", false},
		{"internal/nodecontrol/serving/reader.go", "serving", "stateverify", false},
		{"internal/nodecontrol/statecomposition/x.go", "statecomposition", "statepolicy", false},
	} {
		if got := isolatedImportAllowed(filepath.FromSlash(tc.path), tc.pkg, "talenro.local/platform/internal/nodecontrol/"+tc.imported); got != tc.want {
			t.Errorf("%s %s: %v", tc.path, tc.imported, got)
		}
	}
}
func TestStateverifyBoundary(t *testing.T) {
	for _, path := range []string{"net/http", "database/sql", "encoding/json", "github.com/gowebpki/jcs", "talenro.local/platform/internal/nodecontrol/serving", "talenro.local/platform/internal/nodecontrol/authority", "talenro.local/platform/internal/store", "talenro.local/platform/internal/trust", "crypto/ed25519/other"} {
		for _, alias := range []string{"", "x ", "_ "} {
			if verifySourceAllowed([]byte("package p; import " + alias + strconv.Quote(path))) {
				t.Fatalf("forbidden import %s", path)
			}
		}
	}
	for _, source := range []string{`package p; import t "time"; func f(){t.Now()}`, `package p; import e "crypto/ed25519"; func f(){e.Sign(nil,nil)}`, `package p; func f(){go f()}`, "broken source"} {
		if verifySourceAllowed([]byte(source)) {
			t.Fatal("forbidden capability", source)
		}
	}
	good := []byte(`package p; import("context";"crypto/ed25519");func f(){ed25519.Verify(nil,nil,nil)}`)
	if !verifySourceAllowed(good) {
		t.Fatal("pure verifier rejected")
	}
	if ownerCleanupBeforeParse([]byte(`package p;func Verify(){o.parse();defer o.Clear()}`)) {
		t.Fatal("late cleanup accepted")
	}
	if !ownerCleanupBeforeParse([]byte(`package p;func Verify(){defer o.Clear();o.parse()}`)) {
		t.Fatal("early cleanup rejected")
	}
	files, e := filepath.Glob("*.go")
	if e != nil {
		t.Fatal(e)
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		source, e := os.ReadFile(file)
		if e != nil {
			t.Fatal(e)
		}
		if !verifySourceAllowed(source) {
			t.Error("production boundary", file)
		}
		if file == "verify.go" && !ownerCleanupBeforeParse(source) {
			t.Error("owner cleanup not registered first")
		}
	}
	root := filepath.Clean("../../..")
	e = filepath.WalkDir(root, func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			if path != root && (strings.HasPrefix(d.Name(), ".") || d.Name() == "testdata" || d.Name() == "vendor" || d.Name() == "node_modules") {
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
			v, e := strconv.Unquote(i.Path.Value)
			if e != nil {
				return e
			}
			relativePath, relErr := filepath.Rel(root, path)
			if relErr != nil {
				return relErr
			}
			if !isolatedImportAllowed(relativePath, f.Name.Name, v) {
				t.Errorf("production caller %s", path)
			}
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
}

// Keep AST imported while the RED guard is a closed scaffold.
var _ ast.Node

func TestStateverifyTimeBindingBoundary(t *testing.T) {
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
		if got := isolatedImportAllowed(filepath.FromSlash(tc.path), tc.pkg, "talenro.local/platform/internal/nodecontrol/statepolicy"); got != tc.want {
			t.Errorf("%s %s: %v", tc.path, tc.pkg, got)
		}
	}
	if isolatedImportAllowed(filepath.FromSlash("internal/nodecontrol/statetimebinding/x.go"), "statetimebinding", "talenro.local/platform/internal/nodecontrol/stateverify") {
		t.Fatal("direct verifier")
	}
}
