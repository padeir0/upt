package lexer

import (
	"strconv"
	lx "upt/core/lexeme"
	T "upt/core/lexeme/lexkind"

	. "upt/core"
	et "upt/core/errorkind"
	sv "upt/core/severity"

	"fmt"
	"strings"
	"unicode/utf8"
)

const IsTracking bool = false

func Track(st *Lexer, s string) {
	if IsTracking {
		fmt.Printf("%v: %v\n", s, st.Word.String())
	}
}

func NewLexerError(st *Lexer, t et.ErrorKind, message string) *Error {
	loc := st.GetSourceLocation()
	return &Error{
		Code:     t,
		Severity: sv.Error,
		Location: loc,
		Message:  message,
	}
}

func AllTokens(s *Lexer) []*lx.Lexeme {
	output := []*lx.Lexeme{}
	err := s.Next()
	if err != nil {
		panic(err)
	}
	for s.Word.Kind != T.EOF {
		output = append(output, s.Word)
		err = s.Next()
		if err != nil {
			panic(err)
		}
	}
	return output
}

type Lexer struct {
	Word *lx.Lexeme

	File                string
	BeginLine, BeginCol int
	EndLine, EndCol     int

	Start, End   int
	LastRuneSize int
	Input        string

	Peeked *lx.Lexeme
}

func NewLexer(filename string, s string) *Lexer {
	st := &Lexer{
		File:  filename,
		Input: s,
	}
	return st
}

func (this *Lexer) GetSourceLocation() *Location {
	rng := this.Range()
	return &Location{
		File:  this.File,
		Range: &rng,
	}
}

func (this *Lexer) Next() *Error {
	if this.Peeked != nil {
		p := this.Peeked
		this.Peeked = nil
		this.Word = p
		return nil
	}
	symbol, err := any(this)
	if err != nil {
		return err
	}
	this.Start = this.End // this shouldn't be here
	this.BeginLine = this.EndLine
	this.BeginCol = this.EndCol
	this.Word = symbol
	return nil
}

func (this *Lexer) Peek() (*lx.Lexeme, *Error) {
	symbol, err := any(this)
	if err != nil {
		return nil, err
	}
	this.Start = this.End
	this.Peeked = symbol
	return symbol, nil
}

func (this *Lexer) ReadAll() ([]*lx.Lexeme, *Error) {
	e := this.Next()
	if e != nil {
		return nil, e
	}
	output := []*lx.Lexeme{}
	for this.Word.Kind != T.EOF {
		output = append(output, this.Word)
		e = this.Next()
		if e != nil {
			return nil, e
		}
	}
	return output, nil
}

func (this *Lexer) Selected() string {
	return this.Input[this.Start:this.End]
}

func (this *Lexer) Range() Range {
	return Range{
		Begin: Position{
			Line:   this.BeginLine,
			Column: this.BeginCol,
		},
		End: Position{
			Line:   this.EndLine,
			Column: this.EndCol,
		},
	}
}

func genNumNode(l *Lexer, tp T.LexKind, value interface{}) *lx.Lexeme {
	text := l.Selected()
	n := &lx.Lexeme{
		Kind:  tp,
		Text:  text,
		Value: value,
		Range: l.Range(),
	}
	return n
}

func genNode(l *Lexer, tp T.LexKind) *lx.Lexeme {
	text := l.Selected()
	n := &lx.Lexeme{
		Kind:  tp,
		Text:  text,
		Range: l.Range(),
	}
	return n
}

func nextRune(l *Lexer) rune {
	r, size := utf8.DecodeRuneInString(l.Input[l.End:])
	if r == utf8.RuneError && size == 1 {
		panic("Invalid UTF8 rune in string")
	}
	l.End += size
	l.LastRuneSize = size

	if r == '\n' {
		l.EndLine++
		l.EndCol = 0
	} else {
		l.EndCol++
	}

	return r
}

func peekRune(l *Lexer) rune {
	r, size := utf8.DecodeRuneInString(l.Input[l.End:])
	if r == utf8.RuneError && size == 1 {
		panic("Invalid UTF8 rune in string")
	}

	return r
}

/*ignore ignores the text previously read*/
func ignore(l *Lexer) {
	l.Start = l.End
	l.BeginLine = l.EndLine
	l.BeginCol = l.EndCol
	l.LastRuneSize = 0
}

func acceptRun(l *Lexer, s string) {
	r := peekRune(l)
	for strings.ContainsRune(s, r) {
		nextRune(l)
		r = peekRune(l)
	}
}

