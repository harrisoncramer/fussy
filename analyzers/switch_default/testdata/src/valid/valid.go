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

type Node interface {
	node()
}

type Leaf struct{}

func (Leaf) node() {}

type MyErr struct {
	Kind Kind
}

func (e *MyErr) Error() string {
	return "unknown kind"
}

func (k Kind) String() string {
	switch k {
	case KindRead:
		return "read"
	case KindWrite:
		return "write"
	default:
		return fmt.Sprintf("Kind(%d)", int(k))
	}
}

func sealed(n Node) (string, error) {
	switch n.(type) {
	case Leaf:
		return "leaf", nil
	default:
		return "", errUnknownKind
	}
}

func openWorld(value any) string {
	switch value.(type) {
	case string:
		return "string"
	default:
		return "something else"
	}
}

func concrete(k Kind) (string, *MyErr) {
	switch k {
	case KindRead:
		return "read", nil
	default:
		return "", &MyErr{Kind: k}
	}
}

func declaredZero(k Kind) (Result, error) {
	switch k {
	case KindRead:
		return Result{Name: "read"}, nil
	default:
		var zero Result

		return zero, errUnknownKind
	}
}

func nestedSwitch(k Kind, other Kind) (string, error) {
	switch k {
	case KindRead:
		return "read", nil
	default:
		switch other {
		case KindWrite:
			return "", errUnknownKind
		default:
			return "", errUnknownKind
		}
	}
}

func reported(k Kind) {
	switch k {
	case KindRead:
		log.Print("read")
	default:
		log.Print("unhandled kind")
	}
}

func localCounters(k Kind, names []string) (string, error) {
	switch k {
	case KindRead:
		return "read", nil
	default:
		seen := 0
		for i := 0; i < len(names); i++ {
			seen++
		}
		_ = seen

		return "", errUnknownKind
	}
}

func labelledLoop(k Kind, names []string) (string, error) {
	switch k {
	case KindRead:
		return "read", nil
	default:
	inner:
		for _, name := range names {
			if name != "" {
				break inner
			}
		}

		return "", errUnknownKind
	}
}

func labelledJump(k Kind) (string, error) {
	switch k {
	case KindRead:
		return "read", nil
	default:
		goto done
	done:

		return "", errUnknownKind
	}
}

func zeroBeforeMutation(k Kind, wanted bool) (Result, error) {
	switch k {
	case KindRead:
		return Result{Name: "read"}, nil
	default:
		var zero Result
		if wanted {
			return zero, errUnknownKind
		}
		zero.Name = "guess"
		_ = zero

		return Result{}, errUnknownKind
	}
}
