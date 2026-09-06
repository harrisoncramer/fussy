package invalid

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

type Result struct {
	Name string
}

var errUnknownKind = errors.New("unknown kind")

func guessed(k Kind) (string, error) {
	switch k {
	case KindRead:
		return "read", nil
	default:
		return "write", nil // want `zero values beside its error` `non-nil error`
	}
}

func swallowed(k Kind) (*Result, error) {
	switch k {
	case KindRead:
		return &Result{}, nil
	default:
		return nil, nil // want `non-nil error`
	}
}

func answered(k Kind) error {
	switch k {
	case KindRead:
		return nil
	default:
		return nil // want `non-nil error`
	}
}

func claimed(k Kind) (Result, bool) {
	switch k {
	case KindRead:
		return Result{Name: "read"}, true
	default:
		return Result{}, true // want `must return false in its last result`
	}
}

func predicate(k Kind) bool {
	switch k {
	case KindRead:
		return false
	default:
		return true // want `must return false in its last result`
	}
}

func assigned(k Kind) string {
	label := ""
	switch k {
	case KindRead:
		label = "read"
	default:
		label = "unknown" // want `must not assign to state read after the switch`
	}

	return label
}

func skipped(kinds []Kind) int {
	count := 0
	for _, k := range kinds {
		switch k {
		case KindRead:
			count++
		default:
			continue // want `must not leave the switch`
		}
	}

	return count
}

func empty(k Kind) error {
	switch k {
	case KindRead:
		return nil
	default: // want `must return an error or panic`
	}

	return nil
}

func logged(k Kind) error {
	switch k {
	case KindRead:
		return nil
	default: // want `must return an error or panic`
		log.Printf("unhandled kind %d", k)
	}

	return nil
}

func typed(node any) (string, error) {
	switch node.(type) {
	case *Result:
		return "result", nil
	default:
		return "something", nil // want `zero values beside its error` `non-nil error`
	}
}

func opaque(k Kind) (string, bool) {
	switch k {
	case KindRead:
		return "read", true
	default:
		return fallback() // want `must return a failure this rule can see`
	}
}

func fallback() (string, bool) {
	return "fallback", true
}

func namedGuess(k Kind) (out string, err error) {
	switch k {
	case KindRead:
		out = "read"
		return out, nil
	default:
		out = "unknown" // want `zero values beside its error`
		return          // want `non-nil error`
	}
}

func fieldSet(k Kind, r *Result) error {
	switch k {
	case KindRead:
		return nil
	default:
		r.Name = "unknown" // want `must not assign to state read after the switch`
		return errUnknownKind
	}
}
