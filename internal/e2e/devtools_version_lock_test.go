//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"
)

type devtoolsVersionRecord struct {
	Path, Version string
	Main          bool
	Error         json.RawMessage
}
type devtoolsVersionSnapshot struct {
	Owner    string
	Modules  []devtoolsVersionRecord
	Packages map[string][]devtoolsPackageRecord
}
type devtoolsVersionPolicy struct {
	SchemaVersion                              int
	Baseline                                   map[string]string
	Required, Exceptions, Membership, Evidence map[string]map[string]string
}

// Independent literal contract: changing policy.json must not expand approvals.
const devtoolsRequiredVersions = `tools golang.org/x/time v0.14.0
tools modernc.org/mathutil v1.7.1
tools modernc.org/sortutil v1.2.1
tools modernc.org/strutil v1.2.1
root cel.dev/expr v0.25.2
root cloud.google.com/go v0.121.2
root github.com/Azure/go-ansiterm v0.0.0-20250102033503-faa5f7b0171c
root github.com/BurntSushi/toml v1.6.0
root github.com/alecthomas/units v0.0.0-20240927000941-0f3dac36c52b
root github.com/ebitengine/purego v0.10.0
root github.com/gorilla/websocket v1.5.3
root github.com/hashicorp/go-version v1.9.0
root github.com/mattn/go-colorable v0.1.15
root github.com/moby/moby/api v1.55.0
root github.com/moby/moby/client v0.5.0
root github.com/moby/term v0.5.2
root github.com/pelletier/go-toml/v2 v2.3.1
root github.com/power-devops/perfstat v0.0.0-20240221224432-82ca36839d55
root github.com/shirou/gopsutil/v4 v4.26.4
root github.com/sirupsen/logrus v1.9.4
root github.com/tklauser/go-sysconf v0.3.16
root github.com/tklauser/numcpus v0.11.0
root go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.69.0
root go.uber.org/zap v1.28.0
root golang.org/x/lint v0.0.0-20190930215403-16217165b5de
root golang.org/x/oauth2 v0.36.0
root golang.org/x/xerrors v0.0.0-20220517211312-f3a8303e98df
root google.golang.org/appengine v1.6.7
root google.golang.org/genproto v0.0.0-20220519153652-3a47de7e79bd
root google.golang.org/genproto/googleapis/api v0.0.0-20260715232425-e75dac1f907d
root gopkg.in/yaml.v2 v2.4.0
root honnef.co/go/tools v0.7.0`

const devtoolsAllowedVersions = `root github.com/chzyer/readline v1.5.1
root github.com/ianlancetaylor/demangle v0.0.0-20250417193237-f615e6bd150b
root go.opentelemetry.io/otel/metric/x v0.66.0
tools golang.org/x/time v0.11.0
tools modernc.org/mathutil v1.6.0
tools modernc.org/sortutil v1.2.0
tools modernc.org/strutil v1.2.0` + "\n" + devtoolsRootGraphCandidates

const devtoolsRootGraphCandidates = `root cel.dev/expr v0.25.1
root cloud.google.com/go v0.34.0
root github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1
root github.com/BurntSushi/toml v1.3.2
root github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137
root github.com/ebitengine/purego v0.8.4
root github.com/gorilla/websocket v1.4.2
root github.com/hashicorp/go-version v1.8.0
root github.com/mattn/go-colorable v0.1.14
root github.com/moby/moby/api v1.54.2
root github.com/moby/moby/client v0.4.1
root github.com/moby/term v0.5.0
root github.com/pelletier/go-toml/v2 v2.2.2
root github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c
root github.com/shirou/gopsutil/v4 v4.25.6
root github.com/sirupsen/logrus v1.9.3
root github.com/tklauser/go-sysconf v0.3.12
root github.com/tklauser/numcpus v0.6.1
root go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.68.0
root go.uber.org/zap v1.27.1
root golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3
root golang.org/x/oauth2 v0.34.0
root golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
root google.golang.org/appengine v1.4.0
root google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
root google.golang.org/genproto/googleapis/api v0.0.0-20260120221211-b8f7ae30c516
root gopkg.in/yaml.v2 v2.2.3
root honnef.co/go/tools v0.0.0-20190523083050-ea95bdfd59fc`

