package analyzers

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

type lookupPart struct {
	key  string
	node ast.Node
}

func newPreferObjectLookup(settings Settings) ruleSpec {
	min := settings.minObjectLookupChainLength()
	return newAnalyzer(
		"LEG024",
		"prefer-object-lookup",
		"Prefer set or map lookups over long equality-or chains.",
		func(pass *analysis.Pass) (any, error) {
			checkObjectLookup(pass, min)
			return nil, nil
		},
	)
}

func checkObjectLookup(pass *analysis.Pass, min int) {
	parents := buildParentMap(pass.Files)
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			expression, ok := node.(*ast.BinaryExpr)
			if ok {
				checkObjectLookupExpression(pass, parents, expression, min)
			}

			return true
		})
	}
}

func checkObjectLookupExpression(
	pass *analysis.Pass,
	parents map[ast.Node]ast.Node,
	expression *ast.BinaryExpr,
	min int,
) {
	if !isTopLevelOrExpression(parents, expression) {
		return
	}

	parts := collectLookupParts(pass, expression)
	if !hasObjectLookupParts(parts, min) {
		return
	}

	reportObjectLookup(pass, expression)
}

func isTopLevelOrExpression(parents map[ast.Node]ast.Node, expression *ast.BinaryExpr) bool {
	if expression.Op != token.LOR {
		return false
	}

	return !isNestedOr(parents, expression)
}

func hasObjectLookupParts(parts []lookupPart, min int) bool {
	hasEnoughParts := len(parts) >= min
	if !hasEnoughParts {
		return false
	}

	return sameLookupKey(parts)
}

func reportObjectLookup(pass *analysis.Pass, expression ast.Expr) {
	report(
		pass,
		expression,
		"LEG024",
		"prefer-object-lookup",
		"Use a set, map, or switch instead of a long equality-or chain.",
	)
}

func collectLookupParts(pass *analysis.Pass, expression ast.Expr) []lookupPart {
	binary, ok := expression.(*ast.BinaryExpr)
	if !ok {
		return nil
	}

	if binary.Op == token.LOR {
		left := collectLookupParts(pass, binary.X)
		return append(left, collectLookupParts(pass, binary.Y)...)
	}

	return equalityLookupPart(pass, binary)
}

func equalityLookupPart(pass *analysis.Pass, expression *ast.BinaryExpr) []lookupPart {
	isEquality := expression.Op == token.EQL
	isInequality := expression.Op == token.NEQ
	isComparison := isEquality || isInequality
	if !isComparison {
		return nil
	}

	if isLiteral(expression.Y) {
		return []lookupPart{{key: nodeText(pass, expression.X), node: expression}}
	}

	if isLiteral(expression.X) {
		return []lookupPart{{key: nodeText(pass, expression.Y), node: expression}}
	}

	return nil
}

func isNestedOr(parents map[ast.Node]ast.Node, expression *ast.BinaryExpr) bool {
	parent, ok := parents[expression].(*ast.BinaryExpr)
	if !ok {
		return false
	}

	return parent.Op == token.LOR
}

func sameLookupKey(parts []lookupPart) bool {
	if len(parts) == 0 {
		return false
	}

	first := parts[0].key
	for _, part := range parts {
		if part.key != first {
			return false
		}
	}

	return true
}

func isLiteral(expression ast.Expr) bool {
	switch typed := expression.(type) {
	case *ast.BasicLit:
		return true
	case *ast.Ident:
		return isLiteralIdentifier(typed.Name)
	default:
		return false
	}
}

func isLiteralIdentifier(name string) bool {
	switch name {
	case "true", "false", "nil":
		return true
	default:
		return false
	}
}
