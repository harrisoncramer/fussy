package silent

import (
	"errors"
	"log"
)

type Kind int

const (
	KindUnknown Kind = iota
	KindRead
	KindWrite
)

var errUnknownKind = errors.New("unknown kind")

func ignored(k Kind) error {
	switch k {
	case KindRead:
		return nil
	default:
	}

	return nil
}

func noted(k Kind) error {
	switch k {
	case KindRead:
		return nil
	default:
		log.Printf("unhandled kind %d", k)
	}

	return nil
}

func stillGuessed(k Kind) (string, error) {
	switch k {
	case KindRead:
		return "read", nil
	default:
		return "write", nil // want `zero values beside its error` `non-nil error`
	}
}

func stillAssigned(k Kind) string {
	label := ""
	switch k {
	case KindRead:
		label = "read"
	default:
		label = "unknown" // want `must not assign to state read after the switch`
	}

	return label
}

func stillErroring(k Kind) error {
	switch k {
	case KindRead:
		return nil
	default:
		return errUnknownKind
	}
}
