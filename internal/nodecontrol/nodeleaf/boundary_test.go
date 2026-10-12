package nodeleaf

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

func leafPrint(node ast.Node) string {
	var b bytes.Buffer
	if format.Node(&b, token.NewFileSet(), node) != nil {
		return ""
	}
	return b.String()
}

func leafImportsAllowed(source []byte) bool {
	file, e := parser.ParseFile(token.NewFileSet(), "memory.go", source, 0)
	if e != nil {
		return false
	}
	allowed := map[string]string{
		"bytes": "Clone Equal", "context": "Context", "errors": "New", "crypto/sha256": "Sum256", "crypto/ecdsa": "PublicKey", "crypto/elliptic": "P256",
		"crypto/x509": "Certificate ParseCertificate ParsePKIXPublicKey KeyUsageDigitalSignature ExtKeyUsageClientAuth", "encoding/asn1": "ObjectIdentifier Unmarshal RawValue", "time": "Time Hour",
		"github.com/google/uuid": "UUID Nil RFC4122", "talenro.local/platform/internal/nodecontrol/contracts": "Digest",
	}
	aliases := map[string]string{}
	paths := map[string]bool{}
	for _, im := range file.Imports {
		path, e := strconv.Unquote(im.Path.Value)
		if e != nil || paths[path] {
			return false
		}
		paths[path] = true
		members, ok := allowed[path]
		if !ok {
			return false
		}
		_ = members
		name := path[strings.LastIndex(path, "/")+1:]
		if im.Name != nil {
			name = im.Name.Name
		}
		if name == "." || name == "_" || aliases[name] != "" {
			return false
		}
		aliases[name] = path
	}
	ok := true
	definition := func(id *ast.Ident) {
		if id != nil && aliases[id.Name] != "" {
			ok = false
		}
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.GoStmt:
			ok = false
		case *ast.SelectorExpr:
			name := n.Sel.Name
			if name == "Verify" || strings.HasPrefix(name, "CheckSignature") || name == "State" || name == "Ready" || name == "Authorize" || name == "Commit" {
				ok = false
			}
			if id, is := n.X.(*ast.Ident); is && aliases[id.Name] != "" {
				members := strings.Fields(allowed[aliases[id.Name]])
				found := false
				for _, member := range members {
					if member == name {
						found = true
					}
				}
				if !found {
					ok = false
				}
			}
		case *ast.AssignStmt:
			if n.Tok == token.DEFINE {
				for _, v := range n.Lhs {
					if id, is := v.(*ast.Ident); is {
						definition(id)
					}
				}
			}
		case *ast.ValueSpec:
			for _, id := range n.Names {
				definition(id)
			}
		case *ast.TypeSpec:
			definition(n.Name)
		case *ast.FuncDecl:
			definition(n.Name)
		case *ast.Field:
			for _, id := range n.Names {
				definition(id)
			}
		}
		return true
	})
	return ok
}

func leafConsumerAllowed(path, pkg, imported string) bool {
	if imported != "talenro.local/platform/internal/nodecontrol/nodeleaf" || (strings.HasSuffix(filepath.ToSlash(path), "_test.go") && pkg == "nodeleaf") {
		return true
	}
	if pkg != "nodeleafkeys" {
		return false
	}
	root, err := filepath.Abs(filepath.Clean(filepath.Join("..", "..", "..")))
	if err != nil {
		return false
	}
	abs, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(root, abs)
	if err != nil {
		return false
	}
	return filepath.Dir(rel) == filepath.Join("internal", "nodecontrol", "nodeleafkeys")
}

