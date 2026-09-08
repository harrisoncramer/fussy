// Package config is the settings the analyzers take from .golangci.yaml, which is how a rule is
// turned off or pointed away from the files it does not apply to.
package config

type Config struct {
	CommentClause    CommentClauseConfig    `json:"comment_clause"`
	CommentLength    CommentLengthConfig    `json:"comment_length"`
	CommentPrefix    CommentPrefixConfig    `json:"comment_prefix"`
	CommentReference CommentReferenceConfig `json:"comment_reference"`
	ContextTimeout   ContextTimeoutConfig   `json:"context_timeout"`
	ForbidGetenv     ForbidGetenvConfig     `json:"forbid_getenv"`
	ForbidNilNil     ForbidNilNilConfig     `json:"forbid_nil_nil"`
	ParamsStruct     ParamsStructConfig     `json:"params_struct"`
	StoreVerb        StoreVerbConfig        `json:"store_verb"`
	SwitchDefault    SwitchDefaultConfig    `json:"switch_default"`
	TableTest        TableTestConfig        `json:"table_test"`
	TestDouble       TestDoubleConfig       `json:"test_double"`
}

type CommentClauseConfig struct {
	Skip    bool     `json:"skip"`
	Exclude []string `json:"exclude"`
	// Clauses overrides the joins reported, and none leaves the built-in list in place.
	Clauses []string `json:"clauses"`
}

type CommentReferenceConfig struct {
	Skip    bool     `json:"skip"`
	Exclude []string `json:"exclude"`
}

type ContextTimeoutConfig struct {
	Skip    bool     `json:"skip"`
	Include []string `json:"include"`
}

type CommentLengthConfig struct {
	Skip    bool     `json:"skip"`
	Exclude []string `json:"exclude"`
	// MaxChars caps a one-sentence comment, and zero leaves the length unchecked.
	MaxChars int `json:"max_chars"`
}

type CommentPrefixConfig struct {
	Skip    bool `json:"skip"`
	Require bool `json:"require"`
}

type ForbidGetenvConfig struct {
	Skip bool `json:"skip"`
}

type ForbidNilNilConfig struct {
	Skip bool `json:"skip"`
}

type ParamsStructConfig struct {
	Skip    bool     `json:"skip"`
	Exclude []string `json:"exclude"`
}

type StoreVerbConfig struct {
	Skip bool `json:"skip"`
	// Include names the paths the rule reaches, and none leaves it switched off.
	Include []string `json:"include"`
	// Verbs replaces the built-in list of verbs a method may open with.
	Verbs []string `json:"verbs"`
}

type TableTestConfig struct {
	Skip    bool     `json:"skip"`
	Exclude []string `json:"exclude"`
}

type TestDoubleConfig struct {
	Skip    bool     `json:"skip"`
	Exclude []string `json:"exclude"`
	// Forbidden replaces the built-in stub, mock and spy prefixes.
	Forbidden []string `json:"forbidden"`
	// Preferred is the word the report points at, and none means fake.
	Preferred string `json:"preferred"`
}

type SwitchDefaultConfig struct {
	Skip               bool     `json:"skip"`
	AllowSilentDefault bool     `json:"allow_silent_default"`
	Exclude            []string `json:"exclude"`
}
