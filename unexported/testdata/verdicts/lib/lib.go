package lib

import "errors"

func InternalOnly() string { return "internal" }

func TestOnly() string { return "test" }

func BlackBoxOnly() string { return "black" }

func Nobody() string { return "nobody" }

func CrossTest() string { return "cross" }

func ForGenerator() string { return "generator" }

func Public() Surface {
	inner, err := inside()
	if err != nil {
		return Surface{Name: InternalOnly()}
	}

	return Surface{Name: inner.value + keeper()}
}

type Surface struct {
	Name string
}

type Inner struct {
	value string
}

type Mode int

const (
	ModeQuiet Mode = iota
	ModeLoud
)

type Level int

const LevelLow Level = 1

var ErrSentinel = errors.New("sentinel")

func inside() (Inner, error) {
	if pick() == LevelLow {
		return Inner{value: "low"}, nil
	}

	return Inner{}, ErrSentinel
}

func pick() Level { return LevelLow }

type SelfRef[T SelfRef[T]] interface {
	Self() T
}

func UseSelfRef[T SelfRef[T]](value T) T { return value.Self() }

func ForNested() string { return "nested" }

func Recursive(n int) int {
	if n <= 0 {
		return 0
	}

	return Recursive(n - 1)
}

func PairA() int { return PairB() }

func PairB() int { return PairA() }

type DeadHolder struct {
	Field DeadField
}

type DeadField struct{}

type DeadWithMethod struct{}

func (d DeadWithMethod) Describe() DeadWithMethod { return d }

func KeptByUnexported() string { return "kept" }

func keeper() string { return KeptByUnexported() }

func TestChainHead() string { return TestChainTail() }

func TestChainTail() string { return "tail" }

func DotImported() string { return "dotted" }

var (
	MixedExported = "exported"
	mixedInternal = KeptByMixedBlock()
)

func KeptByMixedBlock() string { return "mixed" }
