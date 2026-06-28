package analyzers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func newNoQuadraticPatterns() ruleSpec {
	return newAnalyzer(
		"LEG005",
		"no-quadratic-patterns",
		"Flag likely quadratic nested loops.",
		func(pass *analysis.Pass) (any, error) {
			checkNestedLoops(pass)
			return nil, nil
		},
	)
}

func checkNestedLoops(pass *analysis.Pass) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			loop, ok := loopBody(node)
			if ok {
				checkLoopBodyForNestedLoops(pass, loop)
				return false
			}

			return true
		})
	}
}

func checkLoopBodyForNestedLoops(pass *analysis.Pass, body *ast.BlockStmt) {
	ast.Inspect(body, func(node ast.Node) bool {
		if node == body {
			return true
		}

		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}

		if isLoopNode(node) {
			report(
				pass,
				node,
				"LEG005",
				"no-quadratic-patterns",
				"Nested loop detected. Consider pre-indexing with a map.",
			)
			return false
		}

		return true
	})
}

func loopBody(node ast.Node) (*ast.BlockStmt, bool) {
	switch typed := node.(type) {
	case *ast.ForStmt:
		return typed.Body, true
	case *ast.RangeStmt:
		return typed.Body, true
	default:
		return nil, false
	}
}

func isLoopNode(node ast.Node) bool {
	switch node.(type) {
	case *ast.ForStmt, *ast.RangeStmt:
		return true
	default:
		return false
	}
}
