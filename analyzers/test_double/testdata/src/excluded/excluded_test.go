package excluded

import "testing"

type stubStore struct{}

func TestDouble(t *testing.T) {
	_ = stubStore{}
}
