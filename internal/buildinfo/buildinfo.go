// Package buildinfo exposes the immutable identity of the running build.
package buildinfo

// Info contains version-control and build-time identity fields.
type Info struct {
	Version string `json:"version"`
	Commit  string `json:"commit"`
	BuiltAt string `json:"built_at"`
}

var (
	version = "dev"
	commit  = "unknown"
	builtAt = "unknown"
)

// Current returns the identity embedded in the current binary.
func Current() Info {
	return Info{Version: version, Commit: commit, BuiltAt: builtAt}
}
