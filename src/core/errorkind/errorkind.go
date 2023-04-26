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
	VarNotAssignable
	InvalidTypeForCond
	OpUnequalTypes
	ExpectedTypeOp
	InvalidFileName
	NoEntryPoint
	WrongEntryType
	ArgNotAssignable
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
	VarNotAssignable:      "E010",
	InvalidTypeForCond:    "E011",
	OpUnequalTypes:        "E012",
	ExpectedTypeOp:        "E013",
	InvalidFileName:       "E014",
	NoEntryPoint:          "E015",
	WrongEntryType:        "E016",
	ArgNotAssignable:      "E017",
}
