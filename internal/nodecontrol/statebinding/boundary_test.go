package statebinding

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

func TestBindingPolicyBindingBoundary(t *testing.T) {
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
		if got := bindingConsumerAllowed(filepath.FromSlash(tc.path), tc.pkg, "talenro.local/platform/internal/nodecontrol/statebinding"); got != tc.want {
			t.Errorf("%s: %v", tc.path, got)
		}
	}

}

func TestBindingTimeBindingBoundary(t *testing.T) {
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
		if got := bindingConsumerAllowed(filepath.FromSlash(tc.path), tc.pkg, "talenro.local/platform/internal/nodecontrol/statebinding"); got != tc.want {
			t.Errorf("%s %s: %v", tc.path, tc.pkg, got)
		}
	}

}

func TestBindingBoundary(t *testing.T) {
	for _, src := range []string{
		`package p;import "context";var _ context.Context`,
		`package p;import "bytes";var _=bytes.Clone`,
		`package p;import "reflect";var _=reflect.DeepEqual`,
		`package p;import "talenro.local/platform/internal/nodecontrol/statecontract";var _=statecontract.EvaluateModel`,
	} {
		if !bindingImportsAllowed([]byte(src)) {
			t.Fatal("pure dependency rejected", src)
		}
	}
	for _, src := range []string{
		`package p;import "os"`, `package p;import "time"`, `package p;import "crypto/sha256"`, `package p;import "encoding/json"`,
		`package p;import "net/http"`, `package p;import "unsafe"`,
		`package p;import "talenro.local/platform/internal/nodecontrol/stateverify"`,
		`package p;import "talenro.local/platform/internal/nodecontrol/stateformat"`,
		`package p;import "talenro.local/platform/internal/nodecontrol/statepolicy"`,
		`package p;import "context";var _=context.WithCancel`,
		`package p;import "reflect";var _=reflect.ValueOf`,
		`package p;import . "bytes"`, `package p;import _ "bytes"`,
		`package p;import(a "bytes";b "bytes");var _=a.Equal;var _=b.Clone`,
		`package p;import(a "bytes";a "context")`,
		`package p;func f(){go f()}`, `broken`,
	} {
		if bindingImportsAllowed([]byte(src)) {
			t.Fatal("forbidden capability", src)
		}
	}
	for _, path := range []string{"internal/nodecontrol/serving/x.go", "internal/nodecontrol/reader/x.go", "internal/nodecontrol/statebinding/x.go", "internal/nodecontrol/statebinding/sub/x.go", "internal/nodecontrol/statebinding-other/x.go"} {
		if bindingConsumerAllowed(filepath.FromSlash(path), "statebinding", "talenro.local/platform/internal/nodecontrol/statebinding") {
			t.Fatal("production consumer", path)
		}
		if !bindingConsumerAllowed(filepath.FromSlash(path), "statebinding", "context") {
			t.Fatal("unrelated dependency")
		}
	}
}

func TestNodeCertTimeConsumerException(t *testing.T) {
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	dir := filepath.FromSlash("internal/nodecontrol/nodecerttime")
	target := "talenro.local/platform/internal/nodecontrol/statebinding"
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
		if got := bindingConsumerAllowed(c.path, c.pkg, target); got != c.want {
			t.Errorf("%s %s: got %v want %v", c.path, c.pkg, got, c.want)
		}
	}
	if bindingConsumerAllowed(filepath.Join(dir, "a.go"), "nodecerttime", target+"/child") {
		t.Error("subpath")
	}

}

func bindingConsumerAllowed(path, pkg, imported string) bool {
	// Only the new exact consumer is added; all previous rules below remain.
	certTimeTarget := "talenro.local/platform/internal/nodecontrol/statebinding"
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
	return imported != "talenro.local/platform/internal/nodecontrol/statebinding" ||
		(pkg == "statepolicybinding" && filepath.Dir(filepath.Clean(path)) == filepath.FromSlash("internal/nodecontrol/statepolicybinding"))
}

