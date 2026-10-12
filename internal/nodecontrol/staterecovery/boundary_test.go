package staterecovery

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

func recoveryImportsAllowed(src []byte) bool {
	f, e := parser.ParseFile(token.NewFileSet(), "", src, 0)
	if e != nil {
		return false
	}
	allowed := map[string]map[string]bool{
		"context":                {"Context": true, "Canceled": true, "DeadlineExceeded": true},
		"errors":                 {"New": true},
		"github.com/google/uuid": {"Parse": true, "Nil": true, "RFC4122": true},
		"talenro.local/platform/internal/nodecontrol/contracts": {"Digest": true},
	}
	aliases := map[string]map[string]bool{}
	for _, im := range f.Imports {
		p, e := strconv.Unquote(im.Path.Value)
		if e != nil {
			return false
		}
		members, ok := allowed[p]
		if !ok {
			return false
		}
		alias := p[strings.LastIndex(p, "/")+1:]
		if im.Name != nil {
			alias = im.Name.Name
		}
		if alias == "." || alias == "_" {
			return false
		}
		if _, exists := aliases[alias]; exists {
			return false
		}
		aliases[alias] = members
	}
	ok := true
	ast.Inspect(f, func(n ast.Node) bool {
		if sel, yes := n.(*ast.SelectorExpr); yes {
			if ident, yes := sel.X.(*ast.Ident); yes {
				if members, imported := aliases[ident.Name]; imported && !members[sel.Sel.Name] {
					ok = false
				}
			}
		}
		return true
	})
	return ok
}
func recoveryConsumerAllowed(importPath string) bool {
	return importPath != "talenro.local/platform/internal/nodecontrol/staterecovery"
}

func TestRecoveryBoundary(t *testing.T) {
	for _, src := range []string{
		`package p; import "os"`, `package p; import "time"`, `package p; import "net"`, `package p; import "unsafe"`, `package p; import "syscall"`, `package p; import "database/sql"`,
		`package p; import "talenro.local/platform/internal/nodecontrol/statepolicy"`, `package p; import "talenro.local/platform/internal/nodecontrol/statetime"`,
		`package p; import . "github.com/google/uuid"; var x=New`, `package p; import _ "context"`,
		`package p; import "github.com/google/uuid"; var x=uuid.NewRandom`, `package p; import "context"; var x=context.Background`,
		`not valid Go`,
	} {
		if recoveryImportsAllowed([]byte(src)) {
			t.Fatal("forbidden import/capability accepted", src)
		}
	}
	if !recoveryImportsAllowed([]byte(`package p; import("context"; "errors"; "github.com/google/uuid"; "talenro.local/platform/internal/nodecontrol/contracts"); var x context.Context; var y contracts.Digest; var z=errors.New("x"); var a=uuid.Parse; var b=uuid.Nil; var c=uuid.RFC4122`)) {
		t.Fatal("pure imports rejected")
	}
	if recoveryConsumerAllowed("talenro.local/platform/internal/nodecontrol/staterecovery") || !recoveryConsumerAllowed("context") {
		t.Fatal("consumer boundary")
	}
	files, e := filepath.Glob("*.go")
	if e != nil {
		t.Fatal(e)
	}
	for _, p := range files {
		if strings.HasSuffix(p, "_test.go") {
			continue
		}
		src, e := os.ReadFile(p)
		if e != nil {
			t.Fatal(e)
		}
		if !recoveryImportsAllowed(src) {
			t.Errorf("impure model %s", p)
		}
	}
	e = filepath.WalkDir("../../..", func(p string, d fs.DirEntry, e error) error {
		if e != nil {
			return e
		}
		if d.IsDir() {
			if d.Name() == ".git" || d.Name() == ".worktrees" || d.Name() == "vendor" {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return nil
		}
		f, e := parser.ParseFile(token.NewFileSet(), p, nil, parser.ImportsOnly)
		if e != nil {
			return e
		}
		for _, im := range f.Imports {
			path, e := strconv.Unquote(im.Path.Value)
			if e != nil {
				return e
			}
			if !recoveryConsumerAllowed(path) {
				t.Errorf("runtime consumer %s", p)
			}
		}
		return nil
	})
	if e != nil {
		t.Fatal(e)
	}
}
func TestRecoveryBoundaryAliases(t *testing.T) {
	for _, src := range []string{
		`package p; import(a "github.com/google/uuid"; b "github.com/google/uuid"); var x=a.New; var y=b.Parse`,
		`package p; import(a "github.com/google/uuid"; b "github.com/google/uuid"); var x=a.Parse; var y=b.NewRandom()`,
		`package p; import u "github.com/google/uuid"; var x=u.NewUUID`,
		`package p; import(a "context"; b "context"); var x=a.WithCancel; var y b.Context`,
		`package p; import(a "time"; b "time"); var x=a.Now; var y b.Time`,
	} {
		if recoveryImportsAllowed([]byte(src)) {
			t.Fatal("alias bypass", src)
		}
	}
	if !recoveryImportsAllowed([]byte(`package p; import(a "github.com/google/uuid"; b "github.com/google/uuid"); var x=a.Parse; var y=b.Nil`)) {
		t.Fatal("pure alias rejected")
	}
}
