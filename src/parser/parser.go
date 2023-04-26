package parser

import (
	. "upt/core"
	ek "upt/core/errorkind"
	lk "upt/core/lexeme/lexkind"
	nk "upt/core/module/nodekind"
	sv "upt/core/severity"

	lex "upt/core/lexeme"
	mod "upt/core/module"

	lxr "upt/lexer"

	"fmt"
	"strings"
)

func Parse(filename string, contents string) (*mod.Node, *Error) {
	l := lxr.NewLexer(filename, contents)
	err := l.Next()
	if err != nil {
		return nil, err
	}
	n, err := portugol(l)
	if err != nil {
		return nil, err
	}
	if l.Word.Kind != lk.EOF {
		return nil, newError(l, ek.ExpectedEOF, "esperado final do arquivo")
	}
	computeRanges(n)
	return n, nil
}

// Portugol := {Funcao}.
func portugol(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "portugol")
	funcs, err := repeat(l, funcao)
	if err != nil {
		return nil, err
	}
	return &mod.Node{
		Leaves: funcs,
		Kind:   nk.Module,
	}, nil
}

// Funcao := [tipo] ident '(' [ArgList] ')' Bloco.
func funcao(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "funcao")
	if !(isType(l.Word) || l.Word.Kind == lk.Ident) {
		return nil, nil
	}
	var retNode *mod.Node
	var err *Error

	if isType(l.Word) {
		retNode, err = consume(l)
		if err != nil {
			return nil, err
		}
	}
	id, err := expect(l, lk.Ident)
	if err != nil {
		return nil, err
	}
	_, err = expect(l, lk.LeftParen)
	if err != nil {
		return nil, err
	}

	var args *mod.Node
	if l.Word.Kind != lk.RightParen {
		args, err = argList(l)
		if err != nil {
			return nil, err
		}
	}

	_, err = expect(l, lk.RightParen)
	if err != nil {
		return nil, err
	}
	bl, err := expectProd(l, bloco, "bloco")
	if err != nil {
		return nil, err
	}

	return &mod.Node{
		Leaves: []*mod.Node{id, args, retNode, bl},
		Kind:   nk.Procedure,
	}, nil
}

// ArgList := Arg {',' Arg}.
func argList(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "argList")
	arglist, err := repeatCommaList(l, arg)
	if err != nil {
		return nil, err
	}
	return &mod.Node{
		Leaves: arglist,
		Kind:   nk.ArgumentList,
	}, nil
}

// Arg := tipo ident.
func arg(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "arg")
	tipo, err := expect(l, lk.Real, lk.Inteiro, lk.Caractere)
	if err != nil {
		return nil, err
	}
	id, err := expect(l, lk.Ident)
	if err != nil {
		return nil, err
	}
	return &mod.Node{
		Leaves: []*mod.Node{tipo, id},
		Kind:   nk.ArgumentList,
	}, nil
}

// Bloco := '{' {Comando} '}'.
func bloco(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "bloco")
	_, err := expect(l, lk.LeftBrace)
	if err != nil {
		return nil, err
	}
	comandos, err := repeat(l, comando)
	if err != nil {
		return nil, err
	}
	_, err = expect(l, lk.RightBrace)
	if err != nil {
		return nil, err
	}
	return &mod.Node{
		Leaves: comandos,
		Kind:   nk.Block,
	}, nil
}

/*
Comando := Atrib term
         | VarDecl term
         | Expr term
         | Leia term
         | Imprima term
         | Se
         | Enquanto
         | Para.
*/
func comando(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "comando")
	switch l.Word.Kind {
	case lk.Leia:
		return prodSemicolon(l, leia)
	case lk.Imprima:
		return prodSemicolon(l, imprima)
	case lk.Se:
		return se(l)
	case lk.Enquanto:
		return enquanto(l)
	case lk.Para:
		return para(l)
	case lk.Retorne:
		return prodSemicolon(l, retorne)
	case lk.Caractere, lk.Real, lk.Inteiro:
		return prodSemicolon(l, varDecl)
	}
	if l.Word.Kind == lk.Ident {
		peeked, err := l.Peek()
		if err != nil {
			return nil, err
		}
		if peeked.Kind == lk.Assign {
			return prodSemicolon(l, atrib)
		}
	}
	return prodSemicolon(l, expr)
}

