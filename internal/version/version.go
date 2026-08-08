// Package version holds build identity stamped by go-build.mk's ldflags.
package version

// Commit, Version, Dirty, and BuildTime are set via -ldflags -X at build
// time. They are empty in an unstamped `go build` or `go test` invocation.
var (
	Commit    string
	Version   string
	Dirty     string
	BuildTime string
)

// String formats the build identity into the line a --version flag prints.
// Version already carries a "-dirty" suffix when applicable (git describe
// --dirty), so this does not add its own.
func String() string {
	version := Version
	if version == "" {
		version = "dev"
	}
	commit := Commit
	if commit == "" {
		commit = "unknown"
	}
	buildTime := BuildTime
	if buildTime == "" {
		buildTime = "unknown"
	}
	return "statusline " + version + " (" + commit + ", built " + buildTime + ")"
}
