//go:build e2e

package e2e

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestDevtoolsVerificationPowerShell(t *testing.T) {
	testDevtoolsRuntime(t)
	testDevtoolsOwnership(t)
	testDevtoolsMissingModule(t, "ps1")
	testDevtoolsIntegrity(t, "ps1")
	testDevtoolsProcess(t, "ps1")
}

func TestDevtoolsVerificationBash(t *testing.T) {
	if runtime.GOOS == "windows" {
		testDevtoolsBashRejection(t)
		return
	}
	testDevtoolsMissingModule(t, "sh")
	testDevtoolsIntegrity(t, "sh")
	testDevtoolsProcess(t, "sh")
}

// Only process/error contracts use this executable; integrity tests above always
// run the real pinned Go implementation against real module archives.
func testDevtoolsProcess(t *testing.T, extension string) {
	t.Helper()
	if runtime.GOOS != "windows" {
		t.Skip("Windows process fixture")
	}
	profile := filepath.Join(t.TempDir(), "owned profile")
	goPath := filepath.Join(profile, "go", "pkg", "mod", "golang.org", "toolchain@v0.0.1-go1.26.5.windows-amd64", "bin", "go.exe")
	if err := os.MkdirAll(filepath.Dir(goPath), 0700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(t.TempDir(), "process.go")
	if err := os.WriteFile(source, []byte(devtoolsProcessSource), 0600); err != nil {
		t.Fatal(err)
	}
	realGo := filepath.Join(os.Getenv("USERPROFILE"), "go", "pkg", "mod", "golang.org", "toolchain@v0.0.1-go1.26.5.windows-amd64", "bin", "go.exe")
	buildCtx, buildCancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer buildCancel()
	build := exec.CommandContext(buildCtx, realGo, "build", "-o", goPath, source)
	build.Env = devtoolsFixtureEnv(map[string]string{"GOENV": "off", "GOWORK": "off", "GOTOOLCHAIN": "local", "GOOS": "windows", "GOPROXY": "off", "GOSUMDB": "off", "CGO_ENABLED": "0"})
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("process fixture build: %v: %s", err, output)
	}
	for _, tc := range []struct {
		name  string
		want  int
		stage string
	}{
		{"go-exit-7", 7, "toolchain"}, {"wrong-version", 1, "toolchain"}, {"output-canary", 7, "toolchain"},
		{"timeout", 124, "toolchain"}, {"natural-exit", 0, ""},
		{"timeout-with-child", 124, "toolchain"}, {"parent-cancel", -1, ""},
		{"output-limit", 1, "toolchain"}, {"stderr-limit", 1, "toolchain"},
		{"ownership-init-failure", 1, "toolchain"},
		{"environment-restore-success", 0, ""}, {"environment-restore-failure", 7, "toolchain"},
		{"launcher-cancel", -1, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.name == "launcher-cancel" && extension != "sh" {
				t.Skip("Git for Windows launcher boundary only")
			}
			root := filepath.Join(t.TempDir(), "repository with spaces")
			module := filepath.Join(root, "tools", "devtools")
			if err := os.MkdirAll(module, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(root, "scripts"), 0700); err != nil {
				t.Fatal(err)
			}
			for name, content := range map[string]string{"go.mod": "module talenro.local/devtools\n\ngo 1.26.0\n", "go.sum": ""} {
				if err := os.WriteFile(filepath.Join(module, name), []byte(content), 0600); err != nil {
					t.Fatal(err)
				}
			}
			content, err := os.ReadFile(filepath.Join("..", "..", "scripts", "verify-devtools."+extension))
			if err != nil {
				t.Fatal(err)
			}
			// Shorten only this owned test copy, never a public production flag.
			deadline := ".AddSeconds(900)"
			replacement := ".AddSeconds(2)"
			if extension == "sh" {
				deadline = "deadline=$((SECONDS + 900))"
				replacement = "deadline=$((SECONDS + 2))"
			}
			if tc.name == "parent-cancel" || tc.name == "launcher-cancel" || tc.name == "output-limit" || tc.name == "stderr-limit" {
				if extension == "sh" {
					replacement = "deadline=$((SECONDS + 30))"
				} else {
					replacement = ".AddSeconds(30)"
				}
			}
			if bytes.Count(content, []byte(deadline)) != 1 {
				t.Fatal("short-deadline fixture must replace exactly one deadline")
			}
			content = bytes.Replace(content, []byte(deadline), []byte(replacement), 1)
			restoreCase := strings.HasPrefix(tc.name, "environment-restore-")
			if restoreCase && extension == "ps1" {
				const finalOutput = "if ($exitCode -eq 0) { [Console]::Out.WriteLine('verify-devtools: passed') }"
				if bytes.Count(content, []byte(finalOutput)) != 1 {
					t.Fatal("environment observation requires exactly one final-output point")
				}
				observation := "[IO.File]::WriteAllText($env:DEVTOOLS_ENV_MARKER, ($env:GOPROXY + '|' + $env:GOFLAGS + '|' + $env:GOWORK + '|' + $env:GOCACHEPROG))\n"
				content = bytes.Replace(content, []byte(finalOutput), []byte(observation+finalOutput), 1)
			}
			if extension == "sh" && (tc.name == "parent-cancel" || tc.name == "launcher-cancel") {
				content = bytes.Replace(content, []byte("stage=module"), []byte("printf 'script-posix=%s native=%s parent=%s\\n' \"$$\" \"$(</proc/$$/winpid)\" \"$PPID\" >\"${DEVTOOLS_PROCESS_TRACE}\"\nstage=module"), 1)
			}
			entry := filepath.Join(root, "scripts", "verify-devtools."+extension)
			if err := os.WriteFile(entry, content, 0600); err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			var command *exec.Cmd
			if extension == "ps1" {
				copyDevtoolsProcessHelper(t, root)
				if tc.name == "ownership-init-failure" {
					if err := os.WriteFile(filepath.Join(root, "scripts", "private", "devtools-process.ps1"), []byte("function Initialize-DevtoolsProcessOwnership { throw 'OWNERSHIP_PRIVATE_CANARY' }\n"), 0600); err != nil {
						t.Fatal(err)
					}
				}
				command = exec.CommandContext(ctx, devtoolsPowerShell(t), "-NoProfile", "-NonInteractive", "-File", entry)
			} else {
				command = exec.CommandContext(ctx, `C:\Program Files\Git\bin\bash.exe`, filepath.ToSlash(entry))
				if tc.name == "parent-cancel" {
					command = exec.CommandContext(ctx, `C:\Program Files\Git\usr\bin\bash.exe`, filepath.ToSlash(entry))
				}
			}
			marker := filepath.Join(root, "owned-processes")
			// A surviving descendant can inherit the capture pipe; never let Wait
			// hide that leak by waiting until the fixture's natural 60s exit.
			command.WaitDelay = 500 * time.Millisecond
			trace := filepath.Join(root, "process-trace")
			envMarker := filepath.Join(root, "restored-environment")
			env := map[string]string{"USERPROFILE": profile, "DEVTOOLS_PROCESS_CASE": tc.name, "DEVTOOLS_PROCESS_MARKER": marker, "DEVTOOLS_PROCESS_TRACE": trace, "GOENV": "off"}
			if restoreCase {
				env["DEVTOOLS_ENV_MARKER"] = envMarker
				env["GOPROXY"] = "https://fixture.invalid"
				env["GOFLAGS"] = "-trimpath"
				env["GOWORK"] = "fixture-original.work"
				env["GOCACHEPROG"] = "fixture-never-execute"
			}
			command.Env = devtoolsFixtureEnv(env)
			var captured bytes.Buffer
			command.Stdout = &captured
			command.Stderr = &captured
			var foreignDone chan error
			if tc.name == "parent-cancel" || tc.name == "timeout-with-child" {
				foreign := exec.Command(goPath, "child")
				if err := foreign.Start(); err != nil {
					t.Fatal(err)
				}
				foreignDone = make(chan error, 1)
				go func() { foreignDone <- foreign.Wait() }()
				t.Cleanup(func() { _ = foreign.Process.Kill(); <-foreignDone })
			}
			if err := command.Start(); err != nil {
				t.Fatal(err)
			}
			var owned []*os.Process
			if tc.name == "timeout-with-child" || tc.name == "parent-cancel" || tc.name == "launcher-cancel" {
				readyDeadline := time.Now().Add(5 * time.Second)
				for {
					data, readErr := os.ReadFile(marker)
					if readErr == nil {
						fields := strings.Fields(string(data))
						if len(fields) == 2 {
							for _, field := range fields {
								pid, err := strconv.Atoi(field)
								if err != nil {
									t.Fatal(err)
								}
								process, err := os.FindProcess(pid)
								if err != nil {
									t.Fatal(err)
								}
								owned = append(owned, process)
							}
							break
						}
					}
					if time.Now().After(readyDeadline) {
						_ = command.Process.Kill()
						_ = command.Wait()
						t.Fatal("owned child did not become ready")
					}
					time.Sleep(10 * time.Millisecond)
				}
				for _, process := range owned {
					t.Cleanup(func() { _ = process.Kill() })
				}
				if tc.name == "parent-cancel" || tc.name == "launcher-cancel" {
					if err := command.Process.Kill(); err != nil {
						t.Fatal(err)
					}
				}
			}
			err = command.Wait()
			output := captured.Bytes()
			if foreignDone != nil {
				select {
				case foreignErr := <-foreignDone:
					foreignDone <- foreignErr // Leave the completion for cleanup.
					t.Errorf("unrelated canary exited during %s: %v", tc.name, foreignErr)
				default:
				}
			}
			for _, process := range owned {
				done := make(chan struct{})
				go func() { _, _ = process.Wait(); close(done) }()
				select {
				case <-done:
				case <-time.After(2 * time.Second):
					_ = process.Kill()
					<-done
					t.Errorf("owned process %d survived %s; fixture cleaned it up", process.Pid, tc.name)
				}
			}
			if ctx.Err() != nil {
				t.Fatalf("process contract exceeded deadline: %v", ctx.Err())
			}
			if tc.name == "parent-cancel" || tc.name == "launcher-cancel" {
				if traceBytes, traceErr := os.ReadFile(trace); traceErr == nil {
					t.Logf("invoked native PID=%d; %s", command.Process.Pid, traceBytes)
				}
				if err == nil {
					t.Error("canceled verifier unexpectedly succeeded")
				}
				return
			}
			if command.ProcessState == nil || command.ProcessState.ExitCode() != tc.want {
				t.Fatalf("want %d, got %v: %s", tc.want, err, output)
			}
			want := "verify-devtools: passed"
			if tc.want != 0 {
				want = fmt.Sprintf("verify-devtools: %s failed with exit code %d.", tc.stage, tc.want)
			}
			if got := strings.TrimSpace(string(output)); got != want {
				t.Fatalf("expected sanitized output %q, got %q", want, got)
			}
			if restoreCase {
				data, err := os.ReadFile(envMarker)
				if err != nil || string(data) != "https://fixture.invalid|-trimpath|fixture-original.work|fixture-never-execute" {
					t.Fatalf("entry did not restore caller environment after %s: %v: %q", tc.name, err, data)
				}
			}
		})
	}
}

