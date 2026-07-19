package analyzers

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

func newNoRedundantBooleanLogic() ruleSpec {
	return newAnalyzer(
		"LEG006",
		"no-redundant-boolean-logic",
		"Avoid redundant boolean comparisons.",
		func(pass *analysis.Pass) (any, error) {
			checkRedundantBooleanLogic(pass)
			return nil, nil
		},
	)
}

func checkRedundantBooleanLogic(pass *analysis.Pass) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			expression, ok := node.(*ast.BinaryExpr)
			if ok {
				checkBooleanComparison(pass, expression)
			}

			return true
		})
	}
}

func checkBooleanComparison(pass *analysis.Pass, expression *ast.BinaryExpr) {
	if !isBooleanComparison(expression) {
		return
	}

	if !hasBoolLiteralComparison(expression) {
		return
	}

	reportBooleanComparison(pass, expression)
}

func reportBooleanComparison(pass *analysis.Pass, expression *ast.BinaryExpr) {
	report(
		pass,
		expression,
		"LEG006",
		"no-redundant-boolean-logic",
		"Avoid comparing a boolean expression to true or false.",
	)
}

func isBooleanComparison(expression *ast.BinaryExpr) bool {
	isEquality := expression.Op == token.EQL
	isInequality := expression.Op == token.NEQ
	return isEquality || isInequality
}

func hasBoolLiteralComparison(expression *ast.BinaryExpr) bool {
	leftIsBool := isBoolLiteral(expression.X)
	rightIsBool := isBoolLiteral(expression.Y)
	return leftIsBool || rightIsBool
}

func isBoolLiteral(expression ast.Expr) bool {
	identifier, ok := expression.(*ast.Ident)
	if !ok {
		return false
	}

	if identifier.Name == "true" {
		return true
	}

	return identifier.Name == "false"
}
