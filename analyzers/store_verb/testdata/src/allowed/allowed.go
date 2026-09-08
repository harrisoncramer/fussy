package allowed

type Store struct{}

func (s *Store) Rows() int {
	return 0
}

func (s *Store) String() string { // want `String does not open with one of the store's verbs, so name it after what it does`
	return ""
}
