package invalid

// A is one arm of a dispatch, since a row can name two actions. // want `comment tacks a "since" clause`
func A() {}

// B holds the cursor, because the list underneath moves. // want `comment tacks a "because" clause`
func B() {}

// C is declared here, rather than beside each caller. // want `comment tacks a "rather than" clause`
func C() {}

// D is the shared tail, so that a screen cannot forget it. // want `comment tacks a "so that" clause`
func D() {}
