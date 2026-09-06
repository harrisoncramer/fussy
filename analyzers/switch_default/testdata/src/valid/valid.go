package valid

import (
	"errors"
	"fmt"
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

func label(k Kind) (string, error) {
	switch k {
	case KindRead:
		return "read", nil
	case KindWrite:
		return "write", nil
	default:
		return "", fmt.Errorf("kind %d: %w", k, errUnknownKind)
	}
}

func validate(k Kind) error {
	switch k {
	case KindRead, KindWrite:
		return nil
	default:
		return errUnknownKind
	}
}

func lookup(k Kind) (Result, bool) {
	switch k {
	case KindRead:
		return Result{Name: "read"}, true
	default:
		return Result{}, false
	}
}

func mustLabel(k Kind) string {
	switch k {
	case KindRead:
		return "read"
	default:
		panic("unhandled kind")
	}
}

func zeroed(k Kind) string {
	switch k {
	case KindRead:
		return "read"
	default:
		return ""
	}
}

func named(k Kind) (out string, err error) {
	switch k {
	case KindRead:
		out = "read"
		return out, nil
	default:
		err = errUnknownKind
		return
	}
}

func delegated(k Kind) (string, error) {
	switch k {
	case KindRead:
		return "read", nil
	default:
		return unhandled(k)
	}
}

func unhandled(k Kind) (string, error) {
	return "", fmt.Errorf("kind %d: %w", k, errUnknownKind)
}

func zeroConst(k Kind) (Kind, error) {
	switch k {
	case KindRead:
		return KindRead, nil
	default:
		return KindUnknown, errUnknownKind
	}
}

func described(node any) (string, error) {
	switch node.(type) {
	case *Result:
		return "result", nil
	default:
		return "", errUnknownKind
	}
}

func noisy(k Kind) error {
	switch k {
	case KindRead:
		return nil
	default:
		log.Printf("unhandled kind %d", k)
		return errUnknownKind
	}
}

func localLoop(k Kind, names []string) (string, error) {
	switch k {
	case KindRead:
		return "read", nil
	default:
		for _, name := range names {
			if name != "" {
				break
			}
		}

		return "", errUnknownKind
	}
}

func plainString(s string) string {
	switch s {
	case "read":
		return "reading"
	default:
		return "something else"
	}
}

func comparison(k Kind) string {
	label := "small"
	switch {
	case k > KindRead:
		label = "big"
	default:
		label = "small"
	}

	return label
}

func closureInside(k Kind) (string, error) {
	switch k {
	case KindRead:
		return "read", nil
	default:
		render := func(k Kind) string {
			switch k {
			case KindRead:
				return "read"
			default:
				return ""
			}
		}
		_ = render

		return "", errUnknownKind
	}
}
