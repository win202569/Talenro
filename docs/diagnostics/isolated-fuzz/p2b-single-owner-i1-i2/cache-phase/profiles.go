package main

import (
	"errors"
	"flag"
	"io"
)

// Fixed profiles return fresh values; caller mutation cannot broaden later runs.
func ProfileIDs(name string) ([][2]string, error) {
	switch name {
	case "paired-v1":
		return [][2]string{{"golang.org/x/text", "v0.40.0"}, {"modernc.org/sqlite", "v1.49.1"}}, nil
	case "libc-v1":
		return [][2]string{{"modernc.org/libc", "v1.72.1"}}, nil
	default:
		return nil, errors.New("PROFILE")
	}
}
func ParseProfileArgs(args []string) (profile, session, repo string, err error) {
	f := flag.NewFlagSet("cache-phase", flag.ContinueOnError)
	f.SetOutput(io.Discard)
	f.StringVar(&profile, "profile", "paired-v1", "fixed diagnostic profile")
	f.StringVar(&session, "session", "", "event session")
	f.StringVar(&repo, "repo", "", "approved worktree")
	if err = f.Parse(args); err != nil {
		return
	}
	if f.NArg() != 0 {
		err = errors.New("CLI_REJECTED")
		return
	}
	_, err = ProfileIDs(profile)
	return
}
