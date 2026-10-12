package statepolicybinding

import (
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

func policyBindingConsumerAllowed(path, pkg, imported string) bool {
	if pkg == "statetimebinding" && filepath.Dir(filepath.Clean(path)) == filepath.FromSlash("internal/nodecontrol/statetimebinding") {
		return true
	}
	return imported != "talenro.local/platform/internal/nodecontrol/statepolicybinding"
}

func TestPolicyBindingTimeBindingBoundary(t *testing.T) {
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
		if got := policyBindingConsumerAllowed(filepath.FromSlash(tc.path), tc.pkg, "talenro.local/platform/internal/nodecontrol/statepolicybinding"); got != tc.want {
			t.Errorf("%s %s: %v", tc.path, tc.pkg, got)
		}
	}

}

func cleanupExpr(e ast.Expr) string {
	switch x := e.(type) {
	case *ast.Ident:
		return x.Name
	case *ast.SelectorExpr:
		return cleanupExpr(x.X) + "." + x.Sel.Name
	case *ast.UnaryExpr:
		if x.Op == token.AND {
			return "&" + cleanupExpr(x.X)
		}
	case *ast.IndexExpr:
		return cleanupExpr(x.X)
	case *ast.IndexListExpr:
		return cleanupExpr(x.X)
	}
	return ""
}
func cleanupCall(s ast.Stmt, want string, args ...string) bool {
	e, ok := s.(*ast.ExprStmt)
	if !ok {
		return false
	}
	c, ok := e.X.(*ast.CallExpr)
	if !ok || cleanupExpr(c.Fun) != want || len(c.Args) != len(args) {
		return false
	}
	for i, a := range args {
		if cleanupExpr(c.Args[i]) != a {
			return false
		}
	}
	return true
}
func cleanupDeferred(s ast.Stmt, owner string) bool {
	d, ok := s.(*ast.DeferStmt)
	return ok && len(d.Call.Args) == 0 && cleanupExpr(d.Call.Fun) == owner+".Clear"
}
func cleanupGuard(s ast.Stmt, condition, call string, args ...string) bool {
	d, ok := s.(*ast.DeferStmt)
	if !ok || len(d.Call.Args) != 0 {
		return false
	}
	f, ok := d.Call.Fun.(*ast.FuncLit)
	if !ok || len(f.Body.List) != 1 {
		return false
	}
	b, ok := f.Body.List[0].(*ast.IfStmt)
	if !ok || b.Init != nil || b.Else != nil {
		return false
	}
	if condition == "!done" {
		x, ok := b.Cond.(*ast.UnaryExpr)
		if !ok || x.Op != token.NOT || cleanupExpr(x.X) != "done" {
			return false
		}
	} else {
		x, ok := b.Cond.(*ast.BinaryExpr)
		if !ok || x.Op != token.EQL || cleanupExpr(x.X) != "result" || cleanupExpr(x.Y) != "nil" {
			return false
		}
	}
	if len(b.Body.List) < 1 || len(b.Body.List) > 2 || !cleanupCall(b.Body.List[0], call, args...) {
		return false
	}
	if len(b.Body.List) == 2 {
		a, ok := b.Body.List[1].(*ast.AssignStmt)
		if !ok || len(a.Lhs) != 1 || cleanupExpr(a.Lhs[0]) != "out" || len(a.Rhs) != 1 {
			return false
		}
		v, ok := a.Rhs[0].(*ast.CompositeLit)
		if !ok || len(v.Elts) != 0 {
			return false
		}
	}
	return true
}
func cleanupAcquire(s ast.Stmt, callee, owner string) bool {
	a, ok := s.(*ast.AssignStmt)
	if !ok || len(a.Lhs) == 0 || cleanupExpr(a.Lhs[0]) != owner || len(a.Rhs) != 1 {
		return false
	}
	c, ok := a.Rhs[0].(*ast.CallExpr)
	return ok && cleanupExpr(c.Fun) == callee
}
func cleanupImmediate(fn *ast.FuncDecl, callee, owner string) bool {
	matches := 0
	valid := true
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok && cleanupExpr(c.Fun) == callee {
			matches++
		}
		return true
	})
	found := 0
	for i, s := range fn.Body.List {
		if cleanupAcquire(s, callee, owner) {
			found++
			if i+1 >= len(fn.Body.List) || !cleanupDeferred(fn.Body.List[i+1], owner) {
				valid = false
			}
		}
	}
	return valid && matches == 1 && found == 1
}
func cleanupNamedResult(fn *ast.FuncDecl) bool {
	if fn.Type.Results == nil {
		return false
	}
	found := false
	for _, f := range fn.Type.Results.List {
		for _, n := range f.Names {
			if n.Name == "result" {
				found = true
			}
		}
	}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if a, ok := n.(*ast.AssignStmt); ok {
			for _, lhs := range a.Lhs {
				if cleanupExpr(lhs) == "result" {
					found = false
				}
			}
		}
		return true
	})
	return found
}
func cleanupReturn(s ast.Stmt, owner string) bool {
	r, ok := s.(*ast.ReturnStmt)
	return ok && len(r.Results) > 0 && cleanupExpr(r.Results[0]) == owner
}
func cleanupProtectedOut(fn *ast.FuncDecl, call string, args ...string) bool {
	guard := -1
	for i, s := range fn.Body.List {
		if cleanupGuard(s, "!done", call, args...) {
			if guard >= 0 {
				return false
			}
			guard = i
		}
	}
	if guard < 0 || len(fn.Body.List) < 3 || !cleanupReturn(fn.Body.List[len(fn.Body.List)-1], "out") {
		return false
	}
	tail, ok := fn.Body.List[len(fn.Body.List)-2].(*ast.AssignStmt)
	if !ok || len(tail.Lhs) != 1 || cleanupExpr(tail.Lhs[0]) != "done" || len(tail.Rhs) != 1 || cleanupExpr(tail.Rhs[0]) != "true" {
		return false
	}
	trueCount, falseCount := 0, 0
	valid := true
	for i, s := range fn.Body.List {
		ast.Inspect(s, func(n ast.Node) bool {
			if a, ok := n.(*ast.AssignStmt); ok {
				for j, lhs := range a.Lhs {
					if cleanupExpr(lhs) == "done" {
						if len(a.Rhs) != len(a.Lhs) {
							valid = false
							continue
						}
						switch cleanupExpr(a.Rhs[j]) {
						case "true":
							trueCount++
							if i != len(fn.Body.List)-2 {
								valid = false
							}
						case "false":
							falseCount++
							if i >= guard {
								valid = false
							}
						default:
							valid = false
						}
					}
				}
			}
			if c, ok := n.(*ast.CallExpr); ok {
				switch cleanupExpr(c.Fun) {
				case "make", "copySlice", "copyRaw", "copyContent", "copyTable", "ownBinding", "ownPolicy", "stateformat.ParseRootSet", "parsed.RootMaterial":
					if i < guard {
						valid = false
					}
				}
			}
			return true
		})
	}
	return valid && trueCount == 1 && falseCount == 1
}
func cleanupEvaluate(fn *ast.FuncDecl) bool {
	b := fn.Body.List
	if len(b) < 8 || !cleanupNamedResult(fn) || !cleanupAcquire(b[0], "ownInput", "owned") {
		return false
	}
	gate, ok := b[1].(*ast.IfStmt)
	if !ok || gate.Init != nil || gate.Else != nil || len(gate.Body.List) != 1 {
		return false
	}
	cond, ok := gate.Cond.(*ast.BinaryExpr)
	if !ok || cond.Op != token.NEQ || cleanupExpr(cond.X) != "e" || cleanupExpr(cond.Y) != "nil" {
		return false
	}
	r, ok := gate.Body.List[0].(*ast.ReturnStmt)
	if !ok || len(r.Results) != 2 || cleanupExpr(r.Results[0]) != "nil" {
		return false
	}
	mapped, ok := r.Results[1].(*ast.CallExpr)
	if !ok || cleanupExpr(mapped.Fun) != "mapError" {
		return false
	}
	a, ok := b[2].(*ast.AssignStmt)
	if !ok || len(a.Lhs) != 1 || cleanupExpr(a.Lhs[0]) != "model" || len(a.Rhs) != 1 {
		return false
	}
	addr, ok := a.Rhs[0].(*ast.UnaryExpr)
	if !ok || addr.Op != token.AND {
		return false
	}
	lit, ok := addr.X.(*ast.CompositeLit)
	if !ok || cleanupExpr(lit.Type) != "Model" || len(lit.Elts) != 1 {
		return false
	}
	kv, ok := lit.Elts[0].(*ast.KeyValueExpr)
	if !ok || cleanupExpr(kv.Key) != "input" || cleanupExpr(kv.Value) != "owned" {
		return false
	}
	if !cleanupGuard(b[3], "result==nil", "model.Clear") || !cleanupImmediate(fn, "statebinding.Evaluate", "binding") || !cleanupImmediate(fn, "buildRootPath", "path") || !cleanupImmediate(fn, "statepolicy.EvaluateModel", "candidate") || !cleanupImmediate(fn, "candidate.State", "state") {
		return false
	}
	capture := false
	summaryCalls := 0
	valid := true
	for _, s := range b {
		ast.Inspect(s, func(n ast.Node) bool {
			if c, ok := n.(*ast.CallExpr); ok && cleanupExpr(c.Fun) == "binding.Summary" {
				summaryCalls++
			}
			return true
		})
		if a, ok := s.(*ast.AssignStmt); ok && len(a.Lhs) == 1 && cleanupExpr(a.Lhs[0]) == "model.summary.Binding" && len(a.Rhs) == 1 {
			if c, ok := a.Rhs[0].(*ast.CallExpr); ok && cleanupExpr(c.Fun) == "binding.Summary" {
				capture = true
			}
		}
	}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if a, ok := n.(*ast.AssignStmt); ok {
			for _, lhs := range a.Lhs {
				switch cleanupExpr(lhs) {
				case "owned", "model", "binding", "path", "candidate", "state":
					if a.Tok != token.DEFINE {
						valid = false
					}
				case "model.valid":
					if a != b[len(b)-2] {
						valid = false
					}
				}
			}
		}
		if r, ok := n.(*ast.ReturnStmt); ok && cleanupReturn(r, "model") && r != b[len(b)-1] {
			valid = false
		}
		return true
	})
	return valid && capture && summaryCalls == 1 && cleanupReturn(b[len(b)-1], "model")
}
func policyBindingCleanupRegistered(source []byte) bool {
	f, e := parser.ParseFile(token.NewFileSet(), "cleanup.go", source, 0)
	if e != nil {
		return false
	}
	allowedOwners := map[string]bool{"Evaluate": true, "Recheck": true, "ownInput": true, "ownBinding": true, "ownPolicy": true, "copyContent": true, "copyTable": true, "cloneSummary": true, "Summary": true, "buildRootPath": true, "loadRootFact": true, "bindKeyProjection": true}
	acquiring := map[string]bool{"ownInput": true, "ownBinding": true, "ownPolicy": true, "copyContent": true, "copyTable": true, "cloneSummary": true, "buildRootPath": true, "loadRootFact": true, "statebinding.Evaluate": true, "statepolicy.EvaluateModel": true, "stateformat.ParseRootSet": true, "stateformat.ParseEnvelope": true, "parsed.RootMaterial": true, "parsed.MetadataMaterial": true, "candidate.State": true}
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Body == nil {
			continue
		}
		valid := true
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if c, ok := n.(*ast.CallExpr); ok && acquiring[cleanupExpr(c.Fun)] && !allowedOwners[fn.Name.Name] {
				valid = false
			}
			return true
		})
		if !valid {
			return false
		}
		switch fn.Name.Name {
		case "Evaluate":
			if !cleanupEvaluate(fn) {
				return false
			}
		case "Recheck":
			if !cleanupImmediate(fn, "Evaluate", "fresh") {
				return false
			}
		case "ownInput":
			if !cleanupProtectedOut(fn, "clearInput", "&out") {
				return false
			}
		case "ownBinding":
			if !cleanupProtectedOut(fn, "clearBinding", "&out") {
				return false
			}
		case "ownPolicy":
			if !cleanupProtectedOut(fn, "clearPolicy", "&out") {
				return false
			}
		case "copyContent":
			if !cleanupProtectedOut(fn, "clearContent", "&out") {
				return false
			}
		case "copyTable":
			if !cleanupProtectedOut(fn, "clearTable", "&out") {
				return false
			}
		case "cloneSummary":
			if !cleanupProtectedOut(fn, "clearSummary", "&out") {
				return false
			}
		case "loadRootFact":
			if !cleanupProtectedOut(fn, "out.material.Clear") || !cleanupImmediate(fn, "stateformat.ParseRootSet", "parsed") {
				return false
			}
			found := false
			for _, s := range fn.Body.List {
				if cleanupAcquire(s, "parsed.RootMaterial", "out.material") {
					found = true
				}
			}
			if !found {
				return false
			}
		case "bindKeyProjection":
			if !cleanupImmediate(fn, "stateformat.ParseEnvelope", "parsed") || !cleanupImmediate(fn, "parsed.MetadataMaterial", "material") {
				return false
			}
		case "buildRootPath":
			if !cleanupNamedResult(fn) {
				return false
			}
			found := false
			for i, s := range fn.Body.List {
				if a, ok := s.(*ast.AssignStmt); ok && len(a.Lhs) == 1 && cleanupExpr(a.Lhs[0]) == "path" {
					if i+1 < len(fn.Body.List) && cleanupGuard(fn.Body.List[i+1], "result==nil", "path.Clear") {
						found = true
					}
				}
			}
			if !found || !cleanupReturn(fn.Body.List[len(fn.Body.List)-1], "path") {
				return false
			}
		case "makeSummary":
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				if c, ok := n.(*ast.CallExpr); ok && cleanupExpr(c.Fun) != "len" {
					valid = false
				}
				return true
			})
			if !valid {
				return false
			}
		}
	}
	return true
}

