package class

type Class int

func (this Class) String() string {
	switch this {
	case Value:
		return "value"
	case IntLit:
		return "lit"
	case Local:
		return "local"
	case Global:
		return "global"
	case Internal_Global:
		return "_global"
	case Arg:
		return "arg"
	}
	panic("invalid class")
}

const (
	InvalidClass Class = iota

	Value // immutable, used to represent the output value of expressions

	Local           // mutable
	Arg             // mutable
	Global          // immutable
	Internal_Global // immutable, used to initialize constants
	IntLit          // immutable
)
