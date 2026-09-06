package main

import (
	"fmt"

	rootlib "fussytest/wsroot/lib"
	"fussytest/wsshared"
)

func main() {
	fmt.Println(wsshared.OnlyTheAppUses(), rootlib.UsedOnlyByTheAppModule())
}
