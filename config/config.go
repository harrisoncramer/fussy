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
	SwitchDefault    SwitchDefaultConfig    `json:"switch_default"`
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

type SwitchDefaultConfig struct {
	Skip               bool     `json:"skip"`
	AllowSilentDefault bool     `json:"allow_silent_default"`
	Exclude            []string `json:"exclude"`
}
