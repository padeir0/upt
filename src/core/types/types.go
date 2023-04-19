package types

import (
	"strings"
)

type Type struct {
	Basic BasicType
	Proc  *ProcType
}

func (t *Type) String() string {
	if t == nil {
		return "nil"
	}
	switch t.Basic {
	case Inteiro:
		return "inteiro"
	case Real:
		return "real"
	case Caractere:
		return "caractere"
	case Void:
		return "void"
	}
	if t.Proc != nil {
		return t.Proc.String()
	}
	return "invalid type"
}

func (this *Type) Equals(other *Type) bool {
	if IsBasic(this) && IsBasic(other) {
		return this.Basic == other.Basic
	}
	// one is basic and the other is not
	if IsBasic(this) || IsBasic(other) {
		return false
	}
	if this.Proc != nil && other.Proc != nil {
		return this.Proc.Equals(other.Proc)
	}
	panic("cannot compare " + this.String() + " with " + other.String())
}

type BasicType int

const (
	InvalidBasicType BasicType = iota

	Real
	Inteiro
	Caractere

	Void
)

type ProcType struct {
	Args []*Type
	Ret  *Type
}

func (this *ProcType) String() string {
	decls := []string{}
	for _, t := range this.Args {
		decls = append(decls, t.String())
	}
	return "proc(" + strings.Join(decls, ", ") + ")" + this.Ret.String()
}

func (this *ProcType) Equals(other *ProcType) bool {
	if len(this.Args) != len(other.Args) {
		return false
	}
	for i := range this.Args {
		if !this.Args[i].Equals(other.Args[i]) {
			return false
		}
	}
	if !this.Ret.Equals(other.Ret) {
		return false
	}
	return true
}

func IsBasic(tt *Type) bool {
	return tt.Basic != InvalidBasicType
}

func IsProc(tt *Type) bool {
	return tt.Proc != nil
}

func IsVoid(tt *Type) bool {
	return tt.Basic == Void
}

var T_Inteiro = &Type{Basic: Inteiro}
var T_Real = &Type{Basic: Real}
var T_Void = &Type{Basic: Void}
var T_Caractere = &Type{Basic: Caractere}

var T_Sqrt = &Type{
	Proc: &ProcType{
		Args: []*Type{T_Real},
		Ret:  T_Real,
	},
}
var T_Pow = &Type{
	Proc: &ProcType{
		Args: []*Type{T_Real, T_Real},
		Ret:  T_Real,
	},
}
