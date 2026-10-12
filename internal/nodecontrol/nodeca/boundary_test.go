package nodeca

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func boundaryFiles(t testing.TB) map[string]*ast.File {
	t.Helper()
	entries, e := os.ReadDir(".")
	if e != nil {
		t.Fatal(e)
	}
	out := map[string]*ast.File{}
	for _, v := range entries {
		if strings.HasSuffix(v.Name(), ".go") && !strings.HasSuffix(v.Name(), "_test.go") {
			f, e := parser.ParseFile(token.NewFileSet(), v.Name(), nil, 0)
			if e != nil {
				t.Fatal(e)
			}
			out[v.Name()] = f
		}
	}
	return out
}
func boundaryParse(t testing.TB, s string) *ast.File {
	t.Helper()
	f, e := parser.ParseFile(token.NewFileSet(), "injected.go", s, 0)
	if e != nil {
		t.Fatal("injected syntax", e)
	}
	return f
}
func TestCAImportsBoundary(t *testing.T) {
	for name, f := range boundaryFiles(t) {
		if e := guardSource(f); e != "" {
			t.Fatalf("%s: %s", name, e)
		}
	}
	for _, s := range []string{
		`package nodeca;import "os";func f(){os.ReadFile("x")}`,
		`package nodeca;import cert "crypto/x509";func f(){cert.ParseCertificate(nil)}`,
		`package nodeca;import . "crypto/x509";func f(){ParseCertificate(nil)}`,
		`package nodeca;import _ "net/http"`,
		`package nodeca;import "bytes";func f(){bytes:=struct{Clone func([]byte)[]byte}{};bytes.Clone(nil)}`,
		`package nodeca;func f(){fake.ParseCertificate(nil)}`,
		`package nodeca;import "crypto/x509";var f=x509.ParseCertificate`,
		`package nodeca;import "crypto/x509";func f(cert *x509.Certificate){m:=cert.CheckSignature;_ = m}`,
	} {
		if guardSource(boundaryParse(t, s)) == "" {
			t.Fatalf("accepted import/member bypass: %s", s)
		}
	}
	// Input.DER is an input field, not an output method.
	if e := guardSource(boundaryParse(t, `package nodeca;func f(in Input){_ = in.DER}`)); e != "" {
		t.Fatal("Input.DER false positive", e)
	}
	if e := guardSource(boundaryParse(t, `package nodeca;import "crypto/x509";func f(){_ = x509.KeyUsageCertSign | x509.KeyUsageCRLSign;decoded.Equal(parsed)}`)); e != "" {
		t.Fatal("constant/time receiver false positive", e)
	}
	if e := guardSource(boundaryParse(t, `package nodeca;func f(){parsed.UTC().Format("2006")}`)); e != "" {
		t.Fatal("time chain false positive", e)
	}
}
func TestCAAPI(t *testing.T) {
	files := boundaryFiles(t)
	if e := guardAPI(files); e != "" {
		t.Fatal(e)
	}
	for _, s := range []string{`package nodeca;func Ready()bool{return true}`, `package nodeca;func(m *Material) DER()[]byte{return nil}`, `package nodeca;type Capability bool`, `package nodeca;var cache int`, `package nodeca;const Trusted=2`} {
		injected := map[string]*ast.File{}
		for k, v := range files {
			injected[k] = v
		}
		injected["injected.go"] = boundaryParse(t, s)
		if guardAPI(injected) == "" {
			t.Fatalf("accepted API/global: %s", s)
		}
	}
	types := map[string]reflect.Type{"Input": reflect.TypeOf(Input{}), "Summary": reflect.TypeOf(Summary{}), "Material": reflect.TypeOf(Material{}), "KeyIdentifier": reflect.TypeOf(KeyIdentifier{})}
	if types["Input"].NumField() != 2 || types["Summary"].NumField() != 15 || types["Material"].NumField() != 1 || types["KeyIdentifier"].Kind() != reflect.Array || types["KeyIdentifier"].Len() != 32 {
		t.Fatal("value layout")
	}
	if reflect.TypeOf(None).Kind() != reflect.Uint8 || reflect.TypeOf(UnknownObligation).Kind() != reflect.Uint8 || None != 0 || NodeCAMaterialShapeAndDeclaredKeyBound != 1 || expectedPending() != ([6]Obligation{1, 2, 3, 4, 5, 6}) {
		t.Fatal("enum strength")
	}
}
func TestCAConsumers(t *testing.T) {
	target := "talenro.local/platform/internal/nodecontrol/nodeca"
	source := `package nodeca;import "` + target + `";var _=nodeca.Parse`
	cases := []struct {
		path, pkg string
		allow     bool
	}{
		{"internal/nodecontrol/nodeca/parse_test.go", "nodeca", true}, {"internal/nodecontrol/nodeca/external_test.go", "nodeca_test", false},
		{"internal/nodecontrol/nodeca_sibling/a_test.go", "nodeca", false}, {"internal/nodecontrol/nodeca_extra/a_test.go", "nodeca", false},
		{"internal/nodecontrol/nodeca/nested/a_test.go", "nodeca", false}, {"internal/other/a.go", "other", false},
		{"internal/other/a_test.go", "nodeca", false}, {"internal/nodecontrol/nodeca/a.go", "other", false},
	}
	for _, c := range cases {
		f := boundaryParse(t, strings.Replace(source, "package nodeca;", "package "+c.pkg+";", 1))
		if (guardConsumer(c.path, f) == "") != c.allow {
			t.Fatalf("consumer %s/%s", c.path, c.pkg)
		}
	}
	root := filepath.Clean("../../..")
	e := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", ".superpowers", ".worktrees", "node_modules", "vendor":
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") {
			return nil
		}
		f, e := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if e != nil {
			return e
		}
		rel, e := filepath.Rel(root, path)
		if e != nil {
			return e
		}
		if issue := guardConsumer(filepath.ToSlash(rel), f); issue != "" {
			t.Errorf("%s: %s", rel, issue)
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
}
func TestCAFileInventory(t *testing.T) {
	entries, e := os.ReadDir(".")
	if e != nil {
		t.Fatal(e)
	}
	var names []string
	for _, v := range entries {
		names = append(names, v.Name())
	}
	sort.Strings(names)
	if e := guardInventory(names); e != "" {
		t.Fatal(e)
	}
	for _, bad := range [][]string{append(append([]string{}, names...), "extra.go"), names[:len(names)-1], append(append([]string{}, names...), "testdata")} {
		if guardInventory(bad) == "" {
			t.Fatal("inventory bypass")
		}
	}
}
func TestCAParseOwnershipFlow(t *testing.T) {
	for _, name := range []string{"owned.go", "parse.go"} {
		b, e := os.ReadFile(name)
		if e != nil {
			t.Fatal(e)
		}
		if issue := guardOwnership(boundaryParse(t, string(b)), name); issue != "" {
			t.Fatal(name, issue)
		}
		count := 3
		if name == "owned.go" {
			count = 4
		}
		for n := 0; n < count; n++ {
			f := boundaryParse(t, string(b))
			if !mutateOwnership(f, name, n) {
				t.Fatal("ownership mutation not applied")
			}
			if guardOwnership(f, name) == "" {
				t.Fatal("ownership bypass", name, n)
			}
		}
	}
}