const devtoolsRecordedDrift = `root cel.dev/expr v0.25.1
root cloud.google.com/go v0.34.0
root github.com/Azure/go-ansiterm v0.0.0-20210617225240-d185dfc1b5a1
root github.com/BurntSushi/toml v1.3.2
root github.com/alecthomas/units v0.0.0-20211218093645-b94a6e3cc137
root github.com/chzyer/readline v1.5.1
root github.com/ebitengine/purego v0.8.4
root github.com/gorilla/websocket v1.4.2
root github.com/hashicorp/go-version v1.8.0
root github.com/ianlancetaylor/demangle v0.0.0-20250417193237-f615e6bd150b
root github.com/mattn/go-colorable v0.1.14
root github.com/moby/moby/api v1.54.2
root github.com/moby/moby/client v0.4.1
root github.com/moby/term v0.5.0
root github.com/pelletier/go-toml/v2 v2.2.2
root github.com/power-devops/perfstat v0.0.0-20210106213030-5aafc221ea8c
root github.com/shirou/gopsutil/v4 v4.25.6
root github.com/sirupsen/logrus v1.9.3
root github.com/tklauser/go-sysconf v0.3.12
root github.com/tklauser/numcpus v0.6.1
root go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp v0.68.0
root go.opentelemetry.io/otel/metric/x v0.66.0
root go.uber.org/zap v1.27.1
root golang.org/x/lint v0.0.0-20190313153728-d0100b6bd8b3
root golang.org/x/oauth2 v0.34.0
root golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1
root google.golang.org/appengine v1.4.0
root google.golang.org/genproto v0.0.0-20200526211855-cb27e3aa2013
root google.golang.org/genproto/googleapis/api v0.0.0-20260120221211-b8f7ae30c516
root gopkg.in/yaml.v2 v2.2.3
root honnef.co/go/tools v0.0.0-20190523083050-ea95bdfd59fc
tools golang.org/x/time v0.11.0
tools modernc.org/mathutil v1.6.0
tools modernc.org/sortutil v1.2.0
tools modernc.org/strutil v1.2.0`

func devtoolsVersionTuples(s string) map[string]map[string]string {
	r := map[string]map[string]string{"root": {}, "tools": {}}
	for _, line := range strings.Split(s, "\n") {
		f := strings.Fields(line)
		r[f[0]][f[1]] = f[2]
	}
	return r
}

func loadDevtoolsVersionPolicy(t *testing.T) devtoolsVersionPolicy {
	t.Helper()
	data, err := os.ReadFile("testdata/devtools-version-lock/policy.json")
	if err != nil {
		t.Fatal(err)
	}
	var p devtoolsVersionPolicy
	if err := json.Unmarshal(data, &p); err != nil {
		t.Fatal(err)
	}
	if p.SchemaVersion != 1 || !reflect.DeepEqual(p.Required, devtoolsVersionTuples(devtoolsRequiredVersions)) || !reflect.DeepEqual(p.Exceptions, devtoolsVersionTuples(devtoolsAllowedVersions)) {
		t.Fatal("policy changed the approved 32 targets / 35 owner-bound exceptions")
	}
	var baseline struct {
		SchemaVersion int
		Modules       []devtoolsVersionRecord
	}
	data, err = os.ReadFile("testdata/devtools-version-lock/baseline.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &baseline); err != nil {
		t.Fatal(err)
	}
	if baseline.SchemaVersion != 1 || len(baseline.Modules) != 666 {
		t.Fatal("invalid immutable baseline")
	}
	want := map[string]string{}
	for _, m := range baseline.Modules {
		if !m.Main && m.Path != "go" && m.Path != "toolchain" {
			want[m.Path] = m.Version
		}
	}
	if !reflect.DeepEqual(want, p.Baseline) {
		t.Fatal("policy baseline differs from archive")
	}
	return p
}

