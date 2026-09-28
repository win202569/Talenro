//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestDevtoolsModuleBoundary(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		path  string
		tools []string
	}{
		{"go.mod", []string{"github.com/pressly/goose/v3/cmd/goose"}},
		{"tools/devtools/go.mod", []string{"github.com/bufbuild/buf/cmd/buf", "github.com/golangci/golangci-lint/v2/cmd/golangci-lint", "github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen", "github.com/sqlc-dev/sqlc/cmd/sqlc", "google.golang.org/protobuf/cmd/protoc-gen-go"}},
	} {
		t.Run(tc.path, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, devtoolsRealGo(t), "mod", "edit", "-json", filepath.Join(root, filepath.FromSlash(tc.path)))
			cmd.Env = devtoolsFixtureEnv(map[string]string{"GOENV": "off", "GOWORK": "off", "GOTOOLCHAIN": "local", "GOPROXY": "off", "GOSUMDB": "off"})
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("module contract cannot be read: %v: %s", err, out)
			}
			var mod struct {
				Module    struct{ Path string }
				Go        string
				Toolchain string
				Tool      []struct{ Path string }
				Require   []struct{ Path, Version string }
				Replace   []json.RawMessage
				Exclude   []json.RawMessage
			}
			if err := json.Unmarshal(out, &mod); err != nil {
				t.Fatal(err)
			}
			var tools []string
			for _, tool := range mod.Tool {
				tools = append(tools, tool.Path)
			}
			slices.Sort(tools)
			slices.Sort(tc.tools)
			if !slices.Equal(tools, tc.tools) {
				t.Errorf("wrong tool ownership: %v", tools)
			}
			if mod.Go != "1.26.0" || mod.Toolchain != "go1.26.5" || len(mod.Replace) != 0 || len(mod.Exclude) != 0 {
				t.Errorf("module runtime/replace contract: %s %s %d", mod.Go, mod.Toolchain, len(mod.Replace))
			}
			want := map[string]string{"github.com/pressly/goose/v3": "v3.27.1", "google.golang.org/protobuf": "v1.36.11", "github.com/oapi-codegen/runtime": "v1.6.0"}
			if tc.path == "go.mod" && mod.Module.Path != "talenro.local/platform" {
				t.Error("wrong root module identity")
			}
			if tc.path != "go.mod" {
				if mod.Module.Path != "talenro.local/devtools" {
					t.Error("wrong tools module identity")
				}
				want = map[string]string{"github.com/bufbuild/buf": "v1.72.0", "github.com/golangci/golangci-lint/v2": "v2.12.2", "github.com/oapi-codegen/oapi-codegen/v2": "v2.8.0", "github.com/sqlc-dev/sqlc": "v1.31.1", "google.golang.org/protobuf": "v1.36.11"}
			}
			for _, require := range mod.Require {
				if require.Path == "talenro.local/platform" || require.Path == "talenro.local/devtools" {
					t.Error("cross-module require")
				}
				if v, ok := want[require.Path]; ok {
					if v != require.Version {
						t.Errorf("version drift for %s", require.Path)
					}
					delete(want, require.Path)
				}
			}
			if len(want) != 0 {
				t.Errorf("required dependencies missing: %v", want)
			}
		})
	}
	if _, err := os.Stat(filepath.Join(root, "go.work")); !os.IsNotExist(err) {
		t.Error("unexpected workspace")
	}
}

type devtoolsRouteEvent struct {
	Tool                                                                string
	Args                                                                []string
	Dir, Flags                                                          string
	Exe, Cache, GoPath, Proxy, SumDB, GoEnv, Work, Toolchain, Auth, VCS string
}
type devtoolsRouteFixture struct{ root, profile, bin, log, host string }

func devtoolsRealGo(t *testing.T) string {
	t.Helper()
	if runtime.GOOS != "windows" {
		t.Skip("Windows devtools consumers")
	}
	return filepath.Join(os.Getenv("USERPROFILE"), "go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe")
}

