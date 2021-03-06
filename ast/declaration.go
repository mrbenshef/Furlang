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
	Arguments   []*ArgumentDeclaration
	Return      types.Type
	Body        *BlockStatement
}

func (e *FunctionDeclaration) First() lexer.Token { return e.Name.First() }
func (e *FunctionDeclaration) Last() lexer.Token  { return e.Body.Last() }
func (e *FunctionDeclaration) declareNode()       {}

type ArgumentDeclaration struct {
	Name *IdentExpression
	Type types.Type
}

func (e *ArgumentDeclaration) First() lexer.Token { return e.Name.First() }
func (e *ArgumentDeclaration) Last() lexer.Token  { return e.Name.Last() }
func (e *ArgumentDeclaration) declareNode()       {}

// VaribleDeclaration is a declare node in the form:
// ident := expression || type ident = expression
type VaribleDeclaration struct {
	Type  types.Type
	Name  *IdentExpression
	Value Expression
}

func (e *VaribleDeclaration) First() lexer.Token { return e.Name.First() }
func (e *VaribleDeclaration) Last() lexer.Token  { return e.Value.Last() }
func (e *VaribleDeclaration) declareNode()       {}
