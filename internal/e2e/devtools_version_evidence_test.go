//go:build e2e

package e2e

import (
	"encoding/json"
	"fmt"
	"io"
	"reflect"
	"sort"
	"strings"
	"testing"
)

type devtoolsPackageRecord struct {
	ImportPath string
	Module     *devtoolsVersionRecord
	Incomplete bool
	Error      json.RawMessage
	DepsErrors []json.RawMessage
}
type devtoolsEvidenceRecord struct{ ID, Owner, Path, Version, Kind, Sum, GoModSum string }

func parseDevtoolsEvidence(report []byte) ([]devtoolsEvidenceRecord, error) {
	var records []devtoolsEvidenceRecord
	seen := map[string]bool{}
	const prefix = "<!-- devtools-evidence "
	for i, line := range strings.Split(string(report), "\n") {
		if !strings.Contains(line, "<!-- devtools-evidence") {
			continue
		}
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, prefix) || !strings.HasSuffix(line, " -->") {
			return nil, fmt.Errorf("evidence line %d: invalid framing", i+1)
		}
		d := json.NewDecoder(strings.NewReader(strings.TrimSuffix(strings.TrimPrefix(line, prefix), " -->")))
		d.DisallowUnknownFields()
		var r devtoolsEvidenceRecord
		fields := map[string]*string{"ID": &r.ID, "Owner": &r.Owner, "Path": &r.Path, "Version": &r.Version, "Kind": &r.Kind, "Sum": &r.Sum, "GoModSum": &r.GoModSum}
		start, err := d.Token()
		if err != nil || start != json.Delim('{') {
			return nil, fmt.Errorf("evidence line %d: expected object", i+1)
		}
		for d.More() {
			token, err := d.Token()
			key, ok := token.(string)
			field := fields[key]
			if err != nil || !ok || field == nil {
				return nil, fmt.Errorf("evidence line %d: unknown or duplicate field", i+1)
			}
			if err := d.Decode(field); err != nil {
				return nil, fmt.Errorf("evidence line %d: %w", i+1, err)
			}
			delete(fields, key)
		}
		if end, err := d.Token(); err != nil || end != json.Delim('}') {
			return nil, fmt.Errorf("evidence line %d: invalid object end", i+1)
		}
		var extra any
		if err := d.Decode(&extra); err != io.EOF {
			return nil, fmt.Errorf("evidence line %d: trailing data", i+1)
		}
		if strings.TrimSpace(r.ID) == "" || seen[r.ID] || (r.Kind != "exception" && r.Kind != "removal") {
			return nil, fmt.Errorf("evidence line %d: invalid ID or kind", i+1)
		}
		seen[r.ID] = true
		records = append(records, r)
	}
	return records, nil
}
func validateDevtoolsEvidence(s devtoolsVersionSnapshot, p devtoolsVersionPolicy, records []devtoolsEvidenceRecord) []string {
	if s.Owner != "root" && s.Owner != "tools" {
		return []string{"unknown evidence owner"}
	}
	var issues []string
	byID := map[string]devtoolsEvidenceRecord{}
	for _, r := range records {
		if _, exists := byID[r.ID]; exists || strings.TrimSpace(r.ID) == "" {
			issues = append(issues, "duplicate or empty evidence ID: "+r.ID)
		}
		byID[r.ID] = r
	}
	check := func(path, version, kind string) {
		id := p.Evidence[s.Owner][path]
		r, ok := byID[id]
		if strings.TrimSpace(id) == "" || !ok || r.Owner != s.Owner || r.Path != path || r.Version != version || r.Kind != kind {
			issues = append(issues, s.Owner+" "+path+": evidence binding mismatch")
		} else if kind == "exception" && (strings.TrimSpace(r.Sum) == "" || strings.TrimSpace(r.GoModSum) == "") {
			issues = append(issues, s.Owner+" "+path+": missing checksum evidence")
		}
	}
	for _, m := range s.Modules {
		if m.Main || m.Path == "go" || m.Path == "toolchain" {
			continue
		}
		if baseline, exists := p.Baseline[m.Path]; !exists || baseline != m.Version {
			if m.Version == "" || p.Exceptions[s.Owner][m.Path] != m.Version {
				issues = append(issues, s.Owner+" "+m.Path+": unapproved evidence version")
			}
			check(m.Path, m.Version, "exception")
		}
	}
	for path, membership := range p.Membership[s.Owner] {
		if membership == "removed" {
			check(path, p.Baseline[path], "removal")
		}
	}
	sort.Strings(issues)
	return issues
}
func compareDevtoolsPackageUse(s devtoolsVersionSnapshot, p devtoolsVersionPolicy) []string {
	rootGraphPaths := devtoolsVersionTuples(devtoolsRootGraphCandidates)["root"]
	phases := []string{"buf", "protoc", "oapi", "sqlc", "lint"}
	if s.Owner == "root" {
		phases = []string{"root", "integration", "goose"}
	} else if s.Owner != "tools" {
		return []string{"unknown package owner"}
	}
	var issues []string
	hasError := func(raw json.RawMessage) bool { return len(raw) > 0 && strings.TrimSpace(string(raw)) != "null" }
	for _, phase := range phases {
		records := s.Packages[phase]
		if len(records) == 0 {
			issues = append(issues, phase+": missing package result")
		}
		seen := map[string]bool{}
		for _, r := range records {
			if r.ImportPath == "" || seen[r.ImportPath] {
				issues = append(issues, phase+": invalid/duplicate package "+r.ImportPath)
			}
			seen[r.ImportPath] = true
			if r.Incomplete || hasError(r.Error) || len(r.DepsErrors) > 0 || (r.Module != nil && hasError(r.Module.Error)) {
				issues = append(issues, phase+": incomplete package query "+r.ImportPath)
			}
			if s.Owner == "tools" && r.Module != nil {
				switch r.Module.Path {
				case "golang.org/x/time", "modernc.org/mathutil", "modernc.org/sortutil", "modernc.org/strutil":
					issues = append(issues, phase+": graph-only module used by package "+r.ImportPath)
				}
			}
			if s.Owner == "root" && r.Module != nil && rootGraphPaths[r.Module.Path] != "" {
				issues = append(issues, phase+": root graph-only module used by package "+r.ImportPath)
			}
		}
	}
	sort.Strings(issues)
	return issues
}

