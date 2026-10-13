package nodeissuerbinding

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

const issuerBindingGoInventory = "doc.go input.go errors.go budget.go owner.go owned_certificate.go owned_binding.go owned_content.go owned_table.go owned_policy.go owned_time.go owned_summary.go model.go lower.go references.go projection.go evaluate.go input_test.go budget_test.go owned_certificate_test.go owned_binding_test.go owned_content_test.go owned_table_test.go owned_policy_test.go owned_time_test.go owned_summary_test.go model_test.go fixtures_test.go lower_test.go references_test.go projection_test.go evaluate_test.go boundary_test.go ownership_boundary_test.go consumer_boundary_test.go"

func issuerBindingSources(t testing.TB) map[string][]byte {
	t.Helper()
	out := map[string][]byte{}
	if e := filepath.WalkDir(".", func(p string, d os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			if p != "." {
				out[filepath.ToSlash(p)+"/"] = nil
			}
			return nil
		}
		if strings.HasSuffix(p, ".go") {
			b, e := os.ReadFile(p)
			if e != nil {
				return e
			}
			out[filepath.ToSlash(p)] = b
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
	return out
}
func issuerBindingMutation(t testing.TB, sources map[string][]byte, file, old, replacement string) map[string][]byte {
	t.Helper()
	out := map[string][]byte{}
	for k, v := range sources {
		out[k] = bytes.Clone(v)
	}
	b, ok := out[file]
	if !ok || old == replacement || !strings.Contains(string(b), old) {
		t.Fatal("inactive mutation", file, old)
	}
	out[file] = []byte(strings.Replace(string(b), old, replacement, 1))
	if bytes.Equal(out[file], b) {
		t.Fatal("inactive mutation")
	}
	return out
}
func TestIssuerBindingProductBoundary(t *testing.T) {
	src := issuerBindingSources(t)
	for name, b := range src {
		if !strings.HasSuffix(name, "_test.go") && !issuerBindingProductAllowed(b) {
			t.Fatal("actual product", name)
		}
	}
	for _, positive := range []string{
		"package nodeissuerbinding; import c \"context\";func f(ctx c.Context)error{return ctx.Err()}",
		"package nodeissuerbinding;import clock \"time\";func normalizeTime(t clock.Time)clock.Time{return t.Round(0).UTC()}",
	} {
		if !issuerBindingProductAllowed([]byte(positive)) {
			t.Fatal("safe named alias/normalization")
		}
	}
	for _, negative := range []string{
		"package nodeissuerbinding;import \"os\";func f(){os.Getenv(\"X\")}",
		"package nodeissuerbinding;import \"time\";func f(){time.Now()}",
		"package nodeissuerbinding;import \"context\";func f(ctx context.Context){go ctx.Err()}",
		"package nodeissuerbinding;func f(){recover()}",
		"package nodeissuerbinding;import . \"time\";var x Time",
		"package nodeissuerbinding;import _ \"time\"",
		"package nodeissuerbinding;import c \"context\";import d \"context\"",
		"package nodeissuerbinding;import \"context\";func f(ctx context.Context){context:=ctx;context.Err()}",
		"package nodeissuerbinding;import \"time\";func f(t time.Time){t.Round(1).UTC()}",
		"package nodeissuerbinding;import \"time\";func f(t time.Time,n time.Time){t.Add(0);t.Unix();t.Equal(n)}",
		"package nodeissuerbinding;import \"context\";func f(ctx context.Context){g:=ctx.Err;g()}",
		"package nodeissuerbinding;import \"talenro.local/platform/internal/nodecontrol/nodeissuer\";var f=nodeissuer.Evaluate",
		"package nodeissuerbinding;import \"talenro.local/platform/internal/nodecontrol/nodeissuer\";var x=nodeissuer.Material{}",
		"package nodeissuerbinding;import \"talenro.local/platform/internal/nodecontrol/statetime\";var x=statetime.ModelInput{}",
	} {
		if issuerBindingProductAllowed([]byte(negative)) {
			t.Fatal("forbidden capability", negative)
		}
	}
}
func TestIssuerBindingCapabilityBoundary(t *testing.T) {
	src := issuerBindingSources(t)
	if !issuerBindingCapabilityAllowed(src) {
		for name, raw := range src {
			if !strings.HasSuffix(name, "_test.go") {
				t.Logf("schema %s %s", name, issuerBindingASTHash(raw))
			}
		}
		t.Fatal("actual inventory/API/flow")
	}
	alias := issuerBindingMutation(t, src, "lower.go", "\"context\"", "c \"context\"")
	formatted := map[string][]byte{}
	for name, raw := range src {
		out, err := format.Source(raw)
		if err != nil {
			t.Fatal(err)
		}
		formatted[name] = out
	}
	if !issuerBindingCapabilityAllowed(formatted) {
		t.Fatal("benign formatting")
	}
	alias["lower.go"] = []byte(strings.ReplaceAll(string(alias["lower.go"]), "context.Context", "c.Context"))
	if !issuerBindingCapabilityAllowed(alias) {
		t.Fatal("safe named import alias")
	}
	rename := issuerBindingMutation(t, src, "lower.go", "ctx context.Context", "operationCtx context.Context")
	rename["lower.go"] = []byte(strings.ReplaceAll(string(src["lower.go"]), "ctx", "operationCtx"))
	if !issuerBindingCapabilityAllowed(rename) {
		t.Fatal("safe bound local rename")
	}
	for _, tc := range []struct{ file, old, replacement string }{
		{"input.go", "None Proof = iota", "None Proof = 1"},
		{"input.go", "Profile     string", "Profile     string; Extra bool"},
		{"errors.go", "nodeissuerbinding: invalid input", "nodeissuerbinding: changed"},
		{"owner.go", "valid   bool", "valid   bool; hook func()"},
		{"owner.go", "in.Round(0).UTC()", "in.Round(1).UTC()"},
		{"lower.go", "nodeleafbinding.Evaluate(ctx, o.input)", "nodeleafbinding.Evaluate(ctx, nodeleafbinding.Input{})"},
		{"projection.go", "identity.IssuerID != issuer", "false"},
	} {
		if issuerBindingCapabilityAllowed(issuerBindingMutation(t, src, tc.file, tc.old, tc.replacement)) {
			t.Fatal("capability/API mutation", tc.file, tc.old)
		}
	}
	for _, extra := range []string{"extra.go", "nested/extra.go", "nodeissuerbinding_extra.go"} {
		bad := map[string][]byte{}
		for k, v := range src {
			bad[k] = v
		}
		bad[extra] = []byte("package nodeissuerbinding")
		if issuerBindingCapabilityAllowed(bad) {
			t.Fatal("extra inventory", extra)
		}
	}
	extra := issuerBindingMutation(t, src, "doc.go", "package nodeissuerbinding", "package nodeissuerbinding\nvar Hook = func(){}")
	if issuerBindingCapabilityAllowed(extra) {
		t.Fatal("global hook")
	}
}
func issuerBindingUnparen(e ast.Expr) ast.Expr {
	for {
		p, ok := e.(*ast.ParenExpr)
		if !ok {
			return e
		}
		e = p.X
	}
}
func issuerBindingProductAllowed(raw []byte) bool {
	f, e := parser.ParseFile(token.NewFileSet(), "product.go", raw, 0)
	if e != nil || f.Name.Name != "nodeissuerbinding" {
		return false
	}
	allowed := map[string]string{
		"talenro.local/platform/internal/nodecontrol/contracts":       "Digest",
		"talenro.local/platform/internal/nodecontrol/nodecerttime":    "Input Summary",
		"talenro.local/platform/internal/nodecontrol/nodeleafbinding": "Input Summary Projection Evaluate SelectedLeafProjectionConditionalTimeBound",
		"crypto/sha256": "Sum256",
		"context":       "Context", "errors": "New", "bytes": "Clone Compare", "sort": "Slice", "time": "Time",
		"talenro.local/platform/internal/nodecontrol/nodeissuer":         "Input",
		"talenro.local/platform/internal/nodecontrol/statetimebinding":   "Input Summary",
		"talenro.local/platform/internal/nodecontrol/statepolicybinding": "Input",
		"talenro.local/platform/internal/nodecontrol/statebinding":       "Input",
		"talenro.local/platform/internal/nodecontrol/statecontract":      "MainStateContentModel",
		"talenro.local/platform/internal/nodecontrol/statematerials":     "Input RootSet RootRotation",
		"talenro.local/platform/internal/nodecontrol/statepolicy":        "ModelInput StateModel",
	}
	aliases := map[string]string{}
	seen := map[string]bool{}
	for _, im := range f.Imports {
		p, e := strconv.Unquote(im.Path.Value)
		if e != nil || allowed[p] == "" || seen[p] {
			return false
		}
		seen[p] = true
		a := filepath.Base(p)
		if im.Name != nil {
			a = im.Name.Name
		}
		if a == "." || a == "_" || aliases[a] != "" {
			return false
		}
		aliases[a] = p
	}
	parents := map[ast.Node]ast.Node{}
	stack := []ast.Node{}
	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return false
		}
		if len(stack) > 0 {
			parents[n] = stack[len(stack)-1]
		}
		stack = append(stack, n)
		return true
	})
	methods := map[string]int{"Err": 0, "Round": 1, "UTC": 0, "IsZero": 0, "Year": 0, "Before": 1, "Summary": 0, "Clear": 0, "clear": 0}
	ok := true
	ast.Inspect(f, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.GoStmt, *ast.SendStmt, *ast.ChanType, *ast.SelectStmt:
			ok = false
		case *ast.TypeSpec:
			if v.Assign.IsValid() {
				selector, is := v.Type.(*ast.SelectorExpr)
				if !is || v.Name.Name != "Input" || selector.Sel.Name != "Input" {
					ok = false
					break
				}
				id, is := selector.X.(*ast.Ident)
				if !is || aliases[id.Name] != "talenro.local/platform/internal/nodecontrol/nodeleafbinding" {
					ok = false
				}
			}
		case *ast.Field:
			if _, is := v.Type.(*ast.FuncType); is {
				ok = false
			}
		case *ast.Ident:
			if aliases[v.Name] != "" && v.Obj != nil {
				ok = false
			}
			if v.Obj != nil && v.Obj.Kind == ast.Fun {
				if d, is := parents[v].(*ast.FuncDecl); !is || d.Name != v {
					if call, is := parents[v].(*ast.CallExpr); !is || issuerBindingUnparen(call.Fun) != v {
						ok = false
					}
				}
			}
		case *ast.SelectorExpr:
			if x, is := issuerBindingUnparen(v.X).(*ast.Ident); is && aliases[x.Name] != "" && x.Obj == nil {
				if !strings.Contains(" "+allowed[aliases[x.Name]]+" ", " "+v.Sel.Name+" ") {
					ok = false
				}
				if v.Sel.Name == "Sum256" || v.Sel.Name == "Evaluate" || v.Sel.Name == "New" || v.Sel.Name == "Clone" || v.Sel.Name == "Compare" || v.Sel.Name == "Slice" {
					c, is := parents[v].(*ast.CallExpr)
					if !is || issuerBindingUnparen(c.Fun) != v {
						ok = false
					}
				}
			} else if _, is := methods[v.Sel.Name]; is {
				c, is := parents[v].(*ast.CallExpr)
				if !is || issuerBindingUnparen(c.Fun) != v {
					ok = false
				}
			}
		case *ast.CallExpr:
			switch fun := issuerBindingUnparen(v.Fun).(type) {
			case *ast.FuncLit:
				if _, deferred := parents[v].(*ast.DeferStmt); !deferred {
					ok = false
				}
			case *ast.Ident:
				if fun.Name == "recover" || fun.Name == "panic" || fun.Name == "append" {
					ok = false
				}
			case *ast.SelectorExpr:
				if x, is := issuerBindingUnparen(fun.X).(*ast.Ident); is && aliases[x.Name] != "" && x.Obj == nil {
					break
				}
				arity, exists := methods[fun.Sel.Name]
				if !exists || len(v.Args) != arity || v.Ellipsis.IsValid() {
					ok = false
					break
				}
				if fun.Sel.Name == "Round" {
					arg, is := issuerBindingUnparen(v.Args[0]).(*ast.BasicLit)
					if !is || arg.Kind != token.INT || arg.Value != "0" {
						ok = false
					}
				}
			default:
				ok = false
			}
		}
		return ok
	})
	return ok
}