func TestDevtoolsVersionPolicy(t *testing.T) {
	loadDevtoolsVersionPolicy(t)
	t.Run("root-28-candidates", func(t *testing.T) {
		for path, candidate := range devtoolsVersionTuples(devtoolsRootGraphCandidates)["root"] {
			for _, scenario := range []string{"original", "candidate", "wrong-version", "wrong-owner", "missing", "removed", "pending", "no-evidence"} {
				t.Run(path+"/"+scenario, func(t *testing.T) {
					p := loadDevtoolsVersionPolicy(t)
					p.Exceptions["root"][path] = candidate
					p.Evidence["root"][path] = "fixture-" + path
					s := devtoolsVersionSnapshot{Owner: "root"}
					for key, version := range p.Baseline {
						p.Membership["root"][key] = "retained"
						p.Membership["tools"][key] = "retained"
						if key == path {
							if scenario == "missing" || scenario == "removed" {
								continue
							}
							if scenario != "original" {
								version = candidate
							}
							if scenario == "wrong-version" {
								version = "v999.0.0"
							}
						}
						s.Modules = append(s.Modules, devtoolsVersionRecord{Path: key, Version: version})
					}
					switch scenario {
					case "wrong-owner":
						s.Owner = "tools"
					case "removed":
						p.Membership["root"][path] = "removed"
					case "pending":
						p.Membership["root"][path] = "pending"
					case "no-evidence":
						delete(p.Evidence["root"], path)
					}
					issues := compareDevtoolsVersions(s, p)
					wantReject := scenario != "original" && scenario != "candidate"
					if (len(issues) > 0) != wantReject {
						t.Fatalf("reject=%v want=%v: %v", len(issues) > 0, wantReject, issues)
					}
				})
			}
		}
	})
	t.Run("four-tools-candidates", func(t *testing.T) {
		for path, candidate := range map[string]string{"golang.org/x/time": "v0.11.0", "modernc.org/mathutil": "v1.6.0", "modernc.org/sortutil": "v1.2.0", "modernc.org/strutil": "v1.2.0"} {
			p := loadDevtoolsVersionPolicy(t)
			p.Exceptions["tools"][path] = candidate
			p.Evidence["tools"][path] = "fixture-evidence"
			s := devtoolsVersionSnapshot{Owner: "tools"}
			for path, version := range p.Baseline {
				p.Membership["tools"][path] = "retained"
				s.Modules = append(s.Modules, devtoolsVersionRecord{Path: path, Version: version})
			}
			if got := compareDevtoolsVersions(s, p); len(got) != 0 {
				t.Fatal(got)
			}
			for i := range s.Modules {
				if s.Modules[i].Path == path {
					s.Modules[i].Version = candidate
				}
			}
			if got := compareDevtoolsVersions(s, p); len(got) != 0 {
				t.Errorf("%s exact candidate rejected: %v", path, got)
			}
		}
	})
	t.Run("recorded-35-differences", func(t *testing.T) {
		// Saved -e diagnostic differences are rejection inputs, never positive evidence.
		for _, line := range strings.Split(devtoolsRecordedDrift, "\n") {
			p := loadDevtoolsVersionPolicy(t)
			f := strings.Fields(line)
			// Historical diagnostic input has no reviewed evidence attached.
			p.Evidence[f[0]] = map[string]string{}
			s := devtoolsVersionSnapshot{Owner: f[0]}
			for path, version := range p.Baseline {
				p.Membership[f[0]][path] = "retained"
				if path == f[1] {
					version = f[2]
				}
				s.Modules = append(s.Modules, devtoolsVersionRecord{Path: path, Version: version})
			}
			if _, ok := p.Baseline[f[1]]; !ok {
				s.Modules = append(s.Modules, devtoolsVersionRecord{Path: f[1], Version: f[2]})
			}
			matched := false
			for _, issue := range compareDevtoolsVersions(s, p) {
				if strings.Contains(issue, f[1]+":") {
					matched = true
				}
			}
			if !matched {
				t.Errorf("recorded drift escaped rejection: %s", line)
			}
		}
	})
	t.Run("three-exact-candidates-and-natural-disappearance", func(t *testing.T) {
		for path, version := range devtoolsVersionTuples(devtoolsAllowedVersions)["root"] {
			p := loadDevtoolsVersionPolicy(t)
			s := devtoolsVersionSnapshot{Owner: "root"}
			for path, version := range p.Baseline {
				p.Membership["root"][path] = "retained"
				s.Modules = append(s.Modules, devtoolsVersionRecord{Path: path, Version: version})
			}
			if got := compareDevtoolsVersions(s, p); len(got) != 0 {
				t.Fatalf("natural original state rejected: %v", got)
			}
			found := false
			for i := range s.Modules {
				if s.Modules[i].Path == path {
					s.Modules[i].Version = version
					found = true
				}
			}
			if !found {
				s.Modules = append(s.Modules, devtoolsVersionRecord{Path: path, Version: version})
			}
			p.Evidence["root"][path] = "synthetic-reviewed-candidate"
			if got := compareDevtoolsVersions(s, p); len(got) != 0 {
				t.Fatalf("exact candidate rejected: %v", got)
			}
		}
	})
	// A permissive comparator would let every negative case below pass incorrectly.
	for _, tc := range []struct {
		name   string
		mutate func(*devtoolsVersionSnapshot, *devtoolsVersionPolicy)
		reject bool
	}{
		{"baseline", func(*devtoolsVersionSnapshot, *devtoolsVersionPolicy) {}, false},
		{"exact-exception", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy) { s.Modules[1].Version = "v1.5.1" }, false},
		{"new-exception", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy) {
			s.Modules = append(s.Modules, devtoolsVersionRecord{Path: "go.opentelemetry.io/otel/metric/x", Version: "v0.66.0"})
		}, false},
		{"target-downgraded", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy) { s.Modules[0].Version = "v0.1.0" }, true},
		{"target-deleted", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy) { s.Modules = s.Modules[1:] }, true},
		{"target-marked-removed", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy) {
			s.Modules = s.Modules[1:]
			p.Membership["root"]["cel.dev/expr"] = "removed"
		}, true},
		{"wrong-owner", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy) {
			s.Owner = "tools"
			s.Modules[1].Version = "v1.5.1"
		}, true},
		{"wrong-version", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy) { s.Modules[1].Version = "v1.5.2" }, true},
		{"fourth-exception", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy) {
			s.Modules = append(s.Modules, devtoolsVersionRecord{Path: "example.com/extra", Version: "v1.0.0"})
		}, true},
		{"unknown-new", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy) {
			s.Modules = append(s.Modules, devtoolsVersionRecord{Path: "example.com/unknown", Version: "v1.0.0"})
		}, true},
		{"unexplained-removal", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy) {
			s.Modules = s.Modules[:1]
			p.Membership["root"]["github.com/chzyer/readline"] = "removed"
			delete(p.Evidence["root"], "github.com/chzyer/readline")
		}, true},
		{"unreviewed-membership", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy) {
			p.Membership["root"]["cel.dev/expr"] = "pending"
		}, true},
		{"exception-no-evidence", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy) {
			s.Modules[1].Version = "v1.5.1"
			delete(p.Evidence["root"], "github.com/chzyer/readline")
		}, true},
		{"error", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy) {
			s.Modules[0].Error = json.RawMessage(`{"Err":"missing"}`)
		}, true},
		{"duplicate", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy) {
			s.Modules = append(s.Modules, s.Modules[0])
		}, true},
		{"empty", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy) { s.Modules = nil }, true},
		{"unknown-owner", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy) { s.Owner = "other" }, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			p := devtoolsVersionPolicy{Baseline: map[string]string{"cel.dev/expr": "v0.25.2", "github.com/chzyer/readline": "v0.0.0-20180603132655-2972be24d48e"}, Required: map[string]map[string]string{"root": {"cel.dev/expr": "v0.25.2"}, "tools": {}}, Exceptions: devtoolsVersionTuples(devtoolsAllowedVersions), Membership: map[string]map[string]string{"root": {"cel.dev/expr": "retained", "github.com/chzyer/readline": "retained"}, "tools": {"cel.dev/expr": "retained", "github.com/chzyer/readline": "retained"}}, Evidence: map[string]map[string]string{"root": {"github.com/chzyer/readline": "reviewed-readline", "go.opentelemetry.io/otel/metric/x": "reviewed-metric"}, "tools": {}}}
			s := devtoolsVersionSnapshot{Owner: "root", Modules: []devtoolsVersionRecord{{Path: "cel.dev/expr", Version: "v0.25.2"}, {Path: "github.com/chzyer/readline", Version: "v0.0.0-20180603132655-2972be24d48e"}}}
			tc.mutate(&s, &p)
			got := compareDevtoolsVersions(s, p)
			if (len(got) > 0) != tc.reject {
				t.Fatalf("reject=%v, want %v: %v", len(got) > 0, tc.reject, got)
			}
		})
	}
}