func prodSemicolon(l *lxr.Lexer, prod production) (*mod.Node, *Error) {
	n, err := prod(l)
	if err != nil {
		return nil, err
	}
	if n != nil {
		_, err = expect(l, lk.Semicolon)
		if err != nil {
			return nil, err
		}
	}
	return n, nil
}

// Retorne := "retorne" Expr.
func retorne(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "retorne")
	kw, err := expect(l, lk.Retorne)
	if err != nil {
		return nil, err
	}
	exp, err := expectProd(l, expr, "expressão")
	if err != nil {
		return nil, err
	}
	kw.Leaves = []*mod.Node{exp}
	return kw, nil
}

// Expr := AndExpr {'ou' AndExpr}.
func expr(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "expr")
	return repeatBinary(l, andExpr, "expressão", isOu)
}

func isOu(l *lex.Lexeme) bool {
	return l.Kind == lk.Ou
}

// AndExpr := CondExpr {'e' CondExpr}.
func andExpr(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "andExpr")
	return repeatBinary(l, compExpr, "expressão", isE)
}

func isE(l *lex.Lexeme) bool {
	return l.Kind == lk.E
}

// CondExpr := AddExpr {compOp AddExpr}.
func compExpr(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "compExpr")
	return repeatBinary(l, addExpr, "expressão", isCompOp)
}

func isCompOp(l *lex.Lexeme) bool {
	switch l.Kind {
	case lk.Equals, lk.Different, lk.Greater, lk.GreaterOrEquals, lk.Less, lk.LessOrEquals:
		return true
	}
	return false
}

// AddExpr := MultExpr {addOp MultExpr}.
func addExpr(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "addExpr")
	return repeatBinary(l, multExpr, "expressão", isAddOp)
}

func isAddOp(l *lex.Lexeme) bool {
	switch l.Kind {
	case lk.Plus, lk.Minus:
		return true
	}
	return false
}

// MultExpr := Unary {multOp Unary}.
func multExpr(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "multExpr")
	return repeatBinary(l, unary, "expressão unaria", isMultOp)
}

func isMultOp(l *lex.Lexeme) bool {
	switch l.Kind {
	case lk.Star, lk.Division, lk.Remainder:
		return true
	}
	return false
}

// Unary := [unaryOp] Termo [Call].
func unary(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "unary")
	var op *mod.Node
	var err *Error
	if isUnaryOp(l.Word) {
		op, err = consume(l)
		if err != nil {
			return nil, err
		}
	}
	n, err := termo(l)
	if err != nil {
		return nil, err
	}
	var c *mod.Node
	if l.Word.Kind == lk.LeftParen {
		c, err = call(l)
		if err != nil {
			return nil, err
		}
	}
	// preguiça de fazer precedencia
	if c != nil {
		c.Leaves = append([]*mod.Node{n}, c.Leaves...)
		if op != nil {
			op.Leaves = []*mod.Node{c}
			return op, nil
		}
		return c, nil
	}
	if op != nil {
		op.Leaves = []*mod.Node{n}
		return op, nil
	}
	return n, nil
}

// Call := '(' [ExprList] ')'
func call(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "call")
	_, err := expect(l, lk.LeftParen)
	if err != nil {
		return nil, err
	}
	exprs, err := repeatCommaList(l, expr)
	_, err = expect(l, lk.RightParen)
	if err != nil {
		return nil, err
	}
	exprList := &mod.Node{
		Leaves: exprs,
		Kind:   nk.ExpressionList,
	}
	return &mod.Node{
		Leaves: []*mod.Node{exprList},
		Kind:   nk.Call,
	}, nil
}

