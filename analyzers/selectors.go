package analyzers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func newNoDeepSelectorChain(settings Settings) ruleSpec {
	max := settings.maxSelectorChainDepth()
	return newAnalyzer(
		"LEG031",
		"no-deep-selector-chain",
		"Avoid deep selector or index chains without named intermediate values.",
		func(pass *analysis.Pass) (any, error) {
			checkSelectorChains(pass, max)
			return nil, nil
		},
	)
}

func checkSelectorChains(pass *analysis.Pass, max int) {
	parents := buildParentMap(pass.Files)
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			expression, ok := node.(ast.Expr)
			if ok {
				checkSelectorChain(pass, parents, expression, max)
			}

			return true
		})
	}
}

func checkSelectorChain(
	pass *analysis.Pass,
	parents map[ast.Node]ast.Node,
	expression ast.Expr,
	max int,
) {
	if isNestedAccessChain(parents, expression) {
		return
	}

	depth := accessChainDepth(expression)
	if depth <= max {
		return
	}

	report(
		pass,
		expression,
		"LEG031",
		"no-deep-selector-chain",
		"Selector or index chain is too deep. Extract a named intermediate value.",
	)
}

func isNestedAccessChain(parents map[ast.Node]ast.Node, expression ast.Expr) bool {
	parent := parents[expression]
	switch typed := parent.(type) {
	case *ast.SelectorExpr:
		return typed.X == expression
	case *ast.IndexExpr:
		return typed.X == expression
	case *ast.IndexListExpr:
		return typed.X == expression
	default:
		return false
	}
}

func accessChainDepth(expression ast.Expr) int {
	switch typed := expression.(type) {
	case *ast.SelectorExpr:
		return 1 + accessChainDepth(typed.X)
	case *ast.IndexExpr:
		return 1 + accessChainDepth(typed.X)
	case *ast.IndexListExpr:
		return 1 + accessChainDepth(typed.X)
	default:
		return 0
	}
}
