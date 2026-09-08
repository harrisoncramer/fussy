package store

type Store struct{}

func (s *Store) DoesActiveNameExist(name string) bool { // want `DoesActiveNameExist does not open with a verb, so name it after what it does:`
	return name != ""
}

func (s *Store) LatestRound(id string) string { // want `LatestRound does not open with a verb, so name it after what it does:`
	return id
}

func (s *Store) LastReleasedReviewOf(id string) string { // want `LastReleasedReviewOf does not open with a verb, so name it after what it does:`
	return id
}

func (s *Store) NamesHoldingSchedule(id string) []string { // want `NamesHoldingSchedule does not open with a verb, so name it after what it does:`
	return []string{id}
}

func (s *Store) ListAgents() []string {
	return nil
}

func (s *Store) Exists(id string) bool {
	return id != ""
}

func (s *Store) GetAgent(id string) string {
	return id
}

func (s *Store) MarkRead(id string) {
	_ = id
}

// An unexported method is the package's own business.
func (s *Store) nextCursor() string {
	return ""
}

// A plain function is not a method on the store.
func Rows() int {
	return 0
}
