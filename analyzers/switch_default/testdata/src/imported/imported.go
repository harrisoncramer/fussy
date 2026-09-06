package imported

import (
	"errors"

	"kinds"
)

var errUnknownKind = errors.New("unknown kind")

func label(k kinds.Kind) (string, error) {
	switch k {
	case kinds.KindRead:
		return "read", nil
	default:
		return "", errUnknownKind
	}
}

func guessed(k kinds.Kind) (string, error) {
	switch k {
	case kinds.KindRead:
		return "read", nil
	default:
		return "write", nil // want `zero values beside its error` `non-nil error`
	}
}
