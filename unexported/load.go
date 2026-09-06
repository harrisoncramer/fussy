package unexported

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/tools/go/packages"
)

// loadMode asks for the syntax and types of the packages the patterns match and nothing of their
// dependencies, since a dependency of the module cannot be the thing that uses it.
const loadMode = packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles |
	packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedModule

// skippedDirs are the directory names the go tool itself never matches with a ... pattern, which
// the walk for unloaded files has to skip for the same reason.
var skippedDirs = map[string]bool{"testdata": true, "vendor": true, "node_modules": true}

// loaded is the module as the sweep managed to see it, which is the packages it typechecked
// beside an account of what it could not.
type loaded struct {
	pkgs          []*packages.Package
	dir           string
	modules       []string
	errors        []string
	unloaded      []string
	unloadedNames []string
}

// load resolves the patterns into packages, with the tests included so a call from a test is
// visible as a call from a test rather than not at all.
func load(opts Options) (*loaded, error) {
	patterns, err := resolvePatterns(opts)
	if err != nil {
		return nil, err
	}

	cfg := &packages.Config{Mode: loadMode, Dir: opts.Dir, Tests: true}

	pkgs, err := packages.Load(cfg, patterns...)
	if err != nil {
		return nil, fmt.Errorf("loading %s: %w", strings.Join(patterns, " "), err)
	}

	if len(pkgs) == 0 {
		return nil, fmt.Errorf("%s matched no packages", strings.Join(patterns, " "))
	}

	unloaded, err := discoverUnloaded(moduleDirs(pkgs, opts.Dir), pkgs)
	if err != nil {
		return nil, err
	}

	result := &loaded{pkgs: pkgs, dir: opts.Dir, modules: mainModules(pkgs), errors: loadErrors(pkgs)}
	result.unloaded = unloaded
	result.unloadedNames = relativeAll(opts.Dir, unloaded)

	return result, nil
}

// resolvePatterns expands ./... into one pattern per workspace module when the working directory
// is a go.work root, since the go command refuses ./... where the directory holds no module of
// its own and a monorepo is exactly that shape.
func resolvePatterns(opts Options) ([]string, error) {
	patterns := opts.Patterns
	if len(patterns) == 0 {
		patterns = []string{"./..."}
	}

	if len(patterns) != 1 || patterns[0] != "./..." {
		return patterns, nil
	}

	if _, err := os.Stat(filepath.Join(opts.Dir, "go.mod")); err == nil {
		return patterns, nil
	}

	uses, err := workspaceUses(opts.Dir)
	if err != nil || len(uses) == 0 {
		return patterns, err
	}

	return uses, nil
}

// workspaceUses reads the module directories a go.work names, and returns nothing at all when
// the directory is not under one.
func workspaceUses(dir string) ([]string, error) {
	work, err := goEnv(dir, "GOWORK")
	if err != nil || work == "" {
		return nil, err
	}

	cmd := exec.Command("go", "work", "edit", "-json")
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("reading %s: %w", work, err)
	}

	var parsed struct {
		Use []struct {
			DiskPath string
		}
	}
	if err := json.Unmarshal(out, &parsed); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", work, err)
	}

	var patterns []string
	for _, use := range parsed.Use {
		rel, err := filepath.Rel(dir, filepath.Join(filepath.Dir(work), use.DiskPath))
		if err != nil || strings.HasPrefix(rel, "..") {
			continue
		}
		patterns = append(patterns, "./"+filepath.ToSlash(rel)+"/...")
	}

	sort.Strings(patterns)

	return patterns, nil
}

func goEnv(dir, name string) (string, error) {
	cmd := exec.Command("go", "env", name)
	cmd.Dir = dir

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("go env %s: %w", name, err)
	}

	return strings.TrimSpace(string(out)), nil
}

func mainModules(pkgs []*packages.Package) []string {
	seen := map[string]bool{}
	for _, p := range pkgs {
		if p.Module != nil && p.Module.Main {
			seen[p.Module.Path] = true
		}
	}

	return sortedKeys(seen)
}

// moduleDirs is where the walk for unloaded files starts, which is the root of every main module
// the patterns reached into.
func moduleDirs(pkgs []*packages.Package, fallback string) []string {
	seen := map[string]bool{}
	for _, p := range pkgs {
		if p.Module != nil && p.Module.Main && p.Module.Dir != "" {
			seen[p.Module.Dir] = true
		}
	}

	if len(seen) == 0 {
		seen[fallback] = true
	}

	return sortedKeys(seen)
}

// loadErrors gathers the reasons a package did not typecheck, which matter here because a use
// the typechecker never recorded reads as an export nobody wants.
func loadErrors(pkgs []*packages.Package) []string {
	seen := map[string]bool{}
	for _, p := range pkgs {
		for _, err := range p.Errors {
			if strings.HasPrefix(err.Msg, "#") {
				continue
			}
			seen[fmt.Sprintf("%s: %s", p.PkgPath, err.Error())] = true
		}
	}

	return sortedKeys(seen)
}

// discoverUnloaded finds the Go files no build configuration in this sweep compiled, which is
// how a file behind a build tag stops being a caller the sweep cannot see.
func discoverUnloaded(roots []string, pkgs []*packages.Package) ([]string, error) {
	known := map[string]bool{}
	ignored := map[string]bool{}
	for _, p := range pkgs {
		for _, group := range [][]string{p.GoFiles, p.CompiledGoFiles, p.OtherFiles} {
			for _, file := range group {
				known[file] = true
			}
		}
		for _, file := range p.IgnoredFiles {
			if strings.HasSuffix(file, ".go") {
				ignored[file] = true
			}
		}
	}

	found := map[string]bool{}
	for _, file := range sortedKeys(ignored) {
		found[file] = true
	}

	rootSet := map[string]bool{}
	for _, root := range roots {
		rootSet[root] = true
	}

	for _, root := range roots {
		if err := walkForGoFiles(root, rootSet, known, found); err != nil {
			return nil, err
		}
	}

	return sortedKeys(found), nil
}

// walkForGoFiles descends one module, skipping what the go tool skips and the modules nested
// inside it, which are either roots of their own or genuinely not part of this sweep, and gives
// up only where the module root itself cannot be read, since then nothing under it was checked.
func walkForGoFiles(root string, roots, known, found map[string]bool) error {
	return filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			if path == root {
				return fmt.Errorf("walking %s: %w", root, err)
			}

			return nil
		}

		if entry.IsDir() {
			if path == root {
				return nil
			}
			if skipDir(path, entry.Name(), roots) {
				return filepath.SkipDir
			}

			return nil
		}

		if strings.HasSuffix(entry.Name(), ".go") && !known[path] {
			found[path] = true
		}

		return nil
	})
}

func skipDir(path, name string, roots map[string]bool) bool {
	if strings.HasPrefix(name, ".") || strings.HasPrefix(name, "_") || skippedDirs[name] {
		return true
	}

	if roots[path] {
		return false
	}

	_, err := os.Stat(filepath.Join(path, "go.mod"))

	return err == nil
}

func relativeAll(dir string, paths []string) []string {
	out := make([]string, 0, len(paths))
	for _, path := range paths {
		out = append(out, relative(dir, path))
	}

	return out
}

func relative(dir, path string) string {
	rel, err := filepath.Rel(dir, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return path
	}

	return filepath.ToSlash(rel)
}

func sortedKeys(set map[string]bool) []string {
	keys := make([]string, 0, len(set))
	for key := range set {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	return keys
}
