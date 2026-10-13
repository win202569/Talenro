package nodeissuermapping

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
import "go/scanner"

const issuerMappingGoInventory = "doc.go input.go errors.go budget.go owner.go owned_claims.go owned_summary.go model.go claims.go lower.go projection.go check.go input_test.go budget_test.go owned_claims_test.go owned_summary_test.go model_test.go fixtures_test.go claims_test.go lower_test.go projection_test.go check_test.go boundary_test.go ownership_boundary_test.go consumer_boundary_test.go"

func TestIssuerMappingProductBoundary(t *testing.T) {
	for name, b := range issuerMappingSources(t) {
		if !strings.HasSuffix(name, "_test.go") && !issuerMappingProductAllowed(b) {
			t.Fatal("actual product", name)
		}
	}
	for _, positive := range []string{
		"package nodeissuermapping;import c \"context\";func f(ctx c.Context)error{return ctx.Err()}",
		"package nodeissuermapping;import clock \"time\";func normalizeTime(t clock.Time)clock.Time{return t.Round(0).UTC()}",
	} {
		if !issuerMappingProductAllowed([]byte(positive)) {
			t.Fatal("benign", positive)
		}
	}
	for _, negative := range []string{
		"package nodeissuermapping;import \"os\";func f(){os.Getenv(\"X\")}",
		"package nodeissuermapping;import \"time\";func f(){time.Now()}",
		"package nodeissuermapping;import \"context\";func f(ctx context.Context){go ctx.Err()}",
		"package nodeissuermapping;func f(){recover()}",
		"package nodeissuermapping;func f(){append([]int{},1)}",
		"package nodeissuermapping;import . \"time\";var x Time",
		"package nodeissuermapping;import _ \"time\"",
		"package nodeissuermapping;import c \"context\";import d \"context\"",
		"package nodeissuermapping;import \"context\";func f(ctx context.Context){context:=ctx;context.Err()}",
		"package nodeissuermapping;import \"time\";func f(t time.Time){t.Round(1).UTC()}",
		"package nodeissuermapping;import \"time\";func f(t time.Time){t.Add(0)}",
		"package nodeissuermapping;import \"context\";func f(ctx context.Context){g:=ctx.Err;g()}",
		"package nodeissuermapping;import \"talenro.local/platform/internal/nodecontrol/nodeissuerbinding\";var f=nodeissuerbinding.Evaluate",
		"package nodeissuermapping;import \"talenro.local/platform/internal/nodecontrol/nodeleafbinding\";var x=nodeleafbinding.Input{}",
		"package nodeissuermapping;import \"crypto/sha256\";func f(){sha256.Sum256(nil)}",
	} {
		if issuerMappingProductAllowed([]byte(negative)) {
			t.Fatal("forbidden", negative)
		}
	}
}
func TestIssuerMappingCapabilityBoundary(t *testing.T) {
	src := issuerMappingSources(t)
	if !issuerMappingCapabilityAllowed(src) {
		for name, b := range src {
			if !strings.HasSuffix(name, "_test.go") {
				t.Logf("schema %s %s", name, issuerMappingASTHash(b))
			}
		}
		t.Fatal("actual inventory/API/flow")
	}
	formatted := map[string][]byte{}
	for name, raw := range src {
		out, e := format.Source(raw)
		if e != nil {
			t.Fatal(e)
		}
		formatted[name] = out
	}
	if !issuerMappingCapabilityAllowed(formatted) {
		t.Fatal("format")
	}
	alias := issuerMappingMutation(t, src, "lower.go", "\"context\"", "c \"context\"")
	alias["lower.go"] = []byte(strings.ReplaceAll(string(alias["lower.go"]), "context.Context", "c.Context"))
	if !issuerMappingCapabilityAllowed(alias) {
		t.Fatal("named alias")
	}
	rename := issuerMappingMutation(t, src, "lower.go", "ctx context.Context", "operationCtx context.Context")
	rename["lower.go"] = []byte(strings.ReplaceAll(string(src["lower.go"]), "ctx", "operationCtx"))
	if !issuerMappingCapabilityAllowed(rename) {
		t.Fatal("bound local rename")
	}
	for _, c := range []struct{ file, old, replacement string }{
		{"input.go", "None Proof = iota", "None Proof = 1"},
		{"input.go", "type Model struct", "type TrustedRegistry struct"},
		{"errors.go", "nodeissuermapping: invalid input", "nodeissuermapping: changed"},
		{"owner.go", "valid bool", "valid bool; material Input"},
		{"owner.go", "in.Round(0).UTC()", "in.Round(1).UTC()"},
		{"budget.go", "maxDeclaredRawBytes", "maxDeclaredRawBytes+1"},
		{"claims.go", "r.IssuerID==other.IssuerID&&r.CAVersion==other.CAVersion", "false"},
		{"claims.go", "r.IssuerID==a.IssuerID&&r.CAVersion==a.CAVersion", "r.IssuerID==a.IssuerID"},
		{"claims.go", "r.State!=DeclaredEligible", "false"},
		{"projection.go", "selected.CADERDigest!=s.Evidence.Evidence.Certificate.CA.CADERDigest", "false"},
		{"projection.go", "!bytes.Equal(a.SerialBytes,leaf.Serial[:leaf.SerialLength])", "false"},
		{"check.go", "o.summary.Proof=DeclaredIssuerSerialCorrespondenceConditionalTimeBound", "o.summary.Proof=None"},
		{"check.go", "AuthenticatedIssuanceLinkAndRowStatus", "UnknownObligation"},
	} {
		if issuerMappingCapabilityAllowed(issuerMappingMutation(t, src, c.file, c.old, c.replacement)) {
			t.Fatal("mutation", c.file, c.old)
		}
	}
	for _, extra := range []string{"extra.go", "nested/extra.go", "other_test.go"} {
		bad := map[string][]byte{}
		for k, v := range src {
			bad[k] = v
		}
		bad[extra] = []byte("package nodeissuermapping")
		if issuerMappingCapabilityAllowed(bad) {
			t.Fatal(extra)
		}
	}
	bad := issuerMappingMutation(t, src, "doc.go", "package nodeissuermapping", "package nodeissuermapping\nvar Hook = func(){}")
	if issuerMappingCapabilityAllowed(bad) {
		t.Fatal("extra capability")
	}
}
func issuerMappingSources(t testing.TB) map[string][]byte {
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
func issuerMappingMutation(t testing.TB, sources map[string][]byte, file, old, replacement string) map[string][]byte {
	t.Helper()
	out := map[string][]byte{}
	for k, v := range sources {
		out[k] = bytes.Clone(v)
	}
	b, ok := out[file]
	if !ok || old == replacement {
		t.Fatal("inactive mutation", file, old)
	}
	if strings.Contains(string(b), old) {
		out[file] = []byte(strings.Replace(string(b), old, replacement, 1))
	} else {
		sourceTokens, oldTokens := issuerMappingTokens(b), issuerMappingTokens([]byte(old))
		changed := false
		for i := 0; i+len(oldTokens) <= len(sourceTokens) && len(oldTokens) > 0; i++ {
			equal := true
			for j := range oldTokens {
				if sourceTokens[i+j].key != oldTokens[j].key {
					equal = false
					break
				}
			}
			if equal {
				first, last := sourceTokens[i], sourceTokens[i+len(oldTokens)-1]
				out[file] = []byte(string(b[:first.start]) + replacement + string(b[last.end:]))
				changed = true
				break
			}
		}
		if !changed {
			t.Fatal("inactive token mutation", file, old)
		}
	}
	if bytes.Equal(out[file], b) {
		t.Fatal("inactive mutation")
	}
	return out
}
func issuerMappingUnparen(e ast.Expr) ast.Expr {
	for {
		p, ok := e.(*ast.ParenExpr)
		if !ok {
			return e
		}
		e = p.X
	}
}
func issuerMappingProductAllowed(raw []byte) bool {
	f, e := parser.ParseFile(token.NewFileSet(), "product.go", raw, 0)
	if e != nil || f.Name.Name != "nodeissuermapping" {
		return false
	}
	allowed := map[string]string{
		"talenro.local/platform/internal/nodecontrol/contracts":         "Digest",
		"talenro.local/platform/internal/nodecontrol/nodeissuerbinding": "Input Summary Projection Evaluate SelectedIssuerSerialReferencesConditionalTimeBound",
		"context": "Context", "errors": "New", "bytes": "Clone Equal", "time": "Time", "github.com/google/uuid": "UUID Nil RFC4122",
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
	methods := map[string]int{"Err": 0, "Round": 1, "UTC": 0, "IsZero": 0, "Year": 0, "Before": 1, "Summary": 0, "Clear": 0, "clear": 0, "Variant": 0}
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
					if call, is := parents[v].(*ast.CallExpr); !is || issuerMappingUnparen(call.Fun) != v {
						ok = false
					}
				}
			}
		case *ast.SelectorExpr:
			if x, is := issuerMappingUnparen(v.X).(*ast.Ident); is && aliases[x.Name] != "" && x.Obj == nil {
				if !strings.Contains(" "+allowed[aliases[x.Name]]+" ", " "+v.Sel.Name+" ") {
					ok = false
				}
				if v.Sel.Name == "Sum256" || v.Sel.Name == "Evaluate" || v.Sel.Name == "New" || v.Sel.Name == "Clone" || v.Sel.Name == "Equal" || v.Sel.Name == "Slice" {
					c, is := parents[v].(*ast.CallExpr)
					if !is || issuerMappingUnparen(c.Fun) != v {
						ok = false
					}
				}
			} else if _, is := methods[v.Sel.Name]; is {
				c, is := parents[v].(*ast.CallExpr)
				if !is || issuerMappingUnparen(c.Fun) != v {
					ok = false
				}
			}
		case *ast.CallExpr:
			switch fun := issuerMappingUnparen(v.Fun).(type) {
			case *ast.FuncLit:
				if _, deferred := parents[v].(*ast.DeferStmt); !deferred {
					ok = false
				}
			case *ast.Ident:
				if fun.Name == "recover" || fun.Name == "panic" || fun.Name == "append" {
					ok = false
				}
			case *ast.SelectorExpr:
				if x, is := issuerMappingUnparen(fun.X).(*ast.Ident); is && aliases[x.Name] != "" && x.Obj == nil {
					break
				}
				arity, exists := methods[fun.Sel.Name]
				if !exists || len(v.Args) != arity || v.Ellipsis.IsValid() {
					ok = false
					break
				}
				if fun.Sel.Name == "Round" {
					arg, is := issuerMappingUnparen(v.Args[0]).(*ast.BasicLit)
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
func issuerMappingASTHash(raw []byte) string {
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
	return issuerMappingHash(b.Bytes())
}
func issuerMappingSchemas() map[string]string {
	return map[string]string{
		"errors.go":        "b61ada1cbff5ca33c5d11d56972dd8e36b411e0d10c4836c2eada45bcfbe043b",
		"input.go":         "987f8d83ad3cd59cbfc57997569296c4b3560d83d5974e6adb4dec6ee92d7f8f",
		"check.go":         "32c502d6da5bd6cf2fe9d4774387598c9fcff331980b13252794b53c610c62b4",
		"lower.go":         "491d6ea23c372249b3fed3910f7392396213b48543c8ed27cfbfedaafd306dd8",
		"model.go":         "4e4724b04baee70b533c62e95315de1806c3ecdcf7dffd2f9f0deddf6e8ddf2a",
		"owned_claims.go":  "c980eb19275343f80c7cf381339161227edbd66e8dbb16f5f2740b1ab90ed970",
		"budget.go":        "d466364af188cc6477b797671aa2aef3f8efc790b4b889002e0d3cc773935adc",
		"claims.go":        "47f607d5344d5031b3d8c594f70f5bd43ae9fc111fa779ff0a79ebc4d193dea1",
		"owned_summary.go": "8c2e4d9d59f79599d59fd25911d94306659d7d1453b98d311c065a3cc0ba1af6",
		"owner.go":         "573fda6e432ad2e81488c5f480e1bc45871b4f16a186b217700fd675d6c712c8",
		"projection.go":    "7a4b35d767d9fa65e3d0bcc52637ed274c9256d2b18fb0ddb4094abfcd740e57",
		"doc.go":           "edf6b98ac676b80c98898b53f99d9ed0aa82583394c4c6cda319d247b84a55cb",
	}
}
func issuerMappingCapabilityAllowed(sources map[string][]byte) bool {
	want := map[string]bool{}
	for _, n := range strings.Fields(issuerMappingGoInventory) {
		want[n] = true
	}
	if len(sources) != len(want) {
		return false
	}
	schemas := issuerMappingSchemas()
	for name, raw := range sources {
		if !want[name] {
			return false
		}
		f, e := parser.ParseFile(token.NewFileSet(), name, raw, parser.ImportsOnly)
		if e != nil || f.Name.Name != "nodeissuermapping" {
			return false
		}
		for _, im := range f.Imports {
			p, e := strconv.Unquote(im.Path.Value)
			if e != nil {
				return false
			}
			if strings.HasPrefix(p, "talenro.local/") && p != "talenro.local/platform/internal/nodecontrol/contracts" && p != "talenro.local/platform/internal/nodecontrol/nodeissuerbinding" {
				return false
			}
		}
		if !strings.HasSuffix(name, "_test.go") {
			if !issuerMappingProductAllowed(raw) || schemas[name] == "" || issuerMappingASTHash(raw) != schemas[name] {
				return false
			}
		}
	}
	return len(schemas) == 12
}

type issuerMappingToken struct {
	key        string
	start, end int
}

func issuerMappingTokens(raw []byte) []issuerMappingToken {
	var s scanner.Scanner
	file := token.NewFileSet().AddFile("mutation", -1, len(raw))
	s.Init(file, raw, nil, 0)
	var out []issuerMappingToken
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.SEMICOLON && lit == "\n" {
			continue
		}
		text := lit
		if text == "" {
			text = tok.String()
		}
		start := file.Offset(pos)
		out = append(out, issuerMappingToken{key: tok.String() + ":" + lit, start: start, end: start + len(text)})
	}
	return out
}
