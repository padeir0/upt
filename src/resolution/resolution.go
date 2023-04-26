package resolution

import (
	. "upt/core"
	mod "upt/core/module"
	lexer "upt/lexer"

	ek "upt/core/errorkind"
	lk "upt/core/lexeme/lexkind"
	nk "upt/core/module/nodekind"
	sk "upt/core/module/symbolkind"
	sv "upt/core/severity"

	"fmt"
	"strings"
)

func Resolve(fullpath string, root *mod.Node) (*mod.Module, *Error) {
	name, err := extractName(fullpath)
	if err != nil {
		return nil, err
	}
	ctx := newCtx(fullpath, name, root)

	err = declareGlobals(ctx, root)
	if err != nil {
		return nil, err
	}

	err = resolveInnerScopes(ctx)
	if err != nil {
		return nil, err
	}

	sy := ctx.M.Global.Find("entrada")
	if sy == nil {
		return nil, errorEntryPointNotFound(ctx.M)
	}

	return ctx.M, nil
}

func declareGlobals(ctx *context, root *mod.Node) *Error {
	for _, leaf := range root.Leaves {
		if leaf.Kind != nk.Procedure {
			panic("invalid node kind for symbol")
		}
		err := declareProc(ctx, leaf)
		if err != nil {
			return err
		}
	}
	return nil
}

func resolveInnerScopes(ctx *context) *Error {
	for _, sy := range ctx.M.Global.Symbols {
		if sy.Kind != sk.Procedure {
			panic("invalid symbol kind")
		}
		err := resolveProcScopes(ctx, sy)
		if err != nil {
			return err
		}
	}
	return nil
}

func extractName(filePath string) (string, *Error) {
	path := strings.Split(filePath, "/")
	name := strings.Split(path[len(path)-1], ".")
	if lexer.IsValidIdentifier(name[0]) {
		return name[0], nil
	}
	return "", &Error{
		Code:     ek.InvalidFileName,
		Severity: sv.Error,
		Message:  filePath + " : " + name[0],
	}
}

type context struct {
	M            *mod.Module
	ScopeCounter int
}

func newCtx(fullpath, name string, root *mod.Node) *context {
	return &context{
		M: &mod.Module{
			FullPath: fullpath,
			Name:     name,
			Root:     root,
			Global: &mod.Scope{
				ID:      0,
				Parent:  mod.Universe,
				Symbols: map[string]*mod.Symbol{},
			},
		},
	}
}

func (this *context) NewScope(parent *mod.Scope) *mod.Scope {
	id := this.ScopeCounter
	this.ScopeCounter++
	return &mod.Scope{
		ID:      id,
		Parent:  parent,
		Symbols: map[string]*mod.Symbol{},
	}
}

func declareProc(ctx *context, n *mod.Node) *Error {
	id := n.Leaves[0]
	name := id.Lexeme.Text
	sy, ok := ctx.M.Global.Symbols[name]
	if ok {
		return errorNameAlreadyDefined(ctx.M, id)
	}
	sy = &mod.Symbol{
		Kind:    sk.Procedure,
		Name:    name,
		N:       n,
		Builtin: false,
		Args:    []mod.Arg{},
	}
	ctx.M.Global.Add(name, sy)
	return nil
}

func resolveProcScopes(ctx *context, sy *mod.Symbol) *Error {
	argScope := ctx.NewScope(ctx.M.Global)
	sy.N.Scope = argScope
	argMap := []mod.Arg{}

	// proc := {id, args, retNode, bl}
	args := sy.N.Leaves[1]
	if args != nil {
		argMap = make([]mod.Arg, len(args.Leaves))
		for i, arg := range args.Leaves {
			// arg := {tipo, id}
			id := arg.Leaves[1]
			name := id.Lexeme.Text
			_, ok := argScope.Symbols[name]
			if ok {
				return errorNameAlreadyDefined(ctx.M, id)
			}
			sy := &mod.Symbol{
				Kind:    sk.Argument,
				Name:    name,
				N:       arg,
				Builtin: false,
			}
			argScope.Add(name, sy)
			argMap[i] = mod.Arg{
				N:    arg,
				Name: name,
				Pos:  i,
			}
		}
	}
	sy.Args = argMap
	bl := sy.N.Leaves[3]
	err := resolveBlock(ctx, argScope, bl)
	if err != nil {
		return err
	}
	return nil
}

