package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

var boolNeedInstrument bool = false // remember: if we instrument, we always need to import gooracle
var additionalNode ast.Stmt
var currentFSet *token.FileSet
var Uint16OpID uint16

var recordOutputFile string
var records []string

var vecTest []string
var strPkg string

var sliceStrNoInstr = []string{
	"src/runtime",
	"src/gooracle",
	"src/sync",
	"src/reflect",
	"src/syscall",
	"src/bufio",
	"src/fmt",
	"src/os",
	"src/strconv",
	"src/strings",
	"src/time",
	"src/bytes",
	"src/hash",
}

func main() {
	//pProjectPath := flag.String("path","","Full path of the target project")
	//pRelativePath := flag.String("include","","Relative path (what's after /src/) of the target project")
	//pShowCompileError := flag.Bool("compile-error", false, "If fail to compile a package, show the errors of compilation")
	//pExcludePath := flag.String("exclude", "vendor", "Name of directories that you want to ignore, divided by \":\"")
	//pRobustMod := flag.Bool("r", false, "If the main package can't pass compiler, check subdirectories one by one")
	pFile := flag.String("file", "", "Full path of the target file to be parsed")
	pVecTest := flag.String("tests", "", "Unit tests that we have run in fuzzer, divided by \";\"")
	flag.Parse()

	filename := *pFile
	strVecTest := *pVecTest

	if strings.Contains(filename, "_test.go") == false {
		return
	}

	vecTest = strings.Split(strVecTest, ";")


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
	strPkg = oldAST.Name.Name

	newAST := astutil.Apply(oldAST, pre, nil)

	if !boolNeedInstrument {
		return
	}


	buf := &bytes.Buffer{}
	err = format.Node(buf, tokenFSet, newAST)
	if err != nil {
		fmt.Printf("error formatting new code: %s in file:%s\n", err.Error(), filename)
		return
	}

	newSource := buf.Bytes()
	fi, err := os.Stat(filename)
	if err != nil {
		fmt.Printf("Error in os.Stat file: %s\tError:%s", filename, err.Error())
		os.Exit(1)
	}
	ioutil.WriteFile(filename, newSource, fi.Mode())

	if recordOutputFile != "" {
		fmt.Printf("Dump operations to %s", recordOutputFile)
		dir := path.Dir(recordOutputFile)
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0755)
			if err != nil {
				fmt.Printf("failed to create dir at %s: %v", dir, err)
				os.Exit(1)
			}
		}
		outputF, err := os.OpenFile(recordOutputFile, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0666)
		if err != nil {
			fmt.Printf("failed to open file at %s: %v", recordOutputFile, err)
			os.Exit(1)
		}

		defer outputF.Close()

		var buffer bytes.Buffer
		for _, str := range records {
			buffer.WriteString(str)
			buffer.WriteByte('\n')
		}

		if _, err = outputF.Write(buffer.Bytes()); err != nil {
			panic(err)
		}
	}

}

func pre(c *astutil.Cursor) bool {
	defer func() {
		if r := recover(); r != nil { // This is allowed. If we insert node into nodes not in slice, we will meet a panic
			// For example, we may identified a receive in select and wanted to insert a function call before it, then this function will panic

			//fmt.Printf("Recover in pre(): c.Name(): %s\n", c.Name())
			//position := currentFSet.Position(c.Node().Pos())
			//position.Filename, _ = filepath.Abs(position.Filename)
			//fmt.Printf("\tLocation: %s\n", position.Filename + ":" + strconv.Itoa(position.Line))
		}
	}()

	if additionalNode != nil && c.Node() == additionalNode {
		newReturn := &ast.ReturnStmt{
			Return:  0,
			Results: nil,
		}
		c.InsertAfter(newReturn)
		additionalNode = nil
		boolNeedInstrument = true
	}

	switch concrete := c.Node().(type) {


	case *ast.FuncDecl:
		strName := concrete.Name.Name
		if strings.HasPrefix(strName, "Test") {
			boolWeRun := false
			for _, strTest := range vecTest {
				vecPkgAndTest := strings.Split(strTest, ":")
				if vecPkgAndTest[0] == strPkg && vecPkgAndTest[1] == strName {
					boolWeRun = true
					break
				}
			}
			if boolWeRun == false {
				if concrete.Body != nil && len(concrete.Body.List) != 0 {
					firstStmt := concrete.Body.List[0]
					additionalNode = firstStmt
					boolNeedInstrument = true // We need to disable this unit test
				}
			}
		}

	default:
	}

	return true
}