func TestLeafProductImports(t *testing.T) {
	files, e := filepath.Glob("*.go")
	if e != nil {
		t.Fatal(e)
	}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		b, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		if !leafImportsAllowed(b) {
			t.Fatal("product boundary", path)
		}
	}
	for _, good := range []string{`package nodeleaf; import "crypto/x509"; var _ = x509.ParseCertificate`, `package nodeleaf; import cert "crypto/x509"; var _ = cert.ParseCertificate`} {
		if !leafImportsAllowed([]byte(good)) {
			t.Fatal("named alias false positive")
		}
	}
	for _, bad := range []string{
		`package nodeleaf; import . "crypto/x509"`, `package nodeleaf; import _ "crypto/x509"`, `package nodeleaf; import "net"`,
		`package nodeleaf; import x "crypto/x509"; import y "crypto/x509"`, `package nodeleaf; import x "bytes"; import x "errors"`,
		`package nodeleaf; import x "bytes"; func f(){x:=1;_ = x}`,
		`package nodeleaf; import x "crypto/x509"; var f=x.NewCertPool`, `package nodeleaf; import x "crypto/x509"; var f=x.CreateCertificate`,
		`package nodeleaf; import t "time"; var f=t.Now`, `package nodeleaf; func f(c C){v:=c.Verify;_ = v}`, `package nodeleaf; func f(c C){v:=c.CheckSignatureFrom;_ = v}`,
		`package nodeleaf; func f(c C){c.Ready()}`, `package nodeleaf; func f(){go f()}`,
	} {
		if leafImportsAllowed([]byte(bad)) {
			t.Fatal("guard failed to reject", bad)
		}
	}
}