const devtoolsProcessSource = `package main
import("fmt";"os";"os/exec";"time";"strings")
func main(){
 if len(os.Args)>1 && os.Args[1]=="child" { time.Sleep(60*time.Second); return }
 if os.Getenv("GOPROXY")!="off" || os.Getenv("GOENV")!="off" || os.Getenv("GOWORK")!="off" || os.Getenv("GOFLAGS")!="-mod=readonly" { os.Exit(9) }
 switch os.Getenv("DEVTOOLS_PROCESS_CASE") {
 case "go-exit-7", "environment-restore-failure": os.Exit(7)
 case "output-canary": fmt.Fprintln(os.Stderr,"secret-canary-password=must-not-escape"); os.Exit(7)
 case "wrong-version": fmt.Println("go version go1.26.4 windows/amd64"); return
 case "timeout": time.Sleep(10*time.Second)
 case "output-limit": fmt.Print(strings.Repeat("secret-output-canary",300000)); time.Sleep(60*time.Second)
 case "stderr-limit": fmt.Fprint(os.Stderr,strings.Repeat("secret-error-canary",300000)); time.Sleep(60*time.Second)
 case "timeout-with-child", "parent-cancel", "launcher-cancel":
  child:=exec.Command(os.Args[0],"child"); if err:=child.Start(); err!=nil { os.Exit(11) }
  if err:=os.WriteFile(os.Getenv("DEVTOOLS_PROCESS_MARKER"),[]byte(fmt.Sprintf("%d %d",os.Getpid(),child.Process.Pid)),0600); err!=nil { os.Exit(12) }
  time.Sleep(60*time.Second)
 }
 if len(os.Args)>1 && os.Args[1]=="version" { time.Sleep(30*time.Millisecond); fmt.Println("go version go1.26.5 windows/amd64"); return }
 if len(os.Args)>1 && os.Args[1]=="list" { return }
 if len(os.Args)==3 && os.Args[1]=="mod" && os.Args[2]=="verify" { fmt.Println("all modules verified"); return }
 os.Exit(8)
}`

