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
	"strconv"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

var needRuntime bool = false
var additionalNode ast.Stmt

func main() {
	//pProjectPath := flag.String("path","","Full path of the target project")
	//pRelativePath := flag.String("include","","Relative path (what's after /src/) of the target project")
	//pShowCompileError := flag.Bool("compile-error", false, "If fail to compile a package, show the errors of compilation")
	//pExcludePath := flag.String("exclude", "vendor", "Name of directories that you want to ignore, divided by \":\"")
	//pRobustMod := flag.Bool("r", false, "If the main package can't pass compiler, check subdirectories one by one")
	pFile := flag.String("file", "", "Full path of the target file to be parsed")

	flag.Parse()

	filename := *pFile

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

	newAST := astutil.Apply(oldAST, pre, nil)

	if needRuntime {
		ok := astutil.AddNamedImport(tokenFSet, oldAST, "gooracle", "gooracle")
		if !ok {
			fmt.Printf("add import failed when parsing %s\n", filename)
		}
	}

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
}

func pre(c *astutil.Cursor) bool {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recover in pre(): c.Name():", c.Name())
		}
	}()
	if additionalNode != nil && c.Node() == additionalNode {
		newBeforeTestCall := NewArgCallExpr("gooracle", "BeforeRun", nil)
		c.InsertBefore(newBeforeTestCall)

		newAfterTestCall := NewArgCall("gooracle", "AfterRun", nil)
		newDefer := &ast.DeferStmt{
			Defer: 0,
			Call:  newAfterTestCall,
		}
		c.InsertBefore(newDefer)
		additionalNode = nil
		needRuntime = true
	}
	switch concrete := c.Node().(type) {
	//case *ast.GoStmt:
	//	//newCallExpr := NewArgCallExpr("count", "NewGo", nil)
	//	//c.InsertBefore(newCallExpr)
	//	//usePackage = true
	//
	//case *ast.AssignStmt:
	//	//var copyLhs []ast.Expr
	//	//
	//	//rhs := concrete.Rhs
	//	//boolIsMake, boolIsChanRhs := false, false
	//	//if len(rhs) == 1 && len(concrete.Lhs) == 1 {
	//	//	if makeFnCallExpr, ok := rhs[0].(*ast.CallExpr); ok {
	//	//		if makeFnIdent, ok := makeFnCallExpr.Fun.(*ast.Ident); ok {
	//	//			if makeFnIdent.Name == "make" {
	//	//				boolIsMake = true
	//	//			}
	//	//		}
	//	//		if len(makeFnCallExpr.Args) <= 2 && len(makeFnCallExpr.Args) >= 1 { // make chan can have 1 or 2 arguments
	//	//			if _, ok := makeFnCallExpr.Args[0].(*ast.ChanType); ok {
	//	//				boolIsChanRhs = true
	//	//			}
	//	//		}
	//	//	}
	//	//}
	//	//
	//	//copyLhs = append(copyLhs, concrete.Lhs[0])
	//	//
	//	//if boolIsMake && boolIsChanRhs {
	//	//	newCallExpr := NewArgCallExpr("count", "NewCh", copyLhs)
	//	//	c.InsertAfter(newCallExpr)
	//	//	usePackage = true
	//	//}
	//
	//case *ast.SendStmt:
	//	//var copyLhs []ast.Expr
	//	//copyLhs = append(copyLhs, concrete.Chan)
	//	newBeforeExpr := NewArgCallExpr("gooracle", "BeforeBlock", nil)
	//	c.InsertBefore(newBeforeExpr)
	//	newAfterExpr := NewArgCallExpr("gooracle", "AfterBlock", nil)
	//	c.InsertAfter(newAfterExpr)
	//	usePackage = true
	//
	//case *ast.ExprStmt:
	//	//if unary, ok := concrete.X.(*ast.UnaryExpr); ok {
	//	//	if unary.Op == token.ARROW {
	//	//		var copyLhs []ast.Expr
	//	//		copyLhs = append(copyLhs, unary.X)
	//	//		newCallExpr := NewArgCallExpr("count", "NewOp", copyLhs)
	//	//		c.InsertAfter(newCallExpr)
	//	//		usePackage = true
	//	//	}
	//	//}
	//	//
	//	//if call, ok := concrete.X.(*ast.CallExpr); ok {
	//	//	if fun_ident, ok := call.Fun.(*ast.Ident); ok {
	//	//		if fun_ident.Name == "close" {
	//	//			var copyLhs []ast.Expr
	//	//			for _, arg := range call.Args {
	//	//				copyLhs = append(copyLhs, arg)
	//	//			}
	//	//			newCallExpr := NewArgCallExpr("count", "NewOp", copyLhs)
	//	//			c.InsertAfter(newCallExpr)
	//	//			usePackage = true
	//	//		}
	//	//	}
	//	//}
	//
	//case *ast.SelectStmt:
	//	//cases := concrete.Body.List
	//	//for _, case_ := range cases {
	//	//	if case_cc,ok := case_.(*ast.CommClause); ok {
	//	//		case_comm := case_cc.Comm
	//	//		switch case_concrete := case_comm.(type) {
	//	//		case *ast.SendStmt:
	//	//			var copyLhs []ast.Expr
	//	//			copyLhs = append(copyLhs, case_concrete.Chan)
	//	//			newCallExpr := NewArgCallExpr("count", "NewOp", copyLhs)
	//	//			c.InsertBefore(newCallExpr)
	//	//			usePackage = true
	//	//		case *ast.ExprStmt:
	//	//			if unary, ok := case_concrete.X.(*ast.UnaryExpr); ok {
	//	//				if unary.Op == token.ARROW {
	//	//					var copyLhs []ast.Expr
	//	//					copyLhs = append(copyLhs, unary.X)
	//	//					newCallExpr := NewArgCallExpr("count", "NewOp", copyLhs)
	//	//					c.InsertBefore(newCallExpr)
	//	//					usePackage = true
	//	//				}
	//	//			}
	//	//		}
	//	//	}
	//	//}
	//	_ = concrete
	//	newBeforeExpr := NewArgCallExpr("gooracle", "BeforeBlock", nil)
	//	c.InsertBefore(newBeforeExpr)
	//	newAfterExpr := NewArgCallExpr("gooracle", "AfterBlock", nil)
	//	c.InsertAfter(newAfterExpr)

	//case *ast.SwitchStmt:
	//	print()

	case *ast.FuncDecl:
		if strings.HasPrefix(concrete.Name.Name, "Test") {
			if concrete.Body != nil && len(concrete.Body.List) != 0 {
				firstStmt := concrete.Body.List[0]
				additionalNode = firstStmt
			}

		}

	default:
	}

	return true
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