/*
Termo := literalInteiro
       | literalReal
       | literalCaracter
       | ident
       | '(' Expr ')'.
*/
func termo(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "termo")
	switch l.Word.Kind {
	case lk.IntLit, lk.RealLit, lk.CharLit, lk.Ident:
		return consume(l)
	case lk.LeftParen:
		_, err := consume(l)
		if err != nil {
			return nil, err
		}
		exp, err := expectProd(l, expr, "expressão")
		_, err = expect(l, lk.RightParen)
		if err != nil {
			return nil, err
		}
		return exp, nil
	}
	return nil, nil
}

func isUnaryOp(l *lex.Lexeme) bool {
	switch l.Kind {
	case lk.Minus, lk.Nao:
		return true
	}
	return false
}

// Atrib := ident "=" Expr.
func atrib(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "atrib")
	id, err := expect(l, lk.Ident)
	if err != nil {
		return nil, err
	}
	ass, err := expect(l, lk.Assign)
	if err != nil {
		return nil, err
	}
	exp, err := expectProd(l, expr, "expression")
	if err != nil {
		return nil, err
	}
	ass.Leaves = []*mod.Node{id, exp}
	return ass, nil
}

// Enquanto := 'enquanto' '(' Expr ')' Bloco.
func enquanto(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "enquanto")
	kw, err := expect(l, lk.Enquanto)
	if err != nil {
		return nil, err
	}
	_, err = expect(l, lk.LeftParen)
	if err != nil {
		return nil, err
	}
	exp, err := expectProd(l, expr, "expression")
	if err != nil {
		return nil, err
	}
	_, err = expect(l, lk.RightParen)
	if err != nil {
		return nil, err
	}
	bl, err := expectProd(l, bloco, "bloco")
	if err != nil {
		return nil, err
	}
	kw.Leaves = []*mod.Node{exp, bl}
	return kw, nil
}

// Para := 'para' '(' Atrib term Expr term Atrib ')' Bloco.
func para(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "para")
	kw, err := expect(l, lk.Para)
	if err != nil {
		return nil, err
	}
	_, err = expect(l, lk.LeftParen)
	if err != nil {
		return nil, err
	}
	initAtrib, err := expectProd(l, atrib, "atribuição")
	if err != nil {
		return nil, err
	}
	_, err = expect(l, lk.Semicolon)
	if err != nil {
		return nil, err
	}
	expr, err := expectProd(l, expr, "expressão")
	if err != nil {
		return nil, err
	}
	_, err = expect(l, lk.Semicolon)
	if err != nil {
		return nil, err
	}
	repeatAtrib, err := expectProd(l, atrib, "atribuição")
	if err != nil {
		return nil, err
	}
	_, err = expect(l, lk.RightParen)
	if err != nil {
		return nil, err
	}
	bl, err := expectProd(l, bloco, "bloco")
	if err != nil {
		return nil, err
	}
	kw.Leaves = []*mod.Node{initAtrib, expr, repeatAtrib, bl}
	return kw, nil
}

// Imprima := 'imprima' '(' ImpArg ')'.
func imprima(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "imprima")
	kw, err := expect(l, lk.Imprima)
	if err != nil {
		return nil, err
	}
	_, err = expect(l, lk.LeftParen)
	if err != nil {
		return nil, err
	}
	arg, err := expectProd(l, impArg, "mensagem ou expressão")
	if err != nil {
		return nil, err
	}
	_, err = expect(l, lk.RightParen)
	if err != nil {
		return nil, err
	}
	kw.Leaves = []*mod.Node{arg}
	return kw, nil
}

// ImpArg := mensagem | Expr.
func impArg(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "imparg")
	if l.Word.Kind == lk.StringLit {
		return consume(l)
	}
	return expr(l)
}

