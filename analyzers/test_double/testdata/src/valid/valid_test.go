package valid

import "testing"

type fakeStore struct{}

// A word that only opens with a forbidden one is a different word.
type mockingbirdCall struct{}

type stubbornRetry struct{}

func TestDoublesAreNamed(t *testing.T) {
	_ = fakeStore{}
	_ = mockingbirdCall{}
	_ = stubbornRetry{}
}
