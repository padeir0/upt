package linearization

import (
	ir "github.com/padeir0/pir"
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
	LocalMap  map[scopedSymbol]ir.Operand
}

func newCtx(M *mod.Module) *context {
	p := ir.NewProgram()
	p.Name = M.Name
	return &context{
		M:         M,
		P:         p,
		Sy:        nil,
		GlobalMap: map[string]ir.SymbolID{},
		LocalMap:  map[scopedSymbol]ir.Operand{},
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
