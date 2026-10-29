package buildinfo

import (
	"runtime"
	"runtime/debug"
	"testing"
)

func fakeRead(bi *debug.BuildInfo) func() (*debug.BuildInfo, bool) {
	return func() (*debug.BuildInfo, bool) { return bi, bi != nil }
}

func TestResolve(t *testing.T) {
	vcs := &debug.BuildInfo{
		GoVersion: "go1.27.3",
		Main:      debug.Module{Version: "(devel)"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "0123456789abcdef0123"},
			{Key: "vcs.time", Value: "2026-10-29T10:00:00Z"},
			{Key: "vcs.modified", Value: "true"},
		},
	}
	tests := []struct {
		name    string
		v, c, d string
		bi      *debug.BuildInfo
		want    Info
	}{
		{
			name: "ldflags win over vcs",
			v:    "v0.1.0", c: "abc1234", d: "2026-11-01",
			bi: vcs,
			want: Info{Version: "v0.1.0", Commit: "abc1234", Date: "2026-11-01", Modified: true,
				GoVersion: "go1.27.3"},
		},
		{
			name: "devel build falls back to vcs",
			bi:   vcs,
			want: Info{Version: "dev", Commit: "0123456789abcdef0123", Date: "2026-10-29T10:00:00Z",
				Modified: true, GoVersion: "go1.27.3"},
		},
		{
			name: "go install module version",
			bi:   &debug.BuildInfo{GoVersion: "go1.27.3", Main: debug.Module{Version: "v0.3.0"}},
			want: Info{Version: "v0.3.0", GoVersion: "go1.27.3"},
		},
		{
			name: "no build info",
			want: Info{Version: "dev", GoVersion: runtime.Version()},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.want.Platform = runtime.GOOS + "/" + runtime.GOARCH
			if got := resolve(tt.v, tt.c, tt.d, fakeRead(tt.bi)); got != tt.want {
				t.Fatalf("resolve() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name string
		info Info
		want string
	}{
		{
			name: "full",
			info: Info{Version: "v0.1.0", Commit: "0123456789abcdef", Date: "2026-11-01", Modified: true,
				GoVersion: "go1.27.3", Platform: "darwin/arm64"},
			want: "qurl v0.1.0 (0123456789ab-dirty, 2026-11-01) go1.27.3 darwin/arm64",
		},
		{
			name: "minimal",
			info: Info{Version: "dev", GoVersion: "go1.27.3", Platform: "linux/amd64"},
			want: "qurl dev go1.27.3 linux/amd64",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.info.String(); got != tt.want {
				t.Fatalf("String() = %q, want %q", got, tt.want)
			}
		})
	}
}
