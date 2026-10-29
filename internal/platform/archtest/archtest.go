// Package archtest checks qURL's dependency rule (see .docs/ARCHITECTURE.md §2.1).
package archtest

import (
	"fmt"
	"go/parser"
	"go/token"
	"io/fs"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

type Graph map[string][]string

type Violation struct {
	Pkg    string
	Import string
	Reason string
}

func (v Violation) String() string {
	return fmt.Sprintf("%s -> %s: %s", v.Pkg, v.Import, v.Reason)
}

func Load(root, module string) (Graph, error) {
	g := Graph{}
	fset := token.NewFileSet()
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			name := d.Name()
			if p != root && (name == "testdata" || name == "vendor" ||
				strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_")) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(p, ".go") || strings.HasSuffix(p, "_test.go") {
			return nil
		}
		f, err := parser.ParseFile(fset, p, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, filepath.Dir(p))
		if err != nil {
			return err
		}
		pkg := module
		if rel != "." {
			pkg = module + "/" + filepath.ToSlash(rel)
		}
		imports := g[pkg]
		for _, s := range f.Imports {
			imp, err := strconv.Unquote(s.Path.Value)
			if err != nil {
				return err
			}
			if !slices.Contains(imports, imp) {
				imports = append(imports, imp)
			}
		}
		g[pkg] = imports
		return nil
	})
	return g, err
}

func Check(g Graph, module string) []Violation {
	var out []Violation
	for pkg, imports := range g {
		rel, ok := strings.CutPrefix(pkg, module+"/")
		if !ok {
			continue
		}
		for _, imp := range imports {
			if reason := verdict(g, module, rel, imp); reason != "" {
				out = append(out, Violation{Pkg: rel, Import: imp, Reason: reason})
			}
		}
	}
	slices.SortFunc(out, func(a, b Violation) int {
		return strings.Compare(a.Pkg+" "+a.Import, b.Pkg+" "+b.Import)
	})
	return out
}

func verdict(g Graph, module, rel, imp string) string {
	if imp == "C" {
		return "cgo is not allowed (CGO_ENABLED=0)"
	}
	impRel, internal := strings.CutPrefix(imp, module+"/")
	std := !internal && isStd(imp)

	switch layerOf(rel) {
	case "cmd":
		return ""
	case "domain":
		if (std && !isHTTP(imp)) || (internal && within(impRel, "internal/domain")) {
			return ""
		}
		return "domain may import only the standard library (excluding net/http)"
	case "app":
		if (std && !isHTTP(imp)) ||
			(internal && (within(impRel, "internal/domain") || within(impRel, "internal/app"))) {
			return ""
		}
		return "app may import only internal/domain and the standard library (excluding net/http)"
	case "platform":
		if internal && !within(impRel, "internal/platform") {
			return "platform packages must stay generic and import no other internal layer"
		}
		return ""
	case "adapter":
		if !internal {
			return ""
		}
		switch layerOf(impRel) {
		case "domain", "app", "platform":
			return ""
		case "adapter":
			if adapterRoot(g, module, rel) == adapterRoot(g, module, impRel) {
				return ""
			}
			return "adapters must not import other adapters; wire them in cmd/qurl"
		}
		return "adapters may import only domain, app and platform"
	}
	return "package is outside the known layers (cmd, internal/{domain,app,adapter,platform})"
}

func layerOf(rel string) string {
	for _, l := range []string{"cmd", "internal/domain", "internal/app", "internal/adapter", "internal/platform"} {
		if within(rel, l) {
			return path.Base(l)
		}
	}
	return ""
}

// adapterRoot is the shallowest directory under internal/adapter that holds Go files.
func adapterRoot(g Graph, module, rel string) string {
	parts := strings.Split(rel, "/")
	for i := 3; i <= len(parts); i++ {
		p := strings.Join(parts[:i], "/")
		if _, ok := g[module+"/"+p]; ok {
			return p
		}
	}
	return rel
}

func within(rel, prefix string) bool {
	return rel == prefix || strings.HasPrefix(rel, prefix+"/")
}

func isStd(imp string) bool {
	first, _, _ := strings.Cut(imp, "/")
	return !strings.Contains(first, ".")
}

func isHTTP(imp string) bool {
	return imp == "net/http" || strings.HasPrefix(imp, "net/http/")
}
