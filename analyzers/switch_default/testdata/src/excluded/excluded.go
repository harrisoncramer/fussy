package excluded

type Kind int

const (
	KindUnknown Kind = iota
	KindRead
	KindWrite
)

func guessed(k Kind) (string, error) {
	switch k {
	case KindRead:
		return "read", nil
	default:
		return "write", nil
	}
}
