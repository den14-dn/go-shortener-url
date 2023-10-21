// Package osexitcheckanalyzer contains check for usage expression os.Exit in package main in function main.
package osexitcheckanalyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

// OsExitCheckAnalyzer describes analysis.Analyzer in which the object identifier,
// its description and execution function are defined.
var OsExitCheckAnalyzer = &analysis.Analyzer{
	Name: "osexitcheck",
	Doc:  `check for contains expression os.Exit in package main in function main`,
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	isOsExit := func(x *ast.ExprStmt) bool {
		call, ok := x.X.(*ast.CallExpr)
		if !ok {
			return false
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return false
		}

		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return false
		}

		return sel.Sel.Name == "Exit" && ident.Name == "os"
	}

	for _, file := range pass.Files {
		var fName string

		if file.Name.Name != "main" {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			switch x := node.(type) {
			case *ast.FuncDecl:
				fName = x.Name.Name
			case *ast.ExprStmt:
				if fName == "main" && isOsExit(x) {
					pass.Reportf(x.Pos(), "package main in func main contains expression os.Exit")
				}
			}

			return true
		})
	}

	return nil, nil
}
