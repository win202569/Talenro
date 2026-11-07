package main

import (
	"reflect"
	"testing"
)

// These cases catch silently selecting the wrong cache identities, accepting
// unapproved names, and sharing mutable allowlist state between invocations.
func TestProfileDefaults(t *testing.T) {
	want := [][2]string{{"golang.org/x/text", "v0.40.0"}, {"modernc.org/sqlite", "v1.49.1"}}
	for _, args := range [][]string{nil, {"-profile", "paired-v1"}} {
		p, _, _, err := ParseProfileArgs(args)
		if err != nil || p != "paired-v1" {
			t.Fatalf("default: %q %v", p, err)
		}
		got, err := ProfileIDs(p)
		if err != nil || !reflect.DeepEqual(got, want) {
			t.Fatalf("paired: %v %v", got, err)
		}
	}
}
func TestProfileLibc(t *testing.T) {
	got, err := ProfileIDs("libc-v1")
	want := [][2]string{{"modernc.org/libc", "v1.72.1"}}
	if err != nil || !reflect.DeepEqual(got, want) {
		t.Fatalf("libc profile missing: got %v err %v", got, err)
	}
	p, s, r, err := ParseProfileArgs([]string{"-profile", "libc-v1", "-session", "abc", "-repo", "missing:/no-such-path"})
	if err != nil || p != "libc-v1" || s != "abc" || r != "missing:/no-such-path" {
		t.Fatalf("CLI selection: %q %q %q %v", p, s, r, err)
	}
}
func TestProfileReject(t *testing.T) {
	for _, name := range []string{"", "unknown", "PAIRED-V1", "LIBC-V1"} {
		if _, err := ProfileIDs(name); err == nil {
			t.Errorf("unapproved profile accepted: %q", name)
		}
		if _, _, _, err := ParseProfileArgs([]string{"-profile", name, "-repo", "missing:/no-such-path"}); err == nil {
			t.Errorf("CLI accepted: %q", name)
		}
	}
}
func TestProfileCopies(t *testing.T) {
	a, err := ProfileIDs("paired-v1")
	if err != nil {
		t.Fatal(err)
	}
	a[0][0] = "changed"
	b, err := ProfileIDs("paired-v1")
	if err != nil || b[0][0] != "golang.org/x/text" {
		t.Fatal("profile state leaked")
	}
}
func TestProfileCLI(t *testing.T) {
	for _, args := range [][]string{{"-unknown"}, {"positional"}, {"-profile", "libc-v1", "positional"}} {
		if _, _, _, err := ParseProfileArgs(args); err == nil {
			t.Errorf("unexpected CLI accepted: %v", args)
		}
	}
}
