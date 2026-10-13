package nodeissuer

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func issuerSources(t testing.TB) map[string][]byte {
	t.Helper()
	entries, e := os.ReadDir(".")
	if e != nil {
		t.Fatal(e)
	}
	sources := map[string][]byte{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") {
			t.Fatal("extra inventory", entry.Name())
		}
		b, e := os.ReadFile(entry.Name())
		if e != nil {
			t.Fatal(e)
		}
		sources[entry.Name()] = b
	}
	return sources
}
func issuerMutation(s map[string][]byte, file, old, new string) map[string][]byte {
	out := map[string][]byte{}
	for k, v := range s {
		out[k] = bytes.Clone(v)
	}
	var pattern strings.Builder
	for _, c := range old {
		if c == ' ' || c == '\t' || c == '\n' {
			continue
		}
		pattern.WriteString(regexp.QuoteMeta(string(c)))
		pattern.WriteString("\\s*")
	}
	re := regexp.MustCompile(strings.TrimSuffix(pattern.String(), "\\s*"))
	text := string(out[file])
	loc := re.FindStringIndex(text)
	if loc == nil {
		return out
	}
	out[file] = []byte(text[:loc[0]] + new + text[loc[1]:])
	return out
}

// Catches imports/members, concealed selectors and callback/alias bypasses.
func TestIssuerProductBoundary(t *testing.T) {
	for name, b := range issuerSources(t) {
		if !strings.HasSuffix(name, "_test.go") && !issuerProductAllowed(b) {
			t.Fatal("product boundary", name)
		}
	}
	for _, good := range []string{`package nodeissuer; import cert "crypto/x509"; func f(b []byte){_,_=cert.ParseCertificate(b)}`, `package nodeissuer; import b "bytes"; func f(x,y []byte)bool{return b.Equal(x,y)}`} {
		if !issuerProductAllowed([]byte(good)) {
			t.Fatal("named alias false positive")
		}
	}
	for _, bad := range []string{
		`package nodeissuer;import . "crypto/x509"`, `package nodeissuer;import _ "crypto/x509"`, `package nodeissuer;import "net"`, `package nodeissuer;import t "time";func f(){t.Now()}`,
		`package nodeissuer;import c "crypto/x509";func f(){c.Verify()}`, `package nodeissuer;import c "crypto/x509";func f(){c.CheckSignatureFrom()}`,
		`package nodeissuer;import c "crypto/x509";func f(){c.MarshalPKIXPublicKey(nil)}`, `package nodeissuer;import e "crypto/ecdsa";func f(){v:=e.VerifyASN1;_=v}`,
		`package nodeissuer;func f(c C){v:=c.VerifyASN1;_=v}`, `package nodeissuer;func f(c C){c.CheckSignatureFrom(nil)}`, `package nodeissuer;func f(c C){c.DER()}`, `package nodeissuer;func f(c C){c.Commit()}`,
		`package nodeissuer;import x "bytes";import y "bytes"`, `package nodeissuer;import x "bytes";import x "errors"`,
		`package nodeissuer;import x "bytes";func f(){x:=1;_=x}`, `package nodeissuer;import x "bytes";func f(){for x:=range []int{} {_=x}}`,
		`package nodeissuer;import x "bytes";type x int`, `package nodeissuer;import x "bytes";type T struct{x int}`,
		`package nodeissuer;func f(){go f()}`, `package nodeissuer;func f(){recover()}`,
	} {
		if issuerProductAllowed([]byte(bad)) {
			t.Fatal("product bypass", bad)
		}
	}
}

