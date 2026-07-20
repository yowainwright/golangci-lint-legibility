package analyzers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func newMaxArrayChainDepth(settings Settings) ruleSpec {
	max := settings.maxArrayChainDepth()
	return newAnalyzer(
		"LEG011",
		"max-array-chain-depth",
		"Limit consecutive collection-style method chains.",
		func(pass *analysis.Pass) (any, error) {
			checkCallChains(pass, max)
			return nil, nil
		},
	)
}

func checkCallChains(pass *analysis.Pass, max int) {
	parents := buildParentMap(pass.Files)
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if ok {
				checkCallChain(pass, parents, call, max)
			}

			return true
		})
	}
}

func checkCallChain(
	pass *analysis.Pass,
	parents map[ast.Node]ast.Node,
	call *ast.CallExpr,
	max int,
) {
	if isNestedChainCall(parents, call) {
		return
	}

	depth := callChainDepth(call)
	if depth <= max {
		return
	}

	reportCallChain(pass, call)
}

func reportCallChain(pass *analysis.Pass, call *ast.CallExpr) {
	report(
		pass,
		call,
		"LEG011",
		"max-array-chain-depth",
		"Method chain is too deep. Split it into named intermediate values.",
	)
}

func isNestedChainCall(parents map[ast.Node]ast.Node, call *ast.CallExpr) bool {
	_, ok := parents[call].(*ast.SelectorExpr)
	return ok
}

func callChainDepth(expression ast.Expr) int {
	call, ok := expression.(*ast.CallExpr)
	if !ok {
		return 0
	}

	selector, ok := call.Fun.(*ast.SelectorExpr)
	if !ok {
		return 0
	}

	return 1 + callChainDepth(selector.X)
}
