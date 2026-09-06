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

type Node interface {
	node()
}

type Leaf struct{}

func (Leaf) node() {}

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

func pointed(k Kind) (*Result, error) {
	switch k {
	case KindRead:
		return &Result{}, nil
	default:
		return &Result{}, errUnknownKind // want `zero values beside its error`
	}
}

func claimed(k Kind) (Result, bool) {
	switch k {
	case KindRead:
		return Result{Name: "read"}, true
	default:
		return Result{}, true // want `must return false`
	}
}

func predicate(k Kind) bool {
	switch k {
	case KindRead:
		return false
	default:
		return true // want `must return false`
	}
}

func labelled(k Kind) string {
	switch k {
	case KindRead:
		return "read"
	default:
		return "unknown" // want `must return the zero value`
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

func jumped(k Kind) error {
	switch k {
	case KindRead:
		return nil
	default:
		goto done // want `must not leave the switch`
	}

done:

	return nil
}

func fellThrough(k Kind) (string, error) {
	switch k {
	default:
		fallthrough // want `must not leave the switch`
	case KindRead:
		return "read", nil
	}
}

func brokeOutward(kinds []Kind) int {
	count := 0
outer:
	for _, k := range kinds {
		switch k {
		case KindRead:
			count++
		default:
			break outer // want `must not leave the switch`
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

func typed(n Node) (string, error) {
	switch n.(type) {
	case Leaf:
		return "leaf", nil
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

func namedLater(k Kind, fine bool) (out string, err error) {
	switch k {
	case KindRead:
		return "read", nil
	default:
		out = "guess" // want `zero values beside its error`
		if !fine {
			return // want `non-nil error`
		}
		out = ""
		err = errUnknownKind

		return
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

func loopHeader(k Kind, names []string) (string, error) {
	index := 0
	switch k {
	case KindRead:
		return "read", nil
	default:
		for index = 0; index < len(names); index++ { // want `must not assign to state read after the switch` `must not assign to state read after the switch`
		}

		return "", errUnknownKind
	}
}

func rangeHeader(k Kind, names []string) (string, error) {
	var seen string
	switch k {
	case KindRead:
		return "read", nil
	default:
		for _, seen = range names { // want `must not assign to state read after the switch`
		}
		_ = seen

		return "", errUnknownKind
	}
}

func (k Kind) String() string {
	render := func(k Kind) (string, error) {
		switch k {
		case KindRead:
			return "read", nil
		default:
			return "guess", nil // want `zero values beside its error` `non-nil error`
		}
	}
	label, _ := render(k)

	return label
}