func compareDevtoolsVersions(snapshot devtoolsVersionSnapshot, policy devtoolsVersionPolicy) []string {
	owner := snapshot.Owner
	if owner != "root" && owner != "tools" {
		return []string{"unknown owner"}
	}
	var issues []string
	add := func(path, reason string) { issues = append(issues, fmt.Sprintf("%s %s: %s", owner, path, reason)) }
	actual := map[string]string{}
	for _, m := range snapshot.Modules {
		if len(m.Error) > 0 && string(m.Error) != "null" {
			add(m.Path, "query Error")
		}
		if m.Main || m.Path == "go" || m.Path == "toolchain" {
			continue
		}
		if _, ok := actual[m.Path]; ok {
			add(m.Path, "duplicate identity")
		}
		actual[m.Path] = m.Version
		if m.Path == "" || m.Version == "" {
			add(m.Path, "empty identity/version")
		}
		baseline, known := policy.Baseline[m.Path]
		if !known || m.Version != baseline {
			if policy.Exceptions[owner][m.Path] != m.Version || m.Version == "" {
				add(m.Path, "unapproved version "+m.Version)
			} else if strings.TrimSpace(policy.Evidence[owner][m.Path]) == "" {
				add(m.Path, "exception lacks evidence")
			}
		}
	}
	if len(actual) == 0 {
		add("", "empty snapshot")
	}
	for path := range policy.Baseline {
		_, present := actual[path]
		switch policy.Membership[owner][path] {
		case "retained":
			if !present {
				add(path, "retained module missing")
			}
		case "removed":
			if present {
				add(path, "removed module still present")
			}
			if strings.TrimSpace(policy.Evidence[owner][path]) == "" {
				add(path, "removal lacks evidence")
			}
		default:
			add(path, "membership unreviewed")
		}
	}
	for path, want := range policy.Required[owner] {
		candidate := actual[path] != "" && policy.Exceptions[owner][path] == actual[path]
		if actual[path] != want && !candidate {
			add(path, "required exact version "+want+", got "+actual[path])
		}
		if policy.Membership[owner][path] != "retained" {
			add(path, "required target must be retained")
		}
	}
	sort.Strings(issues)
	return issues
}

