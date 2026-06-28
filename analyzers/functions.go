package analyzers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func newMaxFunctionLines(settings Settings) ruleSpec {
	max := settings.maxFunctionLines()
	return newAnalyzer(
		"LEG038",
		"max-function-lines",
		"Limit functions to a focused line budget.",
		func(pass *analysis.Pass) (any, error) {
			checkFunctionLines(pass, max)
			return nil, nil
		},
	)
}

func checkFunctionLines(pass *analysis.Pass, max int) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			checkFunctionLineBudget(pass, node, max)
			return true
		})
	}
}

func checkFunctionLineBudget(pass *analysis.Pass, node ast.Node, max int) {
	body := functionBody(node)
	if body == nil {
		return
	}

	lines := nodeLineSpan(pass, node)
	if lines <= max {
		return
	}

	report(
		pass,
		functionReportNode(node),
		"LEG038",
		"max-function-lines",
		"Function is too long. Split it into focused helpers.",
	)
}

func nodeLineSpan(pass *analysis.Pass, node ast.Node) int {
	start := pass.Fset.Position(node.Pos()).Line
	end := pass.Fset.Position(node.End()).Line
	lineSpan := end - start
	return lineSpan + 1
}

func functionReportNode(node ast.Node) ast.Node {
	decl, ok := node.(*ast.FuncDecl)
	if ok {
		return decl.Name
	}

	return node
}
