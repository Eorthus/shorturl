// Package analyzer implements custom static analysis checks.
//
// Currently implemented analyzers:
//
// # ExitCheckAnalyzer
//
// ExitCheckAnalyzer detects direct calls to os.Exit in the main function of
// main packages.
package analyzers

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

// ExitCheckAnalyzer checks for direct os.Exit calls in main function of main package
var ExitCheckAnalyzer = &analysis.Analyzer{
	Name: "exitcheck",
	Doc:  "checks for direct os.Exit calls in main function",
	Run:  run,
	Requires: []*analysis.Analyzer{
		inspect.Analyzer,
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	inspector := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	nodeFilter := []ast.Node{
		(*ast.CallExpr)(nil),
	}

	inspector.Preorder(nodeFilter, func(node ast.Node) {
		call, ok := node.(*ast.CallExpr)
		if !ok {
			return
		}

		// Проверяем, что это вызов os.Exit
		if !isOsExit(pass.TypesInfo, call) {
			return
		}

		// Проверяем, что мы находимся в функции main пакета main
		if isInMainFunc(pass, node) {
			pass.Reportf(node.Pos(), "os.Exit should not be called directly in main function")
		}
	})

	return nil, nil
}

// isOsExit проверяет, является ли вызов функцией os.Exit
func isOsExit(info *types.Info, call *ast.CallExpr) bool {
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return false
	}

	pkgIdent, ok := sel.X.(*ast.Ident)
	if !ok {
		return false
	}

	if pkgIdent.Name != "os" || sel.Sel.Name != "Exit" {
		return false
	}

	return true
}

// isInMainFunc проверяет, находится ли узел в функции main пакета main
func isInMainFunc(pass *analysis.Pass, node ast.Node) bool {
	// Проверяем, что это пакет main
	if pass.Pkg.Name() != "main" {
		return false
	}

	// Ищем родительскую функцию
	currentNode := node
	for {
		parent := findParentFunc(pass, currentNode)
		if parent == nil {
			return false
		}
		if parent.Name.Name == "main" {
			return true
		}
		currentNode = parent
	}
}

// findParentFunc находит родительскую функцию для узла
func findParentFunc(pass *analysis.Pass, node ast.Node) *ast.FuncDecl {
	var result *ast.FuncDecl
	ast.Inspect(pass.Files[0], func(n ast.Node) bool {
		if n == nil {
			return false
		}
		if f, ok := n.(*ast.FuncDecl); ok {
			if nodeIsInFunc(node, f) {
				result = f
				return false
			}
		}
		return true
	})
	return result
}

// nodeIsInFunc проверяет, находится ли узел внутри функции
func nodeIsInFunc(node ast.Node, f *ast.FuncDecl) bool {
	pos := node.Pos()
	return pos >= f.Pos() && pos <= f.End()
}