func readDevtoolsVersionSnapshot(t *testing.T, owner string) devtoolsVersionSnapshot {
	t.Helper()
	if owner != "root" && owner != "tools" {
		t.Fatal("invalid module owner")
	}
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	args := []string{"list", "-m", "-json"}
	if owner == "tools" {
		args = append(args, "-modfile="+filepath.Join(root, "tools/devtools/go.mod"))
	}
	args = append(args, "all")
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, devtoolsRealGo(t), args...)
	cmd.Dir = root
	cmd.WaitDelay = time.Second
	cmd.Env = devtoolsFixtureEnv(map[string]string{"GOENV": "off", "GOWORK": "off", "GOTOOLCHAIN": "local", "GOPROXY": "off", "GOSUMDB": "off", "GOAUTH": "off", "GOVCS": "all:off", "GOFLAGS": "-mod=readonly"})
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("%s strict offline query failed (not version RED): %v", owner, err)
	}
	d := json.NewDecoder(bytes.NewReader(out))
	s := devtoolsVersionSnapshot{Owner: owner}
	seen := map[string]bool{}
	for {
		var m devtoolsVersionRecord
		err = d.Decode(&m)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if len(m.Error) > 0 && string(m.Error) != "null" {
			t.Fatal("module query returned Error")
		}
		if seen[m.Path] || m.Path == "" {
			t.Fatal("invalid/duplicate module identity")
		}
		seen[m.Path] = true
		s.Modules = append(s.Modules, m)
	}
	if len(s.Modules) < 2 {
		t.Fatal("missing module query result")
	}
	return s
}