func acceptUntil(l *Lexer, s string) {
	r := peekRune(l)
	for !strings.ContainsRune(s, r) {
		nextRune(l)
		r = peekRune(l)
	}
}

const (
	/*eof is equivalent to RuneError, but in this package it only shows up in EoFs
	If the rune is invalid, it panics instead*/
	eof rune = utf8.RuneError
)

const (
	insideStr  = `\"`
	insideChar = `\'`
	digits     = "0123456789"
	hex_digits = "0123456789ABCDEFabcdef"
	bin_digits = "01"
	letters    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ_" // yes _ is a letter, fuck you
)

func isNumber(r rune) bool {
	return strings.ContainsRune(digits, r)
}

func isLetter(r rune) bool {
	return strings.ContainsRune(letters, r)
}

func ignoreWhitespace(st *Lexer) {
	r := peekRune(st)
loop:
	for {
		switch r {
		case ' ', '\t', '\n':
			nextRune(st)
		case '#':
			comment(st)
		default:
			break loop
		}
		r = peekRune(st)
	}
	ignore(st)
}

// refactor this
func any(st *Lexer) (*lx.Lexeme, *Error) {
	var r rune
	var tp T.LexKind

	ignoreWhitespace(st)

	r = peekRune(st)

	if isNumber(r) {
		return number(st), nil
	}
	if isLetter(r) {
		return identifier(st), nil
	}
	if r == '\'' {
		return charLit(st)
	}
	if r == '"' {
		return strLit(st), nil
	}

	switch r {
	case '+':
		nextRune(st)
		tp = T.Plus
	case '-':
		nextRune(st)
		tp = T.Minus
	case '*':
		nextRune(st)
		tp = T.Star
	case '>': // >  >=
		nextRune(st)
		r = peekRune(st)
		switch r {
		case '=':
			nextRune(st)
			tp = T.GreaterOrEquals
		default:
			tp = T.Greater
		}
	case '<': // <  <=
		nextRune(st)
		r = peekRune(st)
		switch r {
		case '=':
			nextRune(st)
			tp = T.LessOrEquals
		default:
			tp = T.Less
		}
	case '!':
		nextRune(st)
		r = peekRune(st)
		switch r {
		case '=':
			nextRune(st)
			tp = T.Different
		default:
			message := fmt.Sprintf("simbolo invalido: %v", string(r))
			err := NewLexerError(st, et.InvalidSymbol, message)
			return nil, err
		}
	case '=':
		nextRune(st)
		r = peekRune(st)
		if r == '=' {
			nextRune(st)
			tp = T.Equals
		} else {
			tp = T.Assign
		}
	case '/':
		nextRune(st)
		tp = T.Division
	case '%':
		nextRune(st)
		tp = T.Remainder
	case '(':
		nextRune(st)
		tp = T.LeftParen
	case ')':
		nextRune(st)
		tp = T.RightParen
	case '{':
		nextRune(st)
		tp = T.LeftBrace
	case '}':
		nextRune(st)
		tp = T.RightBrace
	case ',':
		nextRune(st)
		tp = T.Comma
	case ';':
		nextRune(st)
		tp = T.Semicolon
	case eof:
		nextRune(st)
		return &lx.Lexeme{Kind: T.EOF}, nil
	default:
		message := fmt.Sprintf("simbolo invalido: %v", string(r))
		err := NewLexerError(st, et.InvalidSymbol, message)
		return nil, err
	}
	return genNode(st, tp), nil
}

// sorry
func number(st *Lexer) *lx.Lexeme {
	r := peekRune(st)
	var value interface{}
	if r == '0' {
		nextRune(st)
		r = peekRune(st)
		switch r {
		case 'x': // he x
			nextRune(st)
			acceptRun(st, hex_digits)
			value = parseHex(st.Selected())
		case 'b': // b inary
			nextRune(st)
			acceptRun(st, bin_digits)
			value = parseBin(st.Selected())
		default:
			acceptRun(st, digits)
			value = parseNormal(st.Selected())
		}
	} else {
		acceptRun(st, digits)
		value = parseNormal(st.Selected())
	}
	r = peekRune(st)
	if r == '.' { // real
		nextRune(st)
		acceptRun(st, digits)
		r = peekRune(st)
		if r == 'e' { // exponent
			nextRune(st)
			acceptRun(st, digits)
		}
		value = parseReal(st.Selected())
		return genNumNode(st, T.RealLit, value)
	}
	return genNumNode(st, T.IntLit, value)
}

