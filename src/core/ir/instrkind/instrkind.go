package instrkind

import "strconv"

type InstrKind int

func (i InstrKind) String() string {
	switch i {
	case Add: // [number, number] -> number
		return "add"
	case Sub:
		return "sub"
	case Div:
		return "div"
	case Mult:
		return "mult"
	case Rem:
		return "rem"
	case Eq: // [number, number] -> int
		return "eq"
	case Diff:
		return "diff"
	case Less:
		return "less"
	case More:
		return "more"
	case LessEq:
		return "lessEq"
	case MoreEq:
		return "moreEq"
	case Or: // [int, int] -> int
		return "or"
	case And:
		return "and"
	case Not: // [int] -> int
		return "not"
	case Neg: // [number] -> number
		return "neg"
	case Convert: // [value, type] -> newvalue
		return "convert"
	case Copy: // [source] -> dest
		return "copy"
	case Call: // [proc, arg1, arg2, ...] -> dest
		return "call"
	case Return: // [value]
		return "return"
	}
	panic("Unstringified InstrType: " + strconv.Itoa(int(i)))
}

const (
	InvalidInstr InstrKind = iota

	Add // [number, number] -> number
	Sub
	Div
	Mult
	Rem

	Eq // [number, number] -> int
	Diff
	Less
	More
	LessEq
	MoreEq

	Or // [int, int] -> int
	And

	Not // [int] -> int

	Neg // [number] -> number

	Print   // [value]
	Scan    // [] -> variable
	Convert // [value, type] -> newvalue
	Copy    // [source] -> dest
	Call    // [proc, arg1, arg2, ...] -> dest
	Return  // [value]
)
