package ast

import (
	"github.com/bongo227/Furlang/lexer"
	"github.com/bongo227/Furlang/types"
)

type Declare interface {
	Node
	declareNode()
}

// FunctionDeclaration is a declare node in the form:
// ident :: type ident, ... -> type { statement; ... }
type FunctionDeclaration struct {
	Name        *IdentExpression
	DoubleColon lexer.Token
	Arguments   map[IdentExpression]types.Type
	Return      types.Type
	Body        *BlockStatement
}

func (e *FunctionDeclaration) First() lexer.Token { return e.Name.First() }
func (e *FunctionDeclaration) Last() lexer.Token  { return e.Body.Last() }
func (e *FunctionDeclaration) expressionNode()    {}

// VaribleDeclaration is a declare node in the form:
// ident := expression || type ident = expression
type VaribleDeclaration struct {
	Type  types.Type
	Name  *IdentExpression
	Value Expression
}

func (e *VaribleDeclaration) First() lexer.Token { return e.Name.First() }
func (e *VaribleDeclaration) Last() lexer.Token  { return e.Value.Last() }
func (e *VaribleDeclaration) expressionNode()    {}
