package ir

import (
	irc "upt/core/ir/class"
	FT "upt/core/ir/flowkind"
	IT "upt/core/ir/instrkind"
	T "upt/core/types"

	mod "upt/core/module"

	"fmt"
	"strconv"
)

type SymbolID int

type Program struct {
	Name            string
	Entry           SymbolID
	Symbols         []*Symbol
	InternalSymbols []*Symbol

	// if modules are ever introduced (doubtful), this will be wrong
	M *mod.Module
}

func (this *Program) AddProc(p *Procedure) int {
	index := len(this.Symbols)
	this.Symbols = append(this.Symbols, &Symbol{Proc: p})
	return index
}

func (this *Program) AddMem(m *Memory) int {
	index := len(this.Symbols)
	this.Symbols = append(this.Symbols, &Symbol{Mem: m})
	return index
}

func (this *Program) String() string {
	if this == nil {
		return "nil program"
	}
	output := "Program: " + this.Name + "\n\n"
	for _, sy := range this.Symbols {
		output += sy.String() + "\n"
	}
	return output
}

func NewProgram(m *mod.Module) *Program {
	return &Program{
		Name:            m.Name,
		Entry:           0,
		Symbols:         []*Symbol{},
		InternalSymbols: []*Symbol{},
		M:               m,
	}
}

type Symbol struct {
	Proc *Procedure
	Mem  *Memory
}

func (this *Symbol) String() string {
	if this.Proc != nil {
		return this.Proc.String()
	}
	return this.Mem.String()
}

type BlockID int

type Procedure struct {
	Label string
	Args  []*T.Type
	Ret   *T.Type

	Locals [][]*T.Type

	Start     BlockID
	AllBlocks []*BasicBlock
}

func (this *Procedure) ResetBlocks() {
	for _, b := range this.AllBlocks {
		b.Visited = false
	}
}

func (this *Procedure) FirstBlock() *BasicBlock {
	return this.AllBlocks[this.Start]
}

func (this *Procedure) GetBlock(id BlockID) *BasicBlock {
	return this.AllBlocks[id]
}

func (p *Procedure) StrArgs() string {
	return StrTypes(p.Args)
}

func (p *Procedure) StrLocals() string {
	output := ""
	for _, scope := range p.Locals {
		output += "[" + StrTypes(scope) + "]"
	}
	return output
}

func (this *Procedure) String() string {
	output := this.Label + "{\n"
	output += this.StrArgs() + "\n"
	output += this.Ret.String() + "\n"
	output += this.StrLocals() + "\n"
	output += "}:\n"
	for _, bb := range this.AllBlocks {
		output += bb.String() + "\n"
	}
	return output + "\n"
}

func StrTypes(tps []*T.Type) string {
	if len(tps) == 0 {
		return ""
	}
	if len(tps) == 1 {
		return tps[0].String()
	}
	output := tps[0].String()
	for _, t := range tps {
		output += ", " + t.String()
	}
	return output
}

// we use this to represent strings
type Memory struct {
	Data string
}

func (this *Memory) String() string {
	return "memory: \"" + this.Data + "\""
}

type BasicBlock struct {
	Label   string
	Code    []Instr
	Out     Flow
	Visited bool
}

func (this *BasicBlock) AddInstr(i Instr) {
	this.Code = append(this.Code, i)
}

func (this *BasicBlock) Jmp(id BlockID) {
	this.Out = Flow{
		T:    FT.Jmp,
		V:    nil,
		True: id,
	}
}

func (b *BasicBlock) Branch(cond *Operand, True BlockID, False BlockID) {
	b.Out = Flow{
		T:     FT.If,
		V:     cond,
		True:  True,
		False: False,
	}
}

func (b *BasicBlock) Return(ret *Operand) {
	b.Out = Flow{
		V: ret,
		T: FT.Return,
	}
}

