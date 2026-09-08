package invalid

// A is checked by TestASaysWhatItDoes. // want `comment names the test "TestASaysWhatItDoes"`
func A() {}

// B is the half of the pair that lives in helpers.go. // want `comment names the file "helpers.go"`
func B() {}

// C names a path, internal/ui/help.go, in passing. // want `comment names the file "internal/ui/help.go"`
func C() {}