func identifier(st *Lexer) *lx.Lexeme {
	r := peekRune(st)
	if !isLetter(r) {
		panic("identifier not beginning with letter")
	}
	acceptRun(st, digits+letters)
	selected := st.Selected()
	tp := T.Ident
	switch selected {
	case "retorne":
		tp = T.Retorne
	case "para":
		tp = T.Para
	case "enquanto":
		tp = T.Enquanto
	case "se":
		tp = T.Se
	case "senao":
		tp = T.Senao
	case "real":
		tp = T.Real
	case "inteiro":
		tp = T.Inteiro
	case "caractere":
		tp = T.Caractere
	case "imprima":
		tp = T.Imprima
	case "leia":
		tp = T.Leia
	case "ou":
		tp = T.Ou
	case "e":
		tp = T.E
	case "nao":
		tp = T.Nao
	}
	return genNode(st, tp)
}

func comment(st *Lexer) *Error {
	r := nextRune(st)
	if r != '#' {
		panic("internal error: comment without '#'")
	}
	for !strings.ContainsRune("\n"+string(eof), r) {
		nextRune(st)
		r = peekRune(st)
	}
	nextRune(st)
	return nil
}

func strLit(st *Lexer) *lx.Lexeme {
	r := nextRune(st)
	if r != '"' {
		panic("wong")
	}
	for {
		acceptUntil(st, insideStr)
		r := peekRune(st)
		if r == '"' {
			nextRune(st)
			return &lx.Lexeme{
				Text:  st.Selected(),
				Kind:  T.StringLit,
				Range: st.Range(),
			}
		}
		if r == '\\' {
			nextRune(st) // \
			nextRune(st) // escaped rune
		}
	}
}

func charLit(st *Lexer) (*lx.Lexeme, *Error) {
	r := nextRune(st)
	if r != '\'' {
		panic("wong")
	}
	for {
		acceptUntil(st, insideChar)
		r := peekRune(st)
		if r == '\'' {
			nextRune(st)
			text := st.Selected()
			value, err := parseCharLit(st, text[1:len(text)-1])
			if err != nil {
				return nil, err
			}
			return &lx.Lexeme{
				Text:  text,
				Kind:  T.CharLit,
				Range: st.Range(),
				Value: value,
			}, nil
		}
		if r == '\\' {
			nextRune(st) // \
			nextRune(st) // escaped rune
		}
	}
}

func IsValidIdentifier(s string) bool {
	st := NewLexer("oh no", s)
	tks, err := st.ReadAll()
	if err != nil {
		return false
	}
	if len(tks) != 1 { // we want only ID
		return false
	}
	return tks[0].Kind == T.Ident
}

func parseReal(text string) float64 {
	f, err := strconv.ParseFloat(text, 64)
	if err != nil {
		panic(err) // should never happen
	}
	return f
}

func parseNormal(text string) int64 {
	var output int64 = 0
	for i := range text {
		output *= 10
		char := text[i]
		if char >= '0' || char <= '9' {
			output += int64(char - '0')
		} else {
			panic(text)
		}
	}
	return output
}

func parseHex(oldText string) int64 {
	text := oldText[2:]
	var output int64 = 0
	for i := range text {
		output *= 16
		char := text[i]
		if char >= '0' && char <= '9' {
			output += int64(char - '0')
		} else if char >= 'a' && char <= 'f' {
			output += int64(char-'a') + 10
		} else if char >= 'A' && char <= 'F' {
			output += int64(char-'A') + 10
		} else {
			panic(text)
		}
	}
	return output
}

func parseBin(oldText string) int64 {
	text := oldText[2:]
	var output int64 = 0
	for i := range text {
		output *= 2
		char := text[i]
		if char == '0' || char == '1' {
			output += int64(char - '0')
		} else {
			panic(text)
		}
	}
	return output
}

func parseCharLit(l *Lexer, text string) (int64, *Error) {
	value := int64(text[0])
	if len(text) > 1 {
		switch text {
		case "\\n":
			value = '\n'
		case "\\t":
			value = '\t'
		case "\\r":
			value = '\r'
		case "\\'":
			value = '\''
		case "\\\"":
			value = '"'
		case "\\\\":
			value = '\\'
		default:
			return -1, NewLexerError(l, et.InvalidSymbol, "muitos caracteres no literal de caracteres")
		}
	}
	return value, nil
}
