package analyzers

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

type operatorMode int

type operatorLimit struct {
	max     int
	mode    operatorMode
	code    string
	rule    string
	message string
}

const (
	readabilityOperators operatorMode = iota
	booleanOperators
	computedValueOperators
)

func newMaxExpressionOperators(settings Settings) ruleSpec {
	max := settings.maxExpressionOperators()
	return newAnalyzer(
		"LEG001",
		"max-expression-operators",
		"Limit operators inside a single expression.",
		func(pass *analysis.Pass) (any, error) {
			checkExpressionContexts(pass, max, readabilityOperators)
			return nil, nil
		},
	)
}

func newHoistIfOperators(settings Settings) ruleSpec {
	max := settings.maxIfOperators()
	return newAnalyzer(
		"LEG002",
		"hoist-if-operators",
		"Prefer named booleans before operator-heavy conditions.",
		func(pass *analysis.Pass) (any, error) {
			checkIfConditions(pass, max)
			return nil, nil
		},
	)
}

func newNoComputedValues(settings Settings) ruleSpec {
	max := settings.maxComputedValueOperators()
	return newAnalyzer(
		"LEG012",
		"no-computed-values",
		"Prefer named values before returning computed expressions.",
		func(pass *analysis.Pass) (any, error) {
			checkComputedValues(pass, max)
			return nil, nil
		},
	)
}

func checkExpressionContexts(pass *analysis.Pass, max int, mode operatorMode) {
	seen := make(map[ast.Expr]bool)
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			checkExpressionNode(pass, node, max, mode, seen)
			return true
		})
	}
}

func checkExpressionNode(
	pass *analysis.Pass,
	node ast.Node,
	max int,
	mode operatorMode,
	seen map[ast.Expr]bool,
) {
	limit := expressionOperatorLimit(max, mode)
	for _, expression := range expressionContexts(node) {
		checkOperatorLimit(pass, expression, seen, limit)
	}
}

func expressionOperatorLimit(max int, mode operatorMode) operatorLimit {
	return operatorLimit{
		max:     max,
		mode:    mode,
		code:    "LEG001",
		rule:    "max-expression-operators",
		message: "Expression has too many operators. Extract named values.",
	}
}

func expressionContexts(node ast.Node) []ast.Expr {
	switch typed := node.(type) {
	case *ast.ReturnStmt:
		return typed.Results
	case *ast.AssignStmt:
		return typed.Rhs
	case *ast.ValueSpec:
		return typed.Values
	case *ast.IfStmt:
		return []ast.Expr{typed.Cond}
	case *ast.ForStmt:
		return []ast.Expr{typed.Cond}
	case *ast.CallExpr:
		return typed.Args
	default:
		return nil
	}
}

func checkIfConditions(pass *analysis.Pass, max int) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			checkIfConditionNode(pass, node, max)
			return true
		})
	}
}

func checkIfConditionNode(pass *analysis.Pass, node ast.Node, max int) {
	stmt, ok := node.(*ast.IfStmt)
	if !ok {
		return
	}

	seen := make(map[ast.Expr]bool)
	limit := ifOperatorLimit(max)
	checkOperatorLimit(pass, stmt.Cond, seen, limit)
}

func ifOperatorLimit(max int) operatorLimit {
	return operatorLimit{
		max:     max,
		mode:    booleanOperators,
		code:    "LEG002",
		rule:    "hoist-if-operators",
		message: "If condition has too many boolean operators. Hoist it into a named boolean.",
	}
}

func checkComputedValues(pass *analysis.Pass, max int) {
	seen := make(map[ast.Expr]bool)
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			checkComputedNode(pass, node, max, seen)
			return true
		})
	}
}

func checkComputedNode(pass *analysis.Pass, node ast.Node, max int, seen map[ast.Expr]bool) {
	switch typed := node.(type) {
	case *ast.ReturnStmt:
		checkComputedExpressions(pass, typed.Results, max, seen)
	case *ast.CompositeLit:
		checkCompositeValues(pass, typed.Elts, max, seen)
	}
}

func checkCompositeValues(
	pass *analysis.Pass,
	expressions []ast.Expr,
	max int,
	seen map[ast.Expr]bool,
) {
	for _, expression := range expressions {
		value := compositeValue(expression)
		checkComputedExpression(pass, value, max, seen)
	}
}

func checkComputedExpressions(
	pass *analysis.Pass,
	expressions []ast.Expr,
	max int,
	seen map[ast.Expr]bool,
) {
	for _, expression := range expressions {
		checkComputedExpression(pass, expression, max, seen)
	}
}

func checkComputedExpression(
	pass *analysis.Pass,
	expression ast.Expr,
	max int,
	seen map[ast.Expr]bool,
) {
	limit := computedOperatorLimit(max)
	checkOperatorLimit(pass, expression, seen, limit)
}

func computedOperatorLimit(max int) operatorLimit {
	return operatorLimit{
		max:     max,
		mode:    computedValueOperators,
		code:    "LEG012",
		rule:    "no-computed-values",
		message: "Computed value has too many operators. Extract it into a named value.",
	}
}

func compositeValue(expression ast.Expr) ast.Expr {
	keyValue, ok := expression.(*ast.KeyValueExpr)
	if !ok {
		return expression
	}

	return keyValue.Value
}

func checkOperatorLimit(
	pass *analysis.Pass,
	expression ast.Expr,
	seen map[ast.Expr]bool,
	limit operatorLimit,
) {
	if expressionAlreadyChecked(expression, seen) {
		return
	}

	seen[expression] = true
	if !exceedsOperatorLimit(expression, limit) {
		return
	}

	report(pass, expression, limit.code, limit.rule, limit.message)
}

func expressionAlreadyChecked(expression ast.Expr, seen map[ast.Expr]bool) bool {
	if expression == nil {
		return true
	}

	return seen[expression]
}

func exceedsOperatorLimit(expression ast.Expr, limit operatorLimit) bool {
	count := countOperators(expression, limit.mode)
	return count > limit.max
}

func countOperators(expression ast.Expr, mode operatorMode) int {
	count := 0
	ast.Inspect(expression, func(node ast.Node) bool {
		if _, ok := node.(*ast.FuncLit); ok {
			return false
		}

		count += operatorWeight(node, mode)
		return true
	})

	return count
}

func operatorWeight(node ast.Node, mode operatorMode) int {
	switch typed := node.(type) {
	case *ast.BinaryExpr:
		return binaryOperatorWeight(typed.Op, mode)
	case *ast.UnaryExpr:
		return unaryOperatorWeight(typed.Op, mode)
	default:
		return 0
	}
}

func binaryOperatorWeight(operator token.Token, mode operatorMode) int {
	if mode == booleanOperators {
		return logicalOperatorWeight(operator)
	}

	if isReadabilityOperator(operator) {
		return 1
	}

	if mode != computedValueOperators {
		return 0
	}

	if isArithmeticOperator(operator) {
		return 1
	}

	return 0
}

func unaryOperatorWeight(operator token.Token, mode operatorMode) int {
	if mode == booleanOperators {
		return 0
	}

	if operator == token.NOT {
		return 1
	}

	return 0
}

func logicalOperatorWeight(operator token.Token) int {
	isLogicalOperator := operator == token.LAND || operator == token.LOR
	if isLogicalOperator {
		return 1
	}

	return 0
}

func isReadabilityOperator(operator token.Token) bool {
	isLogicalOperator := logicalOperatorWeight(operator) == 1
	if isLogicalOperator {
		return true
	}

	return isComparisonOperator(operator)
}

func isComparisonOperator(operator token.Token) bool {
	switch operator {
	case token.EQL, token.NEQ, token.LSS, token.LEQ, token.GTR, token.GEQ:
		return true
	default:
		return false
	}
}

func isArithmeticOperator(operator token.Token) bool {
	switch operator {
	case token.ADD, token.SUB, token.MUL, token.QUO, token.REM:
		return true
	default:
		return false
	}
}
