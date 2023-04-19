package nodekind

import "strconv"

type NodeKind int

func (this NodeKind) String() string {
	switch this {
	case Terminal:
		return "term"
	case Block:
		return "block"
	case Module:
		return "module"
	case Procedure:
		return "procedure"
	case IdList:
		return "identifier list"
	case Signature:
		return "signature"
	case ArgumentList:
		return "argument list"
	case Argument:
		return "argument"
	case Call:
		return "call"
	case ExpressionList:
		return "expression list"
	case VarList:
		return "variable list"
	}
	return strconv.FormatInt(int64(this), 10)
}

const (
	InvalidNodeKind NodeKind = iota
	Terminal
	Module
	Procedure
	IdList
	Signature
	ArgumentList
	Argument
	Call
	Block
	ExpressionList
	VarList
)