func exprString(n ast.Node) string {
	var b bytes.Buffer
	if e := format.Node(&b, token.NewFileSet(), n); e != nil {
		panic(e)
	}
	s := b.String()
	s = strings.ReplaceAll(s, " ", "")
	s = strings.ReplaceAll(s, "\n", "")
	return strings.ReplaceAll(s, "\t", "")
}
func rootIdent(e ast.Expr) string {
	switch v := e.(type) {
	case *ast.Ident:
		return v.Name
	case *ast.SelectorExpr:
		return rootIdent(v.X)
	case *ast.IndexExpr:
		return rootIdent(v.X)
	case *ast.ParenExpr:
		return rootIdent(v.X)
	case *ast.CallExpr:
		return rootIdent(v.Fun)
	}
	return ""
}
func stringSet(s string) map[string]bool {
	m := map[string]bool{}
	for _, v := range strings.Fields(s) {
		m[v] = true
	}
	return m
}

// Finite AST boundary, not a Go type checker or a formal proof.
func guardSource(f *ast.File) string {
	imports := map[string]map[string]bool{}
	allowed := map[string]string{
		"bytes": "Clone Equal", "context": "Context", "errors": "New", "crypto/sha256": "Sum256", "crypto/ecdsa": "PublicKey", "crypto/elliptic": "P256",
		"crypto/x509":   "Certificate ParseCertificate ParsePKIXPublicKey MarshalPKIXPublicKey KeyUsageCertSign KeyUsageCRLSign",
		"encoding/asn1": "ObjectIdentifier Unmarshal RawValue", "time": "Time", "talenro.local/platform/internal/nodecontrol/contracts": "Digest",
	}
	types := stringSet("Context PublicKey Certificate ObjectIdentifier RawValue Time Digest KeyUsageCertSign KeyUsageCRLSign")
	for _, v := range f.Imports {
		p, e := strconv.Unquote(v.Path.Value)
		if e != nil {
			return "import syntax"
		}
		members, ok := allowed[p]
		if !ok {
			return "forbidden import: " + p
		}
		name := filepath.Base(p)
		if v.Name != nil && v.Name.Name != name {
			return "import alias/dot/blank"
		}
		if imports[name] != nil {
			return "duplicate import name"
		}
		imports[name] = stringSet(members)
	}
	fields := stringSet("DER DeclaredSPKI Proof Pending CADERDigest SPKIDERDigest SubjectDERDigest IssuerDERDigest SubjectKeyIdentifier AuthorityKeyIdentifier Serial SerialLength NotBefore NotAfter HasPathLen PathLen CRLSignAllowed owned der declaredSPKI summary valid tag full content id critical value tbs serial issuer subject spki notBefore notAfter innerSignatureAlgorithm outerSignatureAlgorithm signature extensions X Y Curve Raw RawTBSCertificate RawSubjectPublicKeyInfo RawSubject RawIssuer PublicKey Version SerialNumber Extensions Id Critical Value BasicConstraintsValid IsCA MaxPathLen MaxPathLenZero KeyUsage SubjectKeyId AuthorityKeyId Signature")
	methods := stringSet("Err clear Year Nanosecond Zone UTC Format Equal Before Sign Bytes Cmp IsOnCurve")
	roots := stringSet("ctx o m k key parsed decoded cert raw v slots extensions fields summary")
	parents := map[ast.Node]ast.Node{}
	var stack []ast.Node
	ast.Inspect(f, func(n ast.Node) bool {
		if n == nil {
			stack = stack[:len(stack)-1]
			return true
		}
		if len(stack) > 0 {
			parents[n] = stack[len(stack)-1]
		}
		stack = append(stack, n)
		return true
	})
	issue := ""
	ast.Inspect(f, func(n ast.Node) bool {
		if issue != "" {
			return false
		}
		switch v := n.(type) {
		case *ast.Ident:
			if imports[v.Name] != nil && v.Obj != nil {
				issue = "import name shadowed"
			}
		case *ast.GoStmt, *ast.ChanType, *ast.SendStmt, *ast.SelectStmt:
			issue = "concurrency/runtime construct"
		case *ast.CallExpr:
			if id, ok := v.Fun.(*ast.Ident); ok && stringSet("recover print println")[id.Name] {
				issue = "runtime builtin"
			}
		case *ast.SelectorExpr:
			if id, ok := v.X.(*ast.Ident); ok && imports[id.Name] != nil {
				if !imports[id.Name][v.Sel.Name] {
					issue = "forbidden imported member: " + v.Sel.Name
					break
				}
				if !types[v.Sel.Name] {
					call, ok := parents[v].(*ast.CallExpr)
					if !ok || call.Fun != v {
						issue = "imported method/function value"
					}
				}
				break
			}
			if methods[v.Sel.Name] {
				call, ok := parents[v].(*ast.CallExpr)
				if !ok || call.Fun != v || !roots[rootIdent(v.X)] {
					issue = "unbound or disguised instance method: " + exprString(v)
				}
			} else if !fields[v.Sel.Name] {
				issue = "forbidden instance selector: " + v.Sel.Name
			}
		case *ast.FuncLit:
			call, ok := parents[v].(*ast.CallExpr)
			if !ok {
				issue = "callback"
				break
			}
			if _, ok = parents[call].(*ast.DeferStmt); !ok {
				issue = "non-cleanup closure"
			}
		}
		return true
	})
	return issue
}