// Leia := 'leia' '(' ident ')'.
func leia(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "leia")
	kw, err := expect(l, lk.Leia)
	if err != nil {
		return nil, err
	}
	_, err = expect(l, lk.LeftParen)
	if err != nil {
		return nil, err
	}
	id, err := expect(l, lk.Ident)
	if err != nil {
		return nil, err
	}
	_, err = expect(l, lk.RightParen)
	if err != nil {
		return nil, err
	}
	kw.Leaves = []*mod.Node{id}
	return kw, nil
}

// Se := 'se' '(' Expr ')' Bloco [Senao].
func se(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "se")
	kw, err := expect(l, lk.Se)
	if err != nil {
		return nil, err
	}
	_, err = expect(l, lk.LeftParen)
	if err != nil {
		return nil, err
	}
	exp, err := expectProd(l, expr, "expression")
	if err != nil {
		return nil, err
	}
	_, err = expect(l, lk.RightParen)
	if err != nil {
		return nil, err
	}
	bl, err := expectProd(l, bloco, "bloco")
	if err != nil {
		return nil, err
	}

	var _senao *mod.Node
	if l.Word.Kind == lk.Senao {
		_senao, err = senao(l)
		if err != nil {
			return nil, err
		}
	}
	kw.Leaves = []*mod.Node{exp, bl, _senao}
	return kw, nil
}

// Senao := 'senao' Bloco.
func senao(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "senao")
	_, err := expect(l, lk.Senao)
	if err != nil {
		return nil, err
	}
	return expectProd(l, bloco, "bloco")
}

// VarDecl := tipo Idlist.
func varDecl(l *lxr.Lexer) (*mod.Node, *Error) {
	lxr.Track(l, "varDecl")
	t, err := expectType(l)
	if err != nil {
		return nil, err
	}
	idlist, err := repeatCommaList(l, ident)
	return &mod.Node{
		Leaves: append([]*mod.Node{t}, idlist...),
		Kind:   nk.VarDecl,
	}, nil
}

func ident(l *lxr.Lexer) (*mod.Node, *Error) {
	if l.Word.Kind == lk.Ident {
		return consume(l)
	}
	return nil, nil
}

func isType(l *lex.Lexeme) bool {
	switch l.Kind {
	case lk.Real, lk.Inteiro, lk.Caractere:
		return true
	}
	return false
}

func consume(l *lxr.Lexer) (*mod.Node, *Error) {
	lexeme := l.Word
	err := l.Next()
	if err != nil {
		return nil, err
	}
	rng := lexeme.Range
	n := &mod.Node{
		Lexeme: lexeme,
		Leaves: []*mod.Node{},
		Kind:   nk.Terminal,
		Range:  &rng,
		T:      nil,
	}
	return n, nil
}

func formatLexKinds(kinds []lk.LexKind) string {
	out := []string{}
	for _, kind := range kinds {
		out = append(out, kind.String())
	}
	return strings.Join(out, ", ")
}

func check(l *lxr.Lexer, tpList ...lk.LexKind) *Error {
	for _, tp := range tpList {
		if l.Word.Kind == tp {
			return nil
		}
	}
	list := formatLexKinds(tpList)
	message := fmt.Sprintf("esperado um de %v: ao invés disso foi achado %v",
		list,
		l.Word.Kind)

	err := newError(l, ek.ExpectedSymbol, message)
	return err
}

func expectType(l *lxr.Lexer) (*mod.Node, *Error) {
	return expect(l, lk.Real, lk.Caractere, lk.Inteiro)
}

func expect(l *lxr.Lexer, tpList ...lk.LexKind) (*mod.Node, *Error) {
	for _, tp := range tpList {
		if l.Word.Kind == tp {
			return consume(l)
		}
	}
	list := formatLexKinds(tpList)
	message := fmt.Sprintf("esperado um de %v: ao invés disso foi achado %v",
		list,
		l.Word.Kind)

	err := newError(l, ek.ExpectedSymbol, message)
	return nil, err
}

