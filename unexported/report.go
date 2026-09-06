package unexported

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"
)

// listLimit is how many files or errors a text report names before it starts counting them
// instead, since a sweep of one subtree can leave most of a repository unloaded.
const listLimit = 20

// headings say what each verdict means in the terms of the edit it asks for, since the whole
// point of separating them is that they call for different edits.
var headings = map[Verdict]string{
	VerdictUnexport: "only their own package uses, so lower case them",
	VerdictDelete:   "nothing outside their own tests uses, so delete them",
}

// WriteJSON writes the result as one object, which is the shape for a CI job or an agent rather
// than for a person.
func (r *Result) WriteJSON(w io.Writer) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(r); err != nil {
		return fmt.Errorf("writing json: %w", err)
	}

	return nil
}

// WriteText writes the result grouped by verdict, since a person reading it has to do a different
// thing to each group.
func (r *Result) WriteText(w io.Writer) error {
	out := &lines{to: w}

	for _, verdict := range []Verdict{VerdictUnexport, VerdictDelete} {
		group := r.group(verdict)
		if len(group) == 0 {
			continue
		}

		out.printf("%s: %d %s %s\n", verdict, len(group), plural(len(group), "identifier"), headings[verdict])
		writeGroup(out, group)
		out.printf("\n")
	}

	r.writeSummary(out)

	return out.err
}

// lines is a writer keeping the first error it met, so a report reads as the run of statements it
// is rather than as an error check after every line of output.
type lines struct {
	to  io.Writer
	err error
}

func (l *lines) printf(format string, args ...any) {
	if l.err == nil {
		_, l.err = fmt.Fprintf(l.to, format, args...)
	}
}

func (r *Result) group(verdict Verdict) []Finding {
	var group []Finding
	for _, finding := range r.Findings {
		if finding.Verdict == verdict {
			group = append(group, finding)
		}
	}

	return group
}

// writeGroup lays one verdict's findings out in columns and trims the padding off the end of each
// line, since the note is empty for the findings that need nothing said about them.
func writeGroup(out *lines, group []Finding) {
	var padded bytes.Buffer

	table := tabwriter.NewWriter(&padded, 0, 0, 2, ' ', 0)
	cells := &lines{to: table}
	for _, finding := range group {
		cells.printf("  %s\t%s\t%s:%d\t%s\n", finding, finding.Kind, finding.File, finding.Line, note(finding))
	}

	if err := table.Flush(); err != nil {
		out.err = fmt.Errorf("laying out findings: %w", err)
		return
	}

	for line := range strings.Lines(padded.String()) {
		out.printf("%s\n", strings.TrimRight(line, " \n"))
	}
}

// note says what the group heading does not, which is why a delete is a delete and where the
// sweep only searched this repository.
func note(finding Finding) string {
	var parts []string
	if finding.Verdict != VerdictUnexport {
		parts = append(parts, finding.Reason)
	}

	if finding.Reach == ReachImportable {
		parts = append(parts, ReachImportable)
	}

	return strings.Join(parts, ", ")
}

// writeSummary says what the sweep covered and what it could not, since a confident empty report
// over ground the loader never saw is the one answer this tool must not give.
func (r *Result) writeSummary(out *lines) {
	out.printf("swept %d packages across %s in %s\n", r.Packages, describeModules(r.Modules), r.Elapsed)
	out.printf("methods are not reported, since unexporting one can break interface satisfaction silently\n")

	if importable := r.countImportable(); importable > 0 {
		out.printf("%d of them are marked importable, meaning another module could import that package, so look for consumers outside the modules swept before acting\n", importable)
	}

	if r.KeptByExternalTest > 0 {
		out.printf("%d exported identifiers are left alone because only an external test package uses them\n", r.KeptByExternalTest)
	}

	if r.SkippedGenerated > 0 {
		out.printf("%d exported identifiers declared in generated files were skipped\n", r.SkippedGenerated)
	}

	writeList(out, r.Nested, "module", "nested inside the sweep and not loaded, since ./... resolves to the module at the working directory, so a caller there was read syntactically")
	writeList(out, r.Unloaded, "file", "no build configuration in this sweep compiled, read for uses syntactically rather than typechecked")
	writeList(out, r.LoadErrors, "package", "did not typecheck, so uses inside them may be missing")
}

// writeList names what the sweep could not read, since a report that hides its own blind spots is
// the one thing worse than a report with nothing in it.
func writeList(out *lines, items []string, noun, why string) {
	if len(items) == 0 {
		return
	}

	out.printf("%d %s %s\n", len(items), plural(len(items), noun), why)
	for i, item := range items {
		if i == listLimit {
			out.printf("  and %d more\n", len(items)-listLimit)
			break
		}
		out.printf("  %s\n", item)
	}
}

func (r *Result) countImportable() int {
	count := 0
	for _, finding := range r.Findings {
		if finding.Reach == ReachImportable {
			count++
		}
	}

	return count
}

func plural(count int, noun string) string {
	if count == 1 {
		return noun
	}

	return noun + "s"
}

func describeModules(modules []string) string {
	if len(modules) == 1 {
		return "1 module (" + modules[0] + ")"
	}

	if len(modules) <= 4 {
		return fmt.Sprintf("%d modules (%s)", len(modules), strings.Join(modules, ", "))
	}

	return fmt.Sprintf("%d modules", len(modules))
}