type SelectStruct struct {
	StmtSelect    *ast.SelectStmt   // StmtSelect.Body.List is a vec of CommClause
	VecCommClause []*ast.CommClause // a CommClause is a case and its content in select
	VecOp         []ast.Stmt        // The operations of cases. Nil is default
	VecBody       [][]ast.Stmt      // The content of cases
}

type SwitchStruct struct {
	StmtSwitch    *ast.SwitchStmt // StmtSwitch.Body.List is a vector of CaseClause
	Tag           ast.Expr
	VecCaseClause []*ast.CaseClause // a CaseClause is a case and its content in switch.
	VecVecExpr    [][]ast.Expr      // The expressions of each case.
	VecBody       [][]ast.Stmt      // The content of cases
}

// Deprecated:
func copyOp(stmtOp ast.Stmt) ast.Stmt {
	var result ast.Stmt
	// the stmtOp is either *ast.SendStmt or *ast.ExprStmt
	switch concrete := stmtOp.(type) {
	// TODO: could be "x := <-ch"
	case *ast.SendStmt:
		oriChanIdent, _ := concrete.Chan.(*ast.Ident)
		newSend := &ast.SendStmt{
			Chan: &ast.Ident{
				NamePos: 0,
				Name:    oriChanIdent.Name,
				Obj:     oriChanIdent.Obj,
			},
			Arrow: 0,
			Value: concrete.Value,
		}
		result = newSend
	case *ast.ExprStmt:
		oriUnaryExpr, _ := concrete.X.(*ast.UnaryExpr)
		newRecv := &ast.ExprStmt{X: &ast.UnaryExpr{
			OpPos: 0,
			Op:    token.ARROW,
			X:     oriUnaryExpr.X,
		}}
		result = newRecv
	}

	return result
}

func copyStmtBody(stmtBody []ast.Stmt) []ast.Stmt {
	result := []ast.Stmt{}
	for _, stmt := range stmtBody {
		result = append(result, stmt)
	}
	return result
}

func copySelect(oriSelect *ast.SelectStmt) *ast.SelectStmt {
	result := &ast.SelectStmt{
		Select: 0,
		Body:   oriSelect.Body,
	}
	return result
}

// imports reports whether f has an import with the specified name and path.
func imports(f *ast.File, name, path string) bool {
	for _, s := range f.Imports {
		importedName := importName(s)
		importedPath := importPath(s)
		if importedName == name && importedPath == path {
			return true
		}
	}
	return false
}

// importName returns the name of s,
// or "" if the import is not named.
func importName(s *ast.ImportSpec) string {
	if s.Name == nil {
		return ""
	}
	return s.Name.Name
}

// importPath returns the unquoted import path of s,
// or "" if the path is not properly quoted.
func importPath(s *ast.ImportSpec) string {
	t, err := strconv.Unquote(s.Path.Value)
	if err != nil {
		return ""
	}
	return t
}

func NewArgCall(strPkg, strCallee string, vecExprArg []ast.Expr) *ast.CallExpr {
	newIdentPkg := &ast.Ident{
		NamePos: token.NoPos,
		Name:    strPkg,
		Obj:     nil,
	}
	newIdentCallee := &ast.Ident{
		NamePos: token.NoPos,
		Name:    strCallee,
		Obj:     nil,
	}
	newCallSelector := &ast.SelectorExpr{
		X:   newIdentPkg,
		Sel: newIdentCallee,
	}
	newCall := &ast.CallExpr{
		Fun:      newCallSelector,
		Lparen:   token.NoPos,
		Args:     vecExprArg,
		Ellipsis: token.NoPos,
		Rparen:   token.NoPos,
	}

	return newCall
}

