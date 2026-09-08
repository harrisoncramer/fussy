package verbs

type Rows struct{}

func (r *Rows) FetchAll() []string {
	return nil
}

func (r *Rows) ListAll() []string { // want `ListAll does not open with a verb, so name it after what it does: Fetch`
	return nil
}
