package statecontract

import (
	"context"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"reflect"
	"strings"
	"testing"
)

// The owner cannot be observed after EvaluateModel panics without adding a
// production hook or unsafe access. This deliberately narrow AST guard checks
// the protected ownership tail, not arbitrary Go control flow or memory erasure.
func contractCleanupRegistered(source []byte) bool {
	f, err := parser.ParseFile(token.NewFileSet(), "evaluate.go", source, 0)
	if err != nil {
		return false
	}
	var fn *ast.FuncDecl
	for _, decl := range f.Decls {
		if candidate, ok := decl.(*ast.FuncDecl); ok && candidate.Name.Name == "EvaluateModel" {
			if fn != nil {
				return false
			}
			fn = candidate
		}
	}
	id := func(x ast.Expr, name string) bool {
		v, ok := x.(*ast.Ident)
		return ok && v.Name == name
	}
	call := func(x ast.Expr, owner, member string) bool {
		v, ok := x.(*ast.CallExpr)
		if !ok || len(v.Args) != 0 || v.Ellipsis.IsValid() {
			return false
		}
		s, ok := v.Fun.(*ast.SelectorExpr)
		return ok && id(s.X, owner) && s.Sel.Name == member
	}
	returns := func(s ast.Stmt, first, second string) bool {
		v, ok := s.(*ast.ReturnStmt)
		return ok && len(v.Results) == 2 && id(v.Results[0], first) && id(v.Results[1], second)
	}
	if fn == nil || fn.Recv != nil || fn.Body == nil || len(fn.Body.List) < 4 || fn.Type.Results == nil {
		return false
	}
	r := fn.Type.Results.List
	if len(r) != 2 || len(r[0].Names) != 1 || r[0].Names[0].Name != "result" || len(r[1].Names) != 1 || r[1].Names[0].Name != "err" || !id(r[1].Type, "error") {
		return false
	}
	resultType, ok := r[0].Type.(*ast.StarExpr)
	if !ok || !id(resultType.X, "ContractModel") {
		return false
	}
	tail := fn.Body.List[len(fn.Body.List)-4:]
	// Reject an earlier owner/result alias or transfer. All post-owner statements
	// must be exactly the allocation, immediate cleanup, final check and transfer.
	cleanPrefix := true
	for _, statement := range fn.Body.List[:len(fn.Body.List)-4] {
		ast.Inspect(statement, func(n ast.Node) bool {
			if v, ok := n.(*ast.Ident); ok && (v.Name == "m" || v.Name == "result") {
				cleanPrefix = false
			}
			return true
		})
	}
	allocation, ok := tail[0].(*ast.AssignStmt)
	if !cleanPrefix || !ok || allocation.Tok != token.DEFINE || len(allocation.Lhs) != 1 || !id(allocation.Lhs[0], "m") || len(allocation.Rhs) != 1 {
		return false
	}
	address, ok := allocation.Rhs[0].(*ast.UnaryExpr)
	if !ok || address.Op != token.AND {
		return false
	}
	model, ok := address.X.(*ast.CompositeLit)
	if !ok || !id(model.Type, "ContractModel") || len(model.Elts) != 2 {
		return false
	}
	fields := map[string]ast.Expr{}
	for _, element := range model.Elts {
		field, ok := element.(*ast.KeyValueExpr)
		if !ok {
			return false
		}
		key, ok := field.Key.(*ast.Ident)
		if !ok || fields[key.Name] != nil {
			return false
		}
		fields[key.Name] = field.Value
	}
	copyCall, ok := fields["input"].(*ast.CallExpr)
	if !id(fields["valid"], "true") || !ok || !id(copyCall.Fun, "cloneInput") || len(copyCall.Args) != 1 || !id(copyCall.Args[0], "in") || copyCall.Ellipsis.IsValid() {
		return false
	}
	cleanup, ok := tail[1].(*ast.DeferStmt)
	if !ok || len(cleanup.Call.Args) != 0 || cleanup.Call.Ellipsis.IsValid() {
		return false
	}
	closure, ok := cleanup.Call.Fun.(*ast.FuncLit)
	if !ok || closure.Type.Params.NumFields() != 0 || closure.Type.Results != nil || len(closure.Body.List) != 1 {
		return false
	}
	conditional, ok := closure.Body.List[0].(*ast.IfStmt)
	if !ok || conditional.Init != nil || conditional.Else != nil || len(conditional.Body.List) != 1 {
		return false
	}
	condition, ok := conditional.Cond.(*ast.BinaryExpr)
	if !ok || condition.Op != token.EQL || !id(condition.X, "result") || !id(condition.Y, "nil") {
		return false
	}
	clearCall, ok := conditional.Body.List[0].(*ast.ExprStmt)
	if !ok || !call(clearCall.X, "m", "Clear") {
		return false
	}
	lastCheck, ok := tail[2].(*ast.IfStmt)
	if !ok || lastCheck.Else != nil || len(lastCheck.Body.List) != 1 || !returns(lastCheck.Body.List[0], "nil", "e") {
		return false
	}
	checkInit, ok := lastCheck.Init.(*ast.AssignStmt)
	if !ok || checkInit.Tok != token.DEFINE || len(checkInit.Lhs) != 1 || !id(checkInit.Lhs[0], "e") || len(checkInit.Rhs) != 1 || !call(checkInit.Rhs[0], "ctx", "Err") {
		return false
	}
	checkCondition, ok := lastCheck.Cond.(*ast.BinaryExpr)
	return ok && checkCondition.Op == token.NEQ && id(checkCondition.X, "e") && id(checkCondition.Y, "nil") && returns(tail[3], "m", "nil")
}