func readDevtoolsPackageSnapshot(t *testing.T, phase string) []devtoolsPackageRecord {
	t.Helper()
	root, err := filepath.Abs("../..")
	if err != nil {
		t.Fatal(err)
	}
	fields := "Dir,ImportPath,Name,Standard,ForTest,Module,Match,DepOnly,GoFiles,CgoFiles,TestGoFiles,XTestGoFiles,Imports,ImportMap,Deps,TestImports,XTestImports,Incomplete,Error,DepsErrors,EmbedFiles,TestEmbedFiles,XTestEmbedFiles"
	args := []string{"list", "-deps", "-json=" + fields, "-mod=readonly"}
	tools := map[string]string{"buf": "github.com/bufbuild/buf/cmd/buf", "protoc": "google.golang.org/protobuf/cmd/protoc-gen-go", "oapi": "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen", "sqlc": "github.com/sqlc-dev/sqlc/cmd/sqlc", "lint": "github.com/golangci/golangci-lint/v2/cmd/golangci-lint"}
	switch phase {
	case "root":
		args = append(args, "-test", "./...")
	case "integration":
		args = append(args, "-test", "-tags=integration", "./...")
	case "goose":
		args = append(args, "github.com/pressly/goose/v3/cmd/goose")
	default:
		pkg, ok := tools[phase]
		if !ok {
			t.Fatalf("unknown package phase %q", phase)
		}
		args = append(args, "-modfile="+filepath.Join(root, "tools/devtools/go.mod"), pkg)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, devtoolsRealGo(t), args...)
	cmd.Dir = root
	cmd.WaitDelay = time.Second
	cmd.Env = devtoolsFixtureEnv(map[string]string{"GOENV": "off", "GOWORK": "off", "GOTOOLCHAIN": "local", "GOPROXY": "off", "GOSUMDB": "off", "GOAUTH": "off", "GOVCS": "all:off", "GOFLAGS": "-mod=readonly"})
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("%s strict package query failed: %v", phase, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(out))
	var records []devtoolsPackageRecord
	seen := map[string]bool{}
	for {
		var r devtoolsPackageRecord
		err := decoder.Decode(&r)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		if r.ImportPath == "" || seen[r.ImportPath] || r.Incomplete || (len(r.Error) > 0 && string(r.Error) != "null") || len(r.DepsErrors) > 0 || (r.Module != nil && len(r.Module.Error) > 0 && string(r.Module.Error) != "null") {
			t.Fatalf("%s invalid package result %q", phase, r.ImportPath)
		}
		seen[r.ImportPath] = true
		records = append(records, r)
	}
	if len(records) == 0 {
		t.Fatalf("%s empty package result", phase)
	}
	return records
}

func TestDevtoolsVersionLock(t *testing.T) {
	p := loadDevtoolsVersionPolicy(t)
	report, err := os.ReadFile("../../docs/roadmap/2026-09-27-devtools-version-lock-evidence.md")
	if err != nil {
		t.Fatal(err)
	}
	evidence, err := parseDevtoolsEvidence(report)
	if err != nil {
		t.Fatal(err)
	}
	var baseline struct {
		Packages map[string][]struct{ ImportPath, Module, Version string }
	}
	data, err := os.ReadFile("testdata/devtools-version-lock/baseline.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &baseline); err != nil {
		t.Fatal(err)
	}
	t.Run("module-contract", TestDevtoolsModuleBoundary)
	for _, owner := range []string{"root", "tools"} {
		t.Run(owner, func(t *testing.T) {
			s := readDevtoolsVersionSnapshot(t, owner)
			s.Packages = map[string][]devtoolsPackageRecord{}
			phases := []string{"root", "integration", "goose"}
			if owner == "tools" {
				phases = []string{"buf", "protoc", "oapi", "sqlc", "lint"}
			}
			for _, phase := range phases {
				s.Packages[phase] = readDevtoolsPackageSnapshot(t, phase)
				actual := map[string]string{}
				want := map[string]string{}
				for _, r := range s.Packages[phase] {
					value := ""
					if r.Module != nil {
						value = r.Module.Path + "@" + r.Module.Version
					}
					actual[r.ImportPath] = value
				}
				for _, r := range baseline.Packages[phase] {
					value := ""
					if r.Module != "" {
						value = r.Module + "@" + r.Version
					}
					if _, exists := want[r.ImportPath]; exists {
						t.Fatal("duplicate baseline package")
					}
					want[r.ImportPath] = value
				}
				if len(want) == 0 || !reflect.DeepEqual(actual, want) {
					t.Errorf("%s package map differs from immutable baseline", phase)
				}
			}
			if issues := compareDevtoolsPackageUse(s, p); len(issues) > 0 {
				t.Errorf("package use rejected: %s", strings.Join(issues, "\n"))
			}
			if issues := validateDevtoolsEvidence(s, p, evidence); len(issues) > 0 {
				t.Errorf("evidence rejected: %s", strings.Join(issues, "\n"))
			}
			if issues := compareDevtoolsVersions(s, p); len(issues) > 0 {
				t.Fatalf("version lock rejected: %s", strings.Join(issues, "\n"))
			}
		})
	}
}