func TestBindingProductImports(t *testing.T) {
	files, e := filepath.Glob("*.go")
	if e != nil {
		t.Fatal(e)
	}
	count := 0
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		count++
		raw, e := os.ReadFile(file)
		if e != nil {
			t.Fatal(e)
		}
		if !bindingImportsAllowed(raw) {
			t.Error("production capability", file)
		}
	}
	if count != 9 {
		t.Fatalf("product scan count %d", count)
	}
}

func TestBindingConsumers(t *testing.T) {
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
			imported, e := strconv.Unquote(i.Path.Value)
			if e != nil {
				return e
			}
			if !bindingConsumerAllowed(rel, f.Name.Name, imported) {
				t.Error("production consumer", rel)
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

func TestBindingCleanupRegistration(t *testing.T) {
	good := `package statebinding
func Evaluate(ctx context.Context,in Input)(result *BindingModel,err error){
    owned,e:=ownInput(ctx,in)
    if e!=nil{return nil,e}
    model:=&BindingModel{input:owned}
    defer func(){if result==nil{model.Clear()}}()
    contract,e:=statecontract.EvaluateModel(ctx,owned.Main)
    defer contract.Clear()
    if e=mapError(ctx,e);e!=nil{return nil,e}
    previous,e:=statematerials.Evaluate(ctx,owned.Previous)
    defer previous.Clear()
    if e=mapError(ctx,e);e!=nil{return nil,e}
    proposed,e:=statecomposition.Evaluate(ctx,owned.Proposed)
    defer proposed.Clear()
    if e=mapError(ctx,e);e!=nil{return nil,e}
    model.contractSummary=contract.Summary()
    model.previousSummary=previous.Summary()
    model.proposedSummary=proposed.Summary()
    if e:=ctx.Err();e!=nil{return nil,e}
    model.valid=true
    return model,nil
}`
	if !bindingCleanupRegistered([]byte(good)) {
		t.Fatal("early protected cleanup rejected")
	}
	cases := map[string]string{
		"model_late":             strings.Replace(good, "defer func(){if result==nil{model.Clear()}}()", "ctx.Err();defer func(){if result==nil{model.Clear()}}()", 1),
		"model_after_lower":      strings.Replace(strings.Replace(good, "defer func(){if result==nil{model.Clear()}}()", "", 1), "defer contract.Clear()", "defer contract.Clear();defer func(){if result==nil{model.Clear()}}()", 1),
		"model_conditional":      strings.Replace(good, "defer func(){if result==nil{model.Clear()}}()", "if enabled{defer func(){if result==nil{model.Clear()}}()}", 1),
		"model_missing":          strings.Replace(good, "defer func(){if result==nil{model.Clear()}}()", "", 1),
		"model_wrong_owner":      strings.Replace(good, "model.Clear()", "other.Clear()", 1),
		"model_conditional_body": strings.Replace(good, "model.Clear()", "if enabled{model.Clear()}", 1),
		"model_wrong_result":     strings.Replace(good, "result==nil", "err!=nil", 1),
		"model_reassigned":       strings.Replace(good, "model.valid=true", "model=&BindingModel{};model.valid=true", 1),
		"owned_shadowed":         strings.Replace(good, "model.valid=true", "{owned:=Input{};_=owned};model.valid=true", 1),
		"result_shadowed":        strings.Replace(good, "model.valid=true", "{result:=model;_=result};model.valid=true", 1),
		"result_early_transfer":  strings.Replace(good, "model.valid=true", "result=model;ctx.Err();model.valid=true", 1),
		"summary_unprotected":    strings.Replace(good, "model.proposedSummary=proposed.Summary()", "temporary:=proposed.Summary();ctx.Err();model.proposedSummary=temporary", 1),
		"summary_wrong_owner":    strings.Replace(good, "model.previousSummary=previous.Summary()", "model.previousSummary=contract.Summary()", 1),
		"summary_wrong_model":    strings.Replace(good, "model.proposedSummary=proposed.Summary()", "other.proposedSummary=proposed.Summary()", 1),
		"summary_conditional":    strings.Replace(good, "model.proposedSummary=proposed.Summary()", "if enabled{model.proposedSummary=proposed.Summary()}", 1),
		"summary_overwritten":    strings.Replace(good, "model.valid=true", "model.proposedSummary=statecomposition.Summary{};model.valid=true", 1),
		"own_input_late_check":   strings.Replace(good, "model:=&BindingModel{input:owned}", "ctx.Err();model:=&BindingModel{input:owned}", 1),
		"broken":                 "broken",
	}
	for _, owner := range []string{"contract", "previous", "proposed"} {
		text := "defer " + owner + ".Clear()"
		cases[owner+"_late"] = strings.Replace(good, text, "ctx.Err();"+text, 1)
		cases[owner+"_map_late"] = strings.Replace(good, text, "mapError(ctx,e);"+text, 1)
		cases[owner+"_conditional"] = strings.Replace(good, text, "if enabled{"+text+"}", 1)
		cases[owner+"_missing"] = strings.Replace(good, text, "", 1)
		cases[owner+"_wrong_owner"] = strings.Replace(good, text, "defer other.Clear()", 1)
		cases[owner+"_reassigned"] = strings.Replace(good, "model.valid=true", owner+"=nil;model.valid=true", 1)
	}
	for name, source := range cases {
		t.Run(name, func(t *testing.T) {
			if bindingCleanupRegistered([]byte(source)) {
				t.Fatal("unsafe cleanup accepted")
			}
		})
	}
	raw, e := os.ReadFile("evaluate.go")
	if e != nil {
		t.Fatal(e)
	}
	if !bindingCleanupRegistered(raw) {
		t.Fatal("production owner cleanup")
	}
}

// This is a deliberately conservative check of Evaluate's approved ownership
// shape, not a general Go control-flow proof. Reject alternate shapes instead of
// accepting a cleanup call merely because it appears somewhere in the AST.
func bindingCleanupRegistered(source []byte) bool {
	f, e := parser.ParseFile(token.NewFileSet(), "cleanup.go", source, 0)
	if e != nil {
		return false
	}
	var fn *ast.FuncDecl
	for _, decl := range f.Decls {
		if candidate, ok := decl.(*ast.FuncDecl); ok && candidate.Name.Name == "Evaluate" {
			if fn != nil {
				return false
			}
			fn = candidate
		}
	}
	if fn == nil || fn.Recv != nil || fn.Body == nil || fn.Type.Results == nil || len(fn.Type.Results.List) != 2 {
		return false
	}
	ident := func(x ast.Expr, name string) bool { id, ok := x.(*ast.Ident); return ok && id.Name == name }
	selector := func(x ast.Expr, owner, member string) bool {
		s, ok := x.(*ast.SelectorExpr)
		return ok && ident(s.X, owner) && s.Sel.Name == member
	}
	results := fn.Type.Results.List
	if len(results[0].Names) != 1 || results[0].Names[0].Name != "result" {
		return false
	}
	resultType, ok := results[0].Type.(*ast.StarExpr)
	if !ok || !ident(resultType.X, "BindingModel") {
		return false
	}
	statements := fn.Body.List
	if len(statements) < 12 {
		return false
	}
	own, ok := statements[0].(*ast.AssignStmt)
	if !ok || own.Tok != token.DEFINE || len(own.Lhs) != 2 || len(own.Rhs) != 1 || !ident(own.Lhs[0], "owned") || !ident(own.Lhs[1], "e") {
		return false
	}
	ownCall, ok := own.Rhs[0].(*ast.CallExpr)
	if !ok || !ident(ownCall.Fun, "ownInput") || len(ownCall.Args) != 2 || !ident(ownCall.Args[0], "ctx") || !ident(ownCall.Args[1], "in") {
		return false
	}
	guard, ok := statements[1].(*ast.IfStmt)
	if !ok || guard.Init != nil || guard.Else != nil || len(guard.Body.List) != 1 {
		return false
	}
	condition, ok := guard.Cond.(*ast.BinaryExpr)
	if !ok || condition.Op != token.NEQ || !ident(condition.X, "e") || !ident(condition.Y, "nil") {
		return false
	}
	failure, ok := guard.Body.List[0].(*ast.ReturnStmt)
	if !ok || len(failure.Results) != 2 || !ident(failure.Results[0], "nil") || !ident(failure.Results[1], "e") {
		return false
	}
	model, ok := statements[2].(*ast.AssignStmt)
	if !ok || model.Tok != token.DEFINE || len(model.Lhs) != 1 || len(model.Rhs) != 1 || !ident(model.Lhs[0], "model") {
		return false
	}
	address, ok := model.Rhs[0].(*ast.UnaryExpr)
	if !ok || address.Op != token.AND {
		return false
	}
	composite, ok := address.X.(*ast.CompositeLit)
	if !ok || !ident(composite.Type, "BindingModel") || len(composite.Elts) != 1 {
		return false
	}
	input, ok := composite.Elts[0].(*ast.KeyValueExpr)
	if !ok || !ident(input.Key, "input") || !ident(input.Value, "owned") {
		return false
	}
	clearCall := func(call *ast.CallExpr, owner string) bool {
		return call != nil && len(call.Args) == 0 && selector(call.Fun, owner, "Clear")
	}
	cleanup, ok := statements[3].(*ast.DeferStmt)
	if !ok || len(cleanup.Call.Args) != 0 {
		return false
	}
	literal, ok := cleanup.Call.Fun.(*ast.FuncLit)
	if !ok || literal.Type.Params.NumFields() != 0 || literal.Type.Results != nil || len(literal.Body.List) != 1 {
		return false
	}
	guarded, ok := literal.Body.List[0].(*ast.IfStmt)
	if !ok || guarded.Init != nil || guarded.Else != nil || len(guarded.Body.List) != 1 {
		return false
	}
	condition, ok = guarded.Cond.(*ast.BinaryExpr)
	if !ok || condition.Op != token.EQL || !ident(condition.X, "result") || !ident(condition.Y, "nil") {
		return false
	}
	clearExpression, ok := guarded.Body.List[0].(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := clearExpression.X.(*ast.CallExpr)
	if !ok || !clearCall(call, "model") {
		return false
	}

	owners := []struct{ owner, pkg, evaluate, input, summary string }{
		{"contract", "statecontract", "EvaluateModel", "Main", "contractSummary"},
		{"previous", "statematerials", "Evaluate", "Previous", "previousSummary"},
		{"proposed", "statecomposition", "Evaluate", "Proposed", "proposedSummary"},
	}
	declarations := map[token.Pos]bool{own.Pos(): true, model.Pos(): true}
	captures := map[token.Pos]bool{}
	allowedCalls := map[token.Pos]bool{}
	lowerIndex := 0
	lastOwner := token.Pos(0)
	summaries := map[string]bool{}
	for index, stmt := range statements[4:] {
		assignment, ok := stmt.(*ast.AssignStmt)
		if !ok || len(assignment.Rhs) != 1 {
			continue
		}
		c, ok := assignment.Rhs[0].(*ast.CallExpr)
		if !ok {
			continue
		}
		for _, entry := range owners {
			if selector(c.Fun, entry.pkg, entry.evaluate) {
				if lowerIndex >= len(owners) || entry.owner != owners[lowerIndex].owner || assignment.Tok != token.DEFINE || len(assignment.Lhs) != 2 || !ident(assignment.Lhs[0], entry.owner) || !ident(assignment.Lhs[1], "e") || len(c.Args) != 2 || !ident(c.Args[0], "ctx") || !selector(c.Args[1], "owned", entry.input) {
					return false
				}
				next := index + 5
				if next >= len(statements) {
					return false
				}
				deferOwner, ok := statements[next].(*ast.DeferStmt)
				if !ok || !clearCall(deferOwner.Call, entry.owner) {
					return false
				}
				declarations[assignment.Pos()] = true
				allowedCalls[c.Pos()] = true
				lastOwner = deferOwner.End()
				lowerIndex++
			}
			if selector(c.Fun, entry.owner, "Summary") {
				if lowerIndex != 3 || assignment.Pos() <= lastOwner || assignment.Tok != token.ASSIGN || len(assignment.Lhs) != 1 || !selector(assignment.Lhs[0], "model", entry.summary) || len(c.Args) != 0 || summaries[entry.owner] {
					return false
				}
				summaries[entry.owner] = true
				captures[assignment.Pos()] = true
				allowedCalls[c.Pos()] = true
			}
		}
	}
	if lowerIndex != 3 || len(summaries) != 3 {
		return false
	}
	finalReturn, ok := statements[len(statements)-1].(*ast.ReturnStmt)
	if !ok || len(finalReturn.Results) != 2 || !ident(finalReturn.Results[0], "model") || !ident(finalReturn.Results[1], "nil") {
		return false
	}
	protected := map[string]bool{"owned": true, "model": true, "result": true, "contract": true, "previous": true, "proposed": true}
	valid := true
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		switch node := n.(type) {
		case *ast.AssignStmt:
			for _, lhs := range node.Lhs {
				if id, ok := lhs.(*ast.Ident); ok && protected[id.Name] && !declarations[node.Pos()] {
					valid = false
				}
				if s, ok := lhs.(*ast.SelectorExpr); ok && ident(s.X, "model") && s.Sel.Name != "valid" && !captures[node.Pos()] {
					valid = false
				}
			}
		case *ast.ValueSpec:
			for _, id := range node.Names {
				if protected[id.Name] {
					valid = false
				}
			}
		case *ast.CallExpr:
			for _, entry := range owners {
				if (selector(node.Fun, entry.pkg, entry.evaluate) || selector(node.Fun, entry.owner, "Summary")) && !allowedCalls[node.Pos()] {
					valid = false
				}
			}
		case *ast.ReturnStmt:
			if node != finalReturn && (len(node.Results) != 2 || !ident(node.Results[0], "nil")) {
				valid = false
			}
		}
		return true
	})
	return valid
}

func bindingImportsAllowed(source []byte) bool {
	allowed := map[string][]string{
		"context": {"Context", "Canceled", "DeadlineExceeded"}, "errors": {"New", "Is"}, "bytes": {"Clone", "Compare", "Equal"}, "sort": {"Slice"}, "reflect": {"DeepEqual"},
		"talenro.local/platform/internal/nodecontrol/contracts":        {"Digest"},
		"talenro.local/platform/internal/nodecontrol/statecontract":    {"ModelInput", "MainStateContentModel", "MainStateScope", "RotationRef", "ArtifactRef", "ArtifactKind", "Coverage", "Present", "NeverEstablished", "RootPayload", "RootRotation", "Metadata", "DesiredSnapshot", "ContractModel", "ContractSummary", "ContractShapeValid", "EvaluateModel", "ErrInvalidInput", "ErrIncompleteCoverage", "ErrBindingMismatch", "ErrRollback", "ErrConflict", "ErrStaleModel"},
		"talenro.local/platform/internal/nodecontrol/statematerials":   {"Input", "Scope", "Material", "Expectation", "MaterialRef", "MaterialKind", "RootSet", "RootRotation", "Metadata", "Desired", "MaterialSubsetModel", "Summary", "SubsetConsistent", "Evaluate", "ErrInvalidInput", "ErrLimitExceeded", "ErrInvalidMaterial", "ErrMissingMaterial", "ErrDuplicateMaterial", "ErrReferenceMismatch", "ErrClearedModel"},
		"talenro.local/platform/internal/nodecontrol/statecomposition": {"Input", "AnchorInput", "Target", "RotationLink", "Coverage", "CompositionModel", "Summary", "ConditionalTargetVerified", "Evaluate", "ErrInvalidInput", "ErrLimitExceeded", "ErrInvalidMaterial", "ErrMissingMaterial", "ErrDuplicateMaterial", "ErrBindingMismatch", "ErrSignatureMismatch", "ErrClearedModel"},
	}
	f, err := parser.ParseFile(token.NewFileSet(), "guard.go", source, 0)
	if err != nil {
		return false
	}
	aliases := map[string]map[string]bool{}
	seen := map[string]bool{}
	for _, i := range f.Imports {
		path, err := strconv.Unquote(i.Path.Value)
		members, ok := allowed[path]
		if err != nil || !ok || seen[path] {
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
		for _, member := range members {
			set[member] = true
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
