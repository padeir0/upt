package flowkind

type FlowKind int

func (f FlowKind) String() string {
	switch f {
	case Jmp:
		return "jmp"
	case If:
		return "if"
	case Return:
		return "ret"
	}
	return "invalid FlowKind"
}

const (
	InvalidFlow FlowKind = iota

	Jmp
	If
	Return
)
