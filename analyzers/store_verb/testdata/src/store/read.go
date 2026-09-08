package store

type Store struct{}

func (s *Store) DoesActiveNameExist(name string) bool { // want `DoesActiveNameExist does not open with one of the store's verbs, so name it after what it does`
	return name != ""
}

func (s *Store) LatestRound(id string) string { // want `LatestRound does not open with one of the store's verbs, so name it after what it does`
	return id
}

func (s *Store) LastReleasedReviewOf(id string) string { // want `LastReleasedReviewOf does not open with one of the store's verbs, so name it after what it does`
	return id
}

func (s *Store) NamesHoldingSchedule(id string) []string { // want `NamesHoldingSchedule does not open with one of the store's verbs, so name it after what it does`
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

// A method whose name an interface has already picked cannot be renamed for this rule.
func (s *Store) Scan(src any) error {
	_ = src

	return nil
}

func (s *Store) String() string {
	return ""
}

func (s *Store) Close() error {
	return nil
}

// A plain function is not a method on the store.
func Rows() int {
	return 0
}
