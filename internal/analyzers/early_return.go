package analyzers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func newPreferEarlyReturn() ruleSpec {
	return newAnalyzer(
		"LEG009",
		"prefer-early-return",
		"Avoid else branches after a branch already exits.",
		func(pass *analysis.Pass) (any, error) {
			checkEarlyReturn(pass)
			return nil, nil
		},
	)
}

func newPreferGuardClauses() ruleSpec {
	return newAnalyzer(
		"LEG010",
		"prefer-guard-clauses",
		"Prefer guard clauses over wrapping the main path in one large if block.",
		func(pass *analysis.Pass) (any, error) {
			checkGuardClauses(pass)
			return nil, nil
		},
	)
}

func checkEarlyReturn(pass *analysis.Pass) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			stmt, ok := node.(*ast.IfStmt)
			if ok {
				checkIfEarlyReturn(pass, stmt)
			}

			return true
		})
	}
}

func checkIfEarlyReturn(pass *analysis.Pass, stmt *ast.IfStmt) {
	if stmt.Else == nil {
		return
	}

	if !blockEndsTerminal(stmt.Body) {
		return
	}

	report(pass, stmt.Else, "LEG009", "prefer-early-return", "Avoid else after an if branch exits.")
}

func checkGuardClauses(pass *analysis.Pass) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			body := functionBody(node)
			if body != nil {
				checkFunctionGuardClause(pass, body)
			}

			return true
		})
	}
}

func checkFunctionGuardClause(pass *analysis.Pass, body *ast.BlockStmt) {
	if len(body.List) != 1 {
		return
	}

	stmt, ok := body.List[0].(*ast.IfStmt)
	if !ok {
		return
	}

	if stmt.Else != nil {
		return
	}

	if len(stmt.Body.List) < 2 {
		return
	}

	reportGuardClause(pass, stmt)
}

func reportGuardClause(pass *analysis.Pass, stmt *ast.IfStmt) {
	report(
		pass,
		stmt,
		"LEG010",
		"prefer-guard-clauses",
		"Invert this condition and return early before the main path.",
	)
}

func blockEndsTerminal(block *ast.BlockStmt) bool {
	if block == nil {
		return false
	}

	hasStatements := len(block.List) > 0
	if !hasStatements {
		return false
	}

	return isTerminalStatement(block.List[len(block.List)-1])
}

func isTerminalStatement(statement ast.Stmt) bool {
	switch typed := statement.(type) {
	case *ast.ReturnStmt:
		return true
	case *ast.BranchStmt:
		return typed.Tok.String() != "fallthrough"
	case *ast.ExprStmt:
		return isPanicCall(typed.X)
	default:
		return false
	}
}

func isPanicCall(expression ast.Expr) bool {
	call, ok := expression.(*ast.CallExpr)
	if !ok {
		return false
	}

	identifier, ok := call.Fun.(*ast.Ident)
	if !ok {
		return false
	}

	return identifier.Name == "panic"
}