func TestDevtoolsVersionEvidence(t *testing.T) {
	t.Run("root-28-bindings", func(t *testing.T) {
		for path, version := range devtoolsVersionTuples(devtoolsRootGraphCandidates)["root"] {
			for _, scenario := range []string{"valid", "owner", "path", "version", "kind", "id", "sum", "modsum"} {
				t.Run(path+"/"+scenario, func(t *testing.T) {
					p := loadDevtoolsVersionPolicy(t)
					// This fixture validates one binding, not the real removal inventory.
					p.Membership["root"] = map[string]string{path: "retained"}
					p.Exceptions["root"][path] = version
					p.Evidence["root"][path] = "fixture"
					s := devtoolsVersionSnapshot{Owner: "root", Modules: []devtoolsVersionRecord{{Path: path, Version: version}}}
					r := devtoolsEvidenceRecord{ID: "fixture", Owner: "root", Path: path, Version: version, Kind: "exception", Sum: "h1:archive", GoModSum: "h1:manifest"}
					switch scenario {
					case "owner":
						r.Owner = "tools"
					case "path":
						r.Path = "example.com/other"
					case "version":
						r.Version = "v999.0.0"
					case "kind":
						r.Kind = "removal"
					case "id":
						r.ID = ""
					case "sum":
						r.Sum = ""
					case "modsum":
						r.GoModSum = ""
					}
					issues := validateDevtoolsEvidence(s, p, []devtoolsEvidenceRecord{r})
					if (len(issues) > 0) != (scenario != "valid") {
						t.Fatalf("unexpected binding result: %v", issues)
					}
				})
			}
		}
	})
	valid := `<!-- devtools-evidence {"ID":"tools-time","Owner":"tools","Path":"golang.org/x/time","Version":"v0.11.0","Kind":"exception","Sum":"h1:archive","GoModSum":"h1:manifest"} -->`
	t.Run("parse-valid", func(t *testing.T) {
		r, err := parseDevtoolsEvidence([]byte("Explanation\n" + valid + "\nMore explanation."))
		want := []devtoolsEvidenceRecord{{ID: "tools-time", Owner: "tools", Path: "golang.org/x/time", Version: "v0.11.0", Kind: "exception", Sum: "h1:archive", GoModSum: "h1:manifest"}}
		if err != nil || !reflect.DeepEqual(r, want) {
			t.Fatalf("parse=%v %v", r, err)
		}
	})
	for name, body := range map[string]string{
		"wrong-field-case": strings.Replace(valid, `"Owner":`, `"owner":`, 1),
		"duplicate-field":  strings.Replace(valid, `"ID":`, `"ID":"first","ID":`, 1),
		"duplicate-id":     valid + "\n" + valid,
		"broken-json":      "<!-- devtools-evidence { -->",
		"missing-close":    strings.TrimSuffix(valid, " -->"),
		"unknown-field":    strings.Replace(valid, `"ID":`, `"Extra":true,"ID":`, 1),
		"trailing-json":    strings.Replace(valid, "} -->", "} {} -->", 1),
		"missing-id":       strings.Replace(valid, `"ID":"tools-time",`, "", 1),
		"bad-kind":         strings.Replace(valid, "exception", "other", 1),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := parseDevtoolsEvidence([]byte(body)); err == nil {
				t.Fatal("malformed evidence accepted")
			}
		})
	}
	for _, tc := range []struct {
		name   string
		mutate func(*devtoolsVersionSnapshot, *devtoolsVersionPolicy, *[]devtoolsEvidenceRecord)
		reject bool
	}{
		{"valid-binding", func(*devtoolsVersionSnapshot, *devtoolsVersionPolicy, *[]devtoolsEvidenceRecord) {}, false},
		{"empty-id", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy, r *[]devtoolsEvidenceRecord) {
			p.Evidence["tools"]["golang.org/x/time"] = ""
		}, true},
		{"missing-id", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy, r *[]devtoolsEvidenceRecord) {
			p.Evidence["tools"]["golang.org/x/time"] = "absent"
		}, true},
		{"duplicate-binding", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy, r *[]devtoolsEvidenceRecord) {
			*r = append(*r, (*r)[0])
		}, true},
		{"wrong-owner", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy, r *[]devtoolsEvidenceRecord) {
			(*r)[0].Owner = "root"
		}, true},
		{"wrong-path", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy, r *[]devtoolsEvidenceRecord) {
			(*r)[0].Path = "modernc.org/mathutil"
		}, true},
		{"wrong-version", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy, r *[]devtoolsEvidenceRecord) {
			(*r)[0].Version = "v0.14.0"
		}, true},
		{"wrong-kind", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy, r *[]devtoolsEvidenceRecord) {
			(*r)[0].Kind = "removal"
		}, true},
		{"missing-sum", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy, r *[]devtoolsEvidenceRecord) {
			(*r)[0].Sum = ""
		}, true},
		{"missing-gomodsum", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy, r *[]devtoolsEvidenceRecord) {
			(*r)[0].GoModSum = ""
		}, true},
		{"unapproved-version", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy, r *[]devtoolsEvidenceRecord) {
			s.Modules[0].Version = "v0.12.0"
			(*r)[0].Version = "v0.12.0"
		}, true},
		{"original-needs-no-exception", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy, r *[]devtoolsEvidenceRecord) {
			s.Modules[0].Version = "v0.14.0"
			p.Evidence["tools"] = nil
			*r = nil
		}, false},
		{"removal-bound-to-original", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy, r *[]devtoolsEvidenceRecord) {
			s.Modules = nil
			p.Membership["tools"]["golang.org/x/time"] = "removed"
			(*r)[0].Kind = "removal"
			(*r)[0].Version = "v0.14.0"
		}, false},
		{"removal-wrong-version", func(s *devtoolsVersionSnapshot, p *devtoolsVersionPolicy, r *[]devtoolsEvidenceRecord) {
			s.Modules = nil
			p.Membership["tools"]["golang.org/x/time"] = "removed"
			(*r)[0].Kind = "removal"
		}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := devtoolsVersionSnapshot{Owner: "tools", Modules: []devtoolsVersionRecord{{Path: "golang.org/x/time", Version: "v0.11.0"}}}
			p := devtoolsVersionPolicy{Baseline: map[string]string{"golang.org/x/time": "v0.14.0"}, Exceptions: map[string]map[string]string{"tools": {"golang.org/x/time": "v0.11.0"}}, Evidence: map[string]map[string]string{"tools": {"golang.org/x/time": "tools-time"}}, Membership: map[string]map[string]string{"tools": {"golang.org/x/time": "retained"}}}
			r := []devtoolsEvidenceRecord{{ID: "tools-time", Owner: "tools", Path: "golang.org/x/time", Version: "v0.11.0", Kind: "exception", Sum: "h1:archive", GoModSum: "h1:manifest"}}
			tc.mutate(&s, &p, &r)
			if got := validateDevtoolsEvidence(s, p, r); (len(got) > 0) != tc.reject {
				t.Fatalf("reject=%v want=%v: %v", len(got) > 0, tc.reject, got)
			}
		})
	}
}

