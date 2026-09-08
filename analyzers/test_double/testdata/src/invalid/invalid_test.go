package invalid

import "testing"

type stubStore struct{} // want `stubStore is a stub: name a test double fakeStore, so one word covers all of them`

type MockClock struct{} // want `MockClock is a mock: name a test double fakeClock, so one word covers all of them`

type spyRecorder struct{} // want `spyRecorder is a spy: name a test double fakeRecorder, so one word covers all of them`

func TestDoublesAreNamed(t *testing.T) {
	_ = stubStore{}
	_ = MockClock{}
	_ = spyRecorder{}
}
