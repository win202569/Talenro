package statetimebinding

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

func timeBindingConsumerAllowed(path, pkg, imported string) bool {
	return imported != "talenro.local/platform/internal/nodecontrol/statetimebinding"
}
func timeBindingImportsAllowed(source []byte) bool {
	allowed := map[string][]string{
		"context": {"Context", "Canceled", "DeadlineExceeded"}, "errors": {"New", "Is", "As"}, "bytes": {"Clone", "Compare", "Equal"}, "sort": {"Slice"}, "reflect": {"DeepEqual"}, "time": {"Time"},
		"talenro.local/platform/internal/nodecontrol/contracts":          {"Digest"},
		"talenro.local/platform/internal/nodecontrol/statepolicybinding": {"Input", "Summary", "Evaluate", "ConditionalSelectedPolicyBound", "PolicyFailure", "ErrInvalidInput", "ErrLimitExceeded", "ErrIncompleteCoverage", "ErrInvalidMaterial", "ErrMissingMaterial", "ErrDuplicateMaterial", "ErrBindingMismatch", "ErrSignatureMismatch", "ErrRollback", "ErrConflict", "ErrClearedModel", "ErrUnsupportedPolicyShape", "ErrPolicyRejected"},
		"talenro.local/platform/internal/nodecontrol/statetime":          {"ModelInput", "EvidenceSummary", "EvaluateModel", "Available", "ErrInvalidInput", "ErrBindingMismatch", "ErrSignatureMismatch", "ErrUnauthorizedKey", "ErrAssuranceMissing", "ErrInvalidTime", "ErrExpired", "ErrCheckpointRollback", "ErrStaleEvidence"},
		"talenro.local/platform/internal/nodecontrol/statebinding":       {"Input"},
		"talenro.local/platform/internal/nodecontrol/statepolicy":        {"ModelInput", "StateModel", "RootRecord", "Available", "Code", "Reason"},
		"talenro.local/platform/internal/nodecontrol/statecontract":      {"MainStateContentModel", "Present", "Metadata", "RootPayload"},
		"talenro.local/platform/internal/nodecontrol/statematerials":     {"Input", "Material", "MaterialRef", "Expectation", "RootSet", "RootRotation", "Metadata"},
	}
	f, e := parser.ParseFile(token.NewFileSet(), "guard.go", source, 0)
	if e != nil {
		return false
	}
	aliases := map[string]map[string]bool{}
	seen := map[string]bool{}
	for _, im := range f.Imports {
		p, e := strconv.Unquote(im.Path.Value)
		members, ok := allowed[p]
		if e != nil || !ok || seen[p] {
			return false
		}
		seen[p] = true
		a := filepath.Base(p)
		if im.Name != nil {
			a = im.Name.Name
		}
		if a == "." || a == "_" || aliases[a] != nil {
			return false
		}
		set := map[string]bool{}
		for _, m := range members {
			set[m] = true
		}
		aliases[a] = set
	}
	valid := true
	ast.Inspect(f, func(n ast.Node) bool {
		if s, ok := n.(*ast.SelectorExpr); ok {
			switch s.Sel.Name {
			case "State", "Commit", "Authorize", "Ready", "RecheckModel", "Verify", "AfterFunc", "NewTimer":
				valid = false
			}
		}
		if _, ok := n.(*ast.GoStmt); ok {
			valid = false
		}
		if s, ok := n.(*ast.SelectorExpr); ok {
			if id, ok := s.X.(*ast.Ident); ok {
				if ms, yes := aliases[id.Name]; yes && !ms[s.Sel.Name] {
					valid = false
				}
			}
		}
		return true
	})
	return valid
}
func TestTimeBindingCapabilityBoundary(t *testing.T) {
	for _, src := range []string{`package p;import "time";var _ time.Time`, `package p;import "talenro.local/platform/internal/nodecontrol/statepolicybinding";var _=statepolicybinding.Evaluate`, `package p;import "talenro.local/platform/internal/nodecontrol/statetime";var _=statetime.EvaluateModel`} {
		if !timeBindingImportsAllowed([]byte(src)) {
			t.Fatal("allowed")
		}
	}
	for _, src := range []string{`package p;import "time";var _=time.Now`, `package p;import "talenro.local/platform/internal/nodecontrol/statetime";var _=statetime.RecheckModel`, `package p;import "talenro.local/platform/internal/nodecontrol/statepolicy";var _=statepolicy.EvaluateModel`, `package p;import "reflect";var _=reflect.ValueOf`, `package p;import "talenro.local/platform/internal/nodecontrol/stateformat"`, `package p;import "os"`, `package p;import "crypto/sha256"`, `package p;import . "bytes"`, `package p;import _ "bytes"`, `package p;import(a "bytes";b "bytes")`, `package p;func f(){go f()}`} {
		if timeBindingImportsAllowed([]byte(src)) {
			t.Fatal("forbidden", src)
		}
	}
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

func TestTimeBindingConsumers(t *testing.T) {
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
			if !timeBindingConsumerAllowed(rel, f.Name.Name, imported) {
				t.Errorf("production consumer: %s", rel)
			}
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
}

func TestTimeBindingProductImports(t *testing.T) {
	names := map[string]bool{}
	for _, n := range []string{"doc.go", "input.go", "errors.go", "budget.go", "owned.go", "binding.go", "observations.go", "checkpoint.go", "evaluate.go", "model.go"} {
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
		if !timeBindingImportsAllowed(raw) || !timeBindingCleanupRegistered(raw) {
			t.Fatal("product boundary", n)
		}
	}
	if count != 10 {
		t.Fatal("product scope", count)
	}
}

func TestTimeBindingCleanupRegistration(t *testing.T) {
	for _, name := range []string{"owned.go", "evaluate.go", "model.go"} {
		if !timeBindingCleanupRegistered([]byte(boundaryCanonical(t, name))) {
			t.Fatal("early cleanup", name)
		}
	}
	source := boundaryCanonical(t, "evaluate.go")
	for _, tc := range [][2]string{
		{"defer policy.Clear()", ""}, {"defer policy.Clear()", "if ctx.Err()!=nil{return nil,ctx.Err()};defer policy.Clear()"}, {"defer policy.Clear()", "if true{defer policy.Clear()}"}, {"defer policy.Clear()", "defer evidence.Clear()"},
		{"defer evidence.Clear()", ""}, {"defer evidence.Clear()", "if ctx.Err()!=nil{return nil,ctx.Err()};defer evidence.Clear()"},
		{"model.summary.Policy = policy.Summary()", "captured:=policy.Summary();model.summary.Policy=captured"},
		{"model.summary.Policy = policy.Summary()", "result=model;model.summary.Policy=policy.Summary()"},
		{"if result == nil {", "if result != nil {"},
		{"model := &Model{input: owned}", "if ctx.Err()!=nil{return nil,ctx.Err()};model := &Model{input: owned}"},
		{"model.valid = true", "model.valid = false"},
	} {
		if !strings.Contains(source, tc[0]) {
			t.Fatal("mutation not applicable", tc[0])
		}
		if timeBindingCleanupRegistered([]byte(strings.Replace(source, tc[0], tc[1], 1))) {
			t.Fatal("unsafe cleanup accepted", tc[0])
		}
	}
	for _, tc := range [][3]string{
		{"owned.go", "clearPolicyBinding(&out)", "clearPolicyBinding(&in)"},
		{"owned.go", "clearTime(&out)", "clearTime(&in)"},
		{"owned.go", "if out.Policy, err = ownPolicyBinding", "done=true;if out.Policy,err=ownPolicyBinding"},
		{"model.go", "defer fresh.Clear()", "if ctx.Err()!=nil{return ctx.Err()};defer fresh.Clear()"},
		{"model.go", "clearPolicySummary(&out)", "clearPolicySummary(&in)"},
	} {
		src := boundaryCanonical(t, tc[0])
		if !strings.Contains(src, tc[1]) {
			t.Fatal("mutation missing", tc[1])
		}
		if timeBindingCleanupRegistered([]byte(strings.Replace(src, tc[1], tc[2], 1))) {
			t.Fatal("unsafe owned copy", tc[1])
		}
	}
	if timeBindingCleanupRegistered([]byte("broken")) || timeBindingCleanupRegistered([]byte("package p;func extra(){m,e:=statetime.EvaluateModel(ctx,in);_=m;_=e}")) {
		t.Fatal("unrecognized owner")
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
				case "make", "copySlice", "copyRaw", "copyContent", "copyTable", "ownInput", "ownPolicyBinding", "ownBinding", "ownPolicy", "ownTime", "clonePolicySummary", "stateformat.ParseRootSet", "parsed.RootMaterial":
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
	if !cleanupGuard(b[3], "result==nil", "model.Clear") || !cleanupImmediate(fn, "statepolicybinding.Evaluate", "policy") || !cleanupImmediate(fn, "statetime.EvaluateModel", "evidence") {
		return false
	}
	capture := false
	summaryCalls := 0
	valid := true
	for _, s := range b {
		ast.Inspect(s, func(n ast.Node) bool {
			if c, ok := n.(*ast.CallExpr); ok && cleanupExpr(c.Fun) == "policy.Summary" {
				summaryCalls++
			}
			return true
		})
		if a, ok := s.(*ast.AssignStmt); ok && len(a.Lhs) == 1 && cleanupExpr(a.Lhs[0]) == "model.summary.Policy" && len(a.Rhs) == 1 {
			if c, ok := a.Rhs[0].(*ast.CallExpr); ok && cleanupExpr(c.Fun) == "policy.Summary" {
				capture = true
			}
		}
	}
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if a, ok := n.(*ast.AssignStmt); ok {
			for _, lhs := range a.Lhs {
				switch cleanupExpr(lhs) {
				case "owned", "model", "policy", "evidence":
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
	tail, ok := b[len(b)-2].(*ast.AssignStmt)
	if !ok || len(tail.Lhs) != 1 || cleanupExpr(tail.Lhs[0]) != "model.valid" || len(tail.Rhs) != 1 || cleanupExpr(tail.Rhs[0]) != "true" {
		return false
	}
	final, ok := b[len(b)-3].(*ast.IfStmt)
	if !ok || final.Init == nil {
		return false
	}
	finalAssign, ok := final.Init.(*ast.AssignStmt)
	if !ok || len(finalAssign.Rhs) != 1 {
		return false
	}
	c, ok := finalAssign.Rhs[0].(*ast.CallExpr)
	if !ok || cleanupExpr(c.Fun) != "ctx.Err" {
		return false
	}
	return valid && capture && summaryCalls == 1 && cleanupReturn(b[len(b)-1], "model")
}

func timeBindingCleanupRegistered(source []byte) bool {
	f, e := parser.ParseFile(token.NewFileSet(), "cleanup.go", source, 0)
	if e != nil {
		return false
	}
	allowedOwners := map[string]bool{"Evaluate": true, "Recheck": true, "ownInput": true, "ownPolicyBinding": true, "ownBinding": true, "ownPolicy": true, "ownTime": true, "copyContent": true, "copyTable": true, "cloneSummary": true, "clonePolicySummary": true, "Summary": true}
	acquiring := map[string]bool{"ownInput": true, "ownPolicyBinding": true, "ownBinding": true, "ownPolicy": true, "ownTime": true, "copyContent": true, "copyTable": true, "cloneSummary": true, "clonePolicySummary": true, "statepolicybinding.Evaluate": true, "statetime.EvaluateModel": true}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
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
		case "ownPolicyBinding":
			if !cleanupProtectedOut(fn, "clearPolicyBinding", "&out") {
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
		case "ownTime":
			if !cleanupProtectedOut(fn, "clearTime", "&out") {
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
		case "clonePolicySummary":
			if !cleanupProtectedOut(fn, "clearPolicySummary", "&out") {
				return false
			}
		}
	}
	return true
}