func devtoolsCleanPackageFixture(owner string) devtoolsVersionSnapshot {
	s := devtoolsVersionSnapshot{Owner: owner, Packages: map[string][]devtoolsPackageRecord{}}
	phases := []string{"buf", "protoc", "oapi", "sqlc", "lint"}
	if owner == "root" {
		phases = []string{"root", "integration", "goose"}
	}
	for _, phase := range phases {
		s.Packages[phase] = []devtoolsPackageRecord{{ImportPath: "fmt"}}
	}
	return s
}

func TestDevtoolsVersionPackageUse(t *testing.T) {
	p := devtoolsVersionPolicy{}
	t.Run("root-28-phases", func(t *testing.T) {
		for path, candidate := range devtoolsVersionTuples(devtoolsRootGraphCandidates)["root"] {
			original := devtoolsVersionTuples(devtoolsRequiredVersions)["root"][path]
			for _, phase := range []string{"root", "integration", "goose"} {
				for _, version := range []string{original, candidate} {
					t.Run(path+"/"+phase+"/"+version, func(t *testing.T) {
						s := devtoolsCleanPackageFixture("root")
						s.Packages[phase] = append(s.Packages[phase], devtoolsPackageRecord{ImportPath: path + "/pkg", Module: &devtoolsVersionRecord{Path: path, Version: version}})
						if len(compareDevtoolsPackageUse(s, p)) == 0 {
							t.Fatal("root graph-only module became a used package")
						}
					})
				}
			}
			t.Run(path+"/tools-allowed", func(t *testing.T) {
				s := devtoolsCleanPackageFixture("tools")
				s.Packages["buf"] = append(s.Packages["buf"], devtoolsPackageRecord{ImportPath: path + "/pkg", Module: &devtoolsVersionRecord{Path: path, Version: original}})
				if issues := compareDevtoolsPackageUse(s, p); len(issues) > 0 {
					t.Fatal(issues)
				}
			})
		}
	})
	t.Run("all-tools-clean", func(t *testing.T) {
		if got := compareDevtoolsPackageUse(devtoolsCleanPackageFixture("tools"), p); len(got) != 0 {
			t.Fatal(got)
		}
	})
	for _, phase := range []string{"buf", "protoc", "oapi", "sqlc", "lint"} {
		for _, path := range []string{"golang.org/x/time", "modernc.org/mathutil", "modernc.org/sortutil", "modernc.org/strutil"} {
			t.Run(phase+"/"+path, func(t *testing.T) {
				s := devtoolsCleanPackageFixture("tools")
				s.Packages[phase] = append(s.Packages[phase], devtoolsPackageRecord{ImportPath: path + "/pkg", Module: &devtoolsVersionRecord{Path: path, Version: "v1.0.0"}})
				if len(compareDevtoolsPackageUse(s, p)) == 0 {
					t.Fatal("graph-only module became a used package")
				}
			})
		}
	}
	for _, tc := range []struct {
		name   string
		mutate func(*devtoolsVersionSnapshot)
	}{
		{"missing-group", func(s *devtoolsVersionSnapshot) { delete(s.Packages, "sqlc") }},
		{"empty-group", func(s *devtoolsVersionSnapshot) { s.Packages["sqlc"] = nil }},
		{"package-error", func(s *devtoolsVersionSnapshot) { s.Packages["sqlc"][0].Error = json.RawMessage(`{"Err":"missing"}`) }},
		{"incomplete", func(s *devtoolsVersionSnapshot) { s.Packages["sqlc"][0].Incomplete = true }},
		{"deps-error", func(s *devtoolsVersionSnapshot) {
			s.Packages["sqlc"][0].DepsErrors = []json.RawMessage{json.RawMessage(`{"Err":"missing"}`)}
		}},
		{"module-error", func(s *devtoolsVersionSnapshot) {
			s.Packages["sqlc"][0].Module = &devtoolsVersionRecord{Path: "example.com/pkg", Version: "v1.0.0", Error: json.RawMessage(`{"Err":"missing"}`)}
		}},
		{"duplicate", func(s *devtoolsVersionSnapshot) {
			s.Packages["sqlc"] = append(s.Packages["sqlc"], s.Packages["sqlc"][0])
		}},
		{"empty-identity", func(s *devtoolsVersionSnapshot) { s.Packages["sqlc"][0].ImportPath = "" }},
		{"unknown-owner", func(s *devtoolsVersionSnapshot) { s.Owner = "other" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s := devtoolsCleanPackageFixture("tools")
			tc.mutate(&s)
			if len(compareDevtoolsPackageUse(s, p)) == 0 {
				t.Fatal("invalid package query accepted")
			}
		})
	}
	t.Run("root-goose-mathutil-remains-allowed", func(t *testing.T) {
		s := devtoolsCleanPackageFixture("root")
		s.Packages["goose"] = append(s.Packages["goose"], devtoolsPackageRecord{ImportPath: "modernc.org/mathutil", Module: &devtoolsVersionRecord{Path: "modernc.org/mathutil", Version: "v1.7.1"}})
		if got := compareDevtoolsPackageUse(s, p); len(got) != 0 {
			t.Fatal(got)
		}
	})
}
