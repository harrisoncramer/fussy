package verbs

type Rows struct{}

func (r *Rows) FetchAll() []string {
	return nil
}

func (r *Rows) ListAll() []string { // want `ListAll does not open with one of the store's verbs, so name it after what it does`
	return nil
}
