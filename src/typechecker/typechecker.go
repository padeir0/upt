package typechecker

import (
	. "upt/core"
	colors "upt/core/asciicolors"
	mod "upt/core/module"
	T "upt/core/types"

	ek "upt/core/errorkind"
	lk "upt/core/lexeme/lexkind"
	nk "upt/core/module/nodekind"
	sk "upt/core/module/symbolkind"
)

func Check(M *mod.Module) *Error {
	err := inferGlobals(M)
	if err != nil {
		return err
	}
	sy := M.Global.Find("entrada")
	if !sy.Type.Equals(T.T_Entrada) {
		return errorWrongEntryType(M, sy.N)
	}

	err = checkInnerScopes(M)
	if err != nil {
		return err
	}

	return nil
}

func inferGlobals(M *mod.Module) *Error {
	for _, sy := range M.Global.Symbols {
		if sy.Kind != sk.Procedure {
			panic("invalid node kind for symbol")
		}
		err := inferProc(M, sy)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkInnerScopes(M *mod.Module) *Error {
	for _, sy := range M.Global.Symbols {
		if sy.Kind != sk.Procedure {
			panic("invalid symbol kind")
		}
		err := checkProc(M, sy)
		if err != nil {
			return err
		}
	}
	return nil
}

func inferProc(M *mod.Module, sy *mod.Symbol) *Error {
	// proc := {id, args, retNode, bl}
	scope := sy.N.Scope
	args := sy.N.Leaves[1]
	argTypes := []*T.Type{}
	if args != nil {
		for i, arg := range args.Leaves {
			// arg := {tipo, id}
			tnode := arg.Leaves[0]
			t := t2T(tnode)

			id := arg.Leaves[1]
			name := id.Lexeme.Text

			scope.Symbols[name].Type = t
			sy.Args[i].T = t
			argTypes = append(argTypes, t)
		}
	}
	retNode := sy.N.Leaves[2]
	var retType *T.Type
	if retNode != nil {
		retType = t2T(retNode)
	} else {
		retType = T.T_Inteiro
	}

	sy.Type = T.NewProcType(argTypes, retType)
	return nil
}

func t2T(n *mod.Node) *T.Type {
	switch n.Lexeme.Kind {
	case lk.Caractere:
		return T.T_Caractere
	case lk.Real:
		return T.T_Real
	case lk.Inteiro:
		return T.T_Inteiro
	}
	panic("invalid type")
}

func checkProc(M *mod.Module, sy *mod.Symbol) *Error {
	scope := sy.N.Scope
	bl := sy.N.Leaves[3]
	err := checkBlock(M, sy, scope, bl)
	if err != nil {
		return err
	}
	return nil
}

func checkBlock(M *mod.Module, sy *mod.Symbol, scope *mod.Scope, bl *mod.Node) *Error {
	scope = bl.Scope
	for _, cmd := range bl.Leaves {
		err := checkCmd(M, sy, scope, cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

func checkCmd(M *mod.Module, sy *mod.Symbol, scope *mod.Scope, n *mod.Node) *Error {
	switch n.Kind {
	case nk.Terminal:
		switch n.Lexeme.Kind {
		case lk.Leia:
			return checkLeia(M, scope, n)
		case lk.Imprima:
			return checkImprima(M, scope, n)
		case lk.Se:
			return checkSe(M, sy, scope, n)
		case lk.Enquanto:
			return checkEnquanto(M, sy, scope, n)
		case lk.Para:
			return checkPara(M, sy, scope, n)
		case lk.Retorne:
			return checkRetorne(M, sy, scope, n)
		case lk.Assign:
			return checkAtrib(M, scope, n)
		}
	case nk.Block:
		return checkBlock(M, sy, scope, n)
	case nk.VarDecl:
		return checkVarDecl(M, scope, n)
	}
	return checkExpr(M, scope, n)
}

func checkLeia(M *mod.Module, scope *mod.Scope, n *mod.Node) *Error {
	id := n.Leaves[0]
	name := id.Lexeme.Text
	sy := scope.Find(name)
	id.T = sy.Type
	return nil
}

func checkImprima(M *mod.Module, scope *mod.Scope, n *mod.Node) *Error {
	arg := n.Leaves[0]
	if arg.Lexeme != nil && arg.Lexeme.Kind == lk.StringLit {
		arg.T = T.T_String
		return nil
	}
	err := checkExpr(M, scope, arg)
	if err != nil {
		return err
	}
	return nil
}

func checkPara(M *mod.Module, sy *mod.Symbol, scope *mod.Scope, n *mod.Node) *Error {
	initAtrib := n.Leaves[0]
	var err *Error
	if initAtrib != nil {
		err = checkAtrib(M, scope, initAtrib)
		if err != nil {
			return err
		}
	}

	expr := n.Leaves[1]
	err = checkExpr(M, scope, expr)
	if err != nil {
		return err
	}
	if !T.T_Inteiro.Equals(expr.T) {
		return errorInvalidTypeForCond(M, n, expr.T)
	}

	repeatAtrib := n.Leaves[2]
	err = checkAtrib(M, scope, repeatAtrib)
	if err != nil {
		return err
	}

	bl := n.Leaves[3]
	return checkBlock(M, sy, scope, bl)
}

func checkEnquanto(M *mod.Module, sy *mod.Symbol, scope *mod.Scope, n *mod.Node) *Error {
	expr := n.Leaves[0]
	err := checkExpr(M, scope, expr)
	if err != nil {
		return err
	}
	if !T.T_Inteiro.Equals(expr.T) {
		return errorInvalidTypeForCond(M, n, expr.T)
	}
	bl := n.Leaves[1]
	err = checkBlock(M, sy, scope, bl)
	if err != nil {
		return err
	}
	return nil
}

func checkSe(M *mod.Module, sy *mod.Symbol, scope *mod.Scope, n *mod.Node) *Error {
	expr := n.Leaves[0]
	err := checkExpr(M, scope, expr)
	if err != nil {
		return err
	}
	if !T.T_Inteiro.Equals(expr.T) {
		return errorInvalidTypeForCond(M, n, expr.T)
	}

	bl := n.Leaves[1]
	err = checkBlock(M, sy, scope, bl)
	if err != nil {
		return err
	}

	senao := n.Leaves[2]
	if senao != nil {
		err = checkBlock(M, sy, scope, senao)
		if err != nil {
			return err
		}
	}

	return nil
}

func checkRetorne(M *mod.Module, sy *mod.Symbol, scope *mod.Scope, n *mod.Node) *Error {
	expr := n.Leaves[0]
	err := checkExpr(M, scope, expr)
	if err != nil {
		return err
	}
	retType := sy.Type.Proc.Ret
	if !T.AssignmentTable[retType.Basic][expr.T.Basic] {
		return errorReturnTypeNotAssignable(M, expr, expr.T, sy.Type.Proc.Ret)
	}
	return nil
}

func checkVarDecl(M *mod.Module, scope *mod.Scope, n *mod.Node) *Error {
	// vardecl := {type, id...}
	tNode := n.Leaves[0]
	t := t2T(tNode)
	tNode.T = t
	for _, id := range n.Leaves[1:] {
		name := id.Lexeme.Text
		sy := scope.Symbols[name]
		if sy == nil {
			panic("symbol was nil")
		}
		sy.Type = t
	}
	return nil
}

type typeRule func(a, b *T.Type) *T.Type

func outInt(a, b *T.Type) *T.Type {
	return T.T_Inteiro
}

func convTable(a, b *T.Type) *T.Type {
	return &T.Type{
		Basic: T.ConversionTable[a.Basic][b.Basic],
	}
}

func checkExpr(M *mod.Module, scope *mod.Scope, n *mod.Node) *Error {
	switch n.Kind {
	case nk.Terminal:
		switch n.Lexeme.Kind {
		case lk.Ou, lk.E:
			return checkIntBinExpr(M, scope, n)
		case lk.Equals, lk.Different,
			lk.Greater, lk.GreaterOrEquals, lk.Less, lk.LessOrEquals:
			return checkBinExpr(M, scope, n, outInt)
		case lk.Plus, lk.Star, lk.Division:
			return checkBinExpr(M, scope, n, convTable)
		case lk.Remainder:
			return checkIntBinExpr(M, scope, n)
		case lk.Nao:
			err := checkExpr(M, scope, n.Leaves[0])
			if err != nil {
				return err
			}
			aT := n.Leaves[0].T
			if !aT.Equals(T.T_Inteiro) {
				return errorExpectedType(M, n.Leaves[0], T.T_Inteiro)
			}
			n.T = aT
		case lk.Minus:
			if len(n.Leaves) == 1 {
				err := checkExpr(M, scope, n.Leaves[0])
				if err != nil {
					return err
				}
				n.T = n.Leaves[0].T
			}
			return checkBinExpr(M, scope, n, convTable)
		case lk.IntLit, lk.RealLit, lk.CharLit:
			n.T = litToType(n)
			return nil
		case lk.Ident:
			name := n.Lexeme.Text
			sy := scope.Find(name)
			if sy == nil {
				panic("symbol not found")
			}
			n.T = sy.Type
			return nil
		}
	case nk.Call:
		return checkCall(M, scope, n)
	}
	panic("unreachable")
}

func checkIntBinExpr(M *mod.Module, scope *mod.Scope, n *mod.Node) *Error {
	err := checkExpr(M, scope, n.Leaves[0])
	if err != nil {
		return err
	}
	err = checkExpr(M, scope, n.Leaves[1])
	if err != nil {
		return err
	}
	aT := n.Leaves[0].T
	if !aT.Equals(T.T_Inteiro) {
		return errorExpectedType(M, n.Leaves[0], T.T_Inteiro)
	}
	bT := n.Leaves[1].T
	if !bT.Equals(T.T_Inteiro) {
		return errorExpectedType(M, n.Leaves[1], T.T_Inteiro)
	}
	n.T = T.T_Inteiro
	return nil
}

func checkBinExpr(M *mod.Module, scope *mod.Scope, n *mod.Node, rule typeRule) *Error {
	err := checkExpr(M, scope, n.Leaves[0])
	if err != nil {
		return err
	}
	err = checkExpr(M, scope, n.Leaves[1])
	if err != nil {
		return err
	}
	aT := n.Leaves[0].T
	bT := n.Leaves[1].T
	if !aT.Equals(bT) {
		return errorInvalidOperationUnequalTypes(M, n)
	}
	n.T = rule(aT, bT)
	return nil
}

func litToType(n *mod.Node) *T.Type {
	switch n.Lexeme.Kind {
	case lk.IntLit:
		return T.T_Inteiro
	case lk.RealLit:
		return T.T_Real
	case lk.CharLit:
		return T.T_Caractere
	}
	return nil
}

func checkCall(M *mod.Module, scope *mod.Scope, n *mod.Node) *Error {
	proc := n.Leaves[0]
	err := checkExpr(M, scope, proc)
	if err != nil {
		return err
	}
	if !T.IsProc(proc.T) {
		return errorExpectedProc(M, proc)
	}
	tArgs := proc.T.Proc.Args
	args := n.Leaves[1]
	for i, expr := range args.Leaves {
		err := checkExpr(M, scope, expr)
		if err != nil {
			return err
		}
		if !T.AssignmentTable[tArgs[i].Basic][expr.T.Basic] {
			return errorArgNotAssignable(M, n, tArgs[i])
		}
	}
	n.T = proc.T.Proc.Ret
	return nil
}

func checkAtrib(M *mod.Module, scope *mod.Scope, n *mod.Node) *Error {
	id := n.Leaves[0]
	name := id.Lexeme.Text
	sy := scope.Find(name)
	if sy == nil {
		panic("symbol not found")
	}

	expr := n.Leaves[1]
	err := checkExpr(M, scope, expr)
	if err != nil {
		return err
	}

	if !T.AssignmentTable[sy.Type.Basic][expr.T.Basic] {
		return errorVarNotAssignable(M, n, expr.T, sy.Type)
	}
	return nil
}

func errorVarNotAssignable(M *mod.Module, n *mod.Node, t, u *T.Type) *Error {
	tStr := colors.MakeBlue(t.String())
	uStr := colors.MakeBlue(u.String())
	msg := "a expressão de tipo " + tStr + " não é atribuivel a variavel de tipo " + uStr
	return mod.NewError(M, ek.VarNotAssignable, n, msg)
}

func errorReturnTypeNotAssignable(M *mod.Module, n *mod.Node, t, u *T.Type) *Error {
	tStr := colors.MakeBlue(t.String())
	uStr := colors.MakeBlue(u.String())
	msg := "a expressão de tipo " + tStr + " não é atribuivel ao retorno do procedimento de tipo " + uStr
	return mod.NewError(M, ek.VarNotAssignable, n, msg)
}

func errorInvalidTypeForCond(M *mod.Module, n *mod.Node, t *T.Type) *Error {
	inteiro := colors.MakeBlue(T.T_Inteiro.String())
	tStr := colors.MakeBlue(t.String())
	msg := "expressão condicional deve ser " + inteiro + " não " + tStr
	return mod.NewError(M, ek.InvalidTypeForCond, n, msg)
}

func errorInvalidOperationUnequalTypes(M *mod.Module, n *mod.Node) *Error {
	left := n.Leaves[0]
	right := n.Leaves[1]
	tStr := colors.MakeBlue(left.T.String())
	uStr := colors.MakeBlue(right.T.String())
	msg := "operação entre tipos diferentes " + tStr + " e " + uStr
	return mod.NewError(M, ek.OpUnequalTypes, n, msg)
}

func errorExpectedType(M *mod.Module, n *mod.Node, expected *T.Type) *Error {
	expStr := colors.MakeBlue(expected.String())
	hasStr := colors.MakeBlue(n.T.String())
	msg := "operador espera um valor do tipo " + expStr + " não " + hasStr
	return mod.NewError(M, ek.ExpectedTypeOp, n, msg)
}

func errorArgNotAssignable(M *mod.Module, n *mod.Node, target *T.Type) *Error {
	expStr := colors.MakeBlue(target.String())
	hasStr := colors.MakeBlue(n.T.String())
	msg := "o tipo " + expStr + " não é atribuivel ao argumento " + hasStr
	return mod.NewError(M, ek.ArgNotAssignable, n, msg)
}

func errorExpectedProc(M *mod.Module, n *mod.Node) *Error {
	hasStr := colors.MakeBlue(n.T.String())
	proc := colors.MakeBlue("procedimento")
	msg := "esperado " + proc + " não " + hasStr
	return mod.NewError(M, ek.ExpectedTypeOp, n, msg)
}

func errorWrongEntryType(M *mod.Module, n *mod.Node) *Error {
	msg := "o procedimento de entrada deve receber zero argumentos e retornar um inteiro"
	return mod.NewError(M, ek.WrongEntryType, n, msg)
}