func TestLeafConsumers(t *testing.T) {
	root := filepath.Clean(filepath.Join("..", "..", ".."))
	e := filepath.WalkDir(root, func(path string, d fs.DirEntry, e error) error {
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
		file, e := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if e != nil {
			return e
		}
		for _, im := range file.Imports {
			p, e := strconv.Unquote(im.Path.Value)
			if e != nil {
				return e
			}
			if !leafConsumerAllowed(path, file.Name.Name, p) {
				t.Errorf("unauthorized nodeleaf consumer: %s", path)
			}
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
	exact := filepath.Join(root, "internal", "nodecontrol", "nodeleafkeys")
	abs, err := filepath.Abs(filepath.Join(exact, "a.go"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		path, pkg string
		want      bool
	}{
		{filepath.Join(exact, "a.go"), "nodeleafkeys", true},
		{abs, "nodeleafkeys", true},
		{filepath.Join(exact, "a_test.go"), "nodeleafkeys", true},
		{filepath.Join(root, "internal", "nodecontrol", "nodeleafkeys_extra", "a.go"), "nodeleafkeys", false},
		{filepath.Join(exact, "child", "a.go"), "nodeleafkeys", false},
		{filepath.Join(root, "else", "internal", "nodecontrol", "nodeleafkeys", "a.go"), "nodeleafkeys", false},
		{filepath.Join(exact, "a_test.go"), "nodeleafkeys_test", false},
		{filepath.Join(root, "internal", "nodecontrol", "brother", "a_test.go"), "nodeleafkeys", false},
		{filepath.Join(root, "..", "external", "internal", "nodecontrol", "nodeleafkeys", "a_test.go"), "nodeleafkeys", false},
	} {
		if got := leafConsumerAllowed(tc.path, tc.pkg, "talenro.local/platform/internal/nodecontrol/nodeleaf"); got != tc.want {
			t.Errorf("new consumer guard: %s got %v want %v", tc.path, got, tc.want)
		}
	}
	for _, tc := range []struct {
		path, pkg string
		want      bool
	}{{"nodeleaf/a_test.go", "nodeleaf", true}, {"serving/a.go", "serving", false}, {"statebinding/a_test.go", "statebinding", false}, {"runtime/a.go", "runtime", false}, {"nodeleaf/a.go", "nodeleaf", false}, {"else/a_test.go", "nodeleaf_test", false}} {
		if got := leafConsumerAllowed(tc.path, tc.pkg, "talenro.local/platform/internal/nodecontrol/nodeleaf"); got != tc.want {
			t.Fatal("consumer guard", tc.path)
		}
	}
}

func TestLeafCapabilityBoundary(t *testing.T) {
	wantFiles := strings.Fields("doc.go input.go errors.go budget.go owned.go material.go der.go leaf_fields.go extensions.go parse.go fixtures_test.go errors_test.go budget_test.go owned_test.go material_test.go der_test.go leaf_fields_test.go extensions_test.go parse_test.go boundary_test.go")
	sort.Strings(wantFiles)
	files, e := filepath.Glob("*.go")
	if e != nil {
		t.Fatal(e)
	}
	sort.Strings(files)
	if strings.Join(files, ",") != strings.Join(wantFiles, ",") {
		t.Fatal("file inventory")
	}
	wantTypes := map[string]bool{"Input": true, "Proof": true, "Obligation": true, "Summary": true, "Material": true}
	wantConst := map[string]bool{}
	for _, name := range strings.Fields("None NodeLeafMaterialShapeValid UnknownObligation LeafSignatureAndIssuerTrust KeyIdentifierDerivationAndIssuerBinding TrustedCurrentTime IdentityAndAuthorityProvenance PrivateKeyPossession MainStateAndRuntimeBinding") {
		wantConst[name] = true
	}
	wantVars := map[string]bool{}
	for _, name := range strings.Fields("ErrInvalidInput ErrLimitExceeded ErrInvalidMaterial ErrUnsupportedLeafProfile ErrBindingMismatch") {
		wantVars[name] = true
	}
	wantFunc := map[string]bool{"Parse": true, "Material.Summary": true, "Material.DER": true, "Material.Clear": true}
	structs := map[string]string{
		"Input":    "DER:[]byte,ExpectedSPKI:[]byte,NodeID:uuid.UUID,TrustDomain:string",
		"Summary":  "LeafDERDigest:contracts.Digest,NodeID:uuid.UUID,NotAfter:time.Time,NotBefore:time.Time,Pending:[6]Obligation,Proof:Proof,SPKIDERDigest:contracts.Digest,Serial:[20]byte,SerialLength:uint8,TrustDomain:string",
		"Material": "owned:*owner", "owner": "der:[]byte,expectedSPKI:[]byte,summary:Summary,valid:bool",
	}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		f, e := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if e != nil {
			t.Fatal(e)
		}
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.GenDecl:
				for _, sp := range d.Specs {
					switch s := sp.(type) {
					case *ast.TypeSpec:
						if s.Name.IsExported() {
							if !wantTypes[s.Name.Name] {
								t.Fatal("extra exported type", s.Name.Name)
							}
							delete(wantTypes, s.Name.Name)
						}
						if expected, exists := structs[s.Name.Name]; exists {
							st, ok := s.Type.(*ast.StructType)
							if !ok {
								t.Fatal("struct changed")
							}
							var fields []string
							for _, field := range st.Fields.List {
								if len(field.Names) == 0 {
									t.Fatal("embedded capability")
								}
								for _, name := range field.Names {
									fields = append(fields, name.Name+":"+leafPrint(field.Type))
								}
							}
							sort.Strings(fields)
							if strings.Join(fields, ",") != expected {
								t.Fatal("owner or fields changed", s.Name.Name)
							}
							delete(structs, s.Name.Name)
						}
					case *ast.ValueSpec:
						for _, name := range s.Names {
							if d.Tok == token.VAR {
								if !wantVars[name.Name] {
									t.Fatal("extra package mutable value", name.Name)
								}
								delete(wantVars, name.Name)
							} else if name.IsExported() {
								if !wantConst[name.Name] {
									t.Fatal("extra exported constant")
								}
								delete(wantConst, name.Name)
							}
						}
					}
				}
			case *ast.FuncDecl:
				if d.Name.IsExported() {
					name := d.Name.Name
					if d.Recv != nil {
						name = strings.TrimPrefix(leafPrint(d.Recv.List[0].Type), "*") + "." + name
					}
					if !wantFunc[name] {
						t.Fatal("extra exported function", name)
					}
					delete(wantFunc, name)
				}
			}
		}
	}
	if len(wantTypes)+len(wantConst)+len(wantVars)+len(wantFunc)+len(structs) != 0 {
		t.Fatal("missing fixed API")
	}
}

func leafCleanupRegistered(source []byte) bool {
	f, e := parser.ParseFile(token.NewFileSet(), "memory.go", source, 0)
	if e != nil {
		return false
	}
	found := false
	for _, decl := range f.Decls {
		fn, ok := decl.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || (fn.Name.Name != "Parse" && fn.Name.Name != "ownInput") {
			continue
		}
		found = true
		body := fn.Body.List
		own := fn.Name.Name == "ownInput"
		start := 1
		if own {
			start = 0
		}
		if len(body) < start+5 {
			return false
		}
		prefix := leafPrint(body[start])
		if (own && prefix != "o := &owner{}") || (!own && prefix != "o := ownInput(in)") {
			return false
		}
		if !own {
			if !strings.Contains(leafPrint(body[0]), "preflight(ctx, in)") {
				return false
			}
		}
		if leafPrint(body[start+1]) != "transferred := false" {
			return false
		}
		d, ok := body[start+2].(*ast.DeferStmt)
		if !ok || len(d.Call.Args) != 0 {
			return false
		}
		lit, ok := d.Call.Fun.(*ast.FuncLit)
		if !ok || len(lit.Body.List) != 1 {
			return false
		}
		guard, ok := lit.Body.List[0].(*ast.IfStmt)
		if !ok || guard.Init != nil || guard.Else != nil || leafPrint(guard.Cond) != "!transferred" || len(guard.Body.List) != 1 || leafPrint(guard.Body.List[0]) != "o.clear()" {
			return false
		}
		last := len(body) - 1
		if leafPrint(body[last-1]) != "transferred = true" {
			return false
		}
		if own {
			if leafPrint(body[last]) != "return o" || last != 6 || leafPrint(body[3]) != "o.der = bytes.Clone(in.DER)" || leafPrint(body[4]) != "o.expectedSPKI = bytes.Clone(in.ExpectedSPKI)" {
				return false
			}
		} else {
			if leafPrint(body[last]) != "return &Material{owned: o}, nil" || leafPrint(body[last-2]) != "o.valid = true" || !strings.Contains(leafPrint(body[last-3]), "ctx.Err()") {
				return false
			}
		}
		transfers, valids := 0, 0
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			if a, ok := n.(*ast.AssignStmt); ok {
				if leafPrint(a) == "transferred = true" {
					transfers++
				}
				if leafPrint(a) == "o.valid = true" {
					valids++
				}
			}
			return true
		})
		if transfers != 1 || (!own && valids != 1) {
			return false
		}
	}
	return found
}