func NewArgCallExpr(strPkg, strCallee string, vecExprArg []ast.Expr) *ast.ExprStmt {
	newCall := NewArgCall(strPkg, strCallee, vecExprArg)
	newExpr := &ast.ExprStmt{X: newCall}
	return newExpr
}

func handleCallExpr(ce *ast.CallExpr) (ast.Node, bool) {
	name := getCallExprLiteral(ce)
	switch name {
	case "errors.Wrap":
		return rewriteWrap(ce), true
	case "errors.Wrapf":
		return rewriteWrap(ce), true
	case "errors.Errorf":
		return newErrorfExpr(ce.Args), true
	default:
		return ce, true
	}
}

func handleImportDecl(gd *ast.GenDecl) (ast.Node, bool) {
	// Ignore GenDecl's that aren't imports.
	if gd.Tok != token.IMPORT {
		return gd, true
	}
	// Push "errors" to the front of specs so formatting will sort it with
	// core libraries and discard pkg/errors.
	newSpecs := []ast.Spec{
		&ast.ImportSpec{Path: &ast.BasicLit{Kind: token.STRING, Value: `"errors"`}},
	}
	for _, s := range gd.Specs {
		im, ok := s.(*ast.ImportSpec)
		if !ok {
			continue
		}
		if im.Path.Value == `"github.com/pkg/errors"` {
			continue
		}
		newSpecs = append(newSpecs, s)
	}
	gd.Specs = newSpecs
	return gd, true
}

func rewriteWrap(ce *ast.CallExpr) *ast.CallExpr {
	// Rotate err to the end of a new args list
	newArgs := make([]ast.Expr, len(ce.Args)-1)
	copy(newArgs, ce.Args[1:])
	newArgs = append(newArgs, ce.Args[0])

	// If the format string is a fmt.Sprintf call, we can unwrap it.
	c, name := getCallExpr(newArgs[0])
	if c != nil && name == "fmt.Sprintf" {
		newArgs = append(c.Args, newArgs[1:]...)
	}

	// If the format string is a literal, we can rewrite it:
	//     "......" -> "......: %w"
	// Otherwise, we replace it with a binary op to add the wrap code:
	//     SomeNonLiteral -> SomeNonLiteral + ": %w"
	fmtStr, ok := newArgs[0].(*ast.BasicLit)
	if ok {
		// Strip trailing `"` and append wrap code and new trailing `"`
		fmtStr.Value = fmtStr.Value[:len(fmtStr.Value)-1] + `: %w"`
	} else {
		binOp := &ast.BinaryExpr{
			X:  newArgs[0],
			Op: token.ADD,
			Y:  &ast.BasicLit{Kind: token.STRING, Value: `": %w"`},
		}
		newArgs[0] = binOp
	}

	return newErrorfExpr(newArgs)
}

func getCallExpr(n ast.Node) (*ast.CallExpr, string) {
	c, ok := n.(*ast.CallExpr)
	if !ok {
		return nil, ""
	}
	name := getCallExprLiteral(c)
	if name == "" {
		return nil, ""
	}
	return c, name
}

func getCallExprLiteral(c *ast.CallExpr) string {
	s, ok := c.Fun.(*ast.SelectorExpr)
	if !ok {
		return ""
	}

	i, ok := s.X.(*ast.Ident)
	if !ok {
		return ""
	}

	return i.Name + "." + s.Sel.Name
}

func newErrorfExpr(args []ast.Expr) *ast.CallExpr {
	return &ast.CallExpr{
		Fun: &ast.SelectorExpr{
			X:   &ast.Ident{Name: "fmt"},
			Sel: &ast.Ident{Name: "Errorf"},
		},
		Args: args,
	}
}
