// Command fussy runs the analyzers over the packages it is given, so a repository can be
// swept without adopting the golangci-lint plugin or gaining a config file of its own.
package main

import (
	"fmt"
	"os"

	"github.com/harrisoncramer/fussy/analyzers"
	"github.com/harrisoncramer/fussy/config"

	"golang.org/x/tools/go/analysis/multichecker"
)

// unexportedCommand is the one subcommand that cannot be an analyzer, since it has to see every
// package of the module at once to answer whether anything else uses a symbol.
const unexportedCommand = "unexported"

func main() {
	if len(os.Args) > 1 && os.Args[1] == unexportedCommand {
		os.Exit(runUnexported(os.Args[2:]))
	}

	cfg, err := config.LoadFromEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, "fussy:", err)
		os.Exit(2)
	}

	built := analyzers.BuildAll(cfg)
	if len(built) == 0 {
		fmt.Fprintln(os.Stderr, "fussy: every analyzer is skipped, so there is nothing to run")
		os.Exit(2)
	}

	multichecker.Main(built...)
}
