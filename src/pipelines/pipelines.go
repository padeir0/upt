package pipelines

import (
	"io/ioutil"
	// "os"
	// "os/exec"

	. "upt/core"
	lex "upt/core/lexeme"
	mod "upt/core/module"

	"upt/lexer"
	"upt/parser"
	"upt/resolution"
	"upt/typechecker"
)

// processes a single file and returns all tokens
// or an error
func Lexemes(file string) ([]*lex.Lexeme, *Error) {
	s, err := getFile(file)
	if err != nil {
		return nil, err
	}
	st := lexer.NewLexer(file, s)
	return st.ReadAll()
}

// processes a single file and returns it's AST
// or an error
func Ast(file string) (*mod.Node, *Error) {
	s, err := getFile(file)
	if err != nil {
		return nil, err
	}
	return parser.Parse(file, s)
}

// processes a file and all it's dependencies
// returns a typed Module or an error
func Mod(file string) (*mod.Module, *Error) {
	m, err := resolution.Resolve(file)
	if err != nil {
		return nil, err
	}

	err = typechecker.Check(m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

/*
// processes a Millipascal program and saves a binary
// into disk
func Compile(file string) (string, *Error) {
	fp, err := Mod(file)
	if err != nil {
		return "", err
	}
	ioerr := genBinary(fp)
	if ioerr != nil {
		return "", ProcessFileError(ioerr)
	}
	return fp.Name, nil
}

func genBinary(fp *amd64.FasmProgram) error {
	f, oserr := os.CreateTemp("", "mpc_*")
	if oserr != nil {
		return oserr
	}
	defer os.Remove(f.Name())
	_, oserr = f.WriteString(fp.Contents)
	if oserr != nil {
		return oserr
	}
	cmd := exec.Command("fasm", f.Name(), "./"+fp.Name)
	_, oserr = cmd.Output()
	if oserr != nil {
		return oserr
	}
	return nil
}
*/

func getFile(file string) (string, *Error) {
	text, e := ioutil.ReadFile(file)
	if e != nil {
		return "", ProcessFileError(e)
	}
	return string(text), nil
}
