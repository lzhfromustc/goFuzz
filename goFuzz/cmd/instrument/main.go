package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"hash/fnv"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

var (
	boolNeedInstrument bool = false // remember: if we instrument, we always need to import gooracle
	additionalNode     ast.Stmt
	currentFSet        *token.FileSet
	Uint16OpID         uint16
	recordOutputFile   string
	records            []string
	numOfSelects       uint32
)

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
	pOutput := flag.String("output", "", "primitive operation output")
	flag.Parse()

	recordOutputFile = *pOutput
	filename := *pFile

	// If this file is in SDK and can't be instrumented, skip it
	for _, str := range sliceStrNoInstr {
		if strings.Contains(filename, str) {
			return
		}
	}

	h := fnv.New32a()
	h.Write([]byte(filename))
	Uint16OpID = uint16(h.Sum32())

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

	for _, vecImportSpec := range astutil.Imports(tokenFSet, oldAST) {
		for _, importSpec := range vecImportSpec {
			if importSpec != nil {
				if importSpec.Name != nil {
					if importSpec.Name.Name == "gooracle" { // instrumented before
						fmt.Printf("This file is already instrumented before:%s", filename)
						return
					}
				}
			}

		}
	}

	currentFSet = tokenFSet

	newAST := astutil.Apply(oldAST, preWithoutSelect, nil)
	newASTAfterSelectCopy := astutil.Apply(newAST, preOnlySelect, nil)

	if !boolNeedInstrument {
		return
	}

	ok := astutil.AddNamedImport(tokenFSet, oldAST, "gooracle", "gooracle")
	if !ok {
		fmt.Printf("add import failed when parsing %s\n", filename)
		return
	}

	buf := &bytes.Buffer{}
	err = format.Node(buf, tokenFSet, newASTAfterSelectCopy)
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

	fmt.Printf("number of selects: %d\n", numOfSelects)

}

