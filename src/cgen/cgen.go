package cgen

import (
	mod "upt/core/module"
	T "upt/core/types"

	lk "upt/core/lexeme/lexkind"
	nk "upt/core/module/nodekind"
	sk "upt/core/module/symbolkind"

	"fmt"
	"strconv"
	"strings"
)

//TODO: REQ: gerar código pras funções RAIZ e EXPO
func Gen(m *mod.Module) string {
	ctx := newCtx(m)
	return defaultHeaders +
		forwardDecl(ctx) +
		genMain(ctx) +
		genFunctions(ctx)
}

const defaultHeaders = `
#include <stdio.h>
#include <math.h>
`

func forwardDecl(ctx *context) string {
	output := ""
	for _, sy := range ctx.M.Global.Symbols {
		output += forwardDeclFunc(ctx, sy) + "\n"
	}
	return output
}

func forwardDeclFunc(ctx *context, sy *mod.Symbol) string {
	cID := globalIDtoC(ctx.M, sy)
	ctx.GlobalMap[sy.Name] = cID
	retType := typetoCtype(sy.Type.Proc.Ret)

	args := []string{}
	for _, arg := range sy.Args {
		cType := typetoCtype(arg.T)
		args = append(args, cType)
	}
	return fmt.Sprintf("%v %v(%v);", retType, cID, strings.Join(args, ", "))
}

const defaultMain = `
int main() {
	return %v();
}
`

func genMain(ctx *context) string {
	return fmt.Sprintf(defaultMain, ctx.M.Name+"_entrada")
}

func genFunctions(ctx *context) string {
	output := ""
	for _, sy := range ctx.M.Global.Symbols {
		// precisamos resetar isso pra cada função
		ctx.LocalMap = map[scopedSymbol]string{}
		output += genFunc(ctx, sy) + "\n"
	}
	return output
}

func genFunc(ctx *context, sy *mod.Symbol) string {
	scope := sy.N.Scope

	cID := globalIDtoC(ctx.M, sy)
	ctx.GlobalMap[sy.Name] = cID
	retType := typetoCtype(sy.Type.Proc.Ret)

	args := []string{}
	for _, arg := range sy.Args {
		cArg := ctx.SetLocal(scope, arg.Name)
		cType := typetoCtype(arg.T)
		out := cType + " " + cArg
		args = append(args, out)
	}

	bl := sy.N.Leaves[3]
	block := genBlock(ctx, scope, bl)

	return fmt.Sprintf("%v %v(%v)\n%v",
		retType,
		cID,
		strings.Join(args, ", "),
		block)
}

func genBlock(ctx *context, scope *mod.Scope, bl *mod.Node) string {
	out := ctx.indent() + "{\n"
	ctx.IndentLevel++
	scope = bl.Scope
	for _, cmd := range bl.Leaves {
		out += ctx.indent() + genCmd(ctx, scope, cmd) + "\n"
	}
	ctx.IndentLevel--
	return out + ctx.indent() + "}\n"
}

func genCmd(ctx *context, scope *mod.Scope, n *mod.Node) string {
	switch n.Kind {
	case nk.Terminal:
		switch n.Lexeme.Kind {
		case lk.Leia:
			return genLeia(ctx, scope, n)
		case lk.Imprima:
			return genImprima(ctx, scope, n)
		case lk.Se:
			return genSe(ctx, scope, n)
		case lk.Enquanto:
			return genEnquanto(ctx, scope, n)
		case lk.Para:
			return genPara(ctx, scope, n)
		case lk.Retorne:
			return genRetorne(ctx, scope, n)
		case lk.Assign:
			return genAtrib(ctx, scope, n) + ";"
		}
	case nk.Block:
		return genBlock(ctx, scope, n)
	case nk.VarDecl:
		return genVarDecl(ctx, scope, n)
	}
	return genExpr(ctx, scope, n) + ";"
}

// caracter -> scanf("%c", &variable)
// inteiro  -> scanf("%d", &variable)
// real     -> scanf("%f", &variable)
func genLeia(ctx *context, scope *mod.Scope, n *mod.Node) string {
	arg := n.Leaves[0]
	format := typeToFormat(arg.T)
	name := arg.Lexeme.Text
	_, sc := scope.FindWithScope(name)
	cName := ctx.FindLocal(sc, name)
	return "scanf(\"" + format + "\", &" + cName + ");"
}

