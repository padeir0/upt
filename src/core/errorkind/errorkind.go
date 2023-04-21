package errorkind

import "strconv"

type ErrorKind int

func (et ErrorKind) String() string {
	v, ok := ErrorCodeMap[et]
	if !ok {
		panic(strconv.FormatInt(int64(et), 10) + "is not stringified")
	}
	return v
}

const (
	InvalidErrType ErrorKind = iota
	InternalCompilerError
	FileError
	InvalidSymbol
	ExpectedEOF
	ExpectedSymbol
	ExpectedProd
	NameAlreadyDefined
	SymbolNotDeclared
)

var ErrorCodeMap = map[ErrorKind]string{
	InvalidErrType:        "E001",
	InternalCompilerError: "E002",
	FileError:             "E003",
	InvalidSymbol:         "E004",
	ExpectedEOF:           "E005",
	ExpectedSymbol:        "E006",
	ExpectedProd:          "E007",
	NameAlreadyDefined:    "E008",
	SymbolNotDeclared:     "E009",
}