func fieldTypes(fields *ast.FieldList) []string {
	var out []string
	if fields == nil {
		return out
	}
	for _, f := range fields.List {
		n := len(f.Names)
		if n == 0 {
			n = 1
		}
		for i := 0; i < n; i++ {
			out = append(out, exprString(f.Type))
		}
	}
	return out
}
func structLayout(s *ast.StructType) map[string]string {
	out := map[string]string{}
	for _, f := range s.Fields.List {
		if f.Tag != nil || len(f.Names) == 0 {
			return nil
		}
		for _, n := range f.Names {
			if _, exists := out[n.Name]; exists {
				return nil
			}
			out[n.Name] = exprString(f.Type)
		}
	}
	return out
}
func guardAPI(files map[string]*ast.File) string {
	layouts := map[string]map[string]string{
		"Input":    {"DER": "[]byte", "DeclaredSPKI": "[]byte"},
		"Summary":  {"Proof": "Proof", "Pending": "[6]Obligation", "CADERDigest": "contracts.Digest", "SPKIDERDigest": "contracts.Digest", "SubjectDERDigest": "contracts.Digest", "IssuerDERDigest": "contracts.Digest", "SubjectKeyIdentifier": "KeyIdentifier", "AuthorityKeyIdentifier": "KeyIdentifier", "Serial": "[20]byte", "SerialLength": "uint8", "NotBefore": "time.Time", "NotAfter": "time.Time", "HasPathLen": "bool", "PathLen": "uint8", "CRLSignAllowed": "bool"},
		"Material": {"owned": "*owner"}, "owner": {"der": "[]byte", "declaredSPKI": "[]byte", "summary": "Summary", "valid": "bool"},
	}
	publicTypes := map[string]string{"Proof": "uint8", "Obligation": "uint8", "KeyIdentifier": "[32]byte"}
	privateTypes := stringSet("derElement rawExtension rawCA fieldSummary extensionSummary")
	consts := stringSet("None NodeCAMaterialShapeAndDeclaredKeyBound UnknownObligation CertificateSignatureAndIssuerChain AuthorityIdentifierAndParentBinding TrustedCurrentTime DeploymentAuthorizationAndPurpose BundleStatusAndRollback LeafIssuerRelationAndRuntimeUse")
	errorNames := map[string]string{"ErrInvalidInput": "nodeca: invalid input", "ErrLimitExceeded": "nodeca: limit exceeded", "ErrInvalidMaterial": "nodeca: invalid material", "ErrUnsupportedCAProfile": "nodeca: unsupported CA profile", "ErrBindingMismatch": "nodeca: binding mismatch"}
	seen := map[string]bool{}
	for _, f := range files {
		if f.Name.Name != "nodeca" {
			return "package changed"
		}
		for _, d := range f.Decls {
			switch v := d.(type) {
			case *ast.FuncDecl:
				if v.Name.Name == "init" {
					return "product init"
				}
				if !v.Name.IsExported() {
					continue
				}
				name := v.Name.Name
				if v.Recv != nil {
					if len(v.Recv.List) != 1 || exprString(v.Recv.List[0].Type) != "*Material" {
						return "extra exported receiver"
					}
					name = "Material." + name
				}
				if seen[name] {
					return "duplicate exported function"
				}
				seen[name] = true
				params, results := fieldTypes(v.Type.Params), fieldTypes(v.Type.Results)
				switch name {
				case "Parse":
					if !reflect.DeepEqual(params, []string{"context.Context", "Input"}) || !reflect.DeepEqual(results, []string{"*Material", "error"}) {
						return "Parse signature"
					}
				case "Material.Summary":
					if len(params) != 0 || !reflect.DeepEqual(results, []string{"Summary"}) {
						return "Summary signature"
					}
				case "Material.Clear":
					if len(params) != 0 || len(results) != 0 {
						return "Clear signature"
					}
				default:
					return "extra exported function"
				}
			case *ast.GenDecl:
				for _, s := range v.Specs {
					switch spec := s.(type) {
					case *ast.TypeSpec:
						name := spec.Name.Name
						if seen[name] || spec.Assign.IsValid() {
							return "duplicate or aliased type"
						}
						seen[name] = true
						if layout, ok := layouts[name]; ok {
							st, ok := spec.Type.(*ast.StructType)
							if !ok || !reflect.DeepEqual(structLayout(st), layout) {
								return "layout changed: " + name
							}
						} else if want, ok := publicTypes[name]; ok {
							if exprString(spec.Type) != want {
								return "value type changed"
							}
						} else if !privateTypes[name] {
							return "extra named type"
						}
					case *ast.ValueSpec:
						for _, id := range spec.Names {
							if seen[id.Name] {
								return "duplicate value"
							}
							seen[id.Name] = true
							if v.Tok == token.VAR {
								message, ok := errorNames[id.Name]
								if !ok || len(spec.Names) != 1 || len(spec.Values) != 1 {
									return "extra mutable global"
								}
								call, ok := spec.Values[0].(*ast.CallExpr)
								if !ok || exprString(call.Fun) != "errors.New" || len(call.Args) != 1 {
									return "error not independent sentinel"
								}
								lit, ok := call.Args[0].(*ast.BasicLit)
								if !ok {
									return "error message"
								}
								got, e := strconv.Unquote(lit.Value)
								if e != nil || got != message {
									return "error message changed"
								}
							} else if !consts[id.Name] {
								return "extra constant"
							}
						}
					}
				}
			}
		}
	}
	for _, name := range []string{"Input", "Summary", "Material", "owner", "Proof", "Obligation", "KeyIdentifier", "Parse", "Material.Summary", "Material.Clear"} {
		if !seen[name] {
			return "missing API: " + name
		}
	}
	for name := range consts {
		if !seen[name] {
			return "missing enum"
		}
	}
	for name := range errorNames {
		if !seen[name] {
			return "missing error"
		}
	}
	return ""
}
func guardConsumer(path string, f *ast.File) string {
	target := "talenro.local/platform/internal/nodecontrol/nodeca"
	path = filepath.ToSlash(filepath.Clean(path))
	for _, i := range f.Imports {
		p, e := strconv.Unquote(i.Path.Value)
		if e != nil {
			return "import syntax"
		}
		if p != target && !strings.HasPrefix(p, target+"/") {
			continue
		}
		if filepath.ToSlash(filepath.Dir(path)) != "internal/nodecontrol/nodeca" || f.Name.Name != "nodeca" || !strings.HasSuffix(path, "_test.go") {
			return "nodeca consumer forbidden"
		}
	}
	return ""
}
func guardInventory(names []string) string {
	want := strings.Fields("doc.go input.go errors.go budget.go owned.go material.go spki.go der.go raw_certificate.go ca_fields.go extensions.go parse.go fixtures_test.go budget_test.go owned_test.go material_test.go spki_test.go der_test.go raw_certificate_test.go ca_fields_test.go extensions_test.go parse_test.go boundary_test.go fuzz_test.go")
	sort.Strings(want)
	got := append([]string(nil), names...)
	sort.Strings(got)
	if !reflect.DeepEqual(got, want) {
		return "exact 24-file inventory changed"
	}
	return ""
}