func newDevtoolsRouteFixture(t *testing.T) *devtoolsRouteFixture {
	t.Helper()
	f := &devtoolsRouteFixture{root: filepath.Join(t.TempDir(), "repository with spaces"), profile: filepath.Join(t.TempDir(), "owned profile"), bin: filepath.Join(t.TempDir(), "tool bin"), host: devtoolsPowerShell(t)}
	f.log = filepath.Join(f.root, "events.jsonl")
	for _, dir := range []string{filepath.Join(f.root, "scripts/private"), filepath.Join(f.root, "tools/devtools"), filepath.Join(f.root, "rootonly"), filepath.Join(f.root, "api"), filepath.Join(f.root, "gen"), filepath.Join(f.root, "internal/store"), f.bin} {
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"scripts/verify-devtools.ps1", "scripts/private/devtools-process.ps1", "scripts/check-tools.ps1", "scripts/generate.ps1", "scripts/verify-c11.ps1", "buf.gen.yaml"} {
		data, err := os.ReadFile(filepath.Join("../..", filepath.FromSlash(name)))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(f.root, filepath.FromSlash(name)), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	for path, body := range map[string]string{"rootonly/fixture.go": "package rootonly\n\nconst RootOnly = true\n", "go.mod": "module fixture.local/root\n\ngo 1.26.0\n", "tools/devtools/go.mod": "module talenro.local/devtools\n\ngo 1.26.0\n", "tools/devtools/go.sum": ""} {
		if err := os.WriteFile(filepath.Join(f.root, filepath.FromSlash(path)), []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	pinned := filepath.Join(f.profile, "go/pkg/mod/golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe")
	if err := os.MkdirAll(filepath.Dir(pinned), 0700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(f.bin, "fixture.go")
	if err := os.WriteFile(source, []byte(devtoolsRoutingSource), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, devtoolsRealGo(t), "build", "-o", pinned, source)
	cmd.Env = devtoolsFixtureEnv(map[string]string{"GOENV": "off", "GOWORK": "off", "GOTOOLCHAIN": "local", "GOPROXY": "off", "GOSUMDB": "off", "CGO_ENABLED": "0"})
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("route fixture build: %v: %s", err, out)
	}
	binary, err := os.ReadFile(pinned)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(filepath.Dir(pinned), "gofmt.exe"), binary, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"go.exe", "git.exe", "docker.exe", "gofmt.exe", "pwsh.exe", "powershell.exe"} {
		if err := os.WriteFile(filepath.Join(f.bin, name), binary, 0700); err != nil {
			t.Fatal(err)
		}
	}
	return f
}

func (f *devtoolsRouteFixture) run(t *testing.T, entry string, extra map[string]string) (int, string, []devtoolsRouteEvent) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, f.host, "-NoProfile", "-NonInteractive", "-File", filepath.Join(f.root, "scripts", entry))
	values := map[string]string{"USERPROFILE": f.profile, "PATH": f.bin + string(os.PathListSeparator) + os.Getenv("PATH"), "DEVTOOLS_ROUTE_LOG": f.log, "DEVTOOLS_ROUTE_ROOT": f.root, "GOENV": "off"}
	for k, v := range extra {
		values[k] = v
	}
	cmd.Env = devtoolsFixtureEnv(values)
	cmd.Dir = t.TempDir()
	cmd.WaitDelay = time.Second
	out, err := cmd.CombinedOutput()
	if ctx.Err() != nil || cmd.ProcessState == nil {
		t.Fatalf("consumer deadline/start: %v: %s", err, out)
	}
	var events []devtoolsRouteEvent
	if data, readErr := os.ReadFile(f.log); readErr == nil {
		for _, line := range bytes.Split(bytes.TrimSpace(data), []byte("\n")) {
			if len(line) == 0 {
				continue
			}
			var event devtoolsRouteEvent
			if err := json.Unmarshal(line, &event); err != nil {
				t.Fatal(err)
			}
			events = append(events, event)
		}
	} else if !os.IsNotExist(readErr) {
		t.Fatal(readErr)
	}
	return cmd.ProcessState.ExitCode(), string(out), events
}

