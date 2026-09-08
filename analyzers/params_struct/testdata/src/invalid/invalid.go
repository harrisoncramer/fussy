package invalid

import "other"

type LauncherParams struct {
	Name string
}

// A struct named after the type built rather than the function taking it is the drift this rule
// exists to stop.
func NewLauncher(p LauncherParams) string { // want `NewLauncher takes LauncherParams, so name the struct NewLauncherParams after the function that takes it`
	return p.Name
}

type NewAgentParams struct {
	Name string
}

func NewAgent(params NewAgentParams) string { // want `NewAgentParams is taken as "params", so call it "p" since every params struct is taken the same way`
	return params.Name
}

type WorkflowParams struct {
	Name string
}

type registry struct{}

func (r *registry) Register(opts WorkflowParams) string { // want `Register takes WorkflowParams, so name the struct RegisterParams after the function that takes it` `WorkflowParams is taken as "opts", so call it "p" since every params struct is taken the same way`
	return opts.Name
}

func UseOther(params other.NewThingParams) string { // want `NewThingParams is taken as "params", so call it "p" since every params struct is taken the same way`
	return other.NewThing(params)
}
