package statecomposition

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

const compositionPath = "talenro.local/platform/internal/nodecontrol/statecomposition"

func TestCompositionPolicyBindingBoundary(t *testing.T) {
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
		if got := compositionConsumerAllowed(filepath.FromSlash(tc.path), tc.pkg, "talenro.local/platform/internal/nodecontrol/statecomposition"); got != tc.want {
			t.Errorf("%s: %v", tc.path, got)
		}
	}

}

func TestCompositionTimeBindingBoundary(t *testing.T) {
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
		if got := compositionConsumerAllowed(filepath.FromSlash(tc.path), tc.pkg, "talenro.local/platform/internal/nodecontrol/statecomposition"); got != tc.want {
			t.Errorf("%s %s: %v", tc.path, tc.pkg, got)
		}
	}

}

func TestCompositionBindingBoundary(t *testing.T) {
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
		if got := compositionConsumerAllowed(filepath.FromSlash(tc.path), tc.pkg, compositionPath); got != tc.want {
			t.Errorf("%s: %v", tc.path, got)
		}
	}
}

func TestCompositionBoundary(t *testing.T) {
	for _, src := range []string{
		`package p;import "bytes";var _=bytes.Equal`,
		`package p;import "reflect";var _=reflect.DeepEqual`,
	} {
		if !compositionImportsAllowed([]byte(src)) {
			t.Fatal("pure import rejected", src)
		}
	}
	for _, src := range []string{
		`package p;import "os"`, `package p;import "crypto/sha256"`, `package p;import "encoding/json"`,
		`package p;import "time"`, `package p;import "net/http"`, `package p;import "unsafe"`,
		`package p;import "talenro.local/platform/internal/nodecontrol/statepolicy"`,
		`package p;import "talenro.local/platform/internal/nodecontrol/statecontract"`,
		`package p;import "context";var _=context.WithCancel`,
		`package p;import "reflect";var _=reflect.ValueOf`,
		`package p;import . "bytes"`, `package p;import _ "bytes"`,
		`package p;import(a "bytes";b "bytes");var _=a.Equal;var _=b.Clone`,
		`package p;import(a "bytes";a "context")`, `package p;func f(){go f()}`, `broken`,
	} {
		if compositionImportsAllowed([]byte(src)) {
			t.Fatal("forbidden capability", src)
		}
	}
	if compositionConsumerAllowed("internal/nodecontrol/serving/x.go", "serving", compositionPath) {
		t.Fatal("runtime consumer")
	}
	if !compositionConsumerAllowed("internal/nodecontrol/serving/x.go", "serving", "context") {
		t.Fatal("unrelated dependency")
	}
}