func resolveBlock(ctx *context, scope *mod.Scope, bl *mod.Node) *Error {
	newScope := ctx.NewScope(scope)
	bl.Scope = newScope
	for _, cmd := range bl.Leaves {
		err := resolveCmd(ctx, newScope, cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func resolveCmd(ctx *context, scope *mod.Scope, n *mod.Node) *Error {
	switch n.Kind {
	case nk.Terminal:
		switch n.Lexeme.Kind {
		case lk.Leia:
			return resolveLeia(ctx, scope, n)
		case lk.Imprima:
			return resolveImprima(ctx, scope, n)
		case lk.Se:
			return resolveSe(ctx, scope, n)
		case lk.Enquanto:
			return resolveEnquanto(ctx, scope, n)
		case lk.Para:
			return resolvePara(ctx, scope, n)
		case lk.Retorne:
			return resolveRetorne(ctx, scope, n)
		case lk.Assign:
			return resolveAtrib(ctx, scope, n)
		}
	case nk.Block:
		return resolveBlock(ctx, scope, n)
	case nk.VarDecl:
		return resolveVarDecl(ctx, scope, n)
	}
	return resolveExpr(ctx, scope, n)
}

func resolveLeia(ctx *context, scope *mod.Scope, n *mod.Node) *Error {
	id := n.Leaves[0]
	name := id.Lexeme.Text
	sy := scope.Find(name)
	if sy == nil {
		return errorSymbolNotDeclared(ctx.M, id)
	}
	return nil
}

func resolveImprima(ctx *context, scope *mod.Scope, n *mod.Node) *Error {
	arg := n.Leaves[0]
	if arg.Lexeme != nil && arg.Lexeme.Kind == lk.StringLit {
		return nil
	}
	return resolveExpr(ctx, scope, arg)
}

func resolvePara(ctx *context, scope *mod.Scope, n *mod.Node) *Error {
	initAtrib := n.Leaves[0]
	err := resolveAtrib(ctx, scope, initAtrib)
	if err != nil {
		return err
	}

	expr := n.Leaves[1]
	err = resolveExpr(ctx, scope, expr)
	if err != nil {
		return err
	}

	repeatAtrib := n.Leaves[2]
	err = resolveAtrib(ctx, scope, repeatAtrib)
	if err != nil {
		return err
	}

	bl := n.Leaves[3]
	return resolveBlock(ctx, scope, bl)
}

func resolveEnquanto(ctx *context, scope *mod.Scope, n *mod.Node) *Error {
	expr := n.Leaves[0]
	err := resolveExpr(ctx, scope, expr)
	if err != nil {
		return err
	}
	bl := n.Leaves[1]
	return resolveBlock(ctx, scope, bl)
}

func resolveSe(ctx *context, scope *mod.Scope, n *mod.Node) *Error {
	expr := n.Leaves[0]
	err := resolveExpr(ctx, scope, expr)
	if err != nil {
		return err
	}

	bl := n.Leaves[1]
	err = resolveBlock(ctx, scope, bl)
	if err != nil {
		return err
	}

	senao := n.Leaves[2]
	if senao != nil {
		err = resolveBlock(ctx, scope, senao)
		if err != nil {
			return err
		}
	}

	return nil
}

func resolveRetorne(ctx *context, scope *mod.Scope, n *mod.Node) *Error {
	expr := n.Leaves[0]
	return resolveExpr(ctx, scope, expr)
}

func resolveVarDecl(ctx *context, scope *mod.Scope, n *mod.Node) *Error {
	// vardecl := {type, id...}
	for _, id := range n.Leaves[1:] {
		name := id.Lexeme.Text
		sy := scope.Symbols[name]
		if sy != nil {
			return errorNameAlreadyDefined(ctx.M, id)
		}
		sy = &mod.Symbol{
			Kind:    sk.Local,
			Name:    name,
			N:       id,
			Builtin: false,
		}
		scope.Add(name, sy)
	}
	return nil
}

func resolveExpr(ctx *context, scope *mod.Scope, n *mod.Node) *Error {
	switch n.Kind {
	case nk.Terminal:
		switch n.Lexeme.Kind {
		case lk.Ou, lk.E, lk.Equals, lk.Different,
			lk.Greater, lk.GreaterOrEquals, lk.Less, lk.LessOrEquals,
			lk.Plus, lk.Star, lk.Division, lk.Remainder:
			err := resolveExpr(ctx, scope, n.Leaves[0])
			if err != nil {
				return err
			}
			return resolveExpr(ctx, scope, n.Leaves[1])
		case lk.Nao:
			return resolveExpr(ctx, scope, n.Leaves[0])
		case lk.Minus:
			if len(n.Leaves) == 1 {
				return resolveExpr(ctx, scope, n.Leaves[0])
			}
			err := resolveExpr(ctx, scope, n.Leaves[0])
			if err != nil {
				return err
			}
			return resolveExpr(ctx, scope, n.Leaves[1])
		case lk.IntLit, lk.RealLit, lk.CharLit:
			return nil //nada pra fazer aqui
		case lk.Ident:
			name := n.Lexeme.Text
			sy := scope.Find(name)
			if sy == nil {
				return errorSymbolNotDeclared(ctx.M, n)
			}
			return nil
		}
	case nk.Call:
		proc := n.Leaves[0]
		err := resolveExpr(ctx, scope, proc)
		if err != nil {
			return err
		}
		args := n.Leaves[1]
		for _, expr := range args.Leaves {
			err := resolveExpr(ctx, scope, expr)
			if err != nil {
				return err
			}
		}
		return nil
	}
	fmt.Println(n)
	panic("unreachable")
}

func resolveAtrib(ctx *context, scope *mod.Scope, n *mod.Node) *Error {
	id := n.Leaves[0]
	name := id.Lexeme.Text
	sy := scope.Find(name)
	if sy == nil {
		return errorSymbolNotDeclared(ctx.M, id)
	}
	expr := n.Leaves[1]
	return resolveExpr(ctx, scope, expr)
}

// errors -----------

func errorNameAlreadyDefined(M *mod.Module, newName *mod.Node) *Error {
	return mod.NewError(M, ek.NameAlreadyDefined, newName, "nome já pertence a outro simbolo")
}

func errorSymbolNotDeclared(M *mod.Module, n *mod.Node) *Error {
	return mod.NewError(M, ek.SymbolNotDeclared, n, "simbolo '"+n.Lexeme.Text+"' não foi declarado")
}

func errorEntryPointNotFound(M *mod.Module) *Error {
	return &Error{
		Code:     ek.NoEntryPoint,
		Severity: sv.Error,
		Message:  "programa sem entrada",
	}
}