func TestDevtoolsRouting(t *testing.T) {
	for _, entry := range []string{"check-tools.ps1", "generate.ps1", "verify-c11.ps1"} {
		t.Run(entry, func(t *testing.T) {
			f := newDevtoolsRouteFixture(t)
			code, out, events := f.run(t, entry, nil)
			want := 0
			if entry == "verify-c11.ps1" {
				want = 47
			}
			if code != want {
				t.Fatalf("consumer exit %d want %d: %s; events=%v", code, want, out, events)
			}
			verified := false
			tools := 0
			for _, e := range events {
				if e.Tool == "pwsh" || e.Tool == "powershell" {
					t.Error("PATH interpreter used")
				}
				if e.Tool != "go" {
					continue
				}
				if slices.Equal(e.Args, []string{"mod", "verify"}) {
					verified = true
					continue
				}
				if len(e.Args) > 0 && e.Args[0] == "tool" {
					tools++
					if !verified {
						t.Error("tool ran before verification")
					}
					if !strings.EqualFold(e.Dir, f.root) {
						t.Error("tool did not run in root")
					}
					if len(e.Args) > 1 && e.Args[1] == "goose" {
						continue
					}
					expected := "-modfile=" + filepath.Join(f.root, "tools/devtools/go.mod")
					if len(e.Args) < 3 || !strings.EqualFold(devtoolsRouteModule(e), expected) {
						t.Errorf("wrong module argument: %v", e.Args)
					}
					if strings.Contains(e.Flags, "-modfile") {
						t.Error("module flag leaked to analysis environment")
					}
				}
			}
			if tools == 0 || !verified {
				t.Error("missing actual verification/tool events")
			}
		})
	}
}

// Removing consumer environment binding must expose the foreign cache and PATH
// executable here, even though the real verifier independently uses its own cache.
func TestDevtoolsConsumerCacheBinding(t *testing.T) {
	for _, entry := range []string{"check-tools.ps1", "generate.ps1", "verify-c11.ps1"} {
		t.Run(entry, func(t *testing.T) {
			f := newDevtoolsRouteFixture(t)
			foreign := filepath.Join(t.TempDir(), "unverified cache B")
			code, out, events := f.run(t, entry, map[string]string{
				"GOMODCACHE": foreign, "GOPATH": filepath.Join(foreign, "gopath"),
				"GOPROXY": "https://untrusted.invalid", "GOSUMDB": "off",
				"GOENV": filepath.Join(foreign, "goenv"), "GOWORK": filepath.Join(foreign, "go.work"),
				"GOTOOLCHAIN": "auto", "GOFLAGS": "-tags=cache_canary", "GOAUTH": "netrc", "GOVCS": "*:all",
			})
			want := 0
			if entry == "verify-c11.ps1" {
				want = 47
			}
			if code != want {
				t.Fatalf("consumer exit %d want %d: %s", code, want, out)
			}
			tools, plugins := 0, 0
			for _, e := range events {
				if e.Tool != "go" {
					continue
				}
				if len(e.Args) == 0 || e.Args[0] != "tool" {
					continue
				}
				tools++
				if slices.Contains(e.Args, "protoc-gen-go") {
					plugins++
				}
				cache := filepath.Join(f.profile, "go/pkg/mod")
				pinned := filepath.Join(cache, "golang.org/toolchain@v0.0.1-go1.26.5.windows-amd64/bin/go.exe")
				if !strings.EqualFold(e.Cache, cache) || !strings.EqualFold(e.GoPath, filepath.Join(f.profile, "go")) || !strings.EqualFold(e.Exe, pinned) {
					t.Errorf("verified cache/toolchain not bound to execution: %+v", e)
				}
				if e.Proxy != "off" || e.SumDB != "off" || e.GoEnv != "off" || e.Work != "off" || e.Toolchain != "local" || e.Auth != "off" || e.VCS != "all:off" || e.Flags != "-mod=readonly" {
					t.Errorf("tool inherited unverified configuration: %+v", e)
				}
			}
			if tools == 0 || (entry != "check-tools.ps1" && plugins == 0) {
				t.Fatal("missing tool/plugin execution")
			}
		})
	}
}

