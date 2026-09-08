package configured

import "testing"

type fakeStore struct{} // want `fakeStore is a fake: name a test double mockStore, so one word covers all of them`

type mockClock struct{}

func TestDoublesAreNamed(t *testing.T) {
	_ = fakeStore{}
	_ = mockClock{}
}
