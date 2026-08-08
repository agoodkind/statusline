package version

import "testing"

func TestString(t *testing.T) {
	tests := []struct {
		name      string
		version   string
		commit    string
		dirty     string
		buildTime string
		want      string
	}{
		{
			name:      "stamped",
			version:   "v1.2.3",
			commit:    "abc1234",
			dirty:     "false",
			buildTime: "2026-08-08T12:00:00Z",
			want:      "statusline v1.2.3 (abc1234, built 2026-08-08T12:00:00Z)",
		},
		{
			name:      "stamped dirty",
			version:   "v1.2.3-dirty",
			commit:    "abc1234",
			dirty:     "true",
			buildTime: "2026-08-08T12:00:00Z",
			want:      "statusline v1.2.3-dirty (abc1234, built 2026-08-08T12:00:00Z)",
		},
		{
			name: "unstamped",
			want: "statusline dev (unknown, built unknown)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version, Commit, Dirty, BuildTime := tt.version, tt.commit, tt.dirty, tt.buildTime
			restore := setFields(Version, Commit, Dirty, BuildTime)
			defer restore()

			if got := String(); got != tt.want {
				t.Fatalf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}

// setFields sets the package vars for one test case and returns a func that
// restores their previous values, so tests can run in any order without
// leaking state into each other.
func setFields(version, commit, dirty, buildTime string) func() {
	prevVersion, prevCommit, prevDirty, prevBuildTime := Version, Commit, Dirty, BuildTime
	Version, Commit, Dirty, BuildTime = version, commit, dirty, buildTime
	return func() {
		Version, Commit, Dirty, BuildTime = prevVersion, prevCommit, prevDirty, prevBuildTime
	}
}