func expectProd(l *lxr.Lexer, prod production, name string) (*mod.Node, *Error) {
	n, err := prod(l)
	if err != nil {
		return nil, err
	}
	if n == nil {
		message := fmt.Sprintf("esperado %v ao invés disso foi achado %v", name, l.Word.Kind)
		err := newError(l, ek.ExpectedProd, message)
		return nil, err
	}
	return n, err
}

type production func(l *lxr.Lexer) (*mod.Node, *Error)
type validator func(*lex.Lexeme) bool

/* repeatBinary implements the following pattern
for a given Production and Terminal:

	repeatBinary := Production {Terminal Production}

Validator checks for terminals.
Left to Right precedence
*/
func repeatBinary(l *lxr.Lexer, prod production, name string, v validator) (*mod.Node, *Error) {
	last, err := prod(l)
	if err != nil {
		return nil, err
	}
	if last == nil {
		return nil, nil
	}
	for v(l.Word) {
		parent, err := consume(l)
		if err != nil {
			return nil, err
		}
		parent.AddLeaf(last)

		newLeaf, err := expectProd(l, prod, name)
		if err != nil {
			return nil, err
		}
		parent.AddLeaf(newLeaf)

		last = parent
	}
	return last, nil
}

/* repeat implements the following pattern
for a given Production:

	repeat := {Production}.
*/
func repeat(l *lxr.Lexer, prod production) ([]*mod.Node, *Error) {
	out := []*mod.Node{}
	n, err := prod(l)
	if err != nil {
		return nil, err
	}
	if n == nil {
		return nil, nil
	}
	for n != nil {
		out = append(out, n)
		n, err = prod(l)
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

// Implements the pattern:
//    RepeatCommaList := Production {',' Production} [','].
func repeatCommaList(l *lxr.Lexer, prod production) ([]*mod.Node, *Error) {
	first, err := prod(l)
	if err != nil {
		return nil, err
	}
	if first == nil {
		return nil, nil
	}
	out := []*mod.Node{first}
	for l.Word.Kind == lk.Comma {
		l.Next()
		n, err := prod(l)
		if err != nil {
			return nil, err
		}
		if n != nil {
			out = append(out, n)
		}
	}
	if l.Word.Kind == lk.Comma {
		err := l.Next()
		if err != nil {
			return nil, err
		}
	}
	return out, nil
}

func expectedEOF(l *lxr.Lexer) *Error {
	return newError(l, ek.ExpectedEOF, "simbolo inesperado, esperado fim do arquivo (EOF)")
}

func newError(l *lxr.Lexer, t ek.ErrorKind, message string) *Error {
	return &Error{
		Code:     t,
		Severity: sv.Error,
		Location: &Location{Range: &l.Word.Range, File: l.File},
		Message:  message,
	}
}

func newLexemeError(l *lxr.Lexer, lexeme *lex.Lexeme, t ek.ErrorKind, message string) *Error {
	return &Error{
		Code:     t,
		Severity: sv.Error,
		Location: &Location{Range: &lexeme.Range, File: l.File},
		Message:  message,
	}
}

// TODO: IMPROVE: consider delimiters in the range []{}()
func computeRanges(curr *mod.Node) {
	if curr == nil {
		return
	}
	for _, leaf := range curr.Leaves {
		computeRanges(leaf)
	}
	for _, n := range curr.Leaves {
		if n == nil || n.Range == nil {
			continue
		}
		if curr.Range == nil {
			r := *n.Range
			curr.Range = &r
			continue
		}
		if curr.Range.Begin.MoreThan(n.Range.Begin) {
			curr.Range.Begin = n.Range.Begin
		}
		if curr.Range.End.LessThan(n.Range.End) {
			curr.Range.End = n.Range.End
		}
	}
}