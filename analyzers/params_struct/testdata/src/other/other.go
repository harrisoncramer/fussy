package other

type NewThingParams struct {
	Name string
}

func NewThing(p NewThingParams) string {
	return p.Name
}
