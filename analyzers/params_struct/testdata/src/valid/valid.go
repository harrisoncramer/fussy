package valid

import "other"

type NewMailServiceParams struct {
	Name string
}

// A constructor's struct is named after the constructor and taken as p.
func NewMailService(p NewMailServiceParams) string {
	return p.Name
}

type RecordParams struct {
	Note string
}

type store struct{}

// A method is held to the same rule as a function.
func (s *store) Record(p RecordParams) string {
	return p.Note
}

type SaveParams struct {
	Note string
}

// A large struct taken by reference is still a params struct.
func (s *store) Save(p *SaveParams) string {
	return p.Note
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
func UseOther(p other.NewThingParams) string {
	return other.NewThing(p)
}

type CreateAgentParams struct {
	Name string
}

func CreateAgent(p CreateAgentParams) string {
	return p.Name
}

// A wrapper taking the same params can satisfy no name mentioning one of the two functions.
func CreateAgentWithRetry(p CreateAgentParams) string {
	return CreateAgent(p)
}
