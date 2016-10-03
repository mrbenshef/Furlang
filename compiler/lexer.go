package compiler

import (
	"fmt"
	"strconv"
)

type Token struct {
	tokenType int
	value     interface{}
	line      int
	column    int
	length    int
}

// Token type constants
const (
	tokenName = iota
	tokenNumber
	tokenFloat

	tokenType
	tokenInt32
	tokenFloat32
	tokenReturn
	tokenIf
	tokenElse
	tokenTrue
	tokenFalse

	tokenArrow
	tokenAssign
	tokenDoubleColon

	tokenComma
	tokenSemiColon
	tokenNewLine
	tokenOpenBody
	tokenOpenBracket
	tokenCloseBody
	tokenCloseBracket
	tokenPlus
	tokenMinus
	tokenMultiply
	tokenFloatDivide
	tokenIntDivide
	tokenLessThan
	tokenMoreThan
	tokenColon
	tokenEqual
)

// Lexer maps
var (
	symbolMap = map[string]int{
		"\n": tokenNewLine,
		",":  tokenComma,
		"{":  tokenOpenBody,
		"}":  tokenCloseBody,
		"(":  tokenOpenBracket,
		")":  tokenCloseBracket,
		"+":  tokenPlus,
		"-":  tokenMinus,
		"*":  tokenMultiply,
		"/":  tokenFloatDivide,
		"<":  tokenLessThan,
		">":  tokenMoreThan,
		":":  tokenColon,
		";":  tokenSemiColon,
		"=":  tokenEqual,
	}

	typeMap = map[string]int{
		"i32": typeInt32,
		"f32": typeFloat32,
	}

	nameMap = map[string]int{
		"return": tokenReturn,
		"if":     tokenIf,
		"else":   tokenElse,
		"true":   tokenTrue,
		"false":  tokenFalse,
	}

	multiSymbolMap = map[int][]int{
		tokenArrow:       []int{tokenMinus, tokenMoreThan},
		tokenAssign:      []int{tokenColon, tokenEqual},
		tokenDoubleColon: []int{tokenColon, tokenColon},
		tokenIntDivide:   []int{tokenFloatDivide, tokenFloatDivide},
	}
)

// Converts token to printable string
func (t Token) String() string {
	tokenString := ""
	switch t.tokenType {
	case tokenArrow:
		tokenString = "tokenArrow"
	case tokenAssign:
		tokenString = "tokenAssign"
	case tokenCloseBody:
		tokenString = "tokenCloseBody"
	case tokenComma:
		tokenString = "tokenComma"
	case tokenDoubleColon:
		tokenString = "tokenDoubleColon"
	case tokenInt32:
		tokenString = "tokenInt32"
	case tokenFloat32:
		tokenString = "tokenFloat32"
	case tokenName:
		tokenString = "tokenName"
	case tokenNewLine:
		tokenString = "tokenNewLine"
	case tokenNumber:
		tokenString = "tokenNumber"
	case tokenFloat:
		tokenString = "tokenFloat"
	case tokenOpenBody:
		tokenString = "tokenOpenBody"
	case tokenPlus:
		tokenString = "tokenPlus"
	case tokenMinus:
		tokenString = "tokenMinus"
	case tokenMultiply:
		tokenString = "tokenMultiply"
	case tokenFloatDivide:
		tokenString = "tokenFloatDivide"
	case tokenIntDivide:
		tokenString = "tokenIntDivide"
	case tokenOpenBracket:
		tokenString = "tokenOpenBracket"
	case tokenCloseBracket:
		tokenString = "tokenCloseBracket"
	case tokenReturn:
		tokenString = "tokenReturn"
	case tokenIf:
		tokenString = "tokenIf"
	case tokenElse:
		tokenString = "tokenElse"
	case tokenTrue:
		tokenString = "tokenTrue"
	case tokenFalse:
		tokenString = "tokenFalse"
	case tokenType:
		tokenString = "tokenType"
	default:
		tokenString = "Undefined token"
	}

	return fmt.Sprintf("%s, line: %d, column: %d", tokenString, t.line, t.column)
}

// Parsers what ever is in the the buffer
func parseBuffer(buffer *string, tokens *[]Token, line int, column int) {

	if *buffer != "" {
		bufferLength := len(*buffer)

		if i, err := strconv.Atoi(*buffer); err == nil {
			// Buffer contains a number
			*tokens = append(*tokens, Token{
				tokenType: tokenNumber,
				value:     i,
				line:      line,
				column:    column - bufferLength,
				length:    bufferLength,
			})
		} else if i, err := strconv.ParseFloat(*buffer, 32); err == nil {
			// Buffer contains a float
			*tokens = append(*tokens, Token{
				tokenType: tokenFloat,
				value:     float32(i),
				line:      line,
				column:    column - bufferLength,
				length:    bufferLength,
			})
		} else if val, found := typeMap[*buffer]; found {
			// Buffer contains a type identifyer
			*tokens = append(*tokens, Token{
				tokenType: tokenType,
				value:     val,
				line:      line,
				column:    column - bufferLength,
				length:    bufferLength,
			})
		} else if val, found := nameMap[*buffer]; found {
			// Buffer contains a control name
			*tokens = append(*tokens, Token{
				tokenType: val,
				value:     *buffer,
				line:      line,
				column:    column - bufferLength,
				length:    bufferLength,
			})
		} else {
			// Buffer contains a name
			*tokens = append(*tokens, Token{
				tokenType: tokenName,
				value:     *buffer,
				line:      line,
				column:    column - bufferLength,
				length:    bufferLength,
			})
		}

		*buffer = ""
	}

}

// Lexer returns a sequential list of tokens from the input string
func Lexer(in string) (tokens []Token) {
	buffer := ""

	// Parse all single character tokens, names and numbers
	lineIndex := 1
	columnIndex := 0
characterLoop:
	for _, char := range in {
		columnIndex++

		// Handle whitespace
		if string(char) == " " {
			parseBuffer(&buffer, &tokens, lineIndex, columnIndex)
			continue characterLoop
		}

		// Handle symbol character
		for symbol, symbolToken := range symbolMap {
			if string(char) == symbol {
				parseBuffer(&buffer, &tokens, lineIndex, columnIndex)
				tokens = append(tokens, Token{
					tokenType: symbolToken,
					value:     string(char),
					line:      lineIndex,
					column:    columnIndex,
					length:    1,
				})
				if symbolToken == tokenNewLine {
					lineIndex++
					columnIndex = 0
				}
				continue characterLoop
			}
		}

		// Any other character (number/letter)
		buffer += string(char)
	}

	// Group single character tokens
	for i := 0; i < len(tokens); i++ {
		for symbolsToken, symbols := range multiSymbolMap {
			// Check if tokens can be grouped
			equal := true
			for offset, val := range symbols {
				if len(tokens) > i+offset && tokens[i+offset].tokenType != val {
					equal = false
				}
			}

			// Collapse tokens in group into a single token
			if equal {
				lower := append(tokens[:i], Token{
					tokenType: symbolsToken,
					value:     nil,
					line:      tokens[i].line,
					column:    tokens[i].column,
					length:    2, //TODO: make this work with varible length multisymbols
				})
				tokens = append(lower, tokens[i+len(symbols):]...)
			}
		}
	}

	return tokens
}
