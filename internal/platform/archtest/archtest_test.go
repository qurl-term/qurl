package archtest

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

const mod = "example.com/q"

func TestRepository(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	module := modulePath(t, filepath.Join(root, "go.mod"))
	g, err := Load(root, module)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := g[module+"/internal/platform/archtest"]; !ok {
		t.Fatalf("loader did not find this package; is root %q correct?", root)
	}
	for _, v := range Check(g, module) {
		t.Error(v)
	}
}

func TestCheck(t *testing.T) {
	tests := []struct {
		name  string
		graph Graph
		want  []string
	}{
		{
			name:  "domain uses the standard library",
			graph: Graph{mod + "/internal/domain": {"fmt", "net/url"}},
		},
		{
			name:  "domain rejects net/http and third-party",
			graph: Graph{mod + "/internal/domain": {"net/http", "github.com/x/y"}},
			want:  []string{"internal/domain -> github.com/x/y", "internal/domain -> net/http"},
		},
		{
			name: "app may use domain but not platform",
			graph: Graph{
				mod + "/internal/app":               {mod + "/internal/domain", mod + "/internal/platform/eventbus"},
				mod + "/internal/domain":            nil,
				mod + "/internal/platform/eventbus": nil,
			},
			want: []string{"internal/app -> " + mod + "/internal/platform/eventbus"},
		},
		{
			name: "adapter subpackages share a root",
			graph: Graph{
				mod + "/internal/adapter/tui":            nil,
				mod + "/internal/adapter/tui/components": {mod + "/internal/adapter/tui", mod + "/internal/app"},
				mod + "/internal/app":                    nil,
			},
		},
		{
			name: "adapters must not import each other",
			graph: Graph{
				mod + "/internal/adapter/tui":           {mod + "/internal/adapter/cli"},
				mod + "/internal/adapter/cli":           nil,
				mod + "/internal/adapter/store/yamlfs":  {mod + "/internal/adapter/store/history"},
				mod + "/internal/adapter/store/history": nil,
			},
			want: []string{
				"internal/adapter/store/yamlfs -> " + mod + "/internal/adapter/store/history",
				"internal/adapter/tui -> " + mod + "/internal/adapter/cli",
			},
		},
		{
			name: "platform stays generic",
			graph: Graph{
				mod + "/internal/platform/eventbus": {mod + "/internal/app", "github.com/x/y"},
				mod + "/internal/app":               nil,
			},
			want: []string{"internal/platform/eventbus -> " + mod + "/internal/app"},
		},
		{
			name:  "cgo and unknown layers",
			graph: Graph{mod + "/internal/domain": {"C"}, mod + "/internal/misc": {"fmt"}},
			want:  []string{"internal/domain -> C", "internal/misc -> fmt"},
		},
		{
			name: "only cmd may import cmd",
			graph: Graph{
				mod + "/internal/adapter/cli": {mod + "/cmd/qurl"},
				mod + "/cmd/qurl":             {mod + "/internal/adapter/cli", "github.com/x/y"},
			},
			want: []string{"internal/adapter/cli -> " + mod + "/cmd/qurl"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []string
			for _, v := range Check(tt.graph, mod) {
				got = append(got, v.Pkg+" -> "+v.Import)
			}
			if !slices.Equal(got, tt.want) {
				t.Fatalf("violations\n got: %q\nwant: %q", got, tt.want)
			}
		})
	}
}

func TestLoadSkipsTestFilesAndTestdata(t *testing.T) {
	root := t.TempDir()
	write := func(rel, src string) {
		t.Helper()
		p := filepath.Join(root, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	write("internal/domain/a.go", "package domain\nimport \"fmt\"\nvar _ = fmt.Sprint\n")
	write("internal/domain/a_test.go", "package domain\nimport \"net/http\"\nvar _ = http.MethodGet\n")
	write("internal/domain/testdata/x.go", "package x\nimport \"os\"\n")

	g, err := Load(root, mod)
	if err != nil {
		t.Fatal(err)
	}
	if got := g[mod+"/internal/domain"]; !slices.Equal(got, []string{"fmt"}) {
		t.Errorf("domain imports = %q, want [fmt]", got)
	}
	if _, ok := g[mod+"/internal/domain/testdata"]; ok {
		t.Error("testdata directory was loaded")
	}
}

func modulePath(t *testing.T, gomod string) string {
	t.Helper()
	data, err := os.ReadFile(gomod)
	if err != nil {
		t.Fatal(err)
	}
	for line := range strings.Lines(string(data)) {
		if rest, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			return strings.TrimSpace(rest)
		}
	}
	t.Fatal("no module directive in go.mod")
	return ""
}
