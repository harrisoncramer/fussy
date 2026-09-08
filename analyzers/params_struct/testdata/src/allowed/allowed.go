package allowed

type scriptEnvParams struct {
	Name string
}

// A struct the allow list names is held to no function name, and is still held to the parameter
// name.
func env(params scriptEnvParams) string {
	return params.Name
}

type resumeParams struct {
	Name string
}

func resumeCmd(p resumeParams) string { // want `resumeParams is taken as "p", so call it "params" since every params struct is taken the same way`
	return p.Name
}

type LauncherParams struct {
	Name string
}

func NewLauncher(params LauncherParams) string { // want `NewLauncher takes LauncherParams, so name the struct NewLauncherParams after the function that takes it`
	return params.Name
}
