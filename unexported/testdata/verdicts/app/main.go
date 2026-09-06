package main

import (
	"fmt"

	"fussytest/verdicts/lib"
)

type node struct{}

func (n node) Self() node { return n }

func main() {
	var mode lib.Mode
	fmt.Println(lib.Public().Name, mode, lib.UseSelfRef(node{}), lib.GradeHigh)
}
