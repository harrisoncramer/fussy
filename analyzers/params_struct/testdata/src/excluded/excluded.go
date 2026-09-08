package excluded

type LauncherParams struct {
	Name string
}

func NewLauncher(params LauncherParams) string {
	return params.Name
}
