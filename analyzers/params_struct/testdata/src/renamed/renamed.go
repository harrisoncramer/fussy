package renamed

type NewLauncherParams struct {
	Name string
}

// A repository that has settled on p rather than params configures the name it holds.
func NewLauncher(p NewLauncherParams) string {
	return p.Name
}

type NewAgentParams struct {
	Name string
}

func NewAgent(params NewAgentParams) string { // want `NewAgentParams is taken as "params", so call it "p" since every params struct is taken the same way`
	return params.Name
}