func TestDevtoolsConsumerEnvironmentRestored(t *testing.T) {
	f := newDevtoolsRouteFixture(t)
	entry := filepath.Join(f.root, "scripts", "restore-check.ps1")
	body := `
$ErrorActionPreference='Stop'
function Snapshot { return ((Get-ChildItem Env: | Where-Object { $_.Name -match '^GO' -or $_.Name -eq 'PATH' } | Sort-Object Name | ForEach-Object { $_.Name+'='+$_.Value }) -join "\n") }
$before=Snapshot
& (Join-Path $PSScriptRoot 'check-tools.ps1')
if ($LASTEXITCODE -ne 0) { throw 'consumer failed' }
if ($before -cne (Snapshot)) { throw 'consumer changed caller environment' }
. (Join-Path $PSScriptRoot 'private/devtools-process.ps1')
$saved=Initialize-DevtoolsExecutionEnvironment
try { throw 'simulated caller failure' } catch {} finally { Restore-DevtoolsExecutionEnvironment -Saved $saved }
if ($before -cne (Snapshot)) { throw 'failure path changed caller environment' }
Write-Output 'environment restored'
`
	if err := os.WriteFile(entry, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	code, out, _ := f.run(t, "restore-check.ps1", map[string]string{"GOMODCACHE": filepath.Join(t.TempDir(), "original cache"), "GOFLAGS": "-tags=original", "GOPROXY": "https://untrusted.invalid"})
	if code != 0 || !strings.Contains(out, "environment restored") {
		t.Fatalf("environment restoration failed %d: %s", code, out)
	}
}

func TestDevtoolsFailureStopsConsumers(t *testing.T) {
	for _, entry := range []string{"check-tools.ps1", "generate.ps1", "verify-c11.ps1"} {
		t.Run(entry, func(t *testing.T) {
			for _, kind := range []string{"missing-lock", "module-override"} {
				t.Run(kind, func(t *testing.T) {
					f := newDevtoolsRouteFixture(t)
					extra := map[string]string{}
					if kind == "module-override" {
						extra["GOFLAGS"] = "-modfile=untrusted.mod"
					} else if err := os.Remove(filepath.Join(f.root, "tools/devtools/go.sum")); err != nil {
						t.Fatal(err)
					}
					code, _, events := f.run(t, entry, extra)
					if code == 0 {
						t.Error("missing lock was accepted")
					}
					if len(events) != 0 {
						t.Errorf("work before successful verifier: %v", events)
					}
				})
			}
		})
	}
}

func TestDevtoolsNestedProtoc(t *testing.T) {
	t.Run("failure-stops-generation", func(t *testing.T) {
		f := newDevtoolsRouteFixture(t)
		code, _, events := f.run(t, "generate.ps1", map[string]string{"DEVTOOLS_PLUGIN_FAIL": "1"})
		if code == 0 {
			t.Error("nested plugin failure accepted")
		}
		for _, e := range events {
			if e.Tool == "gofmt" || slices.Contains(e.Args, "oapi-codegen") || slices.Contains(e.Args, "sqlc") {
				t.Errorf("work after plugin failure: %+v", e)
			}
		}
	})
	f := newDevtoolsRouteFixture(t)
	code, out, events := f.run(t, "generate.ps1", nil)
	if code != 0 {
		t.Fatalf("generate failed %d: %s", code, out)
	}
	found := false
	for _, e := range events {
		if len(e.Args) >= 3 && e.Args[2] == "protoc-gen-go" {
			found = true
			if !strings.EqualFold(devtoolsRouteModule(e), "-modfile="+filepath.Join(f.root, "tools/devtools/go.mod")) {
				t.Errorf("nested plugin chose wrong module: %v", e.Args)
			}
		}
	}
	if !found {
		t.Error("Buf did not invoke nested plugin")
	}
}

func devtoolsRouteModule(e devtoolsRouteEvent) string {
	if len(e.Args) < 2 || !strings.HasPrefix(e.Args[1], "-modfile=") {
		return ""
	}
	path := strings.TrimPrefix(e.Args[1], "-modfile=")
	if !filepath.IsAbs(path) {
		path = filepath.Join(e.Dir, path)
	}
	return "-modfile=" + filepath.Clean(path)
}

func TestDevtoolsConsumerRuntime(t *testing.T) {
	for _, entry := range []string{"check-tools.ps1", "generate.ps1", "verify-c11.ps1"} {
		t.Run(entry, func(t *testing.T) {
			for _, kind := range []string{"legacy", "non-windows"} {
				t.Run(kind, func(t *testing.T) {
					f := newDevtoolsRouteFixture(t)
					if kind == "legacy" {
						f.host = filepath.Join(os.Getenv("SystemRoot"), "System32/WindowsPowerShell/v1.0/powershell.exe")
					} else {
						path := filepath.Join(f.root, "scripts", entry)
						data, err := os.ReadFile(path)
						if err != nil {
							t.Fatal(err)
						}
						read := []byte("[Environment]::OSVersion.Platform")
						if bytes.Count(data, read) != 1 {
							t.Fatal("consumer has no unique platform guard for non-Windows rejection")
						}
						if err := os.WriteFile(path, bytes.Replace(data, read, []byte("[PlatformID]::Unix"), 1), 0600); err != nil {
							t.Fatal(err)
						}
					}
					code, out, events := f.run(t, entry, nil)
					if code != 1 || strings.TrimSpace(out) != strings.TrimSuffix(entry, ".ps1")+": PowerShell 7.6.5 required." || len(events) != 0 {
						t.Errorf("runtime started work or wrong rejection: %d %q %v", code, out, events)
					}
				})
			}
		})
	}
}

func TestDevtoolsSameHostDispatch(t *testing.T) {
	for _, entry := range []string{"check-tools.ps1", "generate.ps1", "verify-c11.ps1"} {
		t.Run(entry, func(t *testing.T) {
			f := newDevtoolsRouteFixture(t)
			marker := filepath.Join(f.root, "hosts.jsonl")
			for _, name := range []string{"verify-devtools.ps1", "check-tools.ps1", "generate.ps1", "verify-c11.ps1"} {
				path := filepath.Join(f.root, "scripts", name)
				data, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				prefix := "[IO.File]::AppendAllText($env:DEVTOOLS_HOST_LOG, ((@{Name='" + name + "'; Version=$PSVersionTable.PSVersion.ToString(); Executable=[Environment]::ProcessPath} | ConvertTo-Json -Compress) + \"`n\"))\n"
				if err := os.WriteFile(path, append([]byte(prefix), data...), 0600); err != nil {
					t.Fatal(err)
				}
			}
			code, out, events := f.run(t, entry, map[string]string{"DEVTOOLS_HOST_LOG": marker})
			want := 0
			if entry == "verify-c11.ps1" {
				want = 47
			}
			if code != want {
				t.Fatalf("host fixture failed: %d %s", code, out)
			}
			data, err := os.ReadFile(marker)
			if err != nil {
				t.Fatal(err)
			}
			seen := map[string]int{}
			for _, line := range bytes.Split(bytes.TrimSpace(data), []byte("\n")) {
				var host struct{ Name, Version, Executable string }
				if err := json.Unmarshal(line, &host); err != nil {
					t.Fatal(err)
				}
				seen[host.Name]++
				if host.Version != "7.6.5" || !strings.EqualFold(host.Executable, f.host) {
					t.Errorf("host changed: %+v", host)
				}
			}
			if seen["verify-devtools.ps1"] == 0 {
				t.Error("verifier child was not dispatched")
			}
			if entry == "verify-c11.ps1" && (seen["check-tools.ps1"] != 1 || seen["generate.ps1"] != 1 || seen["verify-devtools.ps1"] != 3) {
				t.Errorf("missing nested entries: %v", seen)
			}
			for _, e := range events {
				if e.Tool == "pwsh" || e.Tool == "powershell" {
					t.Error("PATH interpreter decoy executed")
				}
			}
		})
	}
}

func TestDevtoolsConsumerOwnership(t *testing.T) {
	for _, entry := range []string{"check-tools.ps1", "generate.ps1", "verify-c11.ps1"} {
		t.Run(entry, func(t *testing.T) {
			f := newDevtoolsRouteFixture(t)
			if os.Getenv("DEVTOOLS_TEST_TIMING") == "1" {
				traceDevtoolsFixtureTiming(t, f)
			}
			marker := filepath.Join(f.root, "owned.txt")
			foreign := exec.Command(filepath.Join(f.bin, "go.exe"), "owned-child")
			if err := foreign.Start(); err != nil {
				t.Fatal(err)
			}
			foreignDone := make(chan error, 1)
			go func() { foreignDone <- foreign.Wait() }()
			t.Cleanup(func() { _ = foreign.Process.Kill(); <-foreignDone })
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, f.host, "-NoProfile", "-NonInteractive", "-File", filepath.Join(f.root, "scripts", entry))
			cmd.Env = devtoolsFixtureEnv(map[string]string{"USERPROFILE": f.profile, "PATH": f.bin + string(os.PathListSeparator) + os.Getenv("PATH"), "DEVTOOLS_ROUTE_LOG": f.log, "DEVTOOLS_ROUTE_HOLD": marker, "GOENV": "off"})
			cmd.WaitDelay = 500 * time.Millisecond
			var output bytes.Buffer
			cmd.Stdout = &output
			cmd.Stderr = &output
			if err := cmd.Start(); err != nil {
				t.Fatal(err)
			}
			done := make(chan error, 1)
			go func() { done <- cmd.Wait() }()
			t.Cleanup(func() {
				_ = cmd.Process.Kill()
				select {
				case <-done:
				case <-time.After(time.Second):
				}
			})
			var owned []*os.Process
			readyStarted := time.Now()
			until := readyStarted.Add(12 * time.Second)
			for len(owned) == 0 {
				if data, err := os.ReadFile(marker); err == nil {
					fields := strings.Fields(string(data))
					if len(fields) == 2 {
						for _, field := range fields {
							pid, err := strconv.Atoi(field)
							if err != nil {
								t.Fatal(err)
							}
							p, err := os.FindProcess(pid)
							if err != nil {
								t.Fatal(err)
							}
							owned = append(owned, p)
							t.Cleanup(func() { _ = p.Kill() })
						}
						break
					}
				}
				if time.Now().After(until) {
					// Stop our process and join its output writer before inspecting
					// diagnostics; reading output while Wait runs would race.
					_ = cmd.Process.Kill()
					select {
					case <-done:
						events, _ := os.ReadFile(f.log)
						t.Fatalf("owned tool never became ready within 12s; output=%s; events=%s", output.String(), events)
					case <-time.After(time.Second):
						t.Fatal("owned tool never became ready within 12s; diagnostic output writer did not stop")
					}
				}
				time.Sleep(10 * time.Millisecond)
			}
			t.Logf("owned tool ready after %s", time.Since(readyStarted))
			if err := cmd.Process.Kill(); err != nil {
				t.Fatal(err)
			}
			select {
			case err := <-done:
				if err == nil {
					t.Error("killed consumer succeeded")
				}
			case <-time.After(2 * time.Second):
				t.Error("consumer did not exit")
			}
			for _, p := range owned {
				ended := make(chan struct{})
				go func() { _, _ = p.Wait(); close(ended) }()
				select {
				case <-ended:
				case <-time.After(2 * time.Second):
					_ = p.Kill()
					<-ended
					t.Errorf("owned process survived consumer cancellation: %d", p.Pid)
				}
			}
			select {
			case err := <-foreignDone:
				foreignDone <- err
				t.Error("unrelated process terminated")
			default:
			}
		})
	}
}

