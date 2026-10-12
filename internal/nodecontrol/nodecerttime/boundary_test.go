package nodecerttime

import (
	"bytes"
	"encoding/json"
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

const certTimeGoInventory = "doc.go input.go errors.go budget.go owner.go owned_certificate.go owned_binding.go owned_content.go owned_table.go owned_policy.go owned_time.go owned_summary.go model.go lower.go binding.go window.go evaluate.go input_test.go budget_test.go owned_certificate_test.go owned_binding_test.go owned_content_test.go owned_table_test.go owned_policy_test.go owned_time_test.go owned_summary_test.go model_test.go fixtures_test.go lower_test.go binding_test.go window_test.go evaluate_test.go boundary_test.go ownership_boundary_test.go consumer_boundary_test.go"

func certTimeSources(t testing.TB) map[string][]byte {
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
func certTimeMutation(t testing.TB, sources map[string][]byte, file, old, replacement string) map[string][]byte {
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
func TestCertTimeProductBoundary(t *testing.T) {
	src := certTimeSources(t)
	for name, b := range src {
		if !strings.HasSuffix(name, "_test.go") && !certTimeProductAllowed(b) {
			t.Fatal("actual product", name)
		}
	}
	for _, positive := range []string{
		"package nodecerttime; import c \"context\";func f(ctx c.Context)error{return ctx.Err()}",
		"package nodecerttime;import clock \"time\";func normalizeTime(t clock.Time)clock.Time{return t.Round(0).UTC()}",
	} {
		if !certTimeProductAllowed([]byte(positive)) {
			t.Fatal("safe named alias/normalization")
		}
	}
	for _, negative := range []string{
		"package nodecerttime;import \"os\";func f(){os.Getenv(\"X\")}",
		"package nodecerttime;import \"time\";func f(){time.Now()}",
		"package nodecerttime;import \"context\";func f(ctx context.Context){go ctx.Err()}",
		"package nodecerttime;func f(){recover()}",
		"package nodecerttime;import . \"time\";var x Time",
		"package nodecerttime;import _ \"time\"",
		"package nodecerttime;import c \"context\";import d \"context\"",
		"package nodecerttime;import \"context\";func f(ctx context.Context){context:=ctx;context.Err()}",
		"package nodecerttime;import \"time\";func f(t time.Time){t.Round(1).UTC()}",
		"package nodecerttime;import \"time\";func f(t time.Time,n time.Time){t.Add(0);t.Unix();t.Equal(n)}",
		"package nodecerttime;import \"context\";func f(ctx context.Context){g:=ctx.Err;g()}",
		"package nodecerttime;import \"talenro.local/platform/internal/nodecontrol/nodeissuer\";var f=nodeissuer.Evaluate",
		"package nodecerttime;import \"talenro.local/platform/internal/nodecontrol/nodeissuer\";var x=nodeissuer.Material{}",
		"package nodecerttime;import \"talenro.local/platform/internal/nodecontrol/statetime\";var x=statetime.ModelInput{}",
	} {
		if certTimeProductAllowed([]byte(negative)) {
			t.Fatal("forbidden capability", negative)
		}
	}
}
func TestCertTimeCapabilityBoundary(t *testing.T) {
	src := certTimeSources(t)
	if !certTimeCapabilityAllowed(src) {
		hashes := map[string]string{}
		for k, v := range src {
			if !strings.HasSuffix(k, "_test.go") {
				hashes[k] = certTimeASTHash(v)
			}
		}
		b, _ := json.Marshal(hashes)
		t.Log("AST schema diagnostic", string(b))
		t.Fatal("actual inventory/API/flow")
	}
	alias := certTimeMutation(t, src, "lower.go", "\"context\"", "c \"context\"")
	alias["lower.go"] = []byte(strings.ReplaceAll(string(alias["lower.go"]), "context.Context", "c.Context"))
	if !certTimeCapabilityAllowed(alias) {
		t.Fatal("safe named import alias")
	}
	rename := certTimeMutation(t, src, "lower.go", "ctx context.Context", "operationCtx context.Context")
	rename["lower.go"] = []byte(strings.ReplaceAll(string(src["lower.go"]), "ctx", "operationCtx"))
	if !certTimeCapabilityAllowed(rename) {
		t.Fatal("safe bound local rename")
	}
	for _, tc := range []struct{ file, old, replacement string }{
		{"input.go", "None Proof = iota", "None Proof = 1"},
		{"input.go", "ObservedAt, ValidUntil time.Time", "ObservedAt, ValidUntil time.Time; Extra bool"},
		{"errors.go", "nodecerttime: invalid input", "nodecerttime: changed"},
		{"owner.go", "valid   bool", "valid   bool; hook func()"},
		{"owner.go", "in.Round(0).UTC()", "in.Round(1).UTC()"},
		{"lower.go", "nodeissuer.Evaluate(ctx, o.input.Certificate)", "nodeissuer.Evaluate(ctx, nodeissuer.Input{})"},
		{"binding.go", "node.String()", "node.ID()"},
	} {
		if certTimeCapabilityAllowed(certTimeMutation(t, src, tc.file, tc.old, tc.replacement)) {
			t.Fatal("capability/API mutation", tc.file, tc.old)
		}
	}
	for _, extra := range []string{"extra.go", "nested/extra.go", "nodecerttime_extra.go"} {
		bad := map[string][]byte{}
		for k, v := range src {
			bad[k] = v
		}
		bad[extra] = []byte("package nodecerttime")
		if certTimeCapabilityAllowed(bad) {
			t.Fatal("extra inventory", extra)
		}
	}
	extra := certTimeMutation(t, src, "doc.go", "package nodecerttime", "package nodecerttime\nvar Hook = func(){}")
	if certTimeCapabilityAllowed(extra) {
		t.Fatal("global hook")
	}
}

func certTimeUnparen(e ast.Expr) ast.Expr {
	for {
		p, ok := e.(*ast.ParenExpr)
		if !ok {
			return e
		}
		e = p.X
	}
}
func certTimeProductAllowed(raw []byte) bool {
	f, e := parser.ParseFile(token.NewFileSet(), "product.go", raw, 0)
	if e != nil || f.Name.Name != "nodecerttime" {
		return false
	}
	allowed := map[string]string{
		"context": "Context", "errors": "New", "bytes": "Clone Compare", "sort": "Slice", "time": "Time",
		"talenro.local/platform/internal/nodecontrol/nodeissuer":         "Input Summary Evaluate NodeLeafDirectIssuerSignatureBound",
		"talenro.local/platform/internal/nodecontrol/statetimebinding":   "Input Summary Evaluate ConditionalTimeEvidenceBound",
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
	methods := map[string]int{"Err": 0, "Round": 1, "UTC": 0, "IsZero": 0, "Year": 0, "Before": 1, "Summary": 0, "Clear": 0, "String": 0, "clear": 0}
	ok := true
	ast.Inspect(f, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.GoStmt, *ast.SendStmt, *ast.ChanType, *ast.SelectStmt:
			ok = false
		case *ast.TypeSpec:
			if v.Assign.IsValid() {
				ok = false
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
					if call, is := parents[v].(*ast.CallExpr); !is || certTimeUnparen(call.Fun) != v {
						ok = false
					}
				}
			}
		case *ast.SelectorExpr:
			if x, is := certTimeUnparen(v.X).(*ast.Ident); is && aliases[x.Name] != "" && x.Obj == nil {
				if !strings.Contains(" "+allowed[aliases[x.Name]]+" ", " "+v.Sel.Name+" ") {
					ok = false
				}
				if v.Sel.Name == "Evaluate" || v.Sel.Name == "New" || v.Sel.Name == "Clone" || v.Sel.Name == "Compare" || v.Sel.Name == "Slice" {
					c, is := parents[v].(*ast.CallExpr)
					if !is || certTimeUnparen(c.Fun) != v {
						ok = false
					}
				}
			} else if _, is := methods[v.Sel.Name]; is {
				c, is := parents[v].(*ast.CallExpr)
				if !is || certTimeUnparen(c.Fun) != v {
					ok = false
				}
			}
		case *ast.CallExpr:
			switch fun := certTimeUnparen(v.Fun).(type) {
			case *ast.FuncLit:
				if _, deferred := parents[v].(*ast.DeferStmt); !deferred {
					ok = false
				}
			case *ast.Ident:
				if fun.Name == "recover" || fun.Name == "panic" {
					ok = false
				}
			case *ast.SelectorExpr:
				if x, is := certTimeUnparen(fun.X).(*ast.Ident); is && aliases[x.Name] != "" && x.Obj == nil {
					break
				}
				arity, exists := methods[fun.Sel.Name]
				if !exists || len(v.Args) != arity || v.Ellipsis.IsValid() {
					ok = false
					break
				}
				if fun.Sel.Name == "Round" {
					arg, is := certTimeUnparen(v.Args[0]).(*ast.BasicLit)
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
func certTimeASTHash(raw []byte) string {
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
	return certTimeHash(b.Bytes())
}
func certTimeSchemas() map[string]string {
	return map[string]string{
		"binding.go":           "d3f402317e83fa60b65414b548b7e70ac0cf1e808327e9e635bb2b5c8eadeb96",
		"budget.go":            "c9ab8734875c85905c965710755613f0d1f0dd1292e46e724f94bc26946a4716",
		"doc.go":               "fd679859d467193f5ddc38efc063b1918cdbc2aa28ad90c7fd798aad4e7cbd8e",
		"errors.go":            "b91508a35f09a4f98c7ec6b558306b63f8234fb254e8c24384f34de6950dbec2",
		"evaluate.go":          "be8575f3e5c64b6f522d11840570a4edbaa8ba8882de9361449b8c79411f1b11",
		"input.go":             "cebbecbc5f4945dfb1fc0d8c70da5c8410f318ee2253602c771e31474691b6e4",
		"lower.go":             "6eda3010edf12bcfc6856d7be9c289931debf29f36968b95c96a4d3a0788b270",
		"model.go":             "4ac68834ad30223a563c01b8104e4e9088701c5c508adc0c3585215539d9864e",
		"owned_binding.go":     "a3af2743171453f736a0a84baa94cefbaafa43f0b01aa67b464580fd925e8e80",
		"owned_certificate.go": "6e0ab5d0152099a68d6613763ee9fc73767038106d4b7f6b3b9e883a1ce243c9",
		"owned_content.go":     "f254653904e57470d099b4d839a076602142a768cf12756a191bc9dd4048abda",
		"owned_policy.go":      "f435ce7fad87c525aed9e2c03af7de46d01bc0bb5fffd760684b25c2b117bb21",
		"owned_summary.go":     "5af1d040e1b7de457a8a8cce11069e4bf3059e1b987f46d7f07a7d69970a27b0",
		"owned_table.go":       "10d7c057e6ac97f21f6c9f1a215ffa9f82b8ac1bebdeb7420a534a64f41ca85b",
		"owned_time.go":        "29604eb575fb162ff8ffb7901f22957b380bdc76016a3211cbb270f23ae4be93",
		"owner.go":             "774de2444e671f44c795fbcfc3202c6e9320e37c8bf1957d98d2e9c877bbf31f",
		"window.go":            "c04b5687d99ce6cd4229204f96440021eedd4da41c37279e7c99ebbf1daa339a",
	}
}
func certTimeCapabilityAllowed(sources map[string][]byte) bool {
	want := map[string]bool{}
	for _, n := range strings.Fields(certTimeGoInventory) {
		want[n] = true
	}
	if len(sources) != len(want) {
		return false
	}
	schemas := certTimeSchemas()
	for name, raw := range sources {
		if !want[name] {
			return false
		}
		f, e := parser.ParseFile(token.NewFileSet(), name, raw, parser.ImportsOnly)
		if e != nil || f.Name.Name != "nodecerttime" {
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
			if !certTimeProductAllowed(raw) || schemas[name] == "" || certTimeASTHash(raw) != schemas[name] {
				return false
			}
		}
	}
	return len(schemas) == 17
}
