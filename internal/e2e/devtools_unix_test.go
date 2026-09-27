//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

// Real Bash is under test. Only Go is doubled for controlled failures and
// long-lived children; integrity tests separately use real Go. Removing runtime
// validation, bounded capture or owned-tree cleanup must fail these cases.
// Process objects are retained while the fixture is known to be alive;
// /proc observes exit (including zombies), never discovers kill targets.
func testDevtoolsUnixProcess(t *testing.T) {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("process observer requires Linux; other Unix is unverified")
	}
	bin := t.TempDir()
	source := filepath.Join(bin, "fixture.go")
	if err := os.WriteFile(source, []byte(devtoolsUnixProcessSource), 0600); err != nil {
		t.Fatal(err)
	}
	goPath := filepath.Join(bin, "go")
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", goPath, source)
	build.Env = devtoolsFixtureEnv(map[string]string{"GOENV": "off", "GOWORK": "off", "GOTOOLCHAIN": "local", "GOPROXY": "off", "GOSUMDB": "off", "CGO_ENABLED": "0", "GOCACHE": t.TempDir()})
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("fixture build: %v: %s", err, output)
	}
	for _, tc := range []struct {
		name string
		want int
	}{
		{"natural-exit", 0}, {"go-exit-7", 7}, {"wrong-version", 1}, {"windows-toolchain", 1},
		{"output-canary", 7}, {"timeout-with-child", 124}, {"parent-cancel", -1},
		{"output-limit", 1}, {"stderr-limit", 1}, {"natural-exit-with-child", 0},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := filepath.Join(t.TempDir(), "repository with spaces")
			module := filepath.Join(root, "tools", "devtools")
			for _, dir := range []string{module, filepath.Join(root, "scripts"), filepath.Join(root, "sink")} {
				if err := os.MkdirAll(dir, 0700); err != nil {
					t.Fatal(err)
				}
			}
			for name, value := range map[string]string{"go.mod": "module talenro.local/devtools\n\ngo 1.26.0\n", "go.sum": ""} {
				if err := os.WriteFile(filepath.Join(module, name), []byte(value), 0600); err != nil {
					t.Fatal(err)
				}
			}
			content, err := os.ReadFile(filepath.Join("..", "..", "scripts", "verify-devtools.sh"))
			if err != nil {
				t.Fatal(err)
			}
			const deadline = "deadline=$((SECONDS + 900))"
			if bytes.Count(content, []byte(deadline)) != 1 {
				t.Fatal("expected exactly one private deadline replacement")
			}
			seconds := 30
			if tc.name == "timeout-with-child" {
				seconds = 2
			}
			content = bytes.Replace(content, []byte(deadline), []byte(fmt.Sprintf("deadline=$((SECONDS + %d))", seconds)), 1)
			entry := filepath.Join(root, "scripts", "verify-devtools.sh")
			if err := os.WriteFile(entry, content, 0600); err != nil {
				t.Fatal(err)
			}
			marker := filepath.Join(root, "owned")
			runCtx, runCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer runCancel()
			command := exec.CommandContext(runCtx, "bash", entry)
			command.Env = devtoolsFixtureEnv(map[string]string{"PATH": bin + string(os.PathListSeparator) + os.Getenv("PATH"), "HOME": root, "TMPDIR": filepath.Join(root, "sink"), "DEVTOOLS_UNIX_CASE": tc.name, "DEVTOOLS_UNIX_MARKER": marker})
			command.Dir = t.TempDir()
			command.WaitDelay = 300 * time.Millisecond
			var output bytes.Buffer
			command.Stdout, command.Stderr = &output, &output
			foreign := exec.Command(goPath, "child")
			if err := foreign.Start(); err != nil {
				t.Fatal(err)
			}
			foreignDone := make(chan error, 1)
			go func() { foreignDone <- foreign.Wait() }()
			defer func() { _ = foreign.Process.Kill(); <-foreignDone }()
			if err := command.Start(); err != nil {
				t.Fatal(err)
			}
			defer func() { _ = command.Process.Kill() }()
			var owned []*os.Process
			defer func() {
				for _, p := range owned {
					_ = p.Kill()
					_ = p.Release()
				}
			}()
			withChild := strings.Contains(tc.name, "with-child") || tc.name == "parent-cancel"
			if withChild {
				ready := time.Now().Add(4 * time.Second)
				for {
					data, readErr := os.ReadFile(marker)
					fields := strings.Fields(string(data))
					if readErr == nil && len(fields) == 2 {
						for _, value := range fields {
							pid, err := strconv.Atoi(value)
							if err != nil {
								t.Fatal(err)
							}
							p, err := os.FindProcess(pid)
							if err != nil {
								t.Fatal(err)
							}
							owned = append(owned, p)
						}
						break
					}
					if time.Now().After(ready) {
						_ = command.Process.Kill()
						_ = command.Wait()
						t.Fatal("owned process readiness missing")
					}
					time.Sleep(10 * time.Millisecond)
				}
				if tc.name == "parent-cancel" {
					if err := command.Process.Kill(); err != nil {
						t.Fatal(err)
					}
				}
			}
			err = command.Wait()
			for _, p := range owned {
				end := time.Now().Add(2 * time.Second)
				for !devtoolsLinuxExited(p.Pid) && time.Now().Before(end) {
					time.Sleep(10 * time.Millisecond)
				}
				if !devtoolsLinuxExited(p.Pid) {
					t.Errorf("owned process %d survived %s; retained handle cleanup follows", p.Pid, tc.name)
				}
			}
			select {
			case foreignErr := <-foreignDone:
				foreignDone <- foreignErr
				t.Error("unrelated canary exited")
			default:
			}
			if runCtx.Err() != nil {
				t.Errorf("verifier did not finish within independent 10s test guard: %v", runCtx.Err())
			}
			if tc.want == -1 {
				if err == nil {
					t.Error("forced cancellation succeeded")
				}
			} else {
				if command.ProcessState == nil || command.ProcessState.ExitCode() != tc.want {
					t.Errorf("want exit %d, got %v; output %q", tc.want, err, output.String())
				}
				want := "verify-devtools: passed"
				if tc.want != 0 {
					want = fmt.Sprintf("verify-devtools: toolchain failed with exit code %d.", tc.want)
				}
				if strings.TrimSpace(output.String()) != want {
					t.Errorf("expected sanitized %q, got %q", want, output.String())
				}
			}
			files, err := os.ReadDir(filepath.Join(root, "sink"))
			if err != nil || len(files) != 0 {
				t.Errorf("owned capture sinks not cleaned: entries=%d err=%v", len(files), err)
			}
		})
	}
}