func functionNamed(f *ast.File, name string) *ast.FuncDecl {
	for _, d := range f.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == name {
			return fn
		}
	}
	return nil
}
func cleanupDefer(stmt ast.Stmt) bool {
	d, ok := stmt.(*ast.DeferStmt)
	if !ok {
		return false
	}
	lit, ok := d.Call.Fun.(*ast.FuncLit)
	if !ok || len(lit.Body.List) != 1 {
		return false
	}
	branch, ok := lit.Body.List[0].(*ast.IfStmt)
	if !ok || exprString(branch.Cond) != "!transferred" || branch.Else != nil || len(branch.Body.List) != 1 {
		return false
	}
	x, ok := branch.Body.List[0].(*ast.ExprStmt)
	if !ok {
		return false
	}
	call, ok := x.X.(*ast.CallExpr)
	return ok && exprString(call.Fun) == "o.clear" && len(call.Args) == 0
}
func guardOwnership(f *ast.File, name string) string {
	fnName := "Parse"
	if name == "owned.go" {
		fnName = "ownInput"
	}
	fn := functionNamed(f, fnName)
	if fn == nil {
		return "missing ownership function"
	}
	statements := fn.Body.List
	deferIndex, firstCopy, lastCopy, ownIndex, validIndex, transferIndex := -1, -1, -1, -1, -1, -1
	var copies []string
	preflightPos, ownPos := token.NoPos, token.NoPos
	errCalls := 0
	successReturns := 0
	for i, s := range statements {
		if cleanupDefer(s) {
			deferIndex = i
		}
		if a, ok := s.(*ast.AssignStmt); ok {
			for j, lhs := range a.Lhs {
				if j >= len(a.Rhs) {
					continue
				}
				rhs := a.Rhs[j]
				switch exprString(lhs) {
				case "o.valid":
					if exprString(rhs) != "true" {
						return "valid false handoff"
					}
					validIndex = i
				case "transferred":
					if exprString(rhs) == "true" {
						transferIndex = i
					}
				}
				if call, ok := rhs.(*ast.CallExpr); ok {
					fun := exprString(call.Fun)
					if fun == "bytes.Clone" {
						if firstCopy < 0 {
							firstCopy = i
						}
						lastCopy = i
						if len(call.Args) != 1 {
							return "clone arguments"
						}
						copies = append(copies, exprString(lhs)+"="+exprString(call.Args[0]))
					}
					if fun == "ownInput" {
						ownIndex = i
						ownPos = call.Pos()
					}
				}
			}
		}
	}
	issue := ""
	ast.Inspect(fn.Body, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		switch v := n.(type) {
		case *ast.CallExpr:
			fun := exprString(v.Fun)
			if fun == "preflight" {
				preflightPos = v.Pos()
			}
			if fun == "ctx.Err" {
				errCalls++
			}
			if name == "parse.go" && fun == "x509.ParseCertificate" && (len(v.Args) != 1 || exprString(v.Args[0]) != "o.der") {
				issue = "parser reads caller buffer"
			}
		case *ast.SelectorExpr:
			if name == "parse.go" && rootIdent(v.X) == "in" && v.Pos() > ownPos && ownPos.IsValid() {
				issue = "caller read after ownership"
			}
		case *ast.ReturnStmt:
			if name == "parse.go" && len(v.Results) == 2 && exprString(v.Results[1]) == "nil" {
				successReturns++
				if exprString(v.Results[0]) != "&Material{owned:o}" {
					issue = "unowned success"
				}
			}
		}
		return true
	})
	if issue != "" {
		return issue
	}
	if name == "owned.go" {
		if deferIndex < 0 || deferIndex >= firstCopy || !reflect.DeepEqual(copies, []string{"o.der=in.DER", "o.declaredSPKI=in.DeclaredSPKI"}) || transferIndex <= lastCopy {
			return "cleanup-before-independent-clones missing"
		}
	} else {
		if ownIndex < 0 || deferIndex <= ownIndex || deferIndex > ownIndex+2 || !preflightPos.IsValid() || preflightPos >= ownPos || validIndex <= deferIndex || transferIndex <= validIndex || successReturns != 1 || errCalls < 8 {
			return "Parse lifecycle ordering"
		}
		if validIndex < 1 {
			return "handoff"
		}
		prior, ok := statements[validIndex-1].(*ast.IfStmt)
		if !ok || prior.Init == nil || !strings.Contains(exprString(prior.Init), "ctx.Err()") {
			return "missing final context before valid"
		}
		last, ok := statements[len(statements)-1].(*ast.ReturnStmt)
		if !ok || len(last.Results) != 2 || exprString(last.Results[0]) != "&Material{owned:o}" {
			return "nonfinal handoff"
		}
	}
	return ""
}
func mutateOwnership(f *ast.File, name string, n int) bool {
	fnName := "Parse"
	if name == "owned.go" {
		fnName = "ownInput"
	}
	fn := functionNamed(f, fnName)
	changed := false
	if n == 0 && name == "owned.go" {
		for i, s := range fn.Body.List {
			if d, ok := s.(*ast.DeferStmt); ok {
				fn.Body.List[i] = &ast.ExprStmt{X: d.Call}
				return true
			}
		}
	}
	if n == 3 && name == "owned.go" {
		for i, s := range fn.Body.List {
			if a, ok := s.(*ast.AssignStmt); ok && len(a.Lhs) == 1 && exprString(a.Lhs[0]) == "transferred" && exprString(a.Rhs[0]) == "true" {
				fn.Body.List[i], fn.Body.List[i-1] = fn.Body.List[i-1], fn.Body.List[i]
				return true
			}
		}
	}
	ast.Inspect(fn.Body, func(node ast.Node) bool {
		if changed {
			return false
		}
		switch v := node.(type) {
		case *ast.UnaryExpr:
			if n == 1 && name == "owned.go" || n == 2 && name == "parse.go" {
				if exprString(v) == "!transferred" {
					v.Op = token.ADD
					changed = true
				}
			}
		case *ast.CallExpr:
			if n == 2 && name == "owned.go" && exprString(v.Fun) == "bytes.Clone" && len(v.Args) == 1 && exprString(v.Args[0]) == "in.DeclaredSPKI" {
				v.Fun = ast.NewIdent("identity")
				changed = true
			}
			if n == 0 && name == "parse.go" && exprString(v.Fun) == "x509.ParseCertificate" {
				v.Args[0] = &ast.SelectorExpr{X: ast.NewIdent("in"), Sel: ast.NewIdent("DER")}
				changed = true
			}
		case *ast.AssignStmt:
			if n == 1 && name == "parse.go" && len(v.Lhs) == 1 && exprString(v.Lhs[0]) == "o.valid" {
				v.Rhs[0] = ast.NewIdent("false")
				changed = true
			}
		}
		return true
	})
	return changed
}
func TestCANoRuntimeCalls(t *testing.T) {
	for name, f := range boundaryFiles(t) {
		if issue := guardSource(f); issue != "" {
			t.Fatal(name, issue)
		}
	}
	for _, s := range []string{
		`package nodeca;import "crypto/x509";func f(cert *x509.Certificate){cert.Verify(opts)}`,
		`package nodeca;import "time";func f(){time.Now()}`,
		`package nodeca;import "crypto/x509";func f(){x509.CreateCertificate(nil,nil,nil,nil,nil)}`,
		`package nodeca;func f(){go f()}`, `package nodeca;func f(){_ = recover()}`,
		`package nodeca;func f(){var x chan int;_ = x}`, `package nodeca;func f(){fake.Equal(nil)}`,
	} {
		if guardSource(boundaryParse(t, s)) == "" {
			t.Fatalf("runtime bypass: %s", s)
		}
	}
}
