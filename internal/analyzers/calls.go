package analyzers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func newNoBoolLiteralArgs() ruleSpec {
	return newAnalyzer(
		"LEG035",
		"no-bool-literal-args",
		"Avoid boolean literals as call arguments.",
		func(pass *analysis.Pass) (any, error) {
			checkBoolLiteralArgs(pass)
			return nil, nil
		},
	)
}

func newNoDeepCompositeLiteralArg(settings Settings) ruleSpec {
	max := settings.maxCompositeLiteralArgDepth()
	return newAnalyzer(
		"LEG037",
		"no-deep-composite-literal-arg",
		"Avoid deeply nested composite literals as call arguments.",
		func(pass *analysis.Pass) (any, error) {
			checkDeepCompositeLiteralArgs(pass, max)
			return nil, nil
		},
	)
}

func checkBoolLiteralArgs(pass *analysis.Pass) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if ok {
				checkCallBoolLiteralArgs(pass, call)
			}

			return true
		})
	}
}

func checkCallBoolLiteralArgs(pass *analysis.Pass, call *ast.CallExpr) {
	if isAppendCall(call) {
		return
	}

	for _, arg := range call.Args {
		if isBoolLiteral(arg) {
			report(
				pass,
				arg,
				"LEG035",
				"no-bool-literal-args",
				"Use a named value instead of a boolean literal argument.",
			)
		}
	}
}

func isAppendCall(call *ast.CallExpr) bool {
	identifier, ok := call.Fun.(*ast.Ident)
	if !ok {
		return false
	}

	return identifier.Name == "append"
}

func checkDeepCompositeLiteralArgs(pass *analysis.Pass, max int) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			call, ok := node.(*ast.CallExpr)
			if ok {
				checkCallCompositeLiteralArgs(pass, call, max)
			}

			return true
		})
	}
}

func checkCallCompositeLiteralArgs(pass *analysis.Pass, call *ast.CallExpr, max int) {
	for _, arg := range call.Args {
		depth := compositeLiteralDepth(arg)
		if depth > max {
			report(
				pass,
				arg,
				"LEG037",
				"no-deep-composite-literal-arg",
				"Extract this composite literal before passing it.",
			)
		}
	}
}

func compositeLiteralDepth(expression ast.Expr) int {
	literal, ok := expression.(*ast.CompositeLit)
	if !ok {
		return 0
	}

	maxChildDepth := 0
	for _, element := range literal.Elts {
		childDepth := compositeElementDepth(element)
		if childDepth > maxChildDepth {
			maxChildDepth = childDepth
		}
	}

	return 1 + maxChildDepth
}

func compositeElementDepth(expression ast.Expr) int {
	keyValue, ok := expression.(*ast.KeyValueExpr)
	if ok {
		return compositeLiteralDepth(keyValue.Value)
	}

	return compositeLiteralDepth(expression)
}