func TestLeafCleanupRegistration(t *testing.T) {
	for _, path := range []string{"owned.go", "parse.go"} {
		b, e := os.ReadFile(path)
		if e != nil {
			t.Fatal(e)
		}
		if !leafCleanupRegistered(b) {
			t.Fatal("cleanup registration", path)
		}
		// AST guards use in-memory variants only; real source is never broken.
		text := string(b)
		for _, bad := range []string{
			strings.ReplaceAll(text, "o.clear()", "other.clear()"),
			strings.ReplaceAll(text, "defer func()", "func()"),
			strings.ReplaceAll(text, "transferred := false", "transferred := true"),
			strings.ReplaceAll(text, "if !transferred", "if transferred"),
			strings.ReplaceAll(text, "o := ownInput(in)", "o := ownInput(in)\ntransferred = true"),
			strings.ReplaceAll(text, "o := &owner{}", "o := &owner{}\no.der = bytes.Clone(in.DER)"),
			strings.ReplaceAll(text, "transferred = true", "transferred = true\ntransferred = true"),
		} {
			if bad != text && leafCleanupRegistered([]byte(bad)) {
				t.Fatal("cleanup mutation accepted", path)
			}
		}
	}
	good := []byte("package nodeleaf\nfunc ownInput(in Input)*owner {o:=&owner{};transferred:=false;defer func(){if !transferred{o.clear()}}();o.der=bytes.Clone(in.DER);o.expectedSPKI=bytes.Clone(in.ExpectedSPKI);transferred=true;return o}")
	if !leafCleanupRegistered(good) {
		t.Fatal("valid memory control")
	}
	for _, bad := range []string{
		strings.ReplaceAll(string(good), "defer func(){if !transferred{o.clear()}}();", "if true {defer func(){if !transferred{o.clear()}}()};"),
		strings.ReplaceAll(string(good), "defer func(){if !transferred{o.clear()}}();o.der=bytes.Clone(in.DER);", "o.der=bytes.Clone(in.DER);defer func(){if !transferred{o.clear()}}();"),
	} {
		if leafCleanupRegistered([]byte(bad)) {
			t.Fatal("conditional/late protection accepted")
		}
	}
}
