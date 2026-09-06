package unexported_test

import (
	"path/filepath"
	"slices"
	"testing"

	"github.com/harrisoncramer/fussy/unexported"
)

// TestEveryVerdict pins the whole answer for the fixture module, since the verdicts are only
// useful if the set is exactly right rather than nearly right.
func TestEveryVerdict(t *testing.T) {
	want := map[string]unexported.Verdict{
		"fussytest/verdicts/lib.ErrSentinel":  unexported.VerdictUnexport,
		"fussytest/verdicts/lib.InternalOnly": unexported.VerdictUnexport,
		"fussytest/verdicts/lib.Inner":        unexported.VerdictUnexport,
		"fussytest/verdicts/lib.Level":        unexported.VerdictUnexport,
		"fussytest/verdicts/lib.LevelLow":     unexported.VerdictUnexport,
		"fussytest/verdicts/lib.Nobody":       unexported.VerdictDelete,
		"fussytest/verdicts/lib.TestOnly":     unexported.VerdictDelete,
	}

	got := verdicts(t, sweep(t))

	for name, verdict := range want {
		if got[name] != verdict {
			t.Errorf("%s: got %q, want %q", name, got[name], verdict)
		}
	}

	for name, verdict := range got {
		if _, expected := want[name]; !expected {
			t.Errorf("%s: reported %q, expected no finding", name, verdict)
		}
	}
}

// TestUsesFromAnotherPackagesTestsAreNotFindings pins the reason uses are keyed by package path
// and name rather than by types.Object identity, since loading with the tests included gives one
// declaration several objects and a pointer comparison reports almost every export there is.
func TestUsesFromAnotherPackagesTestsAreNotFindings(t *testing.T) {
	got := verdicts(t, sweep(t))

	if verdict, found := got["fussytest/verdicts/lib.CrossTest"]; found {
		t.Fatalf("CrossTest is called from another package's test variant, so %q is wrong", verdict)
	}
}

// TestOwnTestsMeanDeadRatherThanUnexportable pins the verdict this tool exists for, since
// unexporting something only its own tests call hides it from every tool that would have flagged
// it rather than fixing it.
func TestOwnTestsMeanDeadRatherThanUnexportable(t *testing.T) {
	got := verdicts(t, sweep(t))

	if got["fussytest/verdicts/lib.TestOnly"] != unexported.VerdictDelete {
		t.Fatalf("TestOnly: got %q, want %q", got["fussytest/verdicts/lib.TestOnly"], unexported.VerdictDelete)
	}
}

// TestExternalTestPackagesKeepTheirExport pins that a black-box test counts as another package,
// since unexporting what it calls would stop it compiling.
func TestExternalTestPackagesKeepTheirExport(t *testing.T) {
	result := sweep(t)

	if _, found := verdicts(t, result)["fussytest/verdicts/lib.BlackBoxOnly"]; found {
		t.Error("BlackBoxOnly is called from an external test package, so it must be left alone")
	}

	if result.KeptByExternalTest != 1 {
		t.Errorf("kept by external test: got %d, want 1", result.KeptByExternalTest)
	}
}

// TestBuildTaggedCallersAreFound pins that a file no build configuration compiled is read for the
// names it uses, since a generator behind a go:build ignore tag is a caller the typechecker never
// sees and a confident report without it is wrong.
func TestBuildTaggedCallersAreFound(t *testing.T) {
	result := sweep(t)

	if _, found := verdicts(t, result)["fussytest/verdicts/lib.ForGenerator"]; found {
		t.Error("ForGenerator is called from a build-tagged generator, so it must be left alone")
	}

	if !slices.Contains(result.Unloaded, filepath.ToSlash(filepath.Join("gen", "main.go"))) {
		t.Errorf("unloaded files: got %v, want the build-tagged generator named", result.Unloaded)
	}
}

// TestSelfReferentialGenericsTerminate pins that a type parameter constrained by the type it is a
// parameter of does not walk in a circle, which is a shape real code has and which crashed the
// surface walk before it was guarded.
func TestSelfReferentialGenericsTerminate(t *testing.T) {
	got := verdicts(t, sweep(t))

	if verdict, found := got["fussytest/verdicts/lib.SelfRef"]; found {
		t.Fatalf("SelfRef is used from another package, so %q is wrong", verdict)
	}
}