func TestContractEvaluateCleanupRegistration(t *testing.T) {
	const good = `package statecontract
func EvaluateModel(ctx context.Context,in ModelInput)(result *ContractModel,err error){
    if ctx==nil{return nil,ErrInvalidInput}
    m:=&ContractModel{valid:true,input:cloneInput(in)}
    defer func(){if result==nil{m.Clear()}}()
    if e:=ctx.Err();e!=nil{return nil,e}
    return m,nil
}`
	if !contractCleanupRegistered([]byte(good)) {
		t.Fatal("protected cleanup rejected")
	}
	for name, source := range map[string]string{
		"missing":           strings.Replace(good, "defer func(){if result==nil{m.Clear()}}()", "", 1),
		"late":              strings.Replace(good, "defer func(){if result==nil{m.Clear()}}()", "ctx.Err();defer func(){if result==nil{m.Clear()}}()", 1),
		"wrong_owner":       strings.Replace(good, "m.Clear()", "other.Clear()", 1),
		"error_only":        strings.Replace(good, "result==nil", "err!=nil", 1),
		"conditional_clear": strings.Replace(good, "m.Clear()", "if enabled{m.Clear()}", 1),
		"early_transfer":    strings.Replace(good, "if e:=ctx.Err()", "result=m;if e:=ctx.Err()", 1),
		"reassigned_owner":  strings.Replace(good, "if e:=ctx.Err()", "m=&ContractModel{};if e:=ctx.Err()", 1),
		"no_transfer":       strings.Replace(good, "return m,nil", "return nil,nil", 1),
		"swallowed_error":   strings.Replace(good, "return nil,e", "return m,nil", 1),
		"broken":            "broken",
	} {
		t.Run(name, func(t *testing.T) {
			if contractCleanupRegistered([]byte(source)) {
				t.Fatal("unsafe cleanup accepted")
			}
		})
	}
	source, err := os.ReadFile("evaluate.go")
	if err != nil {
		t.Fatal(err)
	}
	if !contractCleanupRegistered(source) {
		t.Fatal("owner lacks immediate panic-safe deferred cleanup before final context callback")
	}
}

type contractCleanupContext struct {
	context.Context
	calls, at  int
	err        error
	panicValue any
}

func (c *contractCleanupContext) Err() error {
	c.calls++
	if c.calls == c.at {
		if c.panicValue != nil {
			panic(c.panicValue)
		}
		return c.err
	}
	return nil
}

func TestContractEvaluateCleanupOutcomes(t *testing.T) {
	panicValue := &struct{ label string }{"context panic"}
	for _, tc := range []struct {
		name       string
		err        error
		panicValue any
	}{
		{"success", nil, nil},
		{"canceled", context.Canceled, nil},
		{"deadline", context.DeadlineExceeded, nil},
		{"panic", nil, panicValue},
	} {
		t.Run(tc.name, func(t *testing.T) {
			in, expected := ownershipFixture(), ownershipFixture()
			ctx := &contractCleanupContext{Context: context.Background(), at: 2, err: tc.err, panicValue: tc.panicValue}
			var model *ContractModel
			var err error
			var caught any
			func() {
				defer func() { caught = recover() }()
				model, err = EvaluateModel(ctx, in)
			}()
			defer model.Clear()
			if caught != tc.panicValue || !errors.Is(err, tc.err) || ctx.calls != 2 {
				t.Fatalf("wrong outcome: panic=%v error=%v callbacks=%d", caught, err, ctx.calls)
			}
			if !reflect.DeepEqual(in, expected) {
				t.Fatal("exit modified caller input")
			}
			if tc.err != nil || tc.panicValue != nil {
				if model != nil {
					t.Fatal("failure returned model")
				}
				return
			}
			if model == nil || !model.valid || !reflect.DeepEqual(model.input, expected) || model.Summary().Status != ContractShapeValid {
				t.Fatal("success owner cleared before ownership transfer")
			}
		})
	}
}

func TestContractRecheckCleanupPanicPreservesOwner(t *testing.T) {
	for _, at := range []int{3, 4} {
		t.Run(string(rune('0'+at)), func(t *testing.T) {
			in, expected := ownershipFixture(), ownershipFixture()
			model, err := EvaluateModel(context.Background(), in)
			if err != nil {
				t.Fatal(err)
			}
			defer model.Clear()
			panicValue := &struct{ stage int }{at}
			ctx := &contractCleanupContext{Context: context.Background(), at: at, panicValue: panicValue}
			var caught any
			func() {
				defer func() { caught = recover() }()
				_ = RecheckModel(ctx, model, in)
			}()
			if caught != panicValue || ctx.calls != at || !model.valid || !reflect.DeepEqual(model.input, expected) || !reflect.DeepEqual(in, expected) {
				t.Fatal("recheck swallowed panic or cleared borrowed owner/input")
			}
		})
	}
}

func TestContractCleanupOwnedBackingStores(t *testing.T) {
	in, expected := ownershipFixture(), ownershipFixture()
	model, err := EvaluateModel(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	defer model.Clear()
	held := model.input
	paths := collectionPaths(reflect.ValueOf(held), "")
	if len(paths) != 19 {
		t.Fatalf("fixture collection count %d, want 19", len(paths))
	}
	model.Clear()
	if !reflect.DeepEqual(*model, ContractModel{}) || !reflect.DeepEqual(in, expected) {
		t.Fatal("Clear retained model or modified caller")
	}
	for _, path := range paths {
		v := reflect.ValueOf(held)
		for _, field := range strings.Split(path, ".") {
			v = v.FieldByName(field)
		}
		if v.Len() == 0 {
			t.Fatal("unexercised backing store", path)
		}
		for i := 0; i < v.Len(); i++ {
			if !v.Index(i).IsZero() {
				t.Fatal("uncleared backing store", path, i)
			}
		}
	}
}
