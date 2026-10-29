// Package buildinfo reports the version stamped at link time or recorded by the Go toolchain.
package buildinfo

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"strings"
)

// Set by the release build: -X github.com/qurl-term/qurl/internal/platform/buildinfo.version=v1.2.3
var (
	version string
	commit  string
	date    string
)

type Info struct {
	Version   string
	Commit    string
	Date      string
	Modified  bool
	GoVersion string
	Platform  string
}

func Get() Info {
	return resolve(version, commit, date, debug.ReadBuildInfo)
}

func resolve(v, c, d string, read func() (*debug.BuildInfo, bool)) Info {
	info := Info{
		Version:   v,
		Commit:    c,
		Date:      d,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}

	if bi, ok := read(); ok {
		if bi.GoVersion != "" {
			info.GoVersion = bi.GoVersion
		}
		if info.Version == "" && bi.Main.Version != "" && bi.Main.Version != "(devel)" {
			info.Version = bi.Main.Version
		}
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				if info.Commit == "" {
					info.Commit = s.Value
				}
			case "vcs.time":
				if info.Date == "" {
					info.Date = s.Value
				}
			case "vcs.modified":
				info.Modified = s.Value == "true"
			}
		}
	}
	if info.Version == "" {
		info.Version = "dev"
	}
	return info
}

func (i Info) String() string {
	var meta []string

	if i.Commit != "" {
		c := i.Commit
		if len(c) > 12 {
			c = c[:12]
		}
		if i.Modified {
			c += "-dirty"
		}
		meta = append(meta, c)
	}
	if i.Date != "" {
		meta = append(meta, i.Date)
	}
	s := "qurl " + i.Version
	if len(meta) > 0 {
		s += fmt.Sprintf(" (%s)", strings.Join(meta, ", "))
	}

	return s + " " + i.GoVersion + " " + i.Platform
}
