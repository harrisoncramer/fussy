package valid

// A says what it is and nothing about where else to look.
func A() {}

// B mentions Testing without naming a test.
func B() {}

// Explainer: a set-aside paragraph may point at TestSomething and at helpers.go, which is
// what the prefix is for.

// C is fine.
func C() {}

// TestSomething is allowed to name itself, which commentprefix requires it to do.
func TestSomething() {}

// D is published on pkg.go.dev and that is not a file reference.
func D() {}
