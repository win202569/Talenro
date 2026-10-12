package nodeleafkeys

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const keysGoFiles = "doc.go input.go errors.go budget.go owned.go material.go issuer_spki.go key_identifiers.go evaluate.go fixtures_test.go budget_test.go owned_test.go material_test.go issuer_spki_test.go key_identifiers_test.go evaluate_test.go boundary_test.go"

func keysPrint(n ast.Node) string {
	var b bytes.Buffer
	if format.Node(&b, token.NewFileSet(), n) != nil {
		return ""
	}
	return b.String()
}
func keysImportsAllowed(source []byte) bool {
	f, e := parser.ParseFile(token.NewFileSet(), "memory.go", source, 0)
	if e != nil {
		return false
	}
	allowed := map[string]string{
		"bytes": "Clone Equal", "context": "Context", "errors": "New", "crypto/sha256": "Sum256",
		"crypto/rsa": "PublicKey", "crypto/ecdsa": "PublicKey", "crypto/ed25519": "PublicKey PublicKeySize", "crypto/elliptic": "P256 P384 P521",
		"crypto/x509":            "Certificate ParseCertificate ParsePKIXPublicKey MarshalPKIXPublicKey",
		"github.com/google/uuid": "UUID Nil RFC4122", "talenro.local/platform/internal/nodecontrol/contracts": "Digest",
		"talenro.local/platform/internal/nodecontrol/nodeleaf": "Input Summary Obligation Parse None NodeLeafMaterialShapeValid LeafSignatureAndIssuerTrust KeyIdentifierDerivationAndIssuerBinding TrustedCurrentTime IdentityAndAuthorityProvenance PrivateKeyPossession MainStateAndRuntimeBinding ErrInvalidInput ErrLimitExceeded ErrInvalidMaterial ErrUnsupportedLeafProfile ErrBindingMismatch",
	}
	aliases := map[string]string{}
	paths := map[string]bool{}
	for _, im := range f.Imports {
		p, e := strconv.Unquote(im.Path.Value)
		if e != nil || paths[p] {
			return false
		}
		paths[p] = true
		if _, ok := allowed[p]; !ok {
			return false
		}
		name := p[strings.LastIndex(p, "/")+1:]
		if im.Name != nil {
			name = im.Name.Name
		}
		if name == "." || name == "_" || aliases[name] != "" {
			return false
		}
		aliases[name] = p
	}
	ok := true
	definition := func(id *ast.Ident) {
		if id != nil && aliases[id.Name] != "" {
			ok = false
		}
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.GoStmt:
			ok = false
		case *ast.SelectorExpr:
			name := v.Sel.Name
			if name == "Verify" || strings.HasPrefix(name, "CheckSignature") || (name == "DER" && keysPrint(v) != "in.Leaf.DER") || name == "Ready" || name == "Authorize" || name == "State" || name == "Commit" || name == "MarshalJSON" || name == "UnmarshalJSON" || name == "NewCertPool" || name == "SystemCertPool" {
				ok = false
			}
			if id, is := v.X.(*ast.Ident); is && aliases[id.Name] != "" {
				found := false
				for _, m := range strings.Fields(allowed[aliases[id.Name]]) {
					if m == name {
						found = true
					}
				}
				if !found {
					ok = false
				}
			}
		case *ast.AssignStmt:
			if v.Tok == token.DEFINE {
				for _, lhs := range v.Lhs {
					if id, is := lhs.(*ast.Ident); is {
						definition(id)
					}
				}
			}
		case *ast.RangeStmt:
			if v.Tok == token.DEFINE {
				if id, is := v.Key.(*ast.Ident); is {
					definition(id)
				}
				if id, is := v.Value.(*ast.Ident); is {
					definition(id)
				}
			}
		case *ast.ValueSpec:
			for _, id := range v.Names {
				definition(id)
			}
		case *ast.TypeSpec:
			definition(v.Name)
		case *ast.FuncDecl:
			definition(v.Name)
		case *ast.Field:
			for _, id := range v.Names {
				definition(id)
			}
		}
		return true
	})
	return ok
}
func keysSources(t testing.TB) map[string][]byte {
	t.Helper()
	paths, e := filepath.Glob("*.go")
	if e != nil {
		t.Fatal(e)
	}
	out := map[string][]byte{}
	for _, p := range paths {
		b, e := os.ReadFile(p)
		if e != nil {
			t.Fatal(e)
		}
		out[p] = b
	}
	return out
}
func keysCapabilityAllowed(sources map[string][]byte) bool {
	files := make([]string, 0, len(sources))
	for p := range sources {
		files = append(files, p)
	}
	sort.Strings(files)
	expected := strings.Fields(keysGoFiles)
	sort.Strings(expected)
	if strings.Join(files, ",") != strings.Join(expected, ",") {
		return false
	}
	types := map[string]string{"Input": "struct", "KeyIdentifier": "[32]byte", "Proof": "uint8", "Summary": "struct", "Material": "struct"}
	constants := map[string]string{"None": "Proof=iota", "NodeLeafDeclaredKeyIdentifiersBound": "", "keyIdentifierProfile": `"node-key-id.spki-sha256.v1"`}
	vars := map[string]string{
		"ErrInvalidInput": "nodeleaf.ErrInvalidInput", "ErrLimitExceeded": "nodeleaf.ErrLimitExceeded", "ErrInvalidMaterial": "nodeleaf.ErrInvalidMaterial",
		"ErrUnsupportedLeafProfile": "nodeleaf.ErrUnsupportedLeafProfile", "ErrBindingMismatch": "nodeleaf.ErrBindingMismatch",
		"ErrUnsupportedKeyIdentifierProfile": `errors.New("node leaf keys: unsupported key identifier profile")`,
	}
	functions := map[string]string{"Evaluate": "func(ctx context.Context, in Input) (*Material, error)", "Material.Summary": "func() Summary", "Material.Clear": "func()"}
	structs := map[string]string{
		"Input":    "DeclaredIssuerSPKI:[]byte,Leaf:nodeleaf.Input",
		"Summary":  "AuthorityKeyIdentifier:KeyIdentifier,DeclaredIssuerSPKIDERDigest:contracts.Digest,Leaf:nodeleaf.Summary,Proof:Proof,SubjectKeyIdentifier:KeyIdentifier",
		"Material": "owned:*owner", "owner": "declaredIssuerSPKI:[]byte,der:[]byte,expectedSPKI:[]byte,summary:Summary,valid:bool",
	}
	for _, p := range files {
		if strings.HasSuffix(p, "_test.go") {
			continue
		}
		f, e := parser.ParseFile(token.NewFileSet(), p, sources[p], 0)
		if e != nil || f.Name.Name != "nodeleafkeys" || !keysImportsAllowed(sources[p]) {
			return false
		}
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				for _, spec := range d.Specs {
					switch s := spec.(type) {
					case *ast.TypeSpec:
						if s.Name.IsExported() {
							want, ok := types[s.Name.Name]
							if !ok {
								return false
							}
							if want != "struct" && keysPrint(s.Type) != want {
								return false
							}
							delete(types, s.Name.Name)
						}
						if want, exists := structs[s.Name.Name]; exists {
							st, ok := s.Type.(*ast.StructType)
							if !ok {
								return false
							}
							var fields []string
							for _, field := range st.Fields.List {
								if len(field.Names) == 0 || field.Tag != nil {
									return false
								}
								for _, name := range field.Names {
									fields = append(fields, name.Name+":"+keysPrint(field.Type))
								}
							}
							sort.Strings(fields)
							if strings.Join(fields, ",") != want {
								return false
							}
							delete(structs, s.Name.Name)
						}
					case *ast.ValueSpec:
						for _, name := range s.Names {
							if d.Tok == token.VAR {
								want, ok := vars[name.Name]
								if !ok || len(s.Names) != 1 || len(s.Values) != 1 || s.Type != nil || keysPrint(s.Values[0]) != want {
									return false
								}
								delete(vars, name.Name)
							} else if d.Tok == token.CONST {
								want, ok := constants[name.Name]
								if !ok || len(s.Names) != 1 {
									return false
								}
								got := ""
								if len(s.Values) == 1 {
									got = keysPrint(s.Values[0])
								}
								if s.Type != nil {
									got = keysPrint(s.Type) + "=" + got
								}
								if got != want {
									return false
								}
								delete(constants, name.Name)
							}
						}
					}
				}
			case *ast.FuncDecl:
				if d.Name.IsExported() {
					name := d.Name.Name
					if d.Recv != nil {
						if len(d.Recv.List) != 1 {
							return false
						}
						name = strings.TrimPrefix(keysPrint(d.Recv.List[0].Type), "*") + "." + name
					}
					want, ok := functions[name]
					if !ok || keysPrint(d.Type) != want {
						return false
					}
					delete(functions, name)
				}
			}
		}
	}
	return len(types)+len(constants)+len(vars)+len(functions)+len(structs) == 0
}
func keysIDOrder(source []byte) bool {
	f, e := parser.ParseFile(token.NewFileSet(), "memory.go", source, 0)
	if e != nil {
		return false
	}
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "checkKeyIdentifiers" {
			continue
		}
		var conditions []string
		for _, stmt := range fn.Body.List {
			if v, ok := stmt.(*ast.IfStmt); ok {
				conditions = append(conditions, keysPrint(v.Cond))
			}
		}
		return strings.Join(conditions, ";") == "cert == nil;len(cert.SubjectKeyId) != 32;len(cert.AuthorityKeyId) != 32;!bytes.Equal(cert.SubjectKeyId, ski[:]);!bytes.Equal(cert.AuthorityKeyId, aki[:])"
	}
	return false
}
func keysConsumerAllowed(path, pkg, imported string) bool {
	lower := "talenro.local/platform/internal/nodecontrol/nodeleaf"
	self := "talenro.local/platform/internal/nodecontrol/nodeleafkeys"
	if imported != lower && imported != self {
		return true
	}
	if imported == lower && pkg == "nodeleaf" && strings.HasSuffix(filepath.ToSlash(path), "_test.go") {
		return true
	}
	root, e := filepath.Abs(filepath.Clean(filepath.Join("..", "..", "..")))
	if e != nil {
		return false
	}
	abs, e := filepath.Abs(filepath.Clean(path))
	if e != nil {
		return false
	}
	rel, e := filepath.Rel(root, abs)
	if e != nil {
		return false
	}
	exact := filepath.Dir(rel) == filepath.Join("internal", "nodecontrol", "nodeleafkeys") && pkg == "nodeleafkeys"
	if imported == self {
		return exact && strings.HasSuffix(filepath.ToSlash(path), "_test.go")
	}
	return exact
}
func keysCleanupRegistered(source []byte) bool {
	f, e := parser.ParseFile(token.NewFileSet(), "memory.go", source, 0)
	if e != nil {
		return false
	}
	found := false
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil {
			continue
		}
		name := fn.Name.Name
		if name != "ownInput" && name != "Evaluate" && name != "checkDeclaredIssuerSPKI" {
			continue
		}
		found = true
		body := fn.Body.List
		if name == "checkDeclaredIssuerSPKI" {
			seen := false
			for i, stmt := range body {
				if keysPrint(stmt) == "canonical, err := x509.MarshalPKIXPublicKey(key)" {
					if i+1 >= len(body) || keysPrint(body[i+1]) != "defer clear(canonical)" {
						return false
					}
					seen = true
				}
			}
			if !seen {
				return false
			}
			continue
		}
		start := 0
		if name == "Evaluate" {
			if len(body) < 8 || keysPrint(body[0]) != "err := preflight(ctx, in)" {
				return false
			}
			if !strings.Contains(keysPrint(body[1]), "ctx != nil") || !strings.Contains(keysPrint(body[1]), "ctx.Err()") || !strings.Contains(keysPrint(body[2]), "err != nil") {
				return false
			}
			start = 3
		}
		if len(body) < start+5 {
			return false
		}
		expected := "o := &owner{}"
		if name == "Evaluate" {
			expected = "o := ownInput(in)"
		}
		if keysPrint(body[start]) != expected || keysPrint(body[start+1]) != "transferred := false" {
			return false
		}
		deferStmt, ok := body[start+2].(*ast.DeferStmt)
		if !ok || len(deferStmt.Call.Args) != 0 {
			return false
		}
		lit, ok := deferStmt.Call.Fun.(*ast.FuncLit)
		if !ok || len(lit.Body.List) != 1 {
			return false
		}
		guard, ok := lit.Body.List[0].(*ast.IfStmt)
		if !ok || guard.Init != nil || guard.Else != nil || keysPrint(guard.Cond) != "!transferred" || len(guard.Body.List) != 1 || keysPrint(guard.Body.List[0]) != "o.clear()" {
			return false
		}
		last := len(body) - 1
		if keysPrint(body[last-1]) != "transferred = true" {
			return false
		}
		if name == "ownInput" {
			if last != 7 || keysPrint(body[3]) != "o.der = bytes.Clone(in.Leaf.DER)" || keysPrint(body[4]) != "o.expectedSPKI = bytes.Clone(in.Leaf.ExpectedSPKI)" || keysPrint(body[5]) != "o.declaredIssuerSPKI = bytes.Clone(in.DeclaredIssuerSPKI)" || keysPrint(body[last]) != "return o" {
				return false
			}
		} else {
			if keysPrint(body[last]) != "return &Material{owned: o}, nil" || keysPrint(body[last-2]) != "o.valid = true" || !strings.HasPrefix(keysPrint(body[last-3]), "o.summary = Summary{") || !strings.Contains(keysPrint(body[last-4]), "ctx.Err()") {
				return false
			}
			lowerFound := false
			for i, stmt := range body {
				if strings.HasPrefix(keysPrint(stmt), "lower, err := nodeleaf.Parse(") {
					if i+1 >= len(body) || keysPrint(body[i+1]) != "defer lower.Clear()" {
						return false
					}
					text := strings.Join(strings.Fields(keysPrint(stmt)), " ")
					for _, part := range []string{"DER: o.der", "ExpectedSPKI: o.expectedSPKI", "NodeID: in.Leaf.NodeID", "TrustDomain: in.Leaf.TrustDomain"} {
						if !strings.Contains(text, part) {
							return false
						}
					}
					lowerFound = true
				}
			}
			if !lowerFound {
				return false
			}
		}
		transfers, valids := 0, 0
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if a, ok := n.(*ast.AssignStmt); ok {
				if keysPrint(a) == "transferred = true" {
					transfers++
				}
				if keysPrint(a) == "o.valid = true" {
					valids++
				}
			}
			return true
		})
		if transfers != 1 || (name == "Evaluate" && valids != 1) {
			return false
		}
	}
	return found
}
func TestKeysProductImports(t *testing.T) {
	for p, b := range keysSources(t) {
		if !strings.HasSuffix(p, "_test.go") && !keysImportsAllowed(b) {
			t.Fatal("product boundary", p)
		}
	}
	for _, good := range []string{`package nodeleafkeys; import "crypto/x509"; var _=x509.ParseCertificate`, `package nodeleafkeys; import cert "crypto/x509"; var _=cert.ParsePKIXPublicKey`, `package nodeleafkeys; func f(in Input){_ = in.Leaf.DER}`} {
		if !keysImportsAllowed([]byte(good)) {
			t.Fatal("alias/input-field control", good)
		}
	}
	for _, bad := range []string{
		`package nodeleafkeys; import . "crypto/x509"`, `package nodeleafkeys; import _ "crypto/x509"`, `package nodeleafkeys; import "net"`,
		`package nodeleafkeys; import a "bytes"; import b "bytes"`, `package nodeleafkeys; import a "bytes"; import a "errors"`,
		`package nodeleafkeys; import a "bytes"; func f(){a:=1;_ = a}`, `package nodeleafkeys; import a "bytes"; func f(a int){}`,
		`package nodeleafkeys; import x "crypto/x509"; var f=x.CreateCertificate`, `package nodeleafkeys; import x "crypto/x509"; var f=x.NewCertPool`,
		`package nodeleafkeys; import "time"; var f=time.Now`, `package nodeleafkeys; import n "talenro.local/platform/internal/nodecontrol/nodeleaf"; var f=n.Material{}`,
		`package nodeleafkeys; func f(c C){v:=c.Verify;_ = v}`, `package nodeleafkeys; func f(c C){v:=c.CheckSignatureFrom;_ = v}`,
		`package nodeleafkeys; func f(c C){v:=c.DER;_ = v}`, `package nodeleafkeys; func f(c C){c.Authorize()}`, `package nodeleafkeys; func f(){go f()}`,
	} {
		if keysImportsAllowed([]byte(bad)) {
			t.Fatal("bad capability accepted", bad)
		}
	}
}
func TestKeysCapabilityBoundary(t *testing.T) {
	sources := keysSources(t)
	if !keysCapabilityAllowed(sources) {
		t.Fatal("fixed API")
	}
	for _, extra := range []string{"\nvar hook func()\n", "\nfunc Derive(){}\n", "\nfunc(m *Material) DER() []byte{return nil}\n", "\nfunc(m *Material) MarshalJSON()([]byte,error){return nil,nil}\n"} {
		clone := map[string][]byte{}
		for p, b := range sources {
			clone[p] = b
		}
		clone["doc.go"] = append(bytes.Clone(sources["doc.go"]), extra...)
		if keysCapabilityAllowed(clone) {
			t.Fatal("API mutation accepted")
		}
	}
	source := sources["key_identifiers.go"]
	if !keysIDOrder(source) {
		t.Fatal("identifier order")
	}
	text := string(source)
	for _, bad := range []string{
		strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(text, "SubjectKeyId", "TEMP"), "AuthorityKeyId", "SubjectKeyId"), "TEMP", "AuthorityKeyId"),
		strings.ReplaceAll(text, "len(cert.SubjectKeyId) != 32", "len(cert.SubjectKeyId) != 20"),
	} {
		if keysIDOrder([]byte(bad)) {
			t.Fatal("ID order mutation accepted")
		}
	}
}
func TestKeysConsumers(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", "..", ".."))
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == ".superpowers" || d.Name() == "vendor" {
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
		for _, im := range f.Imports {
			p, e := strconv.Unquote(im.Path.Value)
			if e != nil {
				return e
			}
			if !keysConsumerAllowed(path, f.Name.Name, p) {
				t.Errorf("unauthorized consumer %s", path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	exact := filepath.Join(root, "internal", "nodecontrol", "nodeleafkeys")
	abs, e := filepath.Abs(filepath.Join(exact, "a.go"))
	if e != nil {
		t.Fatal(e)
	}
	for _, tc := range []struct {
		path, pkg   string
		lower, self bool
	}{
		{filepath.Join(exact, "a.go"), "nodeleafkeys", true, false},
		{abs, "nodeleafkeys", true, false},
		{filepath.Join(exact, "a_test.go"), "nodeleafkeys", true, true},
		{filepath.Join(exact, "a_test.go"), "nodeleafkeys_test", false, false},
		{filepath.Join(exact, "child", "a_test.go"), "nodeleafkeys", false, false},
		{filepath.Join(root, "internal", "nodecontrol", "nodeleafkeys_extra", "a.go"), "nodeleafkeys", false, false},
		{filepath.Join(root, "else", "internal", "nodecontrol", "nodeleafkeys", "a_test.go"), "nodeleafkeys", false, false},
		{filepath.Join(root, "..", "external", "internal", "nodecontrol", "nodeleafkeys", "a_test.go"), "nodeleafkeys", false, false},
		{"nodeleaf/a_test.go", "nodeleaf", true, false}, {"else/a_test.go", "nodeleaf_test", false, false},
	} {
		if keysConsumerAllowed(tc.path, tc.pkg, "talenro.local/platform/internal/nodecontrol/nodeleaf") != tc.lower || keysConsumerAllowed(tc.path, tc.pkg, "talenro.local/platform/internal/nodecontrol/nodeleafkeys") != tc.self {
			t.Fatal("consumer cases", tc.path)
		}
	}
}
func TestKeysCleanupRegistration(t *testing.T) {
	for _, path := range []string{"owned.go", "evaluate.go", "issuer_spki.go"} {
		b, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		if !keysCleanupRegistered(b) {
			t.Fatal("cleanup shape", path)
		}
		text := string(b)
		for _, bad := range []string{
			strings.ReplaceAll(text, "o.clear()", "other.clear()"),
			strings.ReplaceAll(text, "defer func()", "func()"),
			strings.ReplaceAll(text, "transferred := false", "transferred := true"),
			strings.ReplaceAll(text, "if !transferred", "if transferred"),
			strings.ReplaceAll(text, "o := &owner{}", "o := &owner{}\no.der = bytes.Clone(in.Leaf.DER)"),
			strings.ReplaceAll(text, "o := ownInput(in)", "o := ownInput(in)\ntransferred = true"),
			strings.ReplaceAll(text, "transferred = true", "transferred = true\ntransferred = true"),
			strings.ReplaceAll(text, "defer lower.Clear()", "if true {defer lower.Clear()}"),
			strings.ReplaceAll(text, "defer lower.Clear()", "defer other.Clear()"),
			strings.ReplaceAll(text, "defer clear(canonical)", "defer clear(der)"),
			strings.ReplaceAll(text, "defer clear(canonical)", "if true {defer clear(canonical)}"),
		} {
			if bad != text && keysCleanupRegistered([]byte(bad)) {
				t.Fatal("cleanup mutation accepted", path)
			}
		}
	}
	good := `package nodeleafkeys
 func ownInput(in Input)*owner{o:=&owner{};transferred:=false;defer func(){if !transferred{o.clear()}}();o.der=bytes.Clone(in.Leaf.DER);o.expectedSPKI=bytes.Clone(in.Leaf.ExpectedSPKI);o.declaredIssuerSPKI=bytes.Clone(in.DeclaredIssuerSPKI);transferred=true;return o}`
	if !keysCleanupRegistered([]byte(good)) {
		t.Fatal("memory control")
	}
	for _, bad := range []string{
		strings.ReplaceAll(good, "defer func(){if !transferred{o.clear()}}();", "if true {defer func(){if !transferred{o.clear()}}()};"),
		strings.ReplaceAll(good, "defer func(){if !transferred{o.clear()}}();o.der=bytes.Clone(in.Leaf.DER);", "o.der=bytes.Clone(in.Leaf.DER);defer func(){if !transferred{o.clear()}}();"),
	} {
		if keysCleanupRegistered([]byte(bad)) {
			t.Fatal("late/conditional cleanup accepted")
		}
	}
}