// Opt-in diagnostics instrument only disposable script copies. They do not
// change readiness deadlines or serve as unmodified-entry acceptance evidence.
func traceDevtoolsFixtureTiming(t *testing.T, f *devtoolsRouteFixture) {
	t.Helper()
	dir := filepath.Join(f.root, "timing")
	if err := os.Mkdir(dir, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"verify-c11.ps1", "verify-devtools.ps1", "check-tools.ps1", "generate.ps1"} {
		path := filepath.Join(f.root, "scripts", name)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		trace := func(label string) string {
			return "[IO.File]::AppendAllText('" + strings.ReplaceAll(dir, "'", "''") + "/' + $PID + '.log', [DateTime]::UtcNow.ToString('o') + ' " + name + " " + label + "' + [Environment]::NewLine)\n"
		}
		body := string(data)
		anchor := "  Initialize-DevtoolsProcessOwnership"
		if strings.Count(body, anchor) != 1 {
			t.Fatalf("timing initialization anchor changed: %s", name)
		}
		body = trace("entry") + strings.Replace(body, anchor, trace("init-start")+anchor+"\n"+trace("init-end"), 1)
		if name == "verify-devtools.ps1" {
			anchor = "  $remaining = [Math]::Floor"
			if strings.Count(body, anchor) != 1 {
				t.Fatal("timing command anchor changed")
			}
			body = strings.Replace(body, anchor, trace("command")+anchor, 1)
		}
		if err := os.WriteFile(path, []byte(body), 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Cleanup(func() {
		files, err := os.ReadDir(dir)
		if err != nil {
			t.Error(err)
			return
		}
		for _, file := range files {
			data, err := os.ReadFile(filepath.Join(dir, file.Name()))
			if err != nil {
				t.Error(err)
				continue
			}
			t.Logf("timing %s:\n%s", file.Name(), data)
		}
	})
}

func TestDevtoolsLintTargetsRoot(t *testing.T) {
	f := newDevtoolsRouteFixture(t)
	code, out, events := f.run(t, "verify-c11.ps1", nil)
	if code != 47 {
		t.Fatalf("lint boundary %d: %s", code, out)
	}
	found := false
	for _, e := range events {
		if slices.Contains(e.Args, "golangci-lint") && slices.Contains(e.Args, "run") {
			found = true
			if !devtoolsLintTargetsRoot(e, f.root) {
				t.Errorf("lint target changed: %+v", e)
			}
			if err := verifyDevtoolsRootOnlyPackage(t, e.Dir); err != nil {
				t.Errorf("lint root-only package: %v", err)
			}
		}
	}
	if !found {
		t.Error("lint never ran")
	}
}

func devtoolsLintTargetsRoot(e devtoolsRouteEvent, root string) bool {
	return e.Dir == root && len(e.Args) > 0 && e.Args[len(e.Args)-1] == "./..." && !strings.Contains(e.Flags, "-modfile")
}

func TestDevtoolsLintWrongModuleRejected(t *testing.T) {
	f := newDevtoolsRouteFixture(t)
	path := filepath.Join(f.root, "scripts", "verify-c11.ps1")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	old := "Invoke-C11Stage -Stage 'golangci-lint' -FilePath 'go' -ArgumentList"
	if strings.Count(string(data), old) != 1 {
		t.Fatal("lint mutation anchor changed")
	}
	mutated := strings.Replace(string(data), old, "Invoke-C11Stage -Stage 'golangci-lint' -FilePath 'go' -WorkingDirectory (Join-Path $repoRoot 'tools/devtools') -ArgumentList", 1)
	if err := os.WriteFile(path, []byte(mutated), 0600); err != nil {
		t.Fatal(err)
	}
	code, out, events := f.run(t, "verify-c11.ps1", nil)
	if code != 47 {
		t.Fatalf("mutated lint boundary %d: %s", code, out)
	}
	found := false
	for _, e := range events {
		if slices.Contains(e.Args, "golangci-lint") && slices.Contains(e.Args, "run") {
			found = true
			if e.Dir != filepath.Join(f.root, "tools", "devtools") {
				t.Fatalf("mutation did not reach wrong cwd: %+v", e)
			}
			if devtoolsLintTargetsRoot(e, f.root) {
				t.Error("existing target assertion accepted wrong module")
			}
			if verifyDevtoolsRootOnlyPackage(t, e.Dir) == nil {
				t.Error("root-only package accepted wrong module")
			}
		}
	}
	if !found {
		t.Fatal("mutated lint never ran")
	}
}

func verifyDevtoolsRootOnlyPackage(t *testing.T, dir string) error {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), filepath.Join(dir, "rootonly", "fixture.go"), nil, parser.AllErrors)
	if err != nil {
		return err
	}
	if file.Name.Name != "rootonly" || len(file.Imports) != 0 {
		return fmt.Errorf("unexpected root-only source")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, devtoolsRealGo(t), "list", "-json", "./rootonly")
	cmd.Dir = dir
	cmd.Env = devtoolsFixtureEnv(map[string]string{"GOENV": "off", "GOWORK": "off", "GOTOOLCHAIN": "local", "GOPROXY": "off", "GOSUMDB": "off", "GOAUTH": "off", "GOVCS": "all:off", "GOFLAGS": "-mod=readonly"})
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("root package metadata: %w: %s", err, out)
	}
	var pkg struct {
		ImportPath string
		Module     struct{ Path, Dir string }
	}
	if err := json.Unmarshal(out, &pkg); err != nil {
		return err
	}
	if pkg.ImportPath != "fixture.local/root/rootonly" || pkg.Module.Path != "fixture.local/root" || filepath.Clean(pkg.Module.Dir) != filepath.Clean(dir) {
		return fmt.Errorf("unexpected root package ownership: %+v", pkg)
	}
	return nil
}

