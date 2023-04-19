package lexeme

import (
	. "upt/core"
	lk "upt/core/lexeme/lexkind"
)

type Lexeme struct {
	Text  string
	Kind  lk.LexKind
	Value interface{} // int64 | float64
	Range Range
}

func (this Lexeme) String() string {
	return "(" + this.Text + ", " + this.Kind.String() + ")"
}
