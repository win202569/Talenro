package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	profile, session, repo, err := ParseProfileArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, "CLI_REJECTED")
		os.Exit(1)
	}
	const approvedRepo = "D:/Projects/Talenro/.worktrees/c12-b01-task9-coordinator"
	if !strings.EqualFold(filepath.Clean(repo), filepath.Clean(approvedRepo)) {
		fmt.Fprintln(os.Stderr, "CLI_REJECTED")
		os.Exit(1)
	}
	const cache = "C:/Users/Lenovo/go/pkg/mod"
	var candidates []Candidate
	ids, err := ProfileIDs(profile)
	if err != nil {
		fmt.Fprintln(os.Stderr, "CLI_REJECTED")
		os.Exit(1)
	}
	for _, id := range ids {
		candidates = append(candidates, Candidate{Module: id[0], Version: id[1],
			Dir:     filepath.Join(cache, filepath.FromSlash(id[0]+"@"+id[1])),
			Zip:     filepath.Join(cache, "cache", "download", filepath.FromSlash(id[0]), "@v", id[1]+".zip"),
			SumFile: filepath.Join(repo, "go.sum")})
	}
	encoder := json.NewEncoder(os.Stdout)
	if err := Run(candidates, session, func(e Event) error { return encoder.Encode(e) }); err != nil {
		// Error events identify the failed phase; never print paths or input contents.
		fmt.Fprintln(os.Stderr, "PHASE_STOPPED")
		os.Exit(1)
	}
}
