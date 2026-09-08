package invalid

import "other"

type LauncherParams struct {
	Name string
}

// A struct named after the type built rather than the function taking it is the drift this rule
// exists to stop.
func NewLauncher(p LauncherParams) string { // want `NewLauncher takes LauncherParams, so name the struct NewLauncherParams after the function that takes it` `LauncherParams is taken as "p", so call it "params" since every params struct is taken the same way`
	return p.Name
}

type NewAgentParams struct {
	Name string
}

func NewAgent(opts NewAgentParams) string { // want `NewAgentParams is taken as "opts", so call it "params" since every params struct is taken the same way`
	return opts.Name
}

type WorkflowParams struct {
	Name string
}

type registry struct{}

func (r *registry) Register(opts WorkflowParams) string { // want `Register takes WorkflowParams, so name the struct RegisterParams after the function that takes it` `WorkflowParams is taken as "opts", so call it "params" since every params struct is taken the same way`
	return opts.Name
}

func UseOther(p other.NewThingParams) string { // want `NewThingParams is taken as "p", so call it "params" since every params struct is taken the same way`
	return other.NewThing(p)
}
