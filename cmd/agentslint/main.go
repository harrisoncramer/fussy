// Command agentslint runs the analyzers over the packages it is given, so a repository can be
// swept without adopting the golangci-lint plugin or gaining a config file of its own.
package main

import (
	"fmt"
	"os"

	"github.com/harrisoncramer/agentslinter/analyzers"
	"github.com/harrisoncramer/agentslinter/config"

	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	cfg, err := config.LoadFromEnv()
	if err != nil {
		fmt.Fprintln(os.Stderr, "agentslint:", err)
		os.Exit(2)
	}

	built := analyzers.BuildAll(cfg)
	if len(built) == 0 {
		fmt.Fprintln(os.Stderr, "agentslint: every analyzer is skipped, so there is nothing to run")
		os.Exit(2)
	}

	multichecker.Main(built...)
}
