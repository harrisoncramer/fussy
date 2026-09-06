package main

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/harrisoncramer/fussy/unexported"
)

const unexportedUsage = `usage: fussy unexported [flags] [packages]

Sweeps a whole module for exported identifiers no other package uses, and says whether to
unexport each one or to delete it. Defaults to ./... .

flags:
`

// runUnexported is the unexported subcommand, which returns the process exit status rather than
// leaving with it so the dispatch in main stays in one place.
func runUnexported(args []string) int {
	flags := flag.NewFlagSet("unexported", flag.ContinueOnError)
	asJSON := flags.Bool("json", false, "write the result as JSON rather than as a report")
	generated := flags.Bool("generated", false, "report identifiers declared in generated files too")
	kinds := flags.String("kinds", strings.Join(unexported.AllKinds(), ","), "the declarations to sweep, from func, type, var and const")
	flags.Usage = func() {
		if _, err := fmt.Fprint(flags.Output(), unexportedUsage); err != nil {
			return
		}
		flags.PrintDefaults()
	}

	if err := flags.Parse(args); err != nil {
		return 2
	}

	wanted, err := parseKinds(*kinds)
	if err != nil {
		fmt.Fprintln(os.Stderr, "fussy:", err)
		return 2
	}

	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "fussy:", err)
		return 2
	}

	result, err := unexported.Sweep(unexported.Options{
		Dir:              dir,
		Patterns:         flags.Args(),
		Kinds:            wanted,
		IncludeGenerated: *generated,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, "fussy:", err)
		return 2
	}

	if err := write(result, *asJSON); err != nil {
		fmt.Fprintln(os.Stderr, "fussy:", err)
		return 2
	}

	if len(result.Findings) > 0 {
		return 1
	}

	return 0
}

func write(result *unexported.Result, asJSON bool) error {
	if asJSON {
		return result.WriteJSON(os.Stdout)
	}

	return result.WriteText(os.Stdout)
}

func parseKinds(raw string) ([]string, error) {
	var wanted []string
	for _, kind := range strings.Split(raw, ",") {
		kind = strings.TrimSpace(kind)
		if kind == "" {
			continue
		}
		if !slices.Contains(unexported.AllKinds(), kind) {
			return nil, fmt.Errorf("unknown kind %q, which must be one of %s", kind, strings.Join(unexported.AllKinds(), ", "))
		}
		wanted = append(wanted, kind)
	}

	if len(wanted) == 0 {
		return nil, fmt.Errorf("no kinds left to sweep, which must be some of %s", strings.Join(unexported.AllKinds(), ", "))
	}

	return wanted, nil
}
