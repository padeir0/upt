package symbolkind

type SymbolKind int

const (
	InvalidSymbolKind SymbolKind = iota
	Procedure
	// Local scope
	Argument
	Local
)
