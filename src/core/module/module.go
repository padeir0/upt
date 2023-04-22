package module

import (
	. "upt/core"
	lex "upt/core/lexeme"
	T "upt/core/types"

	ek "upt/core/errorkind"
	nk "upt/core/module/nodekind"
	sk "upt/core/module/symbolkind"
	sv "upt/core/severity"

	"fmt"
	"strings"
)

type Node struct {
	Lexeme *lex.Lexeme
	Leaves []*Node

	Kind  nk.NodeKind
	T     *T.Type
	Range *Range

	// only for scoped nodes
	Scope *Scope
}

func (this *Node) String() string {
	return ast(this, 0)
}

func (this *Node) AddLeaf(other *Node) {
	if this.Leaves == nil {
		this.Leaves = []*Node{other}
	} else {
		this.Leaves = append(this.Leaves, other)
	}
}

func (this *Node) PrependLeaf(other *Node) {
	if this.Leaves == nil {
		this.Leaves = []*Node{other}
	} else {
		this.Leaves = append([]*Node{other}, this.Leaves...)
	}
}

func ast(n *Node, i int) string {
	if n == nil {
		return "nil"
	}
	rng := "nil"
	if n.Range != nil {
		rng = n.Range.String()
	}
	scope := "_"
	if n.Scope != nil {
		scope = fmt.Sprintf("%v", n.Scope.ID)
	}
	output := fmt.Sprintf("{%v, %v, %v, %v, %v",
		n.Lexeme,
		n.Kind,
		n.T.String(),
		rng,
		scope,
	)
	output += "}"
	for _, kid := range n.Leaves {
		if kid == nil {
			output += indent(i) + "nil"
			continue
		}
		output += indent(i) + ast(kid, i+1)
	}
	return output
}

func indent(n int) string {
	output := "\n"
	for i := -1; i < n-1; i++ {
		output += "    "
	}
	output += "└─>"
	return output
}

var Universe *Scope = &Scope{
	Parent: nil,
	Symbols: map[string]*Symbol{
		"raiz": {Name: "raiz", Kind: sk.Procedure, Type: T.T_Sqrt, Builtin: true},
		"expo": {Name: "expo", Kind: sk.Procedure, Type: T.T_Pow, Builtin: true},
	},
}

func Place(M *Module, n *Node) *Location {
	return &Location{
		File:  M.FullPath,
		Range: n.Range,
	}
}

func NewError(M *Module, t ek.ErrorKind, n *Node, message string) *Error {
	loc := Place(M, n)
	return &Error{
		Code:     t,
		Severity: sv.Error,
		Location: loc,
		Message:  message,
	}
}

type Module struct {
	FullPath string
	Name     string
	Root     *Node

	Global *Scope
}

func (this *Module) String() string {
	globals := []string{}
	for name := range this.Global.Symbols {
		globals = append(globals, name)
	}
	return fmt.Sprintf("%v\n", this.FullPath) +
		"globals: " + strings.Join(globals, ", ") + "\n" +
		this.Root.String()
}

type Scope struct {
	ID      int
	Parent  *Scope
	Symbols map[string]*Symbol
}

func (this *Scope) Depth() int {
	s := this
	out := 0
	for s != nil {
		s = s.Parent
		out++
	}
	return out
}

func (this *Scope) String() string {
	output := []string{}
	for _, sy := range this.Symbols {
		output = append(output, sy.String())
	}
	return "{" + strings.Join(output, ", ") + "}"
}

func (this *Scope) Find(name string) *Symbol {
	if this == nil {
		panic("scope was nil")
	}
	v, ok := this.Symbols[name]
	if ok {
		return v
	}
	if this.Parent == nil {
		return nil
	}
	return this.Parent.Find(name)
}

func (this *Scope) Add(name string, sy *Symbol) {
	_, ok := this.Symbols[name]
	if ok {
		panic("name already added")
	}
	this.Symbols[name] = sy
}

type Symbol struct {
	Kind sk.SymbolKind
	Name string
	N    *Node

	Type *T.Type

	Builtin bool

	Args []Arg // mais facil de traduzir
}

func (this *Symbol) String() string {
	switch this.Kind {
	case sk.Procedure:
		return "proc " + this.Name
	case sk.Local:
		return "local " + this.Name
	case sk.Argument:
		return "arg " + this.Name
	default:
		return "invalid"
	}
}

type Arg struct {
	T    *T.Type
	N    *Node
	Name string
	Pos  int
}