func (b *BasicBlock) HasFlow() bool {
	return b.Out.T != FT.InvalidFlow
}

func (b *BasicBlock) IsTerminal() bool {
	return b.Out.T == FT.Return
}

func (b *BasicBlock) String() string {
	output := b.Label + ":\n"
	for _, v := range b.Code {
		output += "\t" + v.String() + "\n"
	}
	output += "\t" + b.Out.String()
	return output
}

type Instr struct {
	T           IT.InstrKind
	Type        *T.Type
	Operands    []*Operand
	Destination *Operand

	// for messages
	Source *mod.Node
}

func (this *Instr) String() string {
	if this == nil {
		return "nil"
	}
	if this.Destination != nil {
		if this.Type != nil {
			return fmt.Sprintf("%v:%v %v -> %v", this.T.String(), this.Type.String(), this.StrOps(), this.Destination.String())
		} else {
			return fmt.Sprintf("%v, %v -> %v", this.T.String(), this.StrOps(), this.Destination.String())
		}
	} else {
		if this.Type != nil {
			return fmt.Sprintf("%v:%v %v", this.T.String(), this.Type.String(), this.StrOps())
		} else {
			return fmt.Sprintf("%v, %v", this.T.String(), this.StrOps())
		}
	}
}

func (this *Instr) StrOps() string {
	if len(this.Operands) == 0 {
		return ""
	}
	output := this.Operands[0].String()
	for _, v := range this.Operands[1:] {
		output += ", " + v.String()
	}
	return output
}

func ProperlyTerminates(proc *Procedure) bool {
	start := proc.FirstBlock()
	proc.ResetBlocks()
	return properlyTerminates(proc, start)
}

func properlyTerminates(proc *Procedure, b *BasicBlock) bool {
	if b.Visited {
		// we just say that this looping branch doesn't matter
		return true
	}
	b.Visited = true
	switch b.Out.T {
	case FT.If:
		t := proc.GetBlock(b.Out.True)
		f := proc.GetBlock(b.Out.False)
		return properlyTerminates(proc, t) && properlyTerminates(proc, f)
	case FT.Jmp:
		t := proc.GetBlock(b.Out.True)
		return properlyTerminates(proc, t)
	case FT.Return:
		return true
	}
	return false
}

type Operand struct {
	IsValue bool
	Type    *T.Type
	Value   Value
}

func (this *Operand) String() string {
	if this.IsValue {
		return this.Value.String()
	} else if this.Type == nil {
		return ":" + this.Type.String()
	}
	panic("contradiction")
}

type Value struct {
	Class irc.Class
	Scope int
	V     int64

	// Local => Procedure.Locals[Value.Scope][Value.V]
	// Arg => Procedure.Args[Value.V]
	// Global => Program.Symbols[Value.V]
	// Internal_Global => Program.InternalSymbols[Value.V]
	// IntLit => Value.V is the value of the literal
}

func (this *Value) String() string {
	vstr := strconv.FormatInt(int64(this.V), 10)
	switch this.Class {
	case irc.Local:
		scpStr := strconv.FormatInt(int64(this.Scope), 10)
		output := this.Class.String() + "'(" + scpStr + ", " + vstr + ")"
		return output
	case irc.IntLit:
		return vstr
	}
	output := this.Class.String() + "'" + vstr
	return output
}

type Flow struct {
	T     FT.FlowKind
	V     *Operand
	True  BlockID
	False BlockID
}

func (this *Flow) String() string {
	switch this.T {
	case FT.Jmp:
		t := strconv.FormatInt(int64(this.True), 10)
		return "jmp .L" + t
	case FT.If:
		t := strconv.FormatInt(int64(this.True), 10)
		f := strconv.FormatInt(int64(this.False), 10)
		return "if " + this.V.String() + "? .L" + t + " : .L" + f
	case FT.Return:
		return "ret " + this.V.String()
	}
	return "invalid FlowType"
}