// Structural schemas lock the reviewed implementation, not arbitrary future
// programs. Parser normalization ignores comments, formatting, import aliases
// and bound local names. Explicit member checks supplement this version lock.
// Neither these schemas nor mutation controls are a formal memory/safety proof.
func issuerBindingASTHash(raw []byte) string {
	f, e := parser.ParseFile(token.NewFileSet(), "schema.go", raw, 0)
	if e != nil {
		return ""
	}
	aliases := map[string]string{}
	for _, im := range f.Imports {
		p, e := strconv.Unquote(im.Path.Value)
		if e != nil {
			return ""
		}
		canonical := filepath.Base(p)
		name := canonical
		if im.Name != nil {
			name = im.Name.Name
		}
		aliases[name] = canonical
		im.Name = nil
	}
	ast.Inspect(f, func(n ast.Node) bool {
		if s, ok := n.(*ast.SelectorExpr); ok {
			if id, ok := s.X.(*ast.Ident); ok && id.Obj == nil {
				if canonical := aliases[id.Name]; canonical != "" {
					id.Name = canonical
				}
			}
		}
		return true
	})
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}
		locals := map[*ast.Object]string{}
		next := 0
		ast.Inspect(fn, func(n ast.Node) bool {
			if id, ok := n.(*ast.Ident); ok && id.Obj != nil && id.Obj.Kind == ast.Var {
				if name, found := locals[id.Obj]; found {
					id.Name = name
				} else {
					next++
					name = fmt.Sprintf("local%d", next)
					locals[id.Obj] = name
					id.Name = name
				}
			}
			return true
		})
	}
	var b bytes.Buffer
	if e = format.Node(&b, token.NewFileSet(), f); e != nil {
		return ""
	}
	return issuerBindingHash(b.Bytes())
}
func issuerBindingSchemas() map[string]string {
	return map[string]string{
		"budget.go":            "a387e0d953d468d75d14b4950b6b0e62d30c81b7d06633583e3981880a3194fe",
		"doc.go":               "c246247fe06f0058f0319a1c09dcf1b7bc5c4963cbcb3161d0f862a0f62fe017",
		"errors.go":            "c751670ed80a547f6c05dcfc18c3b88c3d8ea511bfb60b771101673159873d5b",
		"evaluate.go":          "17ff23f492da1cd51d717c1e741d5c960648d5ce6b27488b9857fafaf75d90cb",
		"input.go":             "12d3dcb901a63ba74934d5f7578a0dc7ead0df7b3189bca05e887ef3681fd5bd",
		"lower.go":             "ba034d7f8691eb1920fa6ec4751bff8600131d642ccc53c998f6b1b6c0c4b653",
		"model.go":             "3f7e129e861ae4a98942a3e97e99bcb7c26f0651a99c1aaefe7e2785093b72a7",
		"owned_binding.go":     "8f95d022e99fc1848663c0b95d6cec9906e55a8f74a01d4a5f1d840fd8f873a9",
		"owned_certificate.go": "0bcf451033603dca958a081d6c62714e7569daa5c14b88c0c0bf792c9febfba7",
		"owned_content.go":     "032900b55aa7e190ed3697319375d89898a86d17b6641d9b66b7a2f4dad38ee6",
		"owned_policy.go":      "52a4f6c145ef02a72d0c9d836a4d2745429002660253bce4035e1e54890a88f3",
		"owned_summary.go":     "509fb59f898960d1c4f0e65807fcb6e086bcdb7092b50a6aee1ba65e3daa73b6",
		"owned_table.go":       "9cd38f3897c39368be9253dcea36a659dd283881084b3c9b69463dd7757fddcd",
		"owned_time.go":        "78558824db2e56737471cba28bc25a2de04c0863e0ad16d0c018fbece80a49b3",
		"owner.go":             "978a89ea0439b6a68acffd643bd466aad589b4d46599b1bdf232c649187b1b8c",
		"projection.go":        "88057267103d0ddac976e68346762a2603be5aca58ab297e733ac8bb1223973a",
		"references.go":        "bb91fe6ecb2e7b4133c974d961230b09a41cf7cafd41e52e8d6a3ddd66cd182c",
	}
}
func issuerBindingCapabilityAllowed(sources map[string][]byte) bool {
	want := map[string]bool{}
	for _, n := range strings.Fields(issuerBindingGoInventory) {
		want[n] = true
	}
	if len(sources) != len(want) {
		return false
	}
	schemas := issuerBindingSchemas()
	for name, raw := range sources {
		if !want[name] {
			return false
		}
		f, e := parser.ParseFile(token.NewFileSet(), name, raw, parser.ImportsOnly)
		if e != nil || f.Name.Name != "nodeissuerbinding" {
			return false
		}
		for _, im := range f.Imports {
			p, e := strconv.Unquote(im.Path.Value)
			if e != nil {
				return false
			}
			for _, bad := range strings.Fields("stateformat statecomposition stateverify statetime nodeca nodeleaf nodeleafkeys") {
				prefix := "talenro.local/platform/internal/nodecontrol/" + bad
				if p == prefix || strings.HasPrefix(p, prefix+"/") {
					return false
				}
			}
		}
		if !strings.HasSuffix(name, "_test.go") {
			if !issuerBindingProductAllowed(raw) || schemas[name] == "" || issuerBindingASTHash(raw) != schemas[name] {
				return false
			}
		}
	}
	return len(schemas) == 17
}
