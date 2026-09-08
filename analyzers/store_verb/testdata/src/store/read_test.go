package store

import "testing"

type fakeRows struct{}

// A double in a test file is not part of the store's surface.
func (f fakeRows) LatestRound() string {
	return ""
}

func TestFakeRows(t *testing.T) {
	_ = fakeRows{}.LatestRound()
}
