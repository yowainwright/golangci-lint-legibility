package analyzers

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/analysis"
)

func newPreferSwitchOverLongIfChain(settings Settings) ruleSpec {
	min := settings.minSwitchChainLength()
	return newAnalyzer(
		"LEG034",
		"prefer-switch-over-long-if-chain",
		"Prefer switch over long if chains that compare the same value.",
		func(pass *analysis.Pass) (any, error) {
			checkSwitchableIfChains(pass, min)
			return nil, nil
		},
	)
}

func newNoComplexIfInit(settings Settings) ruleSpec {
	max := settings.maxIfInitOperators()
	return newAnalyzer(
		"LEG036",
		"no-complex-if-init",
		"Avoid combining an if initializer with an operator-heavy condition.",
		func(pass *analysis.Pass) (any, error) {
			checkComplexIfInit(pass, max)
			return nil, nil
		},
	)
}

func checkSwitchableIfChains(pass *analysis.Pass, min int) {
	parents := buildParentMap(pass.Files)
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			stmt, ok := node.(*ast.IfStmt)
			if ok {
				checkSwitchableIfChain(pass, parents, stmt, min)
			}

			return true
		})
	}
}

func checkSwitchableIfChain(
	pass *analysis.Pass,
	parents map[ast.Node]ast.Node,
	stmt *ast.IfStmt,
	min int,
) {
	if isElseIf(parents, stmt) {
		return
	}

	subjects := collectIfChainSubjects(pass, stmt)
	if len(subjects) < min {
		return
	}

	if !sameSubject(subjects) {
		return
	}

	report(
		pass,
		stmt,
		"LEG034",
		"prefer-switch-over-long-if-chain",
		"Use a switch for this repeated comparison chain.",
	)
}

func isElseIf(parents map[ast.Node]ast.Node, stmt *ast.IfStmt) bool {
	parent, ok := parents[stmt].(*ast.IfStmt)
	if !ok {
		return false
	}

	return parent.Else == stmt
}

func collectIfChainSubjects(pass *analysis.Pass, stmt *ast.IfStmt) []string {
	var subjects []string
	current := stmt
	for current != nil {
		subject, ok := equalitySubject(pass, current.Cond)
		if !ok {
			return subjects
		}

		subjects = append(subjects, subject)
		current, _ = current.Else.(*ast.IfStmt)
	}

	return subjects
}

func equalitySubject(pass *analysis.Pass, expression ast.Expr) (string, bool) {
	binary, ok := expression.(*ast.BinaryExpr)
	if !ok {
		return "", false
	}

	if binary.Op != token.EQL {
		return "", false
	}

	if isLiteral(binary.Y) {
		return nodeText(pass, binary.X), true
	}

	if isLiteral(binary.X) {
		return nodeText(pass, binary.Y), true
	}

	return "", false
}

func sameSubject(subjects []string) bool {
	if len(subjects) == 0 {
		return false
	}

	first := subjects[0]
	for _, subject := range subjects {
		if subject != first {
			return false
		}
	}

	return true
}

func checkComplexIfInit(pass *analysis.Pass, max int) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			stmt, ok := node.(*ast.IfStmt)
			if ok {
				checkIfInitCondition(pass, stmt, max)
			}

			return true
		})
	}
}

func checkIfInitCondition(pass *analysis.Pass, stmt *ast.IfStmt, max int) {
	if stmt.Init == nil {
		return
	}

	operatorCount := countOperators(stmt.Cond, booleanOperators)
	if operatorCount <= max {
		return
	}

	report(
		pass,
		stmt.Cond,
		"LEG036",
		"no-complex-if-init",
		"Move the condition into a named boolean before this if.",
	)
}