func boundaryCanonical(t *testing.T, name string) string {
	t.Helper()
	raw, e := os.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	raw, e = format.Source(raw)
	if e != nil {
		t.Fatal(e)
	}
	return string(raw)
}
func TestPolicyBindingCleanupRegistration(t *testing.T) {
	for _, name := range []string{"owned.go", "roots.go", "projection.go", "evaluate.go", "model.go"} {
		if !policyBindingCleanupRegistered([]byte(boundaryCanonical(t, name))) {
			t.Fatal("missing supported early cleanup", name)
		}
	}
	eval := boundaryCanonical(t, "evaluate.go")
	cases := [][3]string{
		{"missing binding", "defer binding.Clear()", ""},
		{"late binding", "defer binding.Clear()", "if ctx.Err() != nil { return nil, ctx.Err() }; defer binding.Clear()"},
		{"conditional binding", "defer binding.Clear()", "if true { defer binding.Clear() }"},
		{"wrong binding", "defer binding.Clear()", "defer model.Clear()"},
		{"missing path", "defer path.Clear()", ""},
		{"late path", "defer path.Clear()", "if ctx.Err() != nil { return nil, ctx.Err() }; defer path.Clear()"},
		{"wrong candidate", "defer candidate.Clear()", "defer binding.Clear()"},
		{"late candidate", "defer candidate.Clear()", "if ctx.Err() != nil { return nil, ctx.Err() }; defer candidate.Clear()"},
		{"missing state", "defer state.Clear()", ""},
		{"late state", "defer state.Clear()", "if ctx.Err() != nil { return nil, ctx.Err() }; defer state.Clear()"},
		{"unprotected summary", "model.summary.Binding = binding.Summary()", "captured := binding.Summary(); model.summary.Binding = captured"},
		{"early transfer", "model.summary.Binding = binding.Summary()", "result = model; model.summary.Binding = binding.Summary()"},
		{"missing input guard", "if result == nil {", "if result != nil {"},
		{"wrong input owner", "model.Clear()", "binding.Clear()"},
		{"callback before model", "model := &Model{input: owned}", "if ctx.Err() != nil { return nil, ctx.Err() }; model := &Model{input: owned}"},
	}
	for _, tc := range cases {
		t.Run(tc[0], func(t *testing.T) {
			if !strings.Contains(eval, tc[1]) {
				t.Fatal("mutation does not apply")
			}
			bad := strings.Replace(eval, tc[1], tc[2], 1)
			if policyBindingCleanupRegistered([]byte(bad)) {
				t.Fatal("unsafe structure accepted")
			}
		})
	}
	for _, tc := range []struct{ name, from, to string }{
		{"projection.go", "defer parsed.Clear()", "if ctx.Err() != nil { return ctx.Err() }; defer parsed.Clear()"},
		{"projection.go", "defer material.Clear()", "defer parsed.Clear()"},
		{"roots.go", "defer parsed.Clear()", "if ctx.Err() != nil { return out, ctx.Err() }; defer parsed.Clear()"},
		{"roots.go", "out.material, e = parsed.RootMaterial()", "material, e := parsed.RootMaterial(); out.material = material"},
		{"roots.go", "if result == nil {", "if result != nil {"},
		{"owned.go", "clearPolicy(&out)", "clearPolicy(&in)"},
		{"owned.go", "if out.Binding, err = ownBinding", "done = true; if out.Binding, err = ownBinding"},
		{"model.go", "defer fresh.Clear()", "if ctx.Err() != nil { return ctx.Err() }; defer fresh.Clear()"},
		{"model.go", "clearSummary(&out)", "clearSummary(&in)"},
	} {
		source := boundaryCanonical(t, tc.name)
		if !strings.Contains(source, tc.from) {
			t.Fatal("mutation does not apply", tc.name, tc.from)
		}
		bad := strings.Replace(source, tc.from, tc.to, 1)
		if policyBindingCleanupRegistered([]byte(bad)) {
			t.Fatal("unsafe owner shape", tc.name, tc.from)
		}
	}
	if policyBindingCleanupRegistered([]byte("broken")) {
		t.Fatal("syntax")
	}
	if policyBindingCleanupRegistered([]byte("package p; func extra(){c,e:=statepolicy.EvaluateModel(ctx,in);_ = c;_ = e}")) {
		t.Fatal("unrecognized acquiring function")
	}
}
func TestPolicyBindingProductImports(t *testing.T) {
	names := map[string]bool{}
	for _, n := range []string{"doc.go", "input.go", "errors.go", "budget.go", "owned.go", "selection.go", "roots.go", "projection.go", "time.go", "evaluate.go", "model.go"} {
		names[n] = true
	}
	entries, e := os.ReadDir(".")
	if e != nil {
		t.Fatal(e)
	}
	count := 0
	for _, entry := range entries {
		n := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(n, ".go") || strings.HasSuffix(n, "_test.go") {
			continue
		}
		if !names[n] {
			t.Fatal("unexpected product file", n)
		}
		count++
		raw, e := os.ReadFile(n)
		if e != nil {
			t.Fatal(e)
		}
		if !policyBindingImportsAllowed(raw) || !policyBindingCleanupRegistered(raw) {
			t.Fatal("product boundary", n)
		}
	}
	if count != 11 {
		t.Fatal("product scope", count)
	}
}
func TestPolicyBindingConsumers(t *testing.T) {
	repo := filepath.Join("..", "..", "..")
	e := filepath.WalkDir(repo, func(path string, entry os.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if entry.IsDir() {
			if entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		raw, e := os.ReadFile(path)
		if e != nil {
			return e
		}
		f, e := parser.ParseFile(token.NewFileSet(), path, raw, parser.ImportsOnly)
		if e != nil {
			return e
		}
		rel, e := filepath.Rel(repo, path)
		if e != nil {
			return e
		}
		for _, i := range f.Imports {
			imported, e := strconv.Unquote(i.Path.Value)
			if e != nil {
				return e
			}
			if !policyBindingConsumerAllowed(rel, f.Name.Name, imported) {
				t.Errorf("production consumer: %s", rel)
			}
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
}
func policyBindingImportsAllowed(source []byte) bool {
	allowed := map[string][]string{
		"context": {"Context", "Canceled", "DeadlineExceeded"},
		"errors":  {"New", "Is", "As"},
		"bytes":   {"Clone", "Compare", "Equal"},
		"sort":    {"Slice"},
		"reflect": {"DeepEqual"},
		"time":    {"Time", "Unix"},
		"talenro.local/platform/internal/nodecontrol/contracts":        {"Digest"},
		"talenro.local/platform/internal/nodecontrol/statebinding":     {"Input", "Summary", "Evaluate", "ConditionalProjectionBound", "ErrInvalidInput", "ErrLimitExceeded", "ErrIncompleteCoverage", "ErrInvalidMaterial", "ErrMissingMaterial", "ErrDuplicateMaterial", "ErrBindingMismatch", "ErrSignatureMismatch", "ErrRollback", "ErrConflict", "ErrClearedModel"},
		"talenro.local/platform/internal/nodecontrol/statepolicy":      {"ModelInput", "StateModel", "RootRecord", "KeyRecord", "HighWater", "DeploymentObservation", "TimeObservation", "KeyHistoryObservation", "Candidate", "EvaluateModel", "Available", "Unknown", "PolicyError", "Code", "Reason", "Canceled", "Cancellation"},
		"talenro.local/platform/internal/nodecontrol/statecontract":    {"MainStateContentModel", "OptionalTime", "ArtifactRef", "ArtifactKind", "Coverage", "Present", "NeverEstablished", "RootPayload", "RootRotation", "Metadata", "DesiredSnapshot"},
		"talenro.local/platform/internal/nodecontrol/statematerials":   {"Input", "Scope", "Material", "Expectation", "MaterialRef", "MaterialKind", "RootSet", "RootRotation", "Metadata", "Desired"},
		"talenro.local/platform/internal/nodecontrol/statecomposition": {"Summary", "RotationLink", "Coverage"},
		"talenro.local/platform/internal/nodecontrol/stateformat":      {"ParseRootSet", "ParsedRootSet", "RootMaterial", "PublicKeyMaterial", "ParseEnvelope", "ParsedArtifact", "Metadata", "MetadataMaterial", "MetadataKeyMaterial"},
	}
	f, e := parser.ParseFile(token.NewFileSet(), "guard.go", source, 0)
	if e != nil {
		return false
	}
	aliases := map[string]map[string]bool{}
	seen := map[string]bool{}
	for _, i := range f.Imports {
		path, e := strconv.Unquote(i.Path.Value)
		members, ok := allowed[path]
		if e != nil || !ok || seen[path] {
			return false
		}
		seen[path] = true
		alias := filepath.Base(path)
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
				if ms, exists := aliases[id.Name]; exists && !ms[s.Sel.Name] {
					valid = false
				}
			}
		}
		return true
	})
	return valid
}
func TestPolicyBindingBoundary(t *testing.T) {
	for _, src := range []string{"package p;import \"context\";var _ context.Context", "package p;import \"time\";var _ time.Time;var _=time.Unix", "package p;import \"talenro.local/platform/internal/nodecontrol/statebinding\";var _=statebinding.Evaluate", "package p;import \"talenro.local/platform/internal/nodecontrol/statepolicy\";var _=statepolicy.EvaluateModel"} {
		if !policyBindingImportsAllowed([]byte(src)) {
			t.Fatal("allowed pure capability")
		}
	}
	for _, src := range []string{"package p;import \"os\"", "package p;import \"crypto/sha256\"", "package p;import \"encoding/json\"", "package p;import \"time\";var _=time.Now", "package p;import \"reflect\";var _=reflect.ValueOf", "package p;import \"talenro.local/platform/internal/nodecontrol/stateverify\"", "package p;import \"talenro.local/platform/internal/nodecontrol/statetime\"", "package p;import \"talenro.local/platform/internal/nodecontrol/statecontract\";var _=statecontract.EvaluateModel", "package p;import \"talenro.local/platform/internal/nodecontrol/statepolicy\";var _=statepolicy.RecheckModel", "package p;import \"context\";var _=context.WithCancel", "package p;import . \"bytes\"", "package p;import _ \"bytes\"", "package p;import(a \"bytes\";b \"bytes\")", "package p;import(a \"bytes\";a \"context\")", "package p;func f(){go f()}", "broken"} {
		if policyBindingImportsAllowed([]byte(src)) {
			t.Fatal("forbidden capability", src)
		}
	}
	for _, path := range []string{"internal/nodecontrol/serving/x.go", "internal/nodecontrol/statepolicybinding/x.go", "internal/nodecontrol/statepolicybinding/sub/x.go", "internal/nodecontrol/statepolicybinding-other/x.go"} {
		if policyBindingConsumerAllowed(filepath.FromSlash(path), "statepolicybinding", "talenro.local/platform/internal/nodecontrol/statepolicybinding") {
			t.Fatal("consumer", path)
		}
		if !policyBindingConsumerAllowed(path, "p", "context") {
			t.Fatal("unrelated import")
		}
	}
}