// TestTypesOnTheExportedSurfaceAreLeftAlone pins that a type an exported signature names is kept,
// since unexporting it would leave an exported function returning something a caller outside the
// package cannot write down.
func TestTypesOnTheExportedSurfaceAreLeftAlone(t *testing.T) {
	got := verdicts(t, sweep(t))

	if verdict, found := got["fussytest/verdicts/lib.Surface"]; found {
		t.Fatalf("Surface is what the exported Public returns, so %q is wrong", verdict)
	}
}

// TestEnumMembersFollowTheirType pins that the constants of an exported type nobody has proposed
// unexporting are kept, since an enum a caller can switch over needs every member it has.
func TestEnumMembersFollowTheirType(t *testing.T) {
	got := verdicts(t, sweep(t))

	for _, name := range []string{"ModeQuiet", "ModeLoud"} {
		if verdict, found := got["fussytest/verdicts/lib."+name]; found {
			t.Errorf("%s is a member of an exported enum, so %q is wrong", name, verdict)
		}
	}
}

// TestGeneratedDeclarationsAreSkipped pins that a generated file is passed over, since nobody can
// act on it without changing the generator that wrote it.
func TestGeneratedDeclarationsAreSkipped(t *testing.T) {
	result := sweep(t)

	if _, found := verdicts(t, result)["fussytest/verdicts/lib.Generated"]; found {
		t.Error("Generated is declared in a generated file, so it must be skipped")
	}

	if result.SkippedGenerated != 1 {
		t.Errorf("skipped generated: got %d, want 1", result.SkippedGenerated)
	}
}

// TestGeneratedDeclarationsCanBeAskedFor pins that the skip is a default rather than a rule, since
// an export a generator wrote and nobody uses is still worth knowing about.
func TestGeneratedDeclarationsCanBeAskedFor(t *testing.T) {
	result := run(t, unexported.Options{Dir: fixture(t, "verdicts"), IncludeGenerated: true})

	if verdicts(t, result)["fussytest/verdicts/lib.Generated"] != unexported.VerdictDelete {
		t.Error("Generated is used by nothing, so asking for generated files should report it")
	}
}

// TestKindsNarrowTheSweep pins that a run can be held to one sort of declaration, since the first
// cut a repository wants is usually the functions alone.
func TestKindsNarrowTheSweep(t *testing.T) {
	result := run(t, unexported.Options{Dir: fixture(t, "verdicts"), Kinds: []string{unexported.KindFunc}})

	for _, finding := range result.Findings {
		if finding.Kind != unexported.KindFunc {
			t.Errorf("%s: got kind %q, want only funcs", finding, finding.Kind)
		}
	}
}

// TestTheSweepLoadsCleanly pins that the fixture typechecks, since a load error would make every
// other assertion here pass for the wrong reason.
func TestTheSweepLoadsCleanly(t *testing.T) {
	if errs := sweep(t).LoadErrors; len(errs) > 0 {
		t.Fatalf("load errors: %v", errs)
	}
}

func sweep(t *testing.T) *unexported.Result {
	t.Helper()

	return run(t, unexported.Options{Dir: fixture(t, "verdicts")})
}

func run(t *testing.T, opts unexported.Options) *unexported.Result {
	t.Helper()

	result, err := unexported.Sweep(opts)
	if err != nil {
		t.Fatalf("sweeping %s: %v", opts.Dir, err)
	}

	return result
}

func fixture(t *testing.T, name string) string {
	t.Helper()

	dir, err := filepath.Abs(filepath.Join("testdata", name))
	if err != nil {
		t.Fatalf("resolving %s: %v", name, err)
	}

	return dir
}

func verdicts(t *testing.T, result *unexported.Result) map[string]unexported.Verdict {
	t.Helper()

	got := map[string]unexported.Verdict{}
	for _, finding := range result.Findings {
		got[finding.String()] = finding.Verdict
	}

	return got
}
