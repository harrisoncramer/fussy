// Package unexported sweeps a whole module for exported identifiers no other package uses, so a
// codebase's exported surface stays an honest claim about what the rest of the repository
// depends on.
//
// This cannot be an analyzer. An analyzer runs over one package at a time and sees its
// dependencies as export data, so the question it would have to answer, whether anything else in
// the repository uses the symbol in front of it, is one it can never see the evidence for. That
// is the limitation that made staticcheck drop the whole-program mode of its unused check. This
// sweep loads every package of the module at once instead, which is why it lives behind a
// subcommand rather than beside the analyzers.
//
// It draws three verdicts rather than two, and the middle one is the reason the tool exists. An
// identifier its own package uses and nothing else does should be unexported. An identifier only
// its own tests use is dead and should be deleted rather than unexported, because golangci-lint
// analyses with the test files included, so a call from a test is enough to keep staticcheck's
// unused check quiet, and unexporting such a thing turns a visibly unclaimed export into a
// package-private function no tool will ever flag again. An identifier an external test package
// uses is left alone, since that package is a real other package and unexporting the identifier
// would stop the tests compiling.
//
// Uses are keyed by package path and name rather than by types.Object identity. Loading with the
// tests included gives a package several variants and its dependencies arrive as export data, so
// one declaration exists as several distinct objects and a use rarely points at the object the
// declaration made. Comparing pointers reports almost every export in the repository; comparing
// strings reports what is actually there.
//
// A module nested inside the one being swept is not swept with it, since ./... resolves to the
// module rooted at the working directory and a nested module has a build of its own. A caller
// there is a caller in another module, which is the plainest reason an identifier has to stay
// exported, so those files are read for the names they use and the module is named in the report
// rather than passed over in silence.
//
// Two things are deliberately out of scope. Methods are not reported, because unexporting one
// can break interface satisfaction silently, including for an interface declared in a module this
// sweep never loads, and there is no implements check here to clear it. Identifiers a program
// reaches by name rather than by reference, through reflection, a struct tag or a generator's
// template, look unused to any tool that reads the source, and this one is no exception.
package unexported

import (
	"fmt"
	"sort"
	"time"
)

// Verdict is what a person should do about one exported identifier.
type Verdict string

const (
	// VerdictUnexport is for an identifier only its own package uses, which is safe to lower
	// case because staticcheck goes on watching it afterwards.
	VerdictUnexport Verdict = "unexport"
	// VerdictDelete is for an identifier nothing outside its own tests uses, which unexporting
	// would hide from every tool rather than fix.
	VerdictDelete Verdict = "delete"
)

// Reaches say how far outside the sweep a package could have been imported from, which is what
// decides whether a verdict on it is the whole answer or only the part this repository can see.
const (
	// ReachModulePrivate is a package nothing outside the module can import, which is an
	// internal package or a main package, and where a verdict is complete.
	ReachModulePrivate = "module-private"
	// ReachImportable is a package another module could import, where the sweep has only
	// looked at this repository and a consumer elsewhere would not show up.
	ReachImportable = "importable"
)

// Kinds are the declarations a sweep can be pointed at, methods excluded.
const (
	KindFunc  = "func"
	KindType  = "type"
	KindVar   = "var"
	KindConst = "const"
)

// AllKinds is what a sweep covers when it is not narrowed.
func AllKinds() []string {
	return []string{KindFunc, KindType, KindVar, KindConst}
}

// Options is what one sweep is pointed at.
type Options struct {
	// Dir is the directory the patterns are resolved from.
	Dir string
	// Patterns are the go package patterns to sweep, defaulting to ./... .
	Patterns []string
	// Kinds narrows the sweep to some of AllKinds.
	Kinds []string
	// IncludeGenerated reports identifiers declared in generated files, which are skipped by
	// default because nobody can act on them without changing the generator.
	IncludeGenerated bool
}

// Finding is one exported identifier and what to do about it.
type Finding struct {
	Verdict Verdict `json:"verdict"`
	Package string  `json:"package"`
	Name    string  `json:"name"`
	Kind    string  `json:"kind"`
	Reach   string  `json:"reach"`
	File    string  `json:"file"`
	Line    int     `json:"line"`
	Reason  string  `json:"reason"`
}

// Result is what one sweep found, beside the ground it could not cover.
type Result struct {
	Findings []Finding `json:"findings"`
	// Packages is how many packages were loaded, test variants included.
	Packages int `json:"packages"`
	// Modules are the main modules the sweep covered, which is the boundary outside of which an
	// export may still be someone's dependency.
	Modules []string `json:"modules"`
	// Unloaded are the files no build configuration in this sweep compiled, whose uses were
	// read syntactically instead.
	Unloaded []string `json:"unloaded_files"`
	// Nested are the modules rooted inside the ones swept, which ./... does not reach into and
	// whose files were read syntactically instead.
	Nested []string `json:"nested_modules"`
	// LoadErrors are the packages that did not typecheck, whose uses may be missing.
	LoadErrors []string `json:"load_errors"`
	// KeptByExternalTest counts the exports left alone because an external test package needs
	// them, which is dead product code held up by a black-box test.
	KeptByExternalTest int `json:"kept_by_external_test"`
	// SkippedGenerated counts the exports passed over because a generator wrote them.
	SkippedGenerated int    `json:"skipped_generated"`
	ElapsedMillis    int64  `json:"elapsed_ms"`
	MethodsChecked   bool   `json:"methods_checked"`
	Elapsed          string `json:"-"`
}

// Sweep loads the packages the options name and returns what nothing outside their own package
// uses.
func Sweep(opts Options) (*Result, error) {
	start := time.Now()

	if len(opts.Kinds) == 0 {
		opts.Kinds = AllKinds()
	}

	loaded, err := load(opts)
	if err != nil {
		return nil, err
	}

	index := newIndex(opts, loaded)
	index.readDeclarations()
	index.readUses()
	index.readUnloadedFiles(loaded.unloaded)
	findings := index.decide()

	result := &Result{
		Findings:           findings,
		Packages:           len(loaded.pkgs),
		Modules:            loaded.modules,
		Unloaded:           loaded.unloadedNames,
		Nested:             loaded.nested,
		LoadErrors:         loaded.errors,
		KeptByExternalTest: index.keptByExternalTest,
		SkippedGenerated:   index.skippedGenerated,
		MethodsChecked:     false,
	}

	elapsed := time.Since(start)
	result.ElapsedMillis = elapsed.Milliseconds()
	result.Elapsed = elapsed.Round(time.Millisecond).String()

	sortFindings(result.Findings)

	return result, nil
}

// sortFindings puts the findings in the order a person reads them, which groups a package's
// identifiers together rather than interleaving two packages by name.
func sortFindings(findings []Finding) {
	sort.Slice(findings, func(i, j int) bool {
		a, b := findings[i], findings[j]
		if a.Verdict != b.Verdict {
			return a.Verdict < b.Verdict
		}
		if a.Package != b.Package {
			return a.Package < b.Package
		}

		return a.Name < b.Name
	})
}

// String names the identifier the way a person would search for it.
func (f Finding) String() string {
	return fmt.Sprintf("%s.%s", f.Package, f.Name)
}