func compositionImportsAllowed(source []byte) bool {
	allowed := map[string][]string{
		"context": {"Context", "Canceled", "DeadlineExceeded"}, "errors": {"New", "Is"}, "bytes": {"Clone", "Compare", "Equal"}, "sort": {"Slice"}, "reflect": {"DeepEqual"},
		"talenro.local/platform/internal/nodecontrol/contracts":      {"Digest"},
		"talenro.local/platform/internal/nodecontrol/statematerials": {"Input", "Scope", "MaterialRef", "MaterialKind", "RootSet", "RootRotation", "Metadata", "Desired", "Material", "Expectation", "MaterialSubsetModel", "Summary", "Evaluate", "ErrInvalidInput", "ErrLimitExceeded", "ErrInvalidMaterial", "ErrMissingMaterial", "ErrDuplicateMaterial", "ErrClearedModel", "ErrReferenceMismatch"},
		"talenro.local/platform/internal/nodecontrol/stateverify":    {"Batch", "SignatureEvidence", "EvidenceCoordinates", "ArtifactIdentity", "Verify", "ErrInvalidInput", "ErrBindingMismatch", "ErrSignatureMismatch", "ErrCanceled"},
		"talenro.local/platform/internal/nodecontrol/stateformat":    {"ParseEnvelope", "InitialRoot", "ParsedArtifact", "RootMaterial", "ErrInvalidFormat"},
	}
	f, e := parser.ParseFile(token.NewFileSet(), "guard.go", source, 0)
	if e != nil {
		return false
	}
	aliases := map[string]map[string]bool{}
	seen := map[string]bool{}
	for _, i := range f.Imports {
		p, e := strconv.Unquote(i.Path.Value)
		members, ok := allowed[p]
		if e != nil || !ok || seen[p] {
			return false
		}
		seen[p] = true
		alias := filepath.Base(p)
		if i.Name != nil {
			alias = i.Name.Name
		}
		if alias == "." || alias == "_" || aliases[alias] != nil {
			return false
		}
		set := map[string]bool{}
		for _, m := range members {
			set[m] = true
		}
		aliases[alias] = set
	}
	valid := true
	ast.Inspect(f, func(n ast.Node) bool {
		if _, ok := n.(*ast.GoStmt); ok {
			valid = false
		}
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
func compositionConsumerAllowed(path, pkg, imported string) bool {
	if pkg == "statetimebinding" && filepath.Dir(filepath.Clean(path)) == filepath.FromSlash("internal/nodecontrol/statetimebinding") {
		return true
	}
	return imported != compositionPath || (pkg == "statebinding" && filepath.Dir(filepath.Clean(path)) == filepath.FromSlash("internal/nodecontrol/statebinding")) ||
		(pkg == "statepolicybinding" && filepath.Dir(filepath.Clean(path)) == filepath.FromSlash("internal/nodecontrol/statepolicybinding"))
}
func TestCompositionProductImports(t *testing.T) {
	files, e := filepath.Glob("*.go")
	if e != nil {
		t.Fatal(e)
	}
	for _, f := range files {
		if strings.HasSuffix(f, "_test.go") {
			continue
		}
		raw, e := os.ReadFile(f)
		if e != nil {
			t.Fatal(e)
		}
		if !compositionImportsAllowed(raw) {
			t.Error("production capability", f)
		}
	}
}

func TestCompositionCleanupRegistration(t *testing.T) {
	good := `package p;func Evaluate(){owned,e:=ownInput(ctx,in);_=e;defer func(){clearInput(&owned)}();ctx.Err();statematerials.Evaluate(ctx,owned.Materials)}`
	if !compositionCleanupRegistered([]byte(good)) {
		t.Fatal("early cleanup rejected")
	}
	for _, src := range []string{
		`package p;func Evaluate(){owned,e:=ownInput(ctx,in);_=e;ctx.Err();defer func(){clearInput(&owned)}();statematerials.Evaluate(ctx,owned.Materials)}`,
		`package p;func Evaluate(){owned,e:=ownInput(ctx,in);_=e;statematerials.Evaluate(ctx,owned.Materials);defer func(){clearInput(&owned)}()}`,
		`package p;func Evaluate(){owned,e:=ownInput(ctx,in);_=e;if enabled{defer func(){clearInput(&owned)}()};statematerials.Evaluate(ctx,owned.Materials)}`,
		`package p;func Evaluate(){owned,e:=ownInput(ctx,in);_=e;defer func(){clearInput(&other)}();statematerials.Evaluate(ctx,owned.Materials)}`,
		`package p;func Evaluate(){owned,e:=ownInput(ctx,in);_=e;statematerials.Evaluate(ctx,owned.Materials)}`,
		`broken`,
	} {
		if compositionCleanupRegistered([]byte(src)) {
			t.Fatal("unsafe cleanup accepted", src)
		}
	}
	raw, e := os.ReadFile("evaluate.go")
	if e != nil {
		t.Fatal(e)
	}
	if !compositionCleanupRegistered(raw) {
		t.Fatal("production owner cleanup")
	}
}
func TestCompositionConsumers(t *testing.T) {
	for _, path := range []string{"internal/nodecontrol/serving/x.go", "internal/nodecontrol/statecomposition/x.go", "internal/nodecontrol/statecomposition/sub/x.go", "internal/nodecontrol/statecomposition-other/x.go"} {
		if compositionConsumerAllowed(filepath.FromSlash(path), "statecomposition", compositionPath) {
			t.Fatal("disguised consumer", path)
		}
	}
	root := filepath.Clean("../../..")
	count := 0
	e := filepath.WalkDir(root, func(path string, d fs.DirEntry, e error) error {
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
		count++
		f, e := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if e != nil {
			return e
		}
		rel, e := filepath.Rel(root, path)
		if e != nil {
			return e
		}
		for _, i := range f.Imports {
			p, e := strconv.Unquote(i.Path.Value)
			if e != nil {
				return e
			}
			if !compositionConsumerAllowed(rel, f.Name.Name, p) {
				t.Error("production consumer", path)
			}
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	if count == 0 {
		t.Fatal("no product scan")
	}
}

func compositionCleanupRegistered(source []byte) bool {
	f, e := parser.ParseFile(token.NewFileSet(), "cleanup.go", source, 0)
	if e != nil {
		return false
	}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "Evaluate" || fn.Body == nil {
			continue
		}
		var owned, cleanup, firstUse token.Pos
		for _, stmt := range fn.Body.List {
			if as, ok := stmt.(*ast.AssignStmt); ok && len(as.Lhs) > 0 && len(as.Rhs) > 0 {
				if id, ok := as.Lhs[0].(*ast.Ident); ok && id.Name == "owned" {
					if c, ok := as.Rhs[0].(*ast.CallExpr); ok {
						if id, ok := c.Fun.(*ast.Ident); ok && id.Name == "ownInput" {
							owned = c.Pos()
						}
					}
				}
			}
			if d, ok := stmt.(*ast.DeferStmt); ok {
				literal, ok := d.Call.Fun.(*ast.FuncLit)
				if !ok {
					continue
				}
				matches := false
				ast.Inspect(literal.Body, func(n ast.Node) bool {
					if c, ok := n.(*ast.CallExpr); ok && len(c.Args) == 1 {
						if id, ok := c.Fun.(*ast.Ident); ok && id.Name == "clearInput" {
							if unary, ok := c.Args[0].(*ast.UnaryExpr); ok && unary.Op == token.AND {
								if id, ok := unary.X.(*ast.Ident); ok && id.Name == "owned" {
									matches = true
								}
							}
						}
					}
					return true
				})
				if matches && (cleanup == 0 || d.Pos() < cleanup) {
					cleanup = d.Pos()
				}
			}
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			c, ok := n.(*ast.CallExpr)
			if !ok || c.Pos() <= owned {
				return true
			}
			use := false
			switch fun := c.Fun.(type) {
			case *ast.Ident:
				use = fun.Name == "selectBatch"
			case *ast.SelectorExpr:
				use = fun.Sel.Name == "Evaluate" || fun.Sel.Name == "Verify" || fun.Sel.Name == "Err"
			}
			if use && (firstUse == 0 || c.Pos() < firstUse) {
				firstUse = c.Pos()
			}
			return true
		})
		return owned != 0 && cleanup > owned && firstUse > cleanup
	}
	return false
}