func typeToFormat(t *T.Type) string {
	switch t.Basic {
	case T.Caractere:
		return "%c"
	case T.Inteiro:
		return "%d"
	case T.Real:
		return "%lf"
	default:
		return "%d"
	}
}

func genImprima(ctx *context, scope *mod.Scope, n *mod.Node) string {
	arg := n.Leaves[0]
	if arg.Lexeme != nil && arg.Lexeme.Kind == lk.StringLit {
		// a gente mantem as aspas no token da string
		return "printf(" + arg.Lexeme.Text + ");"
	}
	CArg := genExpr(ctx, scope, arg)
	format := typeToFormat(arg.T)
	return "printf(\"" + format + "\", " + CArg + ");\n"
}

func genSe(ctx *context, scope *mod.Scope, n *mod.Node) string {
	// se := {cond, block, senao}
	cond := genExpr(ctx, scope, n.Leaves[0])
	block := genBlock(ctx, scope, n.Leaves[1])
	senao := ""
	sn := n.Leaves[2]
	if sn != nil {
		senao = ctx.indent() + "else\n" + genBlock(ctx, scope, sn)
	}
	return fmt.Sprintf("if (%v)\n%v %v",
		cond, block, senao)
}

func genEnquanto(ctx *context, scope *mod.Scope, n *mod.Node) string {
	// enquanto := {cond, block}
	cond := genExpr(ctx, scope, n.Leaves[0])
	block := genBlock(ctx, scope, n.Leaves[1])
	return fmt.Sprintf("while (%v)\n %v", cond, block)
}

func genPara(ctx *context, scope *mod.Scope, n *mod.Node) string {
	// para := {atrib, cond, atrib, block}
	first := ""
	if n.Leaves[0] != nil {
		first = genAtrib(ctx, scope, n.Leaves[0])
	}
	cond := genExpr(ctx, scope, n.Leaves[1])
	second := genAtrib(ctx, scope, n.Leaves[2])
	block := genBlock(ctx, scope, n.Leaves[3])
	return fmt.Sprintf("for (%v; %v; %v)\n%v",
		first, cond, second, block)
}

func genRetorne(ctx *context, scope *mod.Scope, n *mod.Node) string {
	return "return " + genExpr(ctx, scope, n.Leaves[0]) + ";"
}

func genAtrib(ctx *context, scope *mod.Scope, n *mod.Node) string {
	dest := n.Leaves[0]
	name := dest.Lexeme.Text
	expr := n.Leaves[1]
	_, sc := scope.FindWithScope(name)
	cName := ctx.FindLocal(sc, name)
	return cName + " = " + genExpr(ctx, scope, expr)
}

func genVarDecl(ctx *context, scope *mod.Scope, n *mod.Node) string {
	t := n.Leaves[0].T
	cType := typetoCtype(t)
	ids := []string{}
	for _, id := range n.Leaves[1:] {
		name := id.Lexeme.Text
		cName := ctx.SetLocal(scope, name)
		ids = append(ids, cName)
	}
	return cType + " " + strings.Join(ids, ", ") + ";"
}

// geramos todas as expressões com parentesis pra ter certeza de que
// a ordem de precedencia da linguagem fonte é respeitada
func genExpr(ctx *context, scope *mod.Scope, n *mod.Node) string {
	switch n.Kind {
	case nk.Terminal:
		switch n.Lexeme.Kind {
		case lk.Ou, lk.E, lk.Equals, lk.Different,
			lk.Greater, lk.GreaterOrEquals, lk.Less, lk.LessOrEquals,
			lk.Plus, lk.Star, lk.Division,
			lk.Remainder:
			return genBinExpr(ctx, scope, n)
		case lk.Nao:
			return "(!" + genExpr(ctx, scope, n.Leaves[0]) + ")"
		case lk.Minus:
			if len(n.Leaves) == 1 {
				return "(-" + genExpr(ctx, scope, n.Leaves[0]) + ")"
			}
			return genBinExpr(ctx, scope, n)
		case lk.IntLit, lk.RealLit, lk.CharLit:
			return litToC(n)
		case lk.Ident:
			name := n.Lexeme.Text
			sy, sc := scope.FindWithScope(name)
			switch sy.Kind {
			case sk.Local, sk.Argument:
				return ctx.FindLocal(sc, name)
			case sk.Procedure:
				return ctx.GlobalMap[name]
			}
			panic("unreachable")
		}
	case nk.Call:
		return genCall(ctx, scope, n)
	}
	fmt.Println(n)
	panic("unreachable")
}