// The real scripts and private Job helper run unchanged. This executable doubles
// only external Go/tool work and reports the actual process argument boundary.
const devtoolsRoutingSource = `package main
import("encoding/json";"fmt";"os";"os/exec";"path/filepath";"strings";"syscall";"time")
func batchBackend(name string,args []string) {
 dir:=os.Getenv("DEVTOOLS_ROUTE_BACKEND");if dir=="" {return}
 // Match the original C11 batch boundary: quote shell metacharacters as well
 // as spaces. Default native quoting alone loses the fuzz selector's caret.
 parts:=append([]string{filepath.Join(dir,name+".cmd")},args...)
 for i,v:=range parts {if strings.ContainsAny(v,"\r\n\"%!" ){os.Exit(99)};if i==0||v==""||strings.ContainsAny(v," \t&|<>()^"){parts[i]="\""+v+"\""}}
 interpreter:=filepath.Join(os.Getenv("SystemRoot"),"System32","cmd.exe")
 child:=exec.Command(interpreter)
 child.SysProcAttr=&syscall.SysProcAttr{CmdLine:"\""+interpreter+"\" /d /q /v:off /s /c \""+strings.Join(parts," ")+"\""}
 child.Stdout=os.Stdout;child.Stderr=os.Stderr;child.Env=os.Environ()
 if err:=child.Run();err!=nil {if exit,ok:=err.(*exec.ExitError);ok {os.Exit(exit.ExitCode())};os.Exit(99)}
 os.Exit(0)
}
func main(){
 if len(os.Args)>1&&os.Args[1]=="owned-child" {time.Sleep(60*time.Second);return}
 name:=strings.TrimSuffix(strings.ToLower(filepath.Base(os.Args[0])),".exe");args:=os.Args[1:];cwd,_:=os.Getwd()
 exe,_:=os.Executable()
 event:=struct{Tool string;Args []string;Dir,Flags string;Exe,Cache,GoPath,Proxy,SumDB,GoEnv,Work,Toolchain,Auth,VCS string}{name,args,cwd,os.Getenv("GOFLAGS"),exe,os.Getenv("GOMODCACHE"),os.Getenv("GOPATH"),os.Getenv("GOPROXY"),os.Getenv("GOSUMDB"),os.Getenv("GOENV"),os.Getenv("GOWORK"),os.Getenv("GOTOOLCHAIN"),os.Getenv("GOAUTH"),os.Getenv("GOVCS")}
 f,err:=os.OpenFile(os.Getenv("DEVTOOLS_ROUTE_LOG"),os.O_CREATE|os.O_APPEND|os.O_WRONLY,0600);if err!=nil {os.Exit(90)};_ = json.NewEncoder(f).Encode(event);f.Close()
 if name=="pwsh"||name=="powershell"||name=="docker" {os.Exit(91)}
 if name!="go" {if name=="gofmt" {batchBackend(name,args)};return}
 if len(args)==1&&args[0]=="version" {fmt.Println("go version go1.26.5 windows/amd64");return}
 if len(args)>0&&args[0]=="list" {return}
 if len(args)==2&&args[0]=="mod"&&args[1]=="verify" {fmt.Println("all modules verified");return}
 batchBackend(name,args)
 if len(args)<2||args[0]!="tool" {return}
 if marker:=os.Getenv("DEVTOOLS_ROUTE_HOLD");marker!="" {child:=exec.Command(os.Args[0],"owned-child");if child.Start()!=nil {os.Exit(98)};_ = os.WriteFile(marker,[]byte(fmt.Sprintf("%d %d",os.Getpid(),child.Process.Pid)),0600);time.Sleep(60*time.Second);return}
 pos:=1;if strings.HasPrefix(args[pos],"-modfile="){pos++};if pos>=len(args){os.Exit(92)};tool:=args[pos];tail:=args[pos+1:]
 if tool=="protoc-gen-go"&&len(tail)==0&&os.Getenv("DEVTOOLS_PLUGIN_FAIL")=="1" {os.Exit(48)}
 if len(tail)==1&&(tail[0]=="--version"||tail[0]=="version"||tail[0]=="-version") {switch tool {case "buf":fmt.Println("1.72.0");case "protoc-gen-go":fmt.Println("protoc-gen-go v1.36.11");case "oapi-codegen":fmt.Println("v2.8.0");case "sqlc":fmt.Println("v1.31.1");case "goose":fmt.Println("goose version: v3.27.1");case "golangci-lint":fmt.Println("golangci-lint has version 2.12.2 built with go1.26.5")};return}
 if tool=="golangci-lint"&&len(tail)>0&&tail[0]=="run" {os.Exit(47)}
 if tool=="buf"&&len(tail)>0&&tail[0]=="generate" {
  data,err:=os.ReadFile("buf.gen.yaml");if err!=nil {os.Exit(93)}
  for _,line:=range strings.Split(string(data),"\n") {if i:=strings.Index(line,"local:");i>=0 {var command []string;if json.Unmarshal([]byte(strings.TrimSpace(line[i+6:])),&command)!=nil||len(command)<2||command[0]!="go" {os.Exit(94)}
   child:=exec.Command(command[0],command[1:]...);child.Env=os.Environ();child.Stdout=os.Stdout;child.Stderr=os.Stderr;if err:=child.Run();err!=nil {os.Exit(95)};return}}
  os.Exit(96)
 }
}
`