// The proxy and cache are exclusively owned by the test. In particular, these
// corruption cases must never open the user's shared module cache for writing.
func testDevtoolsIntegrity(t *testing.T, extension string) {
	t.Helper()
	if runtime.GOOS != "windows" {
		t.Skip("Windows fixed-toolchain fixture; native Bash fixture is separate")
	}
	for _, tc := range []struct {
		name string
		want int
	}{
		{"valid", 0}, {"missing-cache", 1}, {"tampered-directory", 1},
		{"tampered-zip", 1}, {"missing-ziphash", 1}, {"inherited-module-override", 1},
		{"missing-cache-without-presence-check", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "repository with spaces")
			profile := filepath.Join(t.TempDir(), "private profile")
			cache := filepath.Join(profile, "go", "pkg", "mod")
			module := filepath.Join(root, "tools", "devtools")
			write := func(path string, data []byte) {
				t.Helper()
				if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, data, 0600); err != nil {
					t.Fatal(err)
				}
			}
			const dependency = "example.com/Upper/fixture"
			const escaped = "example.com/!upper/fixture"
			const dependencyMod = "module example.com/Upper/fixture\n\ngo 1.26.0\n"
			proxy := filepath.Join(t.TempDir(), "proxy")
			versions := filepath.Join(proxy, filepath.FromSlash(escaped), "@v")
			write(filepath.Join(versions, "v1.0.0.mod"), []byte(dependencyMod))
			write(filepath.Join(versions, "v1.0.0.info"), []byte(`{"Version":"v1.0.0","Time":"2026-01-01T00:00:00Z"}`))
			write(filepath.Join(versions, "list"), []byte("v1.0.0\n"))
			var archive bytes.Buffer
			zw := zip.NewWriter(&archive)
			for _, file := range []struct{ name, content string }{
				{"go.mod", dependencyMod}, {"fixture.go", "package fixture\nconst Value = 1\n"},
			} {
				writer, err := zw.Create(dependency + "@v1.0.0/" + file.name)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := writer.Write([]byte(file.content)); err != nil {
					t.Fatal(err)
				}
			}
			if err := zw.Close(); err != nil {
				t.Fatal(err)
			}
			write(filepath.Join(versions, "v1.0.0.zip"), archive.Bytes())
			write(filepath.Join(module, "go.mod"), []byte("module talenro.local/devtools\n\ngo 1.26.0\n\nrequire "+dependency+" v1.0.0\n"))
			toolchain := filepath.Join(os.Getenv("USERPROFILE"), "go", "pkg", "mod", "golang.org", "toolchain@v0.0.1-go1.26.5.windows-amd64")
			goPath := filepath.Join(toolchain, "bin", "go.exe")
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			prep := exec.CommandContext(ctx, goPath, "mod", "download", dependency)
			prep.Dir = module
			proxyURL := (&url.URL{Scheme: "file", Path: "/" + filepath.ToSlash(proxy)}).String()
			prep.Env = devtoolsFixtureEnv(map[string]string{"GOMODCACHE": cache, "GOPROXY": proxyURL, "GOSUMDB": "off", "GOENV": "off", "GOWORK": "off", "GOTOOLCHAIN": "local", "GOFLAGS": ""})
			if output, err := prep.CombinedOutput(); err != nil {
				t.Fatalf("local proxy preparation: %v: %s", err, output)
			}
			for _, relative := range []string{"bin/go.exe", "go.env", "VERSION"} {
				data, err := os.ReadFile(filepath.Join(toolchain, filepath.FromSlash(relative)))
				if err != nil {
					t.Fatal(err)
				}
				write(filepath.Join(cache, "golang.org", "toolchain@v0.0.1-go1.26.5.windows-amd64", filepath.FromSlash(relative)), data)
			}
			// cmd/go locates a relocated GOROOT by its pkg/tool directory.
			// These module-only commands do not invoke a compiler.
			if err := os.MkdirAll(filepath.Join(cache, "golang.org", "toolchain@v0.0.1-go1.26.5.windows-amd64", "pkg", "tool"), 0700); err != nil {
				t.Fatal(err)
			}
			zipPath := filepath.Join(cache, "cache", "download", filepath.FromSlash(escaped), "@v", "v1.0.0.zip")
			directory := filepath.Join(cache, filepath.FromSlash(escaped)+"@v1.0.0")
			switch tc.name {
			case "missing-cache", "missing-cache-without-presence-check":
				if err := os.RemoveAll(directory); err != nil {
					t.Fatal(err)
				}
				if err := os.Remove(zipPath); err != nil {
					t.Fatal(err)
				}
			case "tampered-directory":
				file := filepath.Join(directory, "fixture.go")
				if err := os.Chmod(file, 0600); err != nil {
					t.Fatal(err)
				}
				write(file, []byte("package fixture\nconst Value = 2\n"))
			case "tampered-zip":
				if err := os.Chmod(zipPath, 0600); err != nil {
					t.Fatal(err)
				}
				write(zipPath, []byte("invalid archive"))
			case "missing-ziphash":
				if err := os.Remove(zipPath + "hash"); err != nil {
					t.Fatal(err)
				}
			}
			content, err := os.ReadFile(filepath.Join("..", "..", "scripts", "verify-devtools."+extension))
			if err != nil {
				t.Fatal(err)
			}
			entry := filepath.Join(root, "scripts", "verify-devtools."+extension)
			if tc.name == "missing-cache-without-presence-check" {
				if extension != "ps1" {
					t.Skip("PowerShell mutation; Unix mutation requires its native fixture")
				}
				const loop = "foreach ($line in ($modules -split \"`n\")) {"
				if bytes.Count(content, []byte(loop)) != 1 {
					t.Fatal("presence-check mutation must replace exactly one loop")
				}
				content = bytes.Replace(content, []byte(loop), []byte("foreach ($line in @()) {"), 1)
			}
			write(entry, content)
			var command *exec.Cmd
			if extension == "ps1" {
				copyDevtoolsProcessHelper(t, root)
				command = exec.CommandContext(ctx, devtoolsPowerShell(t), "-NoProfile", "-NonInteractive", "-File", entry)
			} else {
				command = exec.CommandContext(ctx, `C:\Program Files\Git\bin\bash.exe`, filepath.ToSlash(entry))
			}
			env := map[string]string{"USERPROFILE": profile, "GOPATH": filepath.Join(profile, "go"), "GOENV": "off", "GOPROXY": "off", "GOSUMDB": "off", "GOTOOLCHAIN": "local", "GOFLAGS": ""}
			if tc.name == "inherited-module-override" {
				env["GOFLAGS"] = "-modfile=untrusted.mod"
			}
			command.Env = devtoolsFixtureEnv(env)
			command.Dir = root
			output, runErr := command.CombinedOutput()
			if ctx.Err() != nil {
				t.Fatalf("fixture deadline: %v", ctx.Err())
			}
			if command.ProcessState == nil || command.ProcessState.ExitCode() != tc.want {
				t.Fatalf("want exit %d, got %v: %s", tc.want, runErr, output)
			}
			got := strings.TrimSpace(string(output))
			if tc.want == 0 && got != "verify-devtools: passed" {
				t.Fatalf("unexpected success output: %q", got)
			}
			if tc.want != 0 && (!strings.HasPrefix(got, "verify-devtools: ") || !strings.HasSuffix(got, fmt.Sprintf(" failed with exit code %d.", tc.want)) || strings.ContainsAny(got, "\r\n")) {
				t.Fatalf("unsanitized failure: %q", got)
			}
			stage := map[string]string{"missing-cache": "dependencies", "missing-ziphash": "dependencies", "tampered-directory": "integrity", "tampered-zip": "integrity", "inherited-module-override": "environment"}[tc.name]
			if tc.want != 0 && got != "verify-devtools: "+stage+" failed with exit code 1." {
				t.Fatalf("failure did not exercise intended %s stage: %q", stage, got)
			}
		})
	}
}

