package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/ast/astutil"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

// This program should be run after the instrumentation step
// It prints the filename:linenumber of all channel operations (make, send, receive, select) into an output file
// For select, it prints the position of "select" keyword, but doesn't print the positions of cases in the select

var currentFSet *token.FileSet
var vecPosition []string

func main() {

	pFile := flag.String("file", "", "Full path of the target file to be parsed")
	pOutputFile := flag.String("output", "", "Full path of the output file")

	flag.Parse()

	filename := *pFile
	outputFile := *pOutputFile

	oldSource, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Error when read file:", err.Error())
		return
	}

	tokenFSet := token.NewFileSet()
	oldAST, err := parser.ParseFile(tokenFSet, filename, oldSource, parser.ParseComments)
	if err != nil {
		fmt.Printf("error parsing %s: %s\n", filename, err.Error())
		return
	}

	currentFSet = tokenFSet

	newAST := astutil.Apply(oldAST, pre, nil)


	buf := &bytes.Buffer{}
	err = format.Node(buf, tokenFSet, newAST)
	if err != nil {
		fmt.Printf("error formatting new code: %s\n", err.Error())
		return
	}

	newSource := buf.Bytes()
	fi, err := os.Stat(filename)
	if err != nil {
		fmt.Printf("Error in os.Stat file: %s\tError:%s", filename, err.Error())
		return
	}
	ioutil.WriteFile(filename, newSource, fi.Mode())

	outputF, err := os.OpenFile(outputFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		panic(err)
	}

	defer outputF.Close()

	strOutput := ""
	for _, str := range vecPosition {
		if _, exist := mapDoNotPrint[str]; exist {
			continue
		}
		strOutput += str + "\n"
	}

	if _, err = outputF.WriteString(strOutput); err != nil {
		panic(err)
	}
}

var mapDoNotPrint map[string]struct{} = make(map[string]struct{})

func pre(c *astutil.Cursor) bool {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recover in pre(): c.Name():", c.Name())
		}
	}()
	switch concrete := c.Node().(type) {

	case *ast.SelectStmt: // for select, just print the location of select, not cases
		positionOriSelect := currentFSet.Position(concrete.Select)
		positionOriSelect.Filename, _ = filepath.Abs(positionOriSelect.Filename)
		str := positionOriSelect.Filename + ":" + strconv.Itoa(positionOriSelect.Line)
		vecPosition = append(vecPosition, str)
		for _, stmtCommClause := range concrete.Body.List {
			commClause, _ := stmtCommClause.(*ast.CommClause)
			positionOriSelect := currentFSet.Position(commClause.Case)
			positionOriSelect.Filename, _ = filepath.Abs(positionOriSelect.Filename)
			str := positionOriSelect.Filename + ":" + strconv.Itoa(positionOriSelect.Line)
			mapDoNotPrint[str] = struct{}{}
		}

	case *ast.SendStmt: // This is a send operation
		positionOriSend := currentFSet.Position(concrete.Arrow)
		positionOriSend.Filename, _ = filepath.Abs(positionOriSend.Filename)
		str := positionOriSend.Filename + ":" + strconv.Itoa(positionOriSend.Line)
		vecPosition = append(vecPosition, str)

	case *ast.AssignStmt:
		if len(concrete.Rhs) == 1 {
			if callExpr, ok := concrete.Rhs[0].(*ast.CallExpr); ok {
				if funcIdent, ok := callExpr.Fun.(*ast.Ident); ok {
					if funcIdent.Name == "make" {
						if len(callExpr.Args) == 1 { // This is a make operation
							if len(callExpr.Args) == 1 { // This is a make operation
								if _, ok := callExpr.Args[0].(*ast.ChanType); ok {
									positionOp := currentFSet.Position(concrete.TokPos)
									positionOp.Filename, _ = filepath.Abs(positionOp.Filename)
									str := positionOp.Filename + ":" + strconv.Itoa(positionOp.Line)
									vecPosition = append(vecPosition, str)
								}
							}
						}
					}
				}
			}
		}

	case *ast.ExprStmt:
		if unaryExpr, ok := concrete.X.(*ast.UnaryExpr); ok {
			if unaryExpr.Op == token.ARROW { // This is a receive operation
				positionOp := currentFSet.Position(unaryExpr.OpPos)
				positionOp.Filename, _ = filepath.Abs(positionOp.Filename)
				str := positionOp.Filename + ":" + strconv.Itoa(positionOp.Line)
				vecPosition = append(vecPosition, str)
			}
		} else if callExpr, ok := concrete.X.(*ast.CallExpr); ok {
			if funcIdent, ok := callExpr.Fun.(*ast.Ident); ok {
				if funcIdent.Name == "close" { // This is a close operation
					positionOp := currentFSet.Position(callExpr.Lparen)
					positionOp.Filename, _ = filepath.Abs(positionOp.Filename)
					str := positionOp.Filename + ":" + strconv.Itoa(positionOp.Line)
					vecPosition = append(vecPosition, str)
				}
			}
		}

	default:
	}

	return true
}