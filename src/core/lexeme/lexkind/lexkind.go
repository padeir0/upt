package lexkind

type LexKind int

func (this LexKind) String() string {
	return mapToString[this]
}

const (
	InvalidLexKind LexKind = iota

	Ident
	IntLit
	StringLit
	CharLit
	RealLit

	Equals
	Different
	Greater
	GreaterOrEquals
	Less
	LessOrEquals

	Plus
	Minus
	Star
	Division
	Remainder

	Comma
	Semicolon
	LeftParen
	RightParen
	LeftBrace
	RightBrace

	Assign

	Para
	Enquanto
	Se
	Senao
	Real
	Inteiro
	Caractere
	Imprima
	Leia
	Ou
	E
	Nao

	EOF
)

var mapToString = map[LexKind]string{
	Ident:     "id",
	IntLit:    "int lit",
	StringLit: "string lit",
	CharLit:   "char lit",
	RealLit:   "real lit",

	Equals:          "==",
	Different:       "!=",
	Greater:         ">",
	GreaterOrEquals: ">=",
	Less:            "<",
	LessOrEquals:    "<=",

	Plus:      "+",
	Minus:     "-",
	Star:      "*",
	Division:  "/",
	Remainder: "%",

	Comma:     ",",
	Semicolon: ";",

	LeftParen:  "(",
	RightParen: ")",
	LeftBrace:  "{",
	RightBrace: "}",

	Assign: "=",

	Para:      "para",
	Enquanto:  "enquanto",
	Se:        "se",
	Senao:     "senao",
	Real:      "real",
	Inteiro:   "inteiro",
	Caractere: "caractere",
	Imprima:   "imprima",
	Leia:      "leia",
	Ou:        "ou",
	E:         "e",
	Nao:       "nao",

	EOF: "EOF",
}
