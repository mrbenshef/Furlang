package compiler

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"
)

// Compiler hold infomation about the file to be compiled
type Compiler struct {
	program string

	// Compiler optional flags
	OutputTokens bool
	OutputAst    bool
	NoCompile    bool
}

// New creates a new compiler for the file at filePath
func New(filePath string) (*Compiler, error) {
	if filePath == "" {
		return nil, fmt.Errorf("No input file")
	}

	// Read the file
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("Problem reading file '%s'", filePath)
	}
	program := string(data)

	return &Compiler{
		program: program,
	}, nil
}

// Compile compiles the file and writes to the outPath
func (c *Compiler) Compile(buildDirector string) error {
	// Start compiler timer
	start := time.Now()

	// Run lexer
	tokens := Lexer(c.program)

	// Optionaly write tokens to file
	if c.OutputTokens {
		f, err := os.Create(buildDirector + "/tokens.txt")
		if err != nil {
			return fmt.Errorf("Problem creating tokens file: %s\n", err.Error())
		}
		defer f.Close()

		for _, t := range tokens {
			f.WriteString(t.String() + "\n")
		}
	}

	// Run parser
	parser := NewParser(tokens)
	ast := parser.Parse()

	// Optionaly write ast to file (and print it)
	if c.OutputAst {
		f, err := os.Create(buildDirector + "/ast.txt")
		if err != nil {
			fmt.Errorf("Problem creating ast file: %s\n", err.Error())
		}
		defer f.Close()

		ast.Print()
		ast.Write(f)
	}

	// Compile ast to llvm
	if !c.NoCompile {
		llvm := Llvm(&ast)
		f, err := os.Create(buildDirector + "/ben.ll")
		if err != nil {
			return fmt.Errorf("Problem creating llvm ir file: %s\n", err.Error())
		}
		defer f.Close()

		f.WriteString(llvm)
	}

	// Output compiler timings
	fmt.Printf("[Compiled in: %fs]\n", time.Since(start).Seconds())

	return nil
}