// Catches expanding the public capability or retaining parser/key/mutable state.
func TestIssuerCapabilityBoundary(t *testing.T) {
	sources := issuerSources(t)
	if !issuerCapabilityAllowed(sources) {
		t.Fatal("normal capabilities")
	}
	controls := [][3]string{{"input.go", "type Input struct", "type Extra struct{ X bool }; type Input struct"}, {"input.go", "CA nodeca.Input", "CA nodeca.Input; Trusted bool"}, {"input.go", "type Proof uint8", "type Proof uint16"}, {"input.go", "NodeLeafDirectIssuerSignatureBound", "NodeLeafDirectIssuerSignatureBound = 2"}, {"material.go", "type Material struct{owned *owner}", "type Material struct{owned *owner; cert *x509.Certificate}"}, {"owned.go", "summary Summary", "summary Summary; cert *x509.Certificate"}, {"errors.go", "node issuer: invalid input", "node issuer: dynamic input"}, {"input.go", "node-leaf-direct-issuer.p256-sha256.v1", "changed-profile"}, {"material.go", "func(m *Material)Clear()", "func(m *Material)DER()[]byte{return nil};func(m *Material)Clear()"}, {"input.go", "type Input struct", "var hook func();type Input struct"}}
	for _, c := range controls {
		mutated := issuerMutation(sources, c[0], c[1], c[2])
		if bytes.Equal(mutated[c[0]], sources[c[0]]) {
			t.Fatal("inactive capability control", c)
		}
		if issuerCapabilityAllowed(mutated) {
			t.Fatal("capability bypass", c)
		}
	}
	extra := map[string][]byte{}
	for k, v := range sources {
		extra[k] = v
	}
	extra["input.go"] = append(bytes.Clone(sources["input.go"]), []byte("\nvar injected = 1\ntype Extra struct{}")...)
	if issuerCapabilityAllowed(extra) {
		t.Fatal("extra capability")
	}
	extra = map[string][]byte{}
	for k, v := range sources {
		extra[k] = v
	}
	extra["input.go"] = append(bytes.Clone(sources["input.go"]), []byte("\nfunc Backdoor(in Summary) {}")...)
	if issuerCapabilityAllowed(extra) {
		t.Fatal("summary capability")
	}
	extra = map[string][]byte{}
	for k, v := range sources {
		extra[k] = v
	}
	extra["extra.go"] = []byte("package nodeissuer")
	if issuerCapabilityAllowed(extra) {
		t.Fatal("extra file")
	}
}
func TestIssuerConsumers(t *testing.T) {
	root := filepath.Clean("../../..")
	exact := filepath.Join(root, "internal", "nodecontrol", "nodeissuer")
	abs, e := filepath.Abs(filepath.Join(exact, "a_test.go"))
	if e != nil {
		t.Fatal(e)
	}
	target := "talenro.local/platform/internal/nodecontrol/nodeissuer"
	for _, c := range []struct {
		path, pkg string
		want      bool
	}{{filepath.Join(exact, "a_test.go"), "nodeissuer", true}, {abs, "nodeissuer", true}, {filepath.Join(exact, "a.go"), "nodeissuer", false}, {filepath.Join(exact, "a_test.go"), "nodeissuer_test", false}, {filepath.Join(exact, "a_test.go"), "other", false}, {filepath.Join(exact, "nested", "a_test.go"), "nodeissuer", false}, {filepath.Join(root, "internal", "nodecontrol", "nodeissuer_extra", "a_test.go"), "nodeissuer", false}, {filepath.Join(root, "else", "internal", "nodecontrol", "nodeissuer", "a_test.go"), "nodeissuer", false}, {filepath.Join(root, "..", "external", "internal", "nodecontrol", "nodeissuer", "a_test.go"), "nodeissuer", false}, {filepath.Join(root, "internal", "brother", "a_test.go"), "nodeissuer", false}} {
		if issuerConsumerAllowed(c.path, c.pkg, target) != c.want {
			t.Fatal("consumer control", c)
		}
	}
	if issuerConsumerAllowed(abs, "nodeissuer", target+"/child") {
		t.Fatal("subpath")
	}
	e = filepath.WalkDir(root, func(path string, entry fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if entry.IsDir() {
			switch entry.Name() {
			case ".git", ".superpowers", ".worktrees", "vendor", "node_modules":
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
			if !issuerConsumerAllowed(path, f.Name.Name, p) {
				t.Errorf("unauthorized consumer %s", path)
			}
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
}

// Catches late/conditional cleanup, wrong Clear target, early handoff and use of
// borrowed input after ownership. Rejection controls operate on memory only.
func TestIssuerOwnershipFlow(t *testing.T) {
	sources := issuerSources(t)
	if !issuerOwnershipAllowed(sources) {
		t.Fatal("normal ownership")
	}
	for _, c := range [][3]string{
		{"owned.go", "o:=&owner{}", "o:=&owner{};o.leafDER=bytes.Clone(in.Leaf.DER)"},
		{"owned.go", "if !transferred", "if transferred"}, {"owned.go", "o.clear()", "other.clear()"}, {"owned.go", "transferred:=false", "transferred:=true"},
		{"evaluate.go", "o:=ownInput(in)", "o:=ownInput(in);transferred=true"}, {"evaluate.go", "if !transferred", "if transferred"}, {"evaluate.go", "o.valid=true", "o.valid=true;transferred=true"},
		{"evaluate.go", "parseLeafEnvelope(owned.Leaf.DER)", "parseLeafEnvelope(in.Leaf.DER)"}, {"evaluate.go", "verifyLeafSignature(certs.ca,raw)", "verifyLeafSignature(certs.leaf,raw)"},
		{"lower.go", "defer ca.Clear()", "defer leaf.Clear()"}, {"lower.go", "defer leaf.Clear()", "if true{defer leaf.Clear()}"},
	} {
		mutated := issuerMutation(sources, c[0], c[1], c[2])
		if bytes.Equal(mutated[c[0]], sources[c[0]]) {
			t.Fatal("inactive ownership control", c)
		}
		if issuerOwnershipAllowed(mutated) {
			t.Fatal("ownership bypass", c)
		}
	}
	// Whitespace-independent parsed mutations remain active after final gofmt.
	for _, file := range []string{"owned.go", "evaluate.go", "lower.go"} {
		f, e := parser.ParseFile(token.NewFileSet(), file, sources[file], 0)
		if e != nil {
			t.Fatal(e)
		}
		changed := false
		ast.Inspect(f, func(n ast.Node) bool {
			if d, ok := n.(*ast.DeferStmt); ok {
				if call, ok := d.Call.Fun.(*ast.SelectorExpr); ok {
					call.Sel.Name = "Summary"
				} else {
					d.Call.Fun = ast.NewIdent("missingCleanup")
				}
				changed = true
			}
			return true
		})
		if !changed {
			t.Fatal("no cleanup control")
		}
		var b bytes.Buffer
		if format.Node(&b, token.NewFileSet(), f) != nil {
			t.Fatal("mutation format")
		}
		mutated := map[string][]byte{}
		for k, v := range sources {
			mutated[k] = v
		}
		mutated[file] = b.Bytes()
		if issuerOwnershipAllowed(mutated) {
			t.Fatal("removed cleanup", file)
		}
	}
	f, e := parser.ParseFile(token.NewFileSet(), "evaluate.go", sources["evaluate.go"], 0)
	if e != nil {
		t.Fatal(e)
	}
	for _, d := range f.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == "Evaluate" {
			fn.Body.List = append(fn.Body.List[:len(fn.Body.List)-4], fn.Body.List[len(fn.Body.List)-3:]...)
		}
	}
	var b bytes.Buffer
	if e = format.Node(&b, token.NewFileSet(), f); e != nil {
		t.Fatal(e)
	}
	mutated := map[string][]byte{}
	for k, v := range sources {
		mutated[k] = v
	}
	mutated["evaluate.go"] = b.Bytes()
	if issuerOwnershipAllowed(mutated) {
		t.Fatal("no final ctx check")
	}
}

func issuerPrint(n ast.Node) string {
	var b bytes.Buffer
	if format.Node(&b, token.NewFileSet(), n) != nil {
		return ""
	}
	return b.String()
}
func TestIssuerLowerParseBoundary(t *testing.T) {
	for _, source := range []string{`package nodeissuer;import "talenro.local/platform/internal/nodecontrol/nodeca";func f(){_,_=nodeca.Parse(nil,nodeca.Input{})}`, `package nodeissuer;import ca "talenro.local/platform/internal/nodecontrol/nodeca";func f(){_,_=ca.Parse(nil,ca.Input{})}`} {
		if !issuerProductAllowed([]byte(source)) {
			t.Fatal("approved real CA Parse rejected")
		}
	}
}
func issuerShape(n ast.Node) string { return strings.Join(strings.Fields(issuerPrint(n)), "") }
func issuerImports() map[string]string {
	return map[string]string{
		"bytes": "Clone Equal Compare", "context": "Context", "errors": "New", "crypto/sha256": "Sum256", "crypto/ecdsa": "PublicKey VerifyASN1", "crypto/elliptic": "P256", "crypto/x509": "Certificate ParseCertificate ECDSAWithSHA256 KeyUsageCRLSign", "github.com/google/uuid": "UUID Nil RFC4122",
		"talenro.local/platform/internal/nodecontrol/contracts":    "Digest",
		"talenro.local/platform/internal/nodecontrol/nodeleaf":     "Input Summary Proof Obligation None NodeLeafMaterialShapeValid LeafSignatureAndIssuerTrust KeyIdentifierDerivationAndIssuerBinding TrustedCurrentTime IdentityAndAuthorityProvenance PrivateKeyPossession MainStateAndRuntimeBinding",
		"talenro.local/platform/internal/nodecontrol/nodeleafkeys": "Input Summary Proof None NodeLeafDeclaredKeyIdentifiersBound Evaluate",
		"talenro.local/platform/internal/nodecontrol/nodeca":       "Input Summary Proof Obligation None NodeCAMaterialShapeAndDeclaredKeyBound Parse CertificateSignatureAndIssuerChain AuthorityIdentifierAndParentBinding TrustedCurrentTime DeploymentAuthorizationAndPurpose BundleStatusAndRollback LeafIssuerRelationAndRuntimeUse",
	}
}
func issuerContains(list, name string) bool {
	for _, v := range strings.Fields(list) {
		if v == name {
			return true
		}
	}
	return false
}
func issuerProductAllowed(source []byte) bool {
	f, e := parser.ParseFile(token.NewFileSet(), "memory.go", source, 0)
	if e != nil || f.Name.Name != "nodeissuer" {
		return false
	}
	allowed := issuerImports()
	aliases := map[string]string{}
	paths := map[string]bool{}
	for _, im := range f.Imports {
		p, e := strconv.Unquote(im.Path.Value)
		if e != nil || paths[p] || allowed[p] == "" {
			return false
		}
		paths[p] = true
		name := filepath.Base(p)
		if im.Name != nil {
			name = im.Name.Name
		}
		if name == "." || name == "_" || aliases[name] != "" {
			return false
		}
		aliases[name] = p
	}
	calls := map[*ast.SelectorExpr]bool{}
	defers := map[*ast.FuncLit]bool{}
	ast.Inspect(f, func(n ast.Node) bool {
		if c, ok := n.(*ast.CallExpr); ok {
			if s, ok := c.Fun.(*ast.SelectorExpr); ok {
				calls[s] = true
			}
		}
		if d, ok := n.(*ast.DeferStmt); ok {
			if lit, ok := d.Call.Fun.(*ast.FuncLit); ok {
				defers[lit] = true
			}
		}
		return true
	})
	valid := true
	definition := func(id *ast.Ident) {
		if id != nil && aliases[id.Name] != "" {
			valid = false
		}
	}
	fields := "Leaf CA DER ExpectedSPKI DeclaredSPKI NodeID TrustDomain Proof Pending LeafDERDigest SPKIDERDigest CADERDigest SubjectDERDigest IssuerDERDigest DeclaredIssuerSPKIDERDigest SubjectKeyIdentifier AuthorityKeyIdentifier Serial SerialLength NotBefore NotAfter HasPathLen PathLen CRLSignAllowed LeafTBSDERDigest LeafIssuerDERDigest Raw RawTBSCertificate RawIssuer RawSubject RawSubjectPublicKeyInfo SubjectKeyId AuthorityKeyId SerialNumber MaxPathLen MaxPathLenZero KeyUsage SignatureAlgorithm Signature PublicKey Curve X Y leafDER expectedSPKI caDER declaredSPKI summary valid owned ca leaf tbs issuer subject spki innerAlgorithm outerAlgorithm signature tag full content"
	ast.Inspect(f, func(n ast.Node) bool {
		switch v := n.(type) {
		case *ast.GoStmt, *ast.SendStmt, *ast.ChanType:
			valid = false
		case *ast.FuncLit:
			if !defers[v] || v.Type.Params.NumFields() != 0 || v.Type.Results != nil {
				valid = false
			}
		case *ast.CallExpr:
			if id, ok := v.Fun.(*ast.Ident); ok && (id.Name == "recover" || id.Name == "panic" || id.Name == "new") {
				valid = false
			}
		case *ast.SelectorExpr:
			name := v.Sel.Name
			if id, ok := v.X.(*ast.Ident); ok && aliases[id.Name] != "" {
				p := aliases[id.Name]
				if !issuerContains(allowed[p], name) {
					valid = false
				}
				if name == "VerifyASN1" {
					inside := false
					for _, d := range f.Decls {
						if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == "verifyLeafSignature" && v.Pos() > fn.Pos() && v.End() < fn.End() {
							inside = true
						}
					}
					if p != "crypto/ecdsa" || !calls[v] || !inside {
						valid = false
					}
				}
			} else if calls[v] || !issuerContains(fields, name) {
				receiver := issuerShape(v.X)
				ok := false
				switch name {
				case "Err":
					ok = receiver == "ctx"
				case "Summary", "Clear":
					ok = receiver == "ca" || receiver == "leaf"
				case "Sign", "Bytes":
					ok = receiver == "cert.SerialNumber"
				case "UTC":
					ok = receiver == "ca.NotBefore" || receiver == "ca.NotAfter" || receiver == "leaf.NotBefore" || receiver == "leaf.NotAfter"
				case "Equal":
					ok = receiver == "ca.NotBefore.UTC()" || receiver == "ca.NotAfter.UTC()" || receiver == "leaf.NotBefore.UTC()" || receiver == "leaf.NotAfter.UTC()"
				case "IsOnCurve":
					ok = receiver == "key.Curve"
				case "Variant":
					ok = receiver == "in.Leaf.NodeID"
				case "input":
					ok = receiver == "o"
				case "clear":
					ok = receiver == "o" || receiver == "m.owned"
				}
				if !ok || !calls[v] {
					valid = false
				}
			}
		case *ast.AssignStmt:
			if v.Tok == token.DEFINE {
				for _, x := range v.Lhs {
					if id, ok := x.(*ast.Ident); ok {
						definition(id)
					}
				}
			}
		case *ast.RangeStmt:
			if v.Tok == token.DEFINE {
				if id, ok := v.Key.(*ast.Ident); ok {
					definition(id)
				}
				if id, ok := v.Value.(*ast.Ident); ok {
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
			if _, ok := v.Type.(*ast.FuncType); ok {
				valid = false
			}
		}
		return true
	})
	return valid
}
func issuerCanonical(source []byte) (*ast.File, bool) {
	f, e := parser.ParseFile(token.NewFileSet(), "memory.go", source, 0)
	if e != nil || !issuerProductAllowed(source) {
		return nil, false
	}
	aliases := map[string]string{}
	for _, im := range f.Imports {
		p, _ := strconv.Unquote(im.Path.Value)
		name := filepath.Base(p)
		if im.Name != nil {
			name = im.Name.Name
		}
		aliases[name] = filepath.Base(p)
	}
	ast.Inspect(f, func(n ast.Node) bool {
		if s, ok := n.(*ast.SelectorExpr); ok {
			if id, ok := s.X.(*ast.Ident); ok && aliases[id.Name] != "" {
				id.Name = aliases[id.Name]
			}
		}
		return true
	})
	return f, true
}
func issuerFields(s *ast.StructType) string {
	var fields []string
	for _, f := range s.Fields.List {
		if f.Tag != nil || len(f.Names) == 0 {
			return "invalid"
		}
		for _, n := range f.Names {
			fields = append(fields, n.Name+":"+issuerShape(f.Type))
		}
	}
	sort.Strings(fields)
	return strings.Join(fields, ",")
}
func issuerCapabilityAllowed(sources map[string][]byte) bool {
	want := strings.Fields("doc.go input.go errors.go budget.go owned.go material.go lower.go snapshot.go der.go signature.go relation.go evaluate.go fixtures_test.go lower_test.go budget_test.go owned_test.go material_test.go snapshot_test.go boundary_test.go fuzz_test.go der_test.go signature_test.go relation_test.go evaluate_test.go")
	if len(sources) != len(want) {
		return false
	}
	for _, n := range want {
		if sources[n] == nil {
			return false
		}
	}
	types := map[string]string{
		"Input": "CA:nodeca.Input,Leaf:nodeleaf.Input", "Proof": "uint8", "Summary": "CA:nodeca.Summary,Leaf:nodeleafkeys.Summary,LeafIssuerDERDigest:contracts.Digest,LeafTBSDERDigest:contracts.Digest,Proof:Proof",
		"Material": "owned:*owner", "owner": "caDER:[]byte,declaredSPKI:[]byte,expectedSPKI:[]byte,leafDER:[]byte,summary:Summary,valid:bool",
		"lowerSummaries": "ca:nodeca.Summary,leaf:nodeleafkeys.Summary", "certificateSnapshots": "ca:*x509.Certificate,leaf:*x509.Certificate", "derElement": "content:[]byte,full:[]byte,tag:byte",
		"signedLeaf": "innerAlgorithm:derElement,issuer:derElement,outerAlgorithm:derElement,signature:derElement,spki:derElement,subject:derElement,tbs:derElement",
	}
	errors := map[string]string{"ErrInvalidInput": "node issuer: invalid input", "ErrLimitExceeded": "node issuer: limit exceeded", "ErrInvalidMaterial": "node issuer: invalid material", "ErrUnsupportedSignatureProfile": "node issuer: unsupported leaf signature profile", "ErrIssuerRelationMismatch": "node issuer: issuer relation mismatch", "ErrInvalidSignature": "node issuer: invalid leaf signature"}
	constants := map[string]bool{"None": false, "NodeLeafDirectIssuerSignatureBound": false, "issuerRelationProfile": false}
	exported := map[string]string{"Evaluate": "func(ctxcontext.Context,inInput)(*Material,error)", "Material.Summary": "func()Summary", "Material.Clear": "func()"}
	private := "preflight validDomain lowerDigit ownInput owner.input owner.clear stageError checkCAProof checkLowerProofs evaluateLower serialSnapshotMatches parseSnapshots readElement singleElement takeElement minimalInteger minimalOID parseLeafEnvelope checkLeafSignatureProfile signatureIntegerInRange checkSignatureDER verifyLeafSignature checkIssuerRelation"
	seenFunctions := map[string]bool{}
	for name, b := range sources {
		if strings.HasSuffix(name, "_test.go") {
			continue
		}
		f, ok := issuerCanonical(b)
		if !ok {
			return false
		}
		for _, d := range f.Decls {
			switch v := d.(type) {
			case *ast.GenDecl:
				if v.Tok == token.IMPORT {
					continue
				}
				for _, sp := range v.Specs {
					switch s := sp.(type) {
					case *ast.TypeSpec:
						expected, ok := types[s.Name.Name]
						if !ok || s.Assign.IsValid() {
							return false
						}
						actual := issuerShape(s.Type)
						if st, ok := s.Type.(*ast.StructType); ok {
							actual = issuerFields(st)
						}
						if actual != expected {
							return false
						}
						delete(types, s.Name.Name)
					case *ast.ValueSpec:
						if len(s.Names) != 1 {
							return false
						}
						n := s.Names[0].Name
						if v.Tok == token.VAR {
							expected, ok := errors[n]
							if !ok || s.Type != nil || len(s.Values) != 1 {
								return false
							}
							call, ok := s.Values[0].(*ast.CallExpr)
							if !ok || issuerShape(call.Fun) != "errors.New" || len(call.Args) != 1 {
								return false
							}
							lit, ok := call.Args[0].(*ast.BasicLit)
							if !ok {
								return false
							}
							str, e := strconv.Unquote(lit.Value)
							if e != nil || str != expected {
								return false
							}
							delete(errors, n)
						} else if v.Tok == token.CONST {
							already, ok := constants[n]
							if !ok || already {
								return false
							}
							constants[n] = true
							switch n {
							case "None":
								if issuerShape(s.Type) != "Proof" || len(s.Values) != 1 || issuerShape(s.Values[0]) != "iota" {
									return false
								}
							case "NodeLeafDirectIssuerSignatureBound":
								if s.Type != nil || len(s.Values) != 0 {
									return false
								}
							case "issuerRelationProfile":
								if len(s.Values) != 1 {
									return false
								}
								lit, ok := s.Values[0].(*ast.BasicLit)
								if !ok {
									return false
								}
								str, e := strconv.Unquote(lit.Value)
								if e != nil || str != "node-leaf-direct-issuer.p256-sha256.v1" {
									return false
								}
							}
						} else {
							return false
						}
					default:
						return false
					}
				}
			case *ast.FuncDecl:
				n := v.Name.Name
				if v.Recv != nil {
					if len(v.Recv.List) != 1 {
						return false
					}
					n = strings.TrimPrefix(issuerShape(v.Recv.List[0].Type), "*") + "." + n
				}
				if seenFunctions[n] {
					return false
				}
				seenFunctions[n] = true
				if v.Name.IsExported() {
					expected, ok := exported[n]
					if !ok || issuerShape(v.Type) != expected {
						return false
					}
					delete(exported, n)
				} else if !issuerContains(private, n) {
					return false
				}
				if n == "verifyLeafSignature" && name != "signature.go" {
					return false
				}
			default:
				return false
			}
		}
	}
	for _, seen := range constants {
		if !seen {
			return false
		}
	}
	return len(types)+len(errors)+len(exported) == 0
}
func TestNodeLeafBindingConsumerException(t *testing.T) {
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	dir := filepath.FromSlash("internal/nodecontrol/nodeleafbinding")
	target := "talenro.local/platform/internal/nodecontrol/nodeissuer"
	for _, c := range []struct {
		path, pkg string
		want      bool
	}{
		{filepath.Join(dir, "a.go"), "nodeleafbinding", true},
		{filepath.Join(root, dir, "a.go"), "nodeleafbinding", true},
		{filepath.Join(dir, "a_test.go"), "nodeleafbinding", true},
		{filepath.Join(root, dir, "a_test.go"), "nodeleafbinding", true},
		{filepath.Join(dir, "nested", "a.go"), "nodeleafbinding", false},
		{filepath.FromSlash("internal/nodecontrol/nodeleafbinding_extra/a.go"), "nodeleafbinding", false},
		{filepath.Join("else", dir, "a.go"), "nodeleafbinding", false},
		{filepath.Join(root, "..", "external", dir, "a.go"), "nodeleafbinding", false},
		{filepath.Join(dir, "a.go"), "other", false},
		{filepath.Join(dir, "a_test.go"), "nodeleafbinding_test", false},
		{"internal/nodecontrol/brother/a_test.go", "nodeleafbinding", false},
	} {
		if got := issuerConsumerAllowed(c.path, c.pkg, target); got != c.want {
			t.Errorf("%s %s: got %v want %v", c.path, c.pkg, got, c.want)
		}
	}
	if issuerConsumerAllowed(filepath.Join(dir, "a.go"), "nodeleafbinding", target+"/child") {
		t.Error("subpath")
	}
}

func TestNodeCertTimeConsumerException(t *testing.T) {
	root, e := filepath.Abs("../../..")
	if e != nil {
		t.Fatal(e)
	}
	dir := filepath.FromSlash("internal/nodecontrol/nodecerttime")
	target := "talenro.local/platform/internal/nodecontrol/nodeissuer"
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
		if got := issuerConsumerAllowed(c.path, c.pkg, target); got != c.want {
			t.Errorf("%s %s: got %v want %v", c.path, c.pkg, got, c.want)
		}
	}
	if issuerConsumerAllowed(filepath.Join(dir, "a.go"), "nodecerttime", target+"/child") {
		t.Error("subpath")
	}

}

func issuerConsumerAllowed(path, pkg, imported string) bool {
	// Only the exact selected-leaf projection consumer is added.
	leafBindingTarget := "talenro.local/platform/internal/nodecontrol/nodeissuer"
	if imported == leafBindingTarget || strings.HasPrefix(imported, leafBindingTarget+"/") {
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
		exact := e == nil && filepath.Dir(rel) == filepath.FromSlash("internal/nodecontrol/nodeleafbinding")
		if exact || pkg == "nodeleafbinding" || pkg == "nodeleafbinding_test" {
			return imported == leafBindingTarget && exact && pkg == "nodeleafbinding"
		}
	}
	// Only the new exact consumer is added; all previous rules below remain.
	certTimeTarget := "talenro.local/platform/internal/nodecontrol/nodeissuer"
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

	target := "talenro.local/platform/internal/nodecontrol/nodeissuer"
	if imported != target && !strings.HasPrefix(imported, target+"/") {
		return true
	}
	if imported != target || pkg != "nodeissuer" || !strings.HasSuffix(filepath.ToSlash(path), "_test.go") {
		return false
	}
	root, e := filepath.Abs("../../..")
	if e != nil {
		return false
	}
	abs, e := filepath.Abs(filepath.Clean(path))
	if e != nil {
		return false
	}
	rel, e := filepath.Rel(root, abs)
	return e == nil && filepath.Dir(rel) == filepath.Join("internal", "nodecontrol", "nodeissuer")
}
func issuerFunction(f *ast.File, name string) *ast.FuncDecl {
	for _, d := range f.Decls {
		if fn, ok := d.(*ast.FuncDecl); ok && fn.Name.Name == name {
			return fn
		}
	}
	return nil
}
func issuerStatements(fn *ast.FuncDecl) []string {
	if fn == nil || fn.Body == nil {
		return nil
	}
	var s []string
	for _, n := range fn.Body.List {
		s = append(s, issuerShape(n))
	}
	return s
}
func issuerOwnershipAllowed(sources map[string][]byte) bool {
	for _, name := range []string{"owned.go", "evaluate.go", "lower.go", "relation.go", "signature.go", "snapshot.go"} {
		if _, ok := issuerCanonical(sources[name]); !ok {
			return false
		}
	}
	owned, _ := issuerCanonical(sources["owned.go"])
	evaluate, _ := issuerCanonical(sources["evaluate.go"])
	lower, _ := issuerCanonical(sources["lower.go"])
	a := issuerStatements(issuerFunction(owned, "ownInput"))
	want := []string{"o:=&owner{}", "transferred:=false", "deferfunc(){if!transferred{o.clear()}}()", "o.leafDER=bytes.Clone(in.Leaf.DER)", "o.expectedSPKI=bytes.Clone(in.Leaf.ExpectedSPKI)", "o.caDER=bytes.Clone(in.CA.DER)", "o.declaredSPKI=bytes.Clone(in.CA.DeclaredSPKI)", "transferred=true", "returno"}
	if strings.Join(a, ";") != strings.Join(want, ";") {
		return false
	}
	fn := issuerFunction(evaluate, "Evaluate")
	a = issuerStatements(fn)
	if len(a) < 12 || a[0] != "iferr:=preflight(ctx,in);err!=nil{returnnil,err}" || a[1] != "o:=ownInput(in)" || a[2] != "transferred:=false" || a[3] != want[2] || a[4] != "iferr:=stageError(ctx,nil);err!=nil{returnnil,err}" || a[5] != "owned:=o.input(in.Leaf.NodeID,in.Leaf.TrustDomain)" {
		return false
	}
	stages := []string{"s,err:=evaluateLower(ctx,owned)", "certs,err:=parseSnapshots(ctx,owned,s)", "raw,err:=parseLeafEnvelope(owned.Leaf.DER)", "err=checkLeafSignatureProfile(certs.leaf,raw)", "err=checkIssuerRelation(owned,s,certs)", "err=verifyLeafSignature(certs.ca,raw)"}
	at := 6
	for i, s := range stages {
		if i > 0 {
			if a[at] != "iferr=stageError(ctx,nil);err!=nil{returnnil,err}" {
				return false
			}
			at++
		}
		if a[at] != s || a[at+1] != "iferr=stageError(ctx,err);err!=nil{returnnil,err}" {
			return false
		}
		at += 2
	}
	if len(a) != at+5 || a[at] != "o.summary=Summary{Proof:NodeLeafDirectIssuerSignatureBound,Leaf:s.leaf,CA:s.ca,LeafTBSDERDigest:sha256.Sum256(raw.tbs.full),LeafIssuerDERDigest:sha256.Sum256(raw.issuer.full)}" || a[at+1] != "iferr=stageError(ctx,nil);err!=nil{returnnil,err}" || a[at+2] != "o.valid=true" || a[at+3] != "transferred=true" || a[at+4] != "return&Material{owned:o},nil" {
		return false
	}
	l := issuerStatements(issuerFunction(lower, "evaluateLower"))
	ca, leaf := -1, -1
	for i, s := range l {
		if s == "ca,err:=nodeca.Parse(ctx,in.CA)" {
			ca = i
		}
		if s == "leaf,err:=nodeleafkeys.Evaluate(ctx,nodeleafkeys.Input{Leaf:in.Leaf,DeclaredIssuerSPKI:in.CA.DeclaredSPKI})" {
			leaf = i
		}
	}
	if ca < 0 || leaf <= ca || ca+1 >= len(l) || leaf+1 >= len(l) || l[ca+1] != "deferca.Clear()" || l[leaf+1] != "deferleaf.Clear()" {
		return false
	}
	for i, s := range l {
		if strings.HasPrefix(s, "defer") && i != ca+1 && i != leaf+1 {
			return false
		}
	}
	// Context checks surround each real lower result and proof description.
	if l[ca-1] != "iferr:=stageError(ctx,nil);err!=nil{returnout,err}" || l[ca+2] != "iferr=stageError(ctx,err);err!=nil{returnout,err}" || l[leaf-1] != "iferr=stageError(ctx,nil);err!=nil{returnout,err}" || l[leaf+2] != "iferr=stageError(ctx,err);err!=nil{returnout,err}" {
		return false
	}
	relation, _ := issuerCanonical(sources["relation.go"])
	r := issuerStatements(issuerFunction(relation, "checkIssuerRelation"))
	if len(r) != 6 || !strings.HasPrefix(r[1], "if!bytes.Equal(certs.ca.RawSubjectPublicKeyInfo,in.CA.DeclaredSPKI)") || !strings.HasPrefix(r[2], "if!bytes.Equal(certs.leaf.RawIssuer,certs.ca.RawSubject)") || r[3] != "id:=sha256.Sum256(certs.ca.RawSubjectPublicKeyInfo)" || !strings.HasPrefix(r[4], "iflen(certs.ca.SubjectKeyId)!=32||len(certs.leaf.AuthorityKeyId)!=32") {
		return false
	}
	signature, _ := issuerCanonical(sources["signature.go"])
	v := issuerStatements(issuerFunction(signature, "verifyLeafSignature"))
	if len(v) != 6 || v[1] != "key,ok:=ca.PublicKey.(*ecdsa.PublicKey)" || v[3] != "digest:=sha256.Sum256(raw.tbs.full)" || v[4] != "if!ecdsa.VerifyASN1(key,digest[:],raw.signature.content[1:]){returnErrInvalidSignature}" {
		return false
	}
	snapshot, _ := issuerCanonical(sources["snapshot.go"])
	v = issuerStatements(issuerFunction(snapshot, "serialSnapshotMatches"))
	if len(v) < 4 || v[1] != "b:=cert.SerialNumber.Bytes()" || v[2] != "deferclear(b)" {
		return false
	}
	return true
}
