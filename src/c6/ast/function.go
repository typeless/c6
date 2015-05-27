package ast

type Function struct {
	Ident        *Token
	ArgumentList *ArgumentList
	Block        *Block
}

func (f Function) CanBeStatement() {}
func (f Function) String() string  { return "Function.String() is unimplemented." }

func NewFunctionWithToken(tok *Token) *Function {
	return &Function{Ident: tok}
}

// +gen symtable:"SymTable[Function]"
type FunctionSymTable map[string]*Function