func genCall(ctx *context, scope *mod.Scope, n *mod.Node) string {
	proc := n.Leaves[0]
	cProc := genExpr(ctx, scope, proc)

	cArgs := []string{}
	args := n.Leaves[1]
	for _, expr := range args.Leaves {
		carg := genExpr(ctx, scope, expr)
		cArgs = append(cArgs, carg)
	}

	return fmt.Sprintf("%v(%v)", cProc, strings.Join(cArgs, ", "))
}

func genBinExpr(ctx *context, scope *mod.Scope, n *mod.Node) string {
	op := opToC(n.Lexeme.Kind)
	left := n.Leaves[0]
	right := n.Leaves[1]
	return fmt.Sprintf("(%v %v %v)",
		genExpr(ctx, scope, left),
		op,
		genExpr(ctx, scope, right))
}

func opToC(kind lk.LexKind) string {
	switch kind {
	case lk.Ou:
		return "||"
	case lk.E:
		return "&&"
	case lk.Equals:
		return "=="
	case lk.Different:
		return "!="
	case lk.Greater:
		return ">"
	case lk.GreaterOrEquals:
		return ">="
	case lk.Less:
		return "<"
	case lk.LessOrEquals:
		return "<="
	case lk.Plus:
		return "+"
	case lk.Star:
		return "*"
	case lk.Division:
		return "/"
	case lk.Remainder:
		return "%"
	case lk.Minus:
		return "-"
	}
	fmt.Println(kind)
	panic("unreachable")
}

func litToC(n *mod.Node) string {
	switch n.Lexeme.Kind {
	case lk.CharLit, lk.IntLit:
		v := n.Lexeme.Value.(int64)
		return fmt.Sprintf("%v", v)
	case lk.RealLit:
		v := n.Lexeme.Value.(float64)
		return fmt.Sprintf("%v", v)
	}
	panic("unreachable")
}

type scopedSymbol struct {
	ScopeID int
	Name    string
}

type context struct {
	M           *mod.Module
	GlobalMap   map[string]string
	LocalMap    map[scopedSymbol]string
	IndentLevel int
}

func newCtx(M *mod.Module) *context {
	return &context{
		M:           M,
		GlobalMap:   map[string]string{},
		LocalMap:    map[scopedSymbol]string{},
		IndentLevel: 0,
	}
}

func (this *context) indent() string {
	output := ""
	for i := 0; i < this.IndentLevel; i++ {
		output += "\t"
	}
	return output
}

func (this *context) FindLocal(scope *mod.Scope, name string) string {
	ss := scopedSymbol{
		ScopeID: scope.ID,
		Name:    name,
	}
	v, ok := this.LocalMap[ss]
	if !ok {
		fmt.Println(scope.ID, name)
		panic("symbol not found")
	}
	return v
}

func (this *context) SetLocal(scope *mod.Scope, name string) string {
	ss := scopedSymbol{
		ScopeID: scope.ID,
		Name:    name,
	}
	newName := localIDtoC(scope, name)
	this.LocalMap[ss] = newName
	return newName
}

// Não sei se as regras de escopo da linguagem alvo
// vão refletir as mesmas regras da linguagem fonte
// então todos os identificadores terão de ser unicos.
// A maneira mais facil de fazer isso é concatenando
// o identificador do escopo no nome da variavel.
func localIDtoC(scope *mod.Scope, id string) string {
	return id + strconv.Itoa(scope.ID)
}

// Pra ter certeza de que não vai ter conflito
// entre nomes da stdlib de C e globais, todos são
// prefixados com o nome do arquivo.
// (eu verifiquei um por um e nenhum usa undescore '_')
func globalIDtoC(mod *mod.Module, sy *mod.Symbol) string {
	return mod.Name + "_" + sy.Name
}

func typetoCtype(t *T.Type) string {
	if T.IsProc(t) {
		panic("unimplemented")
	}
	switch t.Basic {
	case T.Caractere:
		return "char"
	case T.Real:
		return "double"
	case T.Inteiro:
		return "int"
	case T.Void:
		return "void"
	case T.String:
		return "char *" // talvez não funcione
	}
	panic("unreachable")
}
