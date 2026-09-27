//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestDevtoolsVerificationBash(t *testing.T) {
	testDevtoolsBashEntryRejection(t, "verify-devtools.sh")
}

// Running the real, unmodified entry catches accidental work, fallbacks, and
// argument/environment bypasses. This file is independently runnable on Linux.
func testDevtoolsBashEntryRejection(t *testing.T, entry string) {
	t.Helper()
	interpreters := []string{`C:\Program Files\Git\bin\bash.exe`, `C:\Program Files\Git\usr\bin\bash.exe`}
	if runtime.GOOS != "windows" {
		bash, err := exec.LookPath("bash")
		if err != nil {
			t.Fatal("actual Bash unavailable: ", err)
		}
		interpreters = []string{bash}
	}
	content, err := os.ReadFile(filepath.Join("..", "..", "scripts", entry))
	if err != nil {
		t.Fatal(err)
	}
	for _, interpreter := range interpreters {
		t.Run(filepath.ToSlash(interpreter), func(t *testing.T) {
			if _, err := os.Stat(interpreter); err != nil {
				t.Fatal("required actual Bash unavailable: ", err)
			}
			for _, scenario := range []string{"missing-module", "complete-module", "read-only-module", "outside-cwd", "extra-arguments", "module-override", "cache-program", "spoofed-platform", "marker-only-path", "empty-path"} {
				t.Run(scenario, func(t *testing.T) {
					root := filepath.Join(t.TempDir(), "repository with spaces")
					bin := filepath.Join(root, "marker tools")
					private := filepath.Join(root, "scripts", "private")
					for _, dir := range []string{bin, private} {
						if err := os.MkdirAll(dir, 0700); err != nil {
							t.Fatal(err)
						}
					}
					write := func(path, value string, mode os.FileMode) {
						t.Helper()
						if err := os.WriteFile(path, []byte(value), mode); err != nil {
							t.Fatal(err)
						}
					}
					module := filepath.Join(root, "tools", "devtools")
					if scenario != "missing-module" {
						if err := os.MkdirAll(module, 0700); err != nil {
							t.Fatal(err)
						}
						for name, value := range map[string]string{"go.mod": "module talenro.local/devtools\n\ngo 1.26.0\n", "go.sum": ""} {
							path := filepath.Join(module, name)
							write(path, value, 0600)
							if scenario == "read-only-module" {
								if err := os.Chmod(path, 0400); err != nil {
									t.Fatal(err)
								}
								t.Cleanup(func() { _ = os.Chmod(path, 0600) })
							}
						}
					}
					marker := filepath.Join(root, "work-events")
					tools := []string{"go", "git", "docker", "pwsh", "powershell", "uname", "mktemp"}
					for _, tool := range tools {
						write(filepath.Join(bin, tool), "#!/bin/bash\nprintf '%s\\n' '"+tool+"' >>\"$DEVTOOLS_REJECTION_MARKER\"\nexit 91\n", 0700)
					}
					helper := filepath.Join(private, "devtools-process.sh")
					write(helper, "printf '%s\\n' helper >>\"$DEVTOOLS_REJECTION_MARKER\"\n", 0600)
					write(filepath.Join(private, "devtools-process.ps1"), "[IO.File]::AppendAllText($env:DEVTOOLS_REJECTION_MARKER, 'helper')\n", 0600)
					entryPath := filepath.Join(root, "scripts", entry)
					write(entryPath, string(content), 0600)
					path := devtoolsRejectionShellPath(bin) + ":/usr/bin:/bin"
					if scenario == "marker-only-path" {
						path = devtoolsRejectionShellPath(bin)
					}
					if scenario == "empty-path" {
						path = ""
					}
					values := map[string]string{"DEVTOOLS_REJECTION_MARKER": devtoolsRejectionShellPath(marker), "DEVTOOLS_REJECTION_PATH": path, "GOENV": "off", "GOTOOLCHAIN": "local", "GOPROXY": "off", "GOSUMDB": "off"}
					if scenario == "module-override" {
						values["GOFLAGS"] = "-modfile=untrusted.mod"
					}
					if scenario == "cache-program" {
						values["GOCACHEPROG"] = "go"
					}
					if scenario == "spoofed-platform" {
						values["OSTYPE"] = "linux-gnu"
					}
					run := func(script, cwd string, args ...string) (int, string, string) {
						t.Helper()
						ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
						defer cancel()
						// Git's launcher can rewrite PATH. Set it in the actual shell,
						// then exec that same interpreter with the untouched entry.
						argv := []string{"--noprofile", "--norc", "-c", `export PATH="$DEVTOOLS_REJECTION_PATH"; exec "$BASH" --noprofile --norc "$@"`, "devtools-test", devtoolsRejectionShellPath(script)}
						cmd := exec.CommandContext(ctx, interpreter, append(argv, args...)...)
						cmd.Dir, cmd.Env = cwd, devtoolsRejectionEnv(values)
						cmd.WaitDelay = 300 * time.Millisecond
						var stdout, stderr bytes.Buffer
						cmd.Stdout, cmd.Stderr = &stdout, &stderr
						err := cmd.Run()
						if ctx.Err() != nil || cmd.ProcessState == nil {
							t.Fatalf("entry failed to run: %v / %v", err, ctx.Err())
						}
						return cmd.ProcessState.ExitCode(), stdout.String(), stderr.String()
					}
					// Prove every marker works before relying on absence of events.
					control := filepath.Join(root, "control.sh")
					write(control, strings.Join(tools, "\n")+"\n. '"+devtoolsRejectionShellPath(helper)+"'\n", 0600)
					values["DEVTOOLS_REJECTION_PATH"] = devtoolsRejectionShellPath(bin)
					code, _, stderr := run(control, root)
					data, err := os.ReadFile(marker)
					if code != 0 || err != nil || string(data) != strings.Join(append(tools, "helper"), "\n")+"\n" {
						t.Fatalf("marker control failed: %d %v %q %q", code, err, data, stderr)
					}
					if err := os.Remove(marker); err != nil {
						t.Fatal(err)
					}
					values["DEVTOOLS_REJECTION_PATH"] = path
					cwd := root
					if scenario == "outside-cwd" {
						cwd = t.TempDir()
					}
					var args []string
					if scenario == "extra-arguments" {
						args = []string{"--help", "--skip-verification", "--", "argument with spaces"}
					}
					code, stdout, stderr := run(entryPath, cwd, args...)
					want := strings.TrimSuffix(entry, ".sh") + ": this release requires Windows and PowerShell 7.6.5.\n"
					if code != 1 || stdout != "" || strings.ReplaceAll(stderr, "\r\n", "\n") != want {
						t.Errorf("rejection mismatch: exit=%d stdout=%q stderr=%q", code, stdout, stderr)
					}
					if _, err := os.Stat(marker); !os.IsNotExist(err) {
						t.Errorf("unsupported entry started work: %v", err)
					}
				})
			}
		})
	}
}

func devtoolsRejectionShellPath(path string) string {
	path = filepath.ToSlash(path)
	if runtime.GOOS == "windows" && len(path) > 2 && path[1] == ':' {
		return "/" + strings.ToLower(path[:1]) + path[2:]
	}
	return path
}

func devtoolsRejectionEnv(values map[string]string) []string {
	var env []string
	for _, value := range os.Environ() {
		key, _, _ := strings.Cut(value, "=")
		upper := strings.ToUpper(key)
		if strings.HasPrefix(upper, "BASH_FUNC_") || upper == "BASH_ENV" || upper == "ENV" || upper == "SHELLOPTS" || upper == "BASHOPTS" {
			continue
		}
		if _, replaced := values[upper]; !replaced {
			env = append(env, value)
		}
	}
	for key, value := range values {
		env = append(env, key+"="+value)
	}
	return env
}