func devtoolsFixtureEnv(overrides map[string]string) []string {
	var env []string
	for _, pair := range os.Environ() {
		key, _, _ := strings.Cut(pair, "=")
		if strings.HasPrefix(strings.ToUpper(key), "GO") {
			continue
		}
		replaced := false
		for other := range overrides {
			if strings.EqualFold(key, other) {
				replaced = true
				break
			}
		}
		if !replaced {
			env = append(env, pair)
		}
	}
	for key, value := range overrides {
		env = append(env, key+"="+value)
	}
	return env
}

// A missing tools module must fail before starting Go, without exposing paths or
// inherited output. Removing the module preflight must break this contract.
func testDevtoolsMissingModule(t *testing.T, extension string) {
	t.Helper()
	source := filepath.Join("..", "..", "scripts", "verify-devtools."+extension)
	content, err := os.ReadFile(source)
	if err != nil {
		t.Fatalf("required development-tool verification entry is not implemented: %v", err)
	}
	t.Run("missing-module", func(t *testing.T) {
		root := filepath.Join(t.TempDir(), "repository with spaces")
		scripts := filepath.Join(root, "scripts")
		if err := os.MkdirAll(scripts, 0700); err != nil {
			t.Fatal(err)
		}
		entry := filepath.Join(scripts, "verify-devtools."+extension)
		if err := os.WriteFile(entry, content, 0600); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		var command *exec.Cmd
		if extension == "ps1" {
			command = exec.CommandContext(ctx, devtoolsPowerShell(t), "-NoProfile", "-NonInteractive", "-File", entry)
		} else {
			bash := "bash"
			if runtime.GOOS == "windows" {
				bash = `C:\Program Files\Git\bin\bash.exe`
			}
			command = exec.CommandContext(ctx, bash, filepath.ToSlash(entry))
		}
		command.Dir = t.TempDir() // Resolve relative to the entry, never caller cwd.
		output, err := command.CombinedOutput()
		if ctx.Err() != nil {
			t.Fatalf("missing-module preflight exceeded deadline: %v", ctx.Err())
		}
		if err == nil || command.ProcessState == nil || command.ProcessState.ExitCode() != 1 {
			t.Fatalf("missing-module must exit 1, got %v, output %q", err, output)
		}
		if got := strings.TrimSpace(string(output)); got != "verify-devtools: module failed with exit code 1." {
			t.Fatalf("expected one sanitized failure, got %q", got)
		}
	})
}
