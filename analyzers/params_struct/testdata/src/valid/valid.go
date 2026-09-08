package valid

import "other"

type NewMailServiceParams struct {
	Name string
}

// A constructor's struct is named after the constructor and taken as params.
func NewMailService(params NewMailServiceParams) string {
	return params.Name
}

type RecordParams struct {
	Note string
}

type store struct{}

// A method is held to the same rule as a function.
func (s *store) Record(params RecordParams) string {
	return params.Note
}

type SaveParams struct {
	Note string
}

// A large struct taken by reference is still a params struct.
func (s *store) Save(params *SaveParams) string {
	return params.Note
}

type Config struct {
	Name string
}

// A struct that is not a params struct is nothing to do with this rule.
func Load(cfg Config) string {
	return cfg.Name
}

type Params struct {
	Name string
}

// A type called nothing but Params carries no function name to disagree with.
func Anything(x Params) string {
	return x.Name
}

// A params struct from another package cannot be renamed from here, so only the parameter name
// is held.
func UseOther(params other.NewThingParams) string {
	return other.NewThing(params)
}

type CreateAgentParams struct {
	Name string
}

func CreateAgent(params CreateAgentParams) string {
	return params.Name
}

// A wrapper taking the same params can satisfy no name mentioning one of the two functions.
func CreateAgentWithRetry(params CreateAgentParams) string {
	return CreateAgent(params)
}
