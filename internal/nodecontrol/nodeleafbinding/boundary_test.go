package nodeleafbinding

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

const leafBindingGoInventory = "doc.go input.go errors.go budget.go owner.go owned_certificate.go owned_binding.go owned_content.go owned_table.go owned_policy.go owned_time.go owned_summary.go model.go lower.go projection.go evaluate.go input_test.go budget_test.go owned_certificate_test.go owned_binding_test.go owned_content_test.go owned_table_test.go owned_policy_test.go owned_time_test.go owned_summary_test.go model_test.go fixtures_test.go lower_test.go projection_test.go evaluate_test.go boundary_test.go ownership_boundary_test.go consumer_boundary_test.go"

func leafBindingSources(t testing.TB) map[string][]byte {
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
func leafBindingMutation(t testing.TB, sources map[string][]byte, file, old, replacement string) map[string][]byte {
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
func TestLeafBindingProductBoundary(t *testing.T) {
	src := leafBindingSources(t)
	for name, b := range src {
		if !strings.HasSuffix(name, "_test.go") && !leafBindingProductAllowed(b) {
			t.Fatal("actual product", name)
		}
	}
	for _, positive := range []string{
		"package nodeleafbinding; import c \"context\";func f(ctx c.Context)error{return ctx.Err()}",
		"package nodeleafbinding;import clock \"time\";func normalizeTime(t clock.Time)clock.Time{return t.Round(0).UTC()}",
	} {
		if !leafBindingProductAllowed([]byte(positive)) {
			t.Fatal("safe named alias/normalization")
		}
	}
	for _, negative := range []string{
		"package nodeleafbinding;import \"os\";func f(){os.Getenv(\"X\")}",
		"package nodeleafbinding;import \"time\";func f(){time.Now()}",
		"package nodeleafbinding;import \"context\";func f(ctx context.Context){go ctx.Err()}",
		"package nodeleafbinding;func f(){recover()}",
		"package nodeleafbinding;import . \"time\";var x Time",
		"package nodeleafbinding;import _ \"time\"",
		"package nodeleafbinding;import c \"context\";import d \"context\"",
		"package nodeleafbinding;import \"context\";func f(ctx context.Context){context:=ctx;context.Err()}",
		"package nodeleafbinding;import \"time\";func f(t time.Time){t.Round(1).UTC()}",
		"package nodeleafbinding;import \"time\";func f(t time.Time,n time.Time){t.Add(0);t.Unix();t.Equal(n)}",
		"package nodeleafbinding;import \"context\";func f(ctx context.Context){g:=ctx.Err;g()}",
		"package nodeleafbinding;import \"talenro.local/platform/internal/nodecontrol/nodeissuer\";var f=nodeissuer.Evaluate",
		"package nodeleafbinding;import \"talenro.local/platform/internal/nodecontrol/nodeissuer\";var x=nodeissuer.Material{}",
		"package nodeleafbinding;import \"talenro.local/platform/internal/nodecontrol/statetime\";var x=statetime.ModelInput{}",
	} {
		if leafBindingProductAllowed([]byte(negative)) {
			t.Fatal("forbidden capability", negative)
		}
	}
}
func TestLeafBindingCapabilityBoundary(t *testing.T) {
	src := leafBindingSources(t)
	if !leafBindingCapabilityAllowed(src) {
		for name, raw := range src {
			if !strings.HasSuffix(name, "_test.go") {
				t.Logf("schema %s %s", name, leafBindingASTHash(raw))
			}
		}
		t.Fatal("actual inventory/API/flow")
	}
	alias := leafBindingMutation(t, src, "lower.go", "\"context\"", "c \"context\"")
	alias["lower.go"] = []byte(strings.ReplaceAll(string(alias["lower.go"]), "context.Context", "c.Context"))
	if !leafBindingCapabilityAllowed(alias) {
		t.Fatal("safe named import alias")
	}
	rename := leafBindingMutation(t, src, "lower.go", "ctx context.Context", "operationCtx context.Context")
	rename["lower.go"] = []byte(strings.ReplaceAll(string(src["lower.go"]), "ctx", "operationCtx"))
	if !leafBindingCapabilityAllowed(rename) {
		t.Fatal("safe bound local rename")
	}
	for _, tc := range []struct{ file, old, replacement string }{
		{"input.go", "None Proof = iota", "None Proof = 1"},
		{"input.go", "ObservedAt, ValidUntil time.Time", "ObservedAt, ValidUntil time.Time; Extra bool"},
		{"errors.go", "nodeleafbinding: invalid input", "nodeleafbinding: changed"},
		{"owner.go", "valid   bool", "valid   bool; hook func()"},
		{"owner.go", "in.Round(0).UTC()", "in.Round(1).UTC()"},
		{"lower.go", "nodecerttime.Evaluate(ctx, o.input)", "nodecerttime.Evaluate(ctx, nodecerttime.Input{})"},
		{"projection.go", "leaf.NodeID.String()", "leaf.NodeID.ID()"},
	} {
		if leafBindingCapabilityAllowed(leafBindingMutation(t, src, tc.file, tc.old, tc.replacement)) {
			t.Fatal("capability/API mutation", tc.file, tc.old)
		}
	}
	for _, extra := range []string{"extra.go", "nested/extra.go", "nodeleafbinding_extra.go"} {
		bad := map[string][]byte{}
		for k, v := range src {
			bad[k] = v
		}
		bad[extra] = []byte("package nodeleafbinding")
		if leafBindingCapabilityAllowed(bad) {
			t.Fatal("extra inventory", extra)
		}
	}
	extra := leafBindingMutation(t, src, "doc.go", "package nodeleafbinding", "package nodeleafbinding\nvar Hook = func(){}")
	if leafBindingCapabilityAllowed(extra) {
		t.Fatal("global hook")
	}
}
func leafBindingUnparen(e ast.Expr) ast.Expr {
	for {
		p, ok := e.(*ast.ParenExpr)
		if !ok {
			return e
		}
		e = p.X
	}
}
func leafBindingProductAllowed(raw []byte) bool {
	f, e := parser.ParseFile(token.NewFileSet(), "product.go", raw, 0)
	if e != nil || f.Name.Name != "nodeleafbinding" {
		return false
	}
	allowed := map[string]string{
		"talenro.local/platform/internal/nodecontrol/contracts":    "Digest",
		"talenro.local/platform/internal/nodecontrol/nodecerttime": "Input Summary Evaluate NodeCertificateConditionalTimeBound",
		"context": "Context", "errors": "New", "bytes": "Clone Compare", "sort": "Slice", "time": "Time Unix",
		"talenro.local/platform/internal/nodecontrol/nodeissuer":         "Input",
		"talenro.local/platform/internal/nodecontrol/statetimebinding":   "Input Summary",
		"talenro.local/platform/internal/nodecontrol/statepolicybinding": "Input",
		"talenro.local/platform/internal/nodecontrol/statebinding":       "Input",
		"talenro.local/platform/internal/nodecontrol/statecontract":      "MainStateContentModel TimeWindow DomainIdentity Present CertificateDER",
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
	methods := map[string]int{"Err": 0, "Round": 1, "UTC": 0, "IsZero": 0, "Year": 0, "Before": 1, "Equal": 1, "Summary": 0, "Clear": 0, "String": 0, "clear": 0}
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
				if !is || aliases[id.Name] != "talenro.local/platform/internal/nodecontrol/nodecerttime" {
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
					if call, is := parents[v].(*ast.CallExpr); !is || leafBindingUnparen(call.Fun) != v {
						ok = false
					}
				}
			}
		case *ast.SelectorExpr:
			if x, is := leafBindingUnparen(v.X).(*ast.Ident); is && aliases[x.Name] != "" && x.Obj == nil {
				if !strings.Contains(" "+allowed[aliases[x.Name]]+" ", " "+v.Sel.Name+" ") {
					ok = false
				}
				if v.Sel.Name == "Unix" || v.Sel.Name == "Evaluate" || v.Sel.Name == "New" || v.Sel.Name == "Clone" || v.Sel.Name == "Compare" || v.Sel.Name == "Slice" {
					c, is := parents[v].(*ast.CallExpr)
					if !is || leafBindingUnparen(c.Fun) != v {
						ok = false
					}
				}
			} else if _, is := methods[v.Sel.Name]; is {
				c, is := parents[v].(*ast.CallExpr)
				if !is || leafBindingUnparen(c.Fun) != v {
					ok = false
				}
			}
		case *ast.CallExpr:
			switch fun := leafBindingUnparen(v.Fun).(type) {
			case *ast.FuncLit:
				if _, deferred := parents[v].(*ast.DeferStmt); !deferred {
					ok = false
				}
			case *ast.Ident:
				if fun.Name == "recover" || fun.Name == "panic" {
					ok = false
				}
			case *ast.SelectorExpr:
				if x, is := leafBindingUnparen(fun.X).(*ast.Ident); is && aliases[x.Name] != "" && x.Obj == nil {
					break
				}
				arity, exists := methods[fun.Sel.Name]
				if !exists || len(v.Args) != arity || v.Ellipsis.IsValid() {
					ok = false
					break
				}
				if fun.Sel.Name == "Round" {
					arg, is := leafBindingUnparen(v.Args[0]).(*ast.BasicLit)
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
func leafBindingASTHash(raw []byte) string {
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
	return leafBindingHash(b.Bytes())
}
func leafBindingSchemas() map[string]string {
	return map[string]string{
		"budget.go":            "079e05137cf965e52bc4359ff26d1a1d87651074c006533cb58c4140a62b7f56",
		"doc.go":               "8e6f94c9c619fb92cdcdb43ee758140740062a9cad7b9f0da4f0a0e96a28ecc7",
		"errors.go":            "a8d9015a31ac42b9a45c5a67b78edc3f7f1d93c5a2c35ee57241e5c08eb56e50",
		"evaluate.go":          "5039630e17d57df47493490efe6c7eb84c8e1a47732598c002434b71747c74ac",
		"input.go":             "932ac67b01e261aba73c90f406b3b69564769f649c742e3f9b0e293703b0a6a0",
		"lower.go":             "8c870d3524f71dfd4c44c6d8457022f74b9d5b6acdcc8ea297cf5d9cf21b5ae9",
		"model.go":             "c7ee83283bc4ba8d222404ebde151e9eb2fea117c1ea00f9e35902a7560b7f7d",
		"owned_binding.go":     "6e23570e0202836d8f4dc2477a1bd07501aa390b700070588aa288bfc4e1b080",
		"owned_certificate.go": "feabbadabb4ee40f5349ce53fa9350bf4f678e2d8b6810fb2295b857009eeac9",
		"owned_content.go":     "24b7d5e63e3516f23e56d3bd96d1295b0ee500eb1e63e29dad825220a2a089d2",
		"owned_policy.go":      "418f923025ba9c87f86193410784da192f54ff9a6c6503ed4507edf2c8052b3b",
		"owned_summary.go":     "71ac9c87d0831c53da77dbaa47348a1f3fcc7184e2eed3757baced4cd1b69cdd",
		"owned_table.go":       "8946792572b5ab9a820001bb6eeec7dd7be133296782284bcbbcb62c59cdb8eb",
		"owned_time.go":        "c9e0cb953f9470cb8a2b1e4066c6dcd69a9934aeec11e6fe0f8c3f3a561f71ec",
		"owner.go":             "8fe0b53ea971b991fd44dd531348b72f9486bb4b30a2a7ae5e4a1a98328f494c",
		"projection.go":        "551df18374a6bb3d17dd67a16a45583594b9d1a947a07b342c29579a3e4984e4",
	}
}
func leafBindingCapabilityAllowed(sources map[string][]byte) bool {
	want := map[string]bool{}
	for _, n := range strings.Fields(leafBindingGoInventory) {
		want[n] = true
	}
	if len(sources) != len(want) {
		return false
	}
	schemas := leafBindingSchemas()
	for name, raw := range sources {
		if !want[name] {
			return false
		}
		f, e := parser.ParseFile(token.NewFileSet(), name, raw, parser.ImportsOnly)
		if e != nil || f.Name.Name != "nodeleafbinding" {
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
			if !leafBindingProductAllowed(raw) || schemas[name] == "" || leafBindingASTHash(raw) != schemas[name] {
				return false
			}
		}
	}
	return len(schemas) == 16
}
