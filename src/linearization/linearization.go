package linearization

import (
	ir "upt/core/ir"
	mod "upt/core/module"
)

type scopedSymbol struct {
	ScopeID int
	Name    string
}

type context struct {
	M  *mod.Module
	P  *ir.Program
	Sy *mod.Symbol

	GlobalMap map[string]ir.SymbolID
	LocalMap  map[scopedSymbol]ir.Value
}

func newCtx(M *mod.Module) *context {
	return &context{
		M:         M,
		P:         ir.NewProgram(M),
		Sy:        nil,
		GlobalMap: map[string]ir.SymbolID{},
		LocalMap:  map[scopedSymbol]ir.Value{},
	}
}

func Linearize(m *mod.Module) *ir.Program {
	ctx := newCtx(m)
	for _, sy := range ctx.M.Global.Symbols {
		lnFunc(ctx, sy)
	}
	return ctx.P
}

func lnFunc(ctx *context, sy *mod.Symbol) {
	panic("unimplemented")
}