func preWithoutSelect(c *astutil.Cursor) bool {
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
		newAssign := &ast.AssignStmt{
			Lhs:    nil,
			TokPos: 0,
			Tok:    token.DEFINE,
			Rhs:    nil,
		}
		newObject := &ast.Object{
			Kind: ast.Var,
			Name: "oracleEntry",
			Decl: newAssign,
			Data: nil,
			Type: nil,
		}
		newIdent := &ast.Ident{
			Name: "oracleEntry",
			Obj:  newObject,
		}
		newAssign.Lhs = []ast.Expr{
			newIdent,
		}
		newAssign.Rhs = []ast.Expr{
			NewArgCall("gooracle", "BeforeRun", nil),
		}
		c.InsertBefore(newAssign)

		newAfterTestCall := NewArgCall("gooracle", "AfterRun", []ast.Expr{
			newIdent,
		})
		newDefer := &ast.DeferStmt{
			Defer: 0,
			Call:  newAfterTestCall,
		}
		c.InsertBefore(newDefer)
		additionalNode = nil
		boolNeedInstrument = true
	}

	if cStmt, ok := c.Node().(ast.Stmt); ok {
		for _, recvAndFirstStmt := range vecRecvAndFirstStmt {
			if recvAndFirstStmt.firstStmt == cStmt {
				newCall := NewArgCallExpr("gooracle", "CurrentGoAddValue", []ast.Expr{
					&ast.Ident{
						Name: recvAndFirstStmt.recvName,
						Obj:  recvAndFirstStmt.recvObj,
					},
					&ast.Ident{
						Name: "nil",
						Obj:  nil,
					},
					&ast.BasicLit{
						ValuePos: 0,
						Kind:     token.INT,
						Value:    "0",
					},
				})
				c.InsertBefore(newCall)
				boolNeedInstrument = true // We need to import gooracle
			}
		}
	}

	switch concrete := c.Node().(type) {

	case *ast.SendStmt: // This is a send operation
		intID := int(Uint16OpID)
		newCall := NewArgCallExpr("gooracle", "StoreOpInfo", []ast.Expr{&ast.BasicLit{
			ValuePos: 0,
			Kind:     token.STRING,
			Value:    "\"Send\"",
		}, &ast.BasicLit{
			ValuePos: 0,
			Kind:     token.INT,
			Value:    strconv.Itoa(intID),
		}})
		c.InsertBefore(newCall) // Insert the call to store this operation's type and ID into goroutine local storage
		records = append(records, strconv.Itoa(intID)+":chsend")
		Uint16OpID++

		boolNeedInstrument = true // We need to import gooracle

	case *ast.AssignStmt:
		if len(concrete.Rhs) == 1 {
			if callExpr, ok := concrete.Rhs[0].(*ast.CallExpr); ok { // calling on the right
				if funcIdent, ok := callExpr.Fun.(*ast.Ident); ok {
					if funcIdent.Name == "make" {
						if len(callExpr.Args) == 1 { // This is a make operation
							if _, ok := callExpr.Args[0].(*ast.ChanType); ok {
								intID := int(Uint16OpID)
								newCall := NewArgCallExpr("gooracle", "StoreChMakeInfo", []ast.Expr{
									concrete.Lhs[0],
									&ast.BasicLit{
										ValuePos: 0,
										Kind:     token.INT,
										Value:    strconv.Itoa(intID),
									}})
								c.InsertAfter(newCall)
								boolNeedInstrument = true // We need to import gooracle
								records = append(records, strconv.Itoa(intID)+":chmake")
								Uint16OpID++
							}
						}
					}
				}
			} else {
			}

			// Below is our old instrumentation plan for trad primitives, which cause compilation errors when we take
			// the address of an element in map (not allowed in Go)

			//if compositeLit, ok := concrete.Rhs[0].(*ast.CompositeLit); ok { // like `mu := sync.Mutex{}`
			//	if typeSelectorExpr, ok := compositeLit.Type.(*ast.SelectorExpr); ok { // like `sync.Mutex{}`
			//		if pkgIdent, ok := typeSelectorExpr.X.(*ast.Ident); ok { // like `sync`
			//			if pkgIdent.Name == "sync" {
			//				var boolIsSyncAlloc bool = true
			//				var strCallee string
			//				switch typeSelectorExpr.Sel.Name {
			//				case "WaitGroup":
			//					strCallee = "RecordWgCreate"
			//				case "Mutex":
			//					strCallee = "RecordMutexCreate"
			//				case "RWMutex":
			//					strCallee = "RecordRWMutexCreate"
			//				case "Cond":
			//					strCallee = "RecordCondCreate"
			//				default:
			//					boolIsSyncAlloc = false
			//				}
			//				if boolIsSyncAlloc {
			//					position := currentFSet.Position(concrete.TokPos)
			//					position.Filename, _ = filepath.Abs(position.Filename)
			//					intID := int(Uint16OpID)
			//					newCall := NewArgCallExpr("gooracle", strCallee, []ast.Expr{
			//						&ast.UnaryExpr{
			//							OpPos: 0,
			//							Op:    token.AND,
			//							X:     concrete.Lhs[0],
			//						},
			//						&ast.BasicLit{
			//							ValuePos: 0,
			//							Kind:     token.STRING,
			//							Value:    "\"" + position.Filename + ":" + strconv.Itoa(position.Line) + "\"",
			//						},
			//						&ast.BasicLit{
			//							ValuePos: 0,
			//							Kind:     token.INT,
			//							Value:    strconv.Itoa(intID),
			//						},
			//					})
			//					c.InsertAfter(newCall)
			//					boolNeedInstrument = true // We need to import gooracle
			//					records = append(records, strconv.Itoa(intID)+":"+strCallee)
			//					Uint16OpID++
			//				}
			//			}
			//		}
			//	}
			//}
		}

	case *ast.ExprStmt:
		if unaryExpr, ok := concrete.X.(*ast.UnaryExpr); ok {
			if unaryExpr.Op == token.ARROW { // This is a receive operation
				intID := int(Uint16OpID)
				newCall := NewArgCallExpr("gooracle", "StoreOpInfo", []ast.Expr{&ast.BasicLit{
					ValuePos: 0,
					Kind:     token.STRING,
					Value:    "\"Recv\"",
				}, &ast.BasicLit{
					ValuePos: 0,
					Kind:     token.INT,
					Value:    strconv.Itoa(intID),
				}})
				c.InsertBefore(newCall)
				boolNeedInstrument = true // We need to import gooracle
				records = append(records, strconv.Itoa(intID)+":chrecv")
				Uint16OpID++
			}
		} else if callExpr, ok := concrete.X.(*ast.CallExpr); ok { // like `close(ch)` or `mu.Lock()`
			if funcIdent, ok := callExpr.Fun.(*ast.Ident); ok { // like `close(ch)`
				if funcIdent.Name == "close" { // This is a close operation
					intID := int(Uint16OpID)
					newCall := NewArgCallExpr("gooracle", "StoreOpInfo", []ast.Expr{&ast.BasicLit{
						ValuePos: 0,
						Kind:     token.STRING,
						Value:    "\"Close\"",
					}, &ast.BasicLit{
						ValuePos: 0,
						Kind:     token.INT,
						Value:    strconv.Itoa(intID),
					}})
					c.InsertBefore(newCall)
					boolNeedInstrument = true // We need to import gooracle
					records = append(records, strconv.Itoa(intID)+":chclose")
					Uint16OpID++
				}
			} else if selectorExpr, ok := callExpr.Fun.(*ast.SelectorExpr); ok { // like `mu.Lock()`
				var boolNameMatchSyncCall bool = true
				var strOpType string
				switch selectorExpr.Sel.Name {
				case "Lock":
					strOpType = "Lock"
				case "RUnlock":
					strOpType = "RUnlock"
				case "RLock":
					strOpType = "RLock"
				case "Unlock":
					strOpType = "Unlock"
				case "Add":
					strOpType = "Add"
				case "Done":
					strOpType = "Done"
				case "Broadcast":
					strOpType = "Broadcast"
				case "Signal":
					strOpType = "Signal"
				case "Wait":
					strOpType = "Wait"
				default:
					boolNameMatchSyncCall = false
				}
				if boolNameMatchSyncCall {
					intID := int(Uint16OpID)
					newCall := NewArgCallExpr("gooracle", "StoreOpInfo", []ast.Expr{&ast.BasicLit{
						ValuePos: 0,
						Kind:     token.STRING,
						Value:    "\"" + strOpType + "\"",
					}, &ast.BasicLit{
						ValuePos: 0,
						Kind:     token.INT,
						Value:    strconv.Itoa(intID),
					}})
					c.InsertBefore(newCall)
					boolNeedInstrument = true // We need to import gooracle
					records = append(records, strconv.Itoa(intID)+":trad"+strOpType)
					Uint16OpID++
				}
			}
		}

	//case *ast.SwitchStmt:
	//	positionOriSelect := currentFSet.Position(concrete.Switch)
	//	_ = positionOriSelect

	case *ast.FuncDecl:
		if strings.HasPrefix(concrete.Name.Name, "Test") {
			if concrete.Body != nil && len(concrete.Body.List) != 0 {
				firstStmt := concrete.Body.List[0]
				additionalNode = firstStmt
				boolNeedInstrument = true // We need to import gooracle
			}

		} else if concrete.Recv != nil && concrete.Body != nil {
			if len(concrete.Recv.List) == 1 && len(concrete.Body.List) > 0 {
				recvField := concrete.Recv.List[0]
				if len(recvField.Names) == 1 {
					ident := recvField.Names[0]
					if ident.Name != "" {
						recvAndFirstStmt := &RecvAndFirstStmt{
							recvName:  ident.Name,
							firstStmt: concrete.Body.List[0],
							recvObj:   ident.Obj,
						}
						vecRecvAndFirstStmt = append(vecRecvAndFirstStmt, recvAndFirstStmt)
					}
				}
			}
		}

	default:
	}

	return true
}