func devtoolsLinuxExited(pid int) bool {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
	if os.IsNotExist(err) {
		return true
	}
	if err != nil {
		return false
	}
	end := strings.LastIndexByte(string(data), ')')
	return end >= 0 && len(data) > end+2 && (data[end+2] == 'Z' || data[end+2] == 'X')
}

const devtoolsUnixProcessSource = `package main
import("fmt";"os";"os/exec";"strings";"time")
func main(){
 if len(os.Args)>1 && os.Args[1]=="child" {time.Sleep(60*time.Second);return}
 mode:=os.Getenv("DEVTOOLS_UNIX_CASE")
 if len(os.Args)>1 && os.Args[1]=="version" {
  switch mode {
  case "go-exit-7":os.Exit(7)
  case "output-canary":fmt.Fprintln(os.Stderr,"PRIVATE_UNIX_CANARY");os.Exit(7)
  case "wrong-version":fmt.Println("go version go1.26.4 linux/amd64");return
  case "windows-toolchain":fmt.Println("go version go1.26.5 windows/amd64");return
  case "output-limit","stderr-limit":
   out:=os.Stdout;if mode=="stderr-limit" {out=os.Stderr};fmt.Fprint(out,strings.Repeat("PRIVATE_UNIX_CANARY",300000));time.Sleep(60*time.Second);return
  case "timeout-with-child","parent-cancel","natural-exit-with-child":
   self,_:=os.Executable(); child:=exec.Command(self,"child");if child.Start()!=nil {os.Exit(9)}
   if os.WriteFile(os.Getenv("DEVTOOLS_UNIX_MARKER"),[]byte(fmt.Sprintf("%d %d",os.Getpid(),child.Process.Pid)),0600)!=nil {os.Exit(9)}
   if mode!="natural-exit-with-child" {time.Sleep(60*time.Second)} else {time.Sleep(100*time.Millisecond)}
  }
  fmt.Println("go version go1.26.5 linux/amd64");return
 }
 if len(os.Args)>1 && os.Args[1]=="list" {return}
 if len(os.Args)==3 && os.Args[1]=="mod" && os.Args[2]=="verify" {fmt.Println("all modules verified");return}
 os.Exit(8)
}`
