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

func TestNodeLeafBindingConsumerException(t *testing.T) {
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	dir := filepath.FromSlash("internal/nodecontrol/nodeleafbinding")
	target := "talenro.local/platform/internal/nodecontrol/statecontract"
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
		if got := contractConsumerAllowed(c.path, c.pkg, target); got != c.want {
			t.Errorf("%s %s: got %v want %v", c.path, c.pkg, got, c.want)
		}
	}
	if contractConsumerAllowed(filepath.Join(dir, "a.go"), "nodeleafbinding", target+"/child") {
		t.Error("subpath")
	}
}

func TestNodeCertTimeConsumerException(t *testing.T) {
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	dir := filepath.FromSlash("internal/nodecontrol/nodecerttime")
	target := "talenro.local/platform/internal/nodecontrol/statecontract"
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
		if got := contractConsumerAllowed(c.path, c.pkg, target); got != c.want {
			t.Errorf("%s %s: got %v want %v", c.path, c.pkg, got, c.want)
		}
	}
	if contractConsumerAllowed(filepath.Join(dir, "a.go"), "nodecerttime", target+"/child") {
		t.Error("subpath")
	}

}

func contractConsumerAllowed(path, pkg, imported string) bool {
	issuerBindingTarget := "talenro.local/platform/internal/nodecontrol/statecontract"
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
	// Only the exact selected-leaf projection consumer is added.
	leafBindingTarget := "talenro.local/platform/internal/nodecontrol/statecontract"
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
	certTimeTarget := "talenro.local/platform/internal/nodecontrol/statecontract"
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
	e = walkContractConsumers(root, func(p string) { t.Errorf("production consumer: %s", p) })
	if e != nil {
		t.Fatal(e)
	}
}

func walkContractConsumers(root string, reject func(string)) error {
	return filepath.WalkDir(root, func(p string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			// Retained diagnostic SDKs contain intentional invalid-Go fixtures.
			// Exclude only this root-owned artifact tree, not nested namesakes.
			if p == filepath.Join(filepath.Clean(root), ".superpowers", "diagnostics") {
				return filepath.SkipDir
			}
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
				reject(p)
			}
		}
		return nil
	})
}

func TestContractConsumerScanScope(t *testing.T) {
	const forbidden = `package serving; import "talenro.local/platform/internal/nodecontrol/statecontract"`
	const malformed = `package sample; import "math",`
	for _, tc := range []struct {
		name, path, source   string
		parseError, rejected bool
	}{
		{"diagnostic syntax fixture", ".superpowers/diagnostics/sdk/sample.go", malformed, false, false},
		{"diagnostic import fixture", ".superpowers/diagnostics/sdk/sample.go", forbidden, false, false},
		{"ordinary source", "internal/serving/x.go", `package serving; import "fmt"`, false, false},
		{"ordinary malformed", "internal/serving/x.go", malformed, true, false},
		{"ordinary forbidden", "internal/serving/x.go", forbidden, false, true},
		{"nested diagnostic malformed", "internal/.superpowers/diagnostics/x.go", malformed, true, false},
		{"nested diagnostic forbidden", "internal/.superpowers/diagnostics/x.go", forbidden, false, true},
		{"similar root malformed", ".superpowers-other/diagnostics/x.go", malformed, true, false},
		{"similar root forbidden", ".superpowers-other/diagnostics/x.go", forbidden, false, true},
		{"other private source", ".superpowers/source/x.go", forbidden, false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := t.TempDir()
			path := filepath.Join(root, filepath.FromSlash(tc.path))
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte(tc.source), 0600); err != nil {
				t.Fatal(err)
			}
			var rejected []string
			err := walkContractConsumers(root, func(p string) { rejected = append(rejected, p) })
			if (err != nil) != tc.parseError {
				t.Fatalf("parse error = %v, want error %v", err, tc.parseError)
			}
			if tc.rejected {
				if len(rejected) != 1 || rejected[0] != path {
					t.Fatalf("rejected = %v, want [%s]", rejected, path)
				}
			} else if len(rejected) != 0 {
				t.Fatalf("unexpected rejected consumers: %v", rejected)
			}
		})
	}
}
func TestContractIssuerBindingBoundary(t *testing.T) {
	const target = "talenro.local/platform/internal/nodecontrol/statecontract"
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
		if got := contractConsumerAllowed(c.path, c.pkg, c.imp); got != c.want {
			t.Fatal("exact issuer binding consumer", c, got)
		}
	}
}
