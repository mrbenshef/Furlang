package irgen

import (
	"fmt"

	"github.com/bongo227/Furlang/ast"
	"github.com/bongo227/Furlang/lexer"
	"github.com/bongo227/goory"
	goorytypes "github.com/bongo227/goory/types"
	gooryvalues "github.com/bongo227/goory/value"
	"github.com/k0kubun/pp"
)

type scope map[string]gooryvalues.Value

type Irgen struct {
	root            ast.Ast
	currentFunction ast.Function

	scopes       []scope
	currentScope int

	module *goory.Module
	block  *goory.Block
}

func NewIrgen(ast *ast.Ast) *Irgen {
	scopes := make([]scope, 1000)
	for i := range scopes {
		scopes[i] = make(scope)
	}

	return &Irgen{
		root:         *ast,
		scopes:       scopes,
		currentScope: 0,
		module:       goory.NewModule("furlang"),
	}
}

// Finds a scoped value
func (g *Irgen) find(v string) gooryvalues.Value {
	// Start at current scope and work backwords until the value is found
	search := g.currentScope
	for search >= 0 {
		if value, ok := g.scopes[search][v]; ok {
			return value
		}
		search--
	}

	panic(fmt.Sprintf("Varible not in scope: %s", v))
}

func (g *Irgen) Generate() string {
	for _, function := range g.root.Functions {
		g.function(function)
	}

	return g.module.LLVM()
}

func (g *Irgen) typ(node ast.Type) goorytypes.Type {
	return node.Llvm()
}

func (g *Irgen) function(node ast.Function) {
	g.currentFunction = node
	g.currentScope = 0
	// Add function to module
	function := g.module.NewFunction(node.Name.Value, g.typ(node.Type.Returns[0]))
	g.block = function.Entry()

	// Add function to scope
	g.scopes[g.currentScope][node.Name.Value] = function

	// Add arguments to scope
	g.currentScope++
	for _, arg := range node.Type.Parameters {
		argType := g.typ(arg.Type)
		argValue := function.AddArgument(argType, arg.Ident.Value)
		argAlloc := g.block.Alloca(argType)
		g.block.Store(argAlloc, argValue)
		g.scopes[g.currentScope][arg.Ident.Value] = argAlloc
	}

	// Add expressions to function body
	for _, exp := range node.Body.Expressions {
		g.expression(exp)
	}
}

func (g *Irgen) expression(node ast.Expression) gooryvalues.Value {
	switch node := node.(type) {
	case ast.Return:
		g.ret(node)
		return nil
	case ast.Assignment:
		g.assignment(node)
		return nil
	case ast.Call:
		return g.call(node)
	case *ast.If:
		next := g.ifNode(node)
		g.block = next
		return nil
	case ast.Binary:
		return g.binary(node)
	case ast.Integer:
		return g.integer(node)
	case ast.Float:
		return g.float(node)
	case ast.Ident:
		if node.Value == "true" || node.Value == "false" {
			return g.boolNode(node)
		}

		return g.block.Load(g.find(node.Value).(gooryvalues.Pointer))
	}

	panic(fmt.Sprintf("Node not handled: %s", pp.Sprint(node)))
}

// block compiles a block and returns the start block and the end block (if it branches elsewhere)
func (g *Irgen) blockNode(node ast.Block) (*goory.Block, *goory.Block) {
	oldScope := g.currentScope
	oldBlock := g.block

	// Set new scope/block
	newBlock := g.block.Function().AddBlock()
	g.currentScope++
	g.block = newBlock

	// Compile expressions in block
	for _, e := range node.Expressions {
		g.expression(e)
	}

	// Restore scope/block
	g.currentScope = oldScope
	g.block = oldBlock

	return newBlock, g.block
}

func (g *Irgen) assignment(node ast.Assignment) {
	var typ goorytypes.Type
	value := g.expression(node.Expression)

	if node.Type == nil {
		// typ = goory.IntType(32)
		typ = value.Type()
	} else {
		typ = g.typ(node.Type)
		value = g.block.Cast(value, typ)
	}

	alloc := g.block.Alloca(typ)
	// Store value in current scope
	g.scopes[g.currentScope][node.Name.Value] = alloc
	g.block.Store(alloc, value)
}

func (g *Irgen) call(node ast.Call) gooryvalues.Value {
	// Find function in scope
	function := g.find(node.Function.Value)

	// Get argument values
	var args []gooryvalues.Value
	argTypes := function.Type().(goorytypes.Function).Arguments()
	for i, a := range node.Arguments {
		value := g.expression(a)
		value = g.block.Cast(value, argTypes[i])
		args = append(args, value)
	}

	// Call function with values
	call := g.block.Call(function, args...)
	return call
}

func (g *Irgen) ifNode(node *ast.If) (next *goory.Block) {
	trueStart, trueEnd := g.blockNode(node.Block)
	condition := g.expression(node.Condition)

	var falseStart, falseEnd *goory.Block
	nextBlock := g.block.Function().AddBlock()
	if node.Else == nil {
		// If node has no else so condition branch should branch to next block
		falseStart = nextBlock
		falseEnd = nextBlock
	} else {
		falseStart, falseEnd = g.blockNode(node.Else.Block)

		// If else node has a condition insert a new block with a conditional check
		if node.Else.Condition != nil {
			falseConditionCheck := g.block.Function().AddBlock()
			falseConditionCheck.CondBr(
				g.expression(node.Else.Condition),
				falseStart,
				nextBlock)
		}
	}

	g.block.CondBr(condition, trueStart, falseStart)

	// Check if else if chain continues
	if (node.Else != nil) && (node.Else.Else != nil) {
		g.ifNode(node.Else.Else)
	}

	// If blocks dont terminate brach to next block
	if !trueEnd.Terminated() {
		trueEnd.Br(nextBlock)
	}
	if !falseEnd.Terminated() {
		falseEnd.Br(nextBlock)
	}

	return nextBlock
}

func (g *Irgen) binary(node ast.Binary) gooryvalues.Value {
	lhs := g.expression(node.Lhs)
	rhs := g.expression(node.Rhs)

	switch node.Op {
	case lexer.ADD:
		return g.block.Add(lhs, rhs)
	case lexer.QUO:
		return g.block.Fdiv(lhs, rhs)
	}

	panic("Unhandled binary operator")
}

func (g *Irgen) ret(node ast.Return) {
	value := g.expression(node.Value)
	value = g.block.Cast(value, g.currentFunction.Type.Returns[0].Llvm())
	g.block.Ret(value)
}

func (g *Irgen) integer(node ast.Integer) gooryvalues.Value {
	return goory.Constant(goory.IntType(64), node.Value)
}

func (g *Irgen) float(node ast.Float) gooryvalues.Value {
	return goory.Constant(goory.DoubleType(), node.Value)
}

func (g *Irgen) boolNode(node ast.Ident) gooryvalues.Value {
	switch node.Value {
	case "true":
		return goory.Constant(goory.BoolType(), true)
	case "false":
		return goory.Constant(goory.BoolType(), false)
	}

	panic("Ident is not a bool")
}
