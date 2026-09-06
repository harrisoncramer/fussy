package kinds

type Kind int

const (
	KindUnknown Kind = iota
	KindRead
	KindWrite
)
