// Package plugin registers the analyzers with golangci-lint, which is what makes them run as
// part of one lint pass rather than as a second tool.
package plugin

import (
	"github.com/harrisoncramer/agentslinter/analyzers"
	"github.com/harrisoncramer/agentslinter/config"

	"github.com/golangci/plugin-module-register/register"
	"golang.org/x/tools/go/analysis"
)

func init() {
	register.Plugin("agentslinter", func(settings any) (register.LinterPlugin, error) {
		s, err := register.DecodeSettings[config.Config](settings)
		if err != nil {
			return nil, err
		}

		return &agentsLinter{config: s}, nil
	})
}

type agentsLinter struct {
	config config.Config
}

// BuildAnalyzers returns every analyzer this plugin contributes.
func (a *agentsLinter) BuildAnalyzers() ([]*analysis.Analyzer, error) {
	return analyzers.BuildAll(a.config), nil
}

// GetLoadMode tells golangci-lint the analyzers need full type information.
func (a *agentsLinter) GetLoadMode() string {
	return register.LoadModeTypesInfo
}
