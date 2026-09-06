//go:build ignore

package main

import (
	"fmt"

	"fussytest/verdicts/lib"
	. "fussytest/verdicts/lib"
)

func main() {
	fmt.Println(lib.ForGenerator(), DotImported())
}