type RecvAndFirstStmt struct {
	recvName  string
	firstStmt ast.Stmt
	recvObj   *ast.Object
}

var vecRecvAndFirstStmt []*RecvAndFirstStmt

func preOnlySelect(c *astutil.Cursor) bool {
	defer func() {
		if r := recover(); r != nil { // This is allowed. If we insert node into nodes not in slice, we will meet a panic
			// For example, we may identified a receive in select and wanted to insert a function call before it, then this function will panic

			//fmt.Printf("Recover in pre(): c.Name(): %s\n", c.Name())
			//position := currentFSet.Position(c.Node().Pos())
			//position.Filename, _ = filepath.Abs(position.Filename)
			//fmt.Printf("\tLocation: %s\n", position.Filename + ":" + strconv.Itoa(position.Line))
		}
	}()

	switch concrete := c.Node().(type) {

	case *ast.SelectStmt:
		numOfSelects += 1
		positionOriSelect := currentFSet.Position(concrete.Select)
		positionOriSelect.Filename, _ = filepath.Abs(positionOriSelect.Filename)
		// store the original select
		oriSelect := SelectStruct{
			StmtSelect:    concrete,
			VecCommClause: nil,
			VecOp:         nil,
			VecBody:       nil,
		}
		for _, stmtCommClause := range concrete.Body.List {
			commClause, _ := stmtCommClause.(*ast.CommClause)
			oriSelect.VecCommClause = append(oriSelect.VecCommClause, commClause)
			oriSelect.VecOp = append(oriSelect.VecOp, commClause.Comm)
			vecContent := []ast.Stmt{}
			for _, stmt := range commClause.Body {
				vecContent = append(vecContent, stmt)
			}
			oriSelect.VecBody = append(oriSelect.VecBody, vecContent)
		}

		// create a switch
		newSwitch := &ast.SwitchStmt{
			Switch: 0,
			Init:   nil,
			Tag: NewArgCall("gooracle", "ReadSelect", []ast.Expr{
				&ast.BasicLit{ // first parameter: filename
					ValuePos: 0,
					Kind:     token.STRING,
					Value:    "\"" + positionOriSelect.Filename + "\"",
				}, &ast.BasicLit{ // second parameter: linenumber of original select
					ValuePos: 0,
					Kind:     token.INT,
					Value:    strconv.Itoa(positionOriSelect.Line),
				}, &ast.BasicLit{
					ValuePos: 0,
					Kind:     token.INT,
					Value:    strconv.Itoa(len(oriSelect.VecCommClause)),
				}}),
			Body: &ast.BlockStmt{
				Lbrace: 0,
				List:   nil,
				Rbrace: 0,
			},
		}
		vecCaseClause := []ast.Stmt{}
		// The number of switch case is (the number of non-default select cases + 1)
		for i, stmtOp := range oriSelect.VecOp {

			// if the case's expression is nil, it means this case is a default. We don't intrument this case
			// but we instrument other cases
			if stmtOp == nil {
				continue
			}

			newCaseClause := &ast.CaseClause{
				Case:  0,
				List:  nil,
				Colon: 0,
				Body:  nil,
			}
			newBasicLit := &ast.BasicLit{
				ValuePos: 0,
				Kind:     token.INT,
				Value:    strconv.Itoa(i),
			}
			newCaseClause.List = []ast.Expr{newBasicLit}

			// the case's content is one select statement
			newSelect := &ast.SelectStmt{
				Select: 0,
				Body:   &ast.BlockStmt{},
			}
			firstSelectCase := &ast.CommClause{
				Case:  0,
				Comm:  oriSelect.VecOp[i],
				Colon: 0,
				Body:  copyStmtBody(oriSelect.VecBody[i]),
			}
			secondSelectCase := &ast.CommClause{
				Case: 0,
				Comm: &ast.ExprStmt{X: &ast.UnaryExpr{
					OpPos: 0,
					Op:    token.ARROW,
					X:     NewArgCall("gooracle", "SelectTimeout", nil),
				}},
				Colon: 0,
				Body: []ast.Stmt{
					// The first line is a call to gooracle.StoreLastMySwitchChoice(-1)
					// The second line is a copy of original select
					&ast.ExprStmt{X: NewArgCall("gooracle", "StoreLastMySwitchChoice", []ast.Expr{&ast.UnaryExpr{
						OpPos: 0,
						Op:    token.SUB,
						X: &ast.BasicLit{
							ValuePos: 0,
							Kind:     token.INT,
							Value:    "1",
						},
					}})},
					copySelect(oriSelect.StmtSelect)},
			}
			newSelect.Body.List = append(newSelect.Body.List, firstSelectCase, secondSelectCase)

			newCaseClause.Body = []ast.Stmt{newSelect}

			// add the created case to vector
			vecCaseClause = append(vecCaseClause, newCaseClause)
		}

		// add one default case to switch
		newCaseClauseDefault := &ast.CaseClause{
			Case:  0,
			List:  nil,
			Colon: 0,
			Body: []ast.Stmt{
				// The first line is a call to gooracle.StoreLastMySwitchChoice(-1)
				// The second line is a copy of original select
				&ast.ExprStmt{X: NewArgCall("gooracle", "StoreLastMySwitchChoice", []ast.Expr{&ast.UnaryExpr{
					OpPos: 0,
					Op:    token.SUB,
					X: &ast.BasicLit{
						ValuePos: 0,
						Kind:     token.INT,
						Value:    "1",
					},
				}})},
				copySelect(oriSelect.StmtSelect)},
		}
		vecCaseClause = append(vecCaseClause, newCaseClauseDefault)

		newSwitch.Body.List = vecCaseClause

		// Insert the new switch before the select
		c.InsertBefore(newSwitch)

		// Delete the original select
		c.Delete()

		boolNeedInstrument = true // We need to import gooracle

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
