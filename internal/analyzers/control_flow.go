package analyzers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func newMaxControlFlowDepth(settings Settings) ruleSpec {
	max := settings.maxControlFlowDepth()
	return newAnalyzer(
		"LEG003",
		"max-control-flow-depth",
		"Limit nested control-flow depth.",
		func(pass *analysis.Pass) (any, error) {
			checkControlFlowDepth(pass, max)
			return nil, nil
		},
	)
}

func checkControlFlowDepth(pass *analysis.Pass, max int) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			body := functionBody(node)
			if body == nil {
				return true
			}

			walkBlockDepth(pass, body, 0, max)
			return false
		})
	}
}

func walkBlockDepth(pass *analysis.Pass, block *ast.BlockStmt, depth int, max int) {
	for _, statement := range block.List {
		walkStatementDepth(pass, statement, depth, max)
	}
}

func walkStatementDepth(pass *analysis.Pass, statement ast.Stmt, depth int, max int) {
	switch typed := statement.(type) {
	case *ast.IfStmt:
		walkIfDepth(pass, typed, depth, max)
	case *ast.ForStmt:
		walkNestedBlock(pass, typed, typed.Body, depth, max)
	case *ast.RangeStmt:
		walkNestedBlock(pass, typed, typed.Body, depth, max)
	case *ast.SwitchStmt:
		walkSwitchDepth(pass, typed, depth, max)
	case *ast.TypeSwitchStmt:
		walkTypeSwitchDepth(pass, typed, depth, max)
	case *ast.SelectStmt:
		walkSelectDepth(pass, typed, depth, max)
	}
}

func walkIfDepth(pass *analysis.Pass, stmt *ast.IfStmt, depth int, max int) {
	nextDepth := reportDepth(pass, stmt, depth, max)
	walkBlockDepth(pass, stmt.Body, nextDepth, max)

	switch elseNode := stmt.Else.(type) {
	case *ast.BlockStmt:
		walkBlockDepth(pass, elseNode, nextDepth, max)
	case *ast.IfStmt:
		walkIfDepth(pass, elseNode, depth, max)
	}
}

func walkNestedBlock(pass *analysis.Pass, node ast.Node, block *ast.BlockStmt, depth int, max int) {
	nextDepth := reportDepth(pass, node, depth, max)
	walkBlockDepth(pass, block, nextDepth, max)
}

func walkSwitchDepth(pass *analysis.Pass, stmt *ast.SwitchStmt, depth int, max int) {
	nextDepth := reportDepth(pass, stmt, depth, max)
	walkCaseClauses(pass, stmt.Body.List, nextDepth, max)
}

func walkTypeSwitchDepth(pass *analysis.Pass, stmt *ast.TypeSwitchStmt, depth int, max int) {
	nextDepth := reportDepth(pass, stmt, depth, max)
	walkCaseClauses(pass, stmt.Body.List, nextDepth, max)
}

func walkSelectDepth(pass *analysis.Pass, stmt *ast.SelectStmt, depth int, max int) {
	nextDepth := reportDepth(pass, stmt, depth, max)
	walkCaseClauses(pass, stmt.Body.List, nextDepth, max)
}

func walkCaseClauses(pass *analysis.Pass, clauses []ast.Stmt, depth int, max int) {
	for _, clause := range clauses {
		body, ok := caseClauseBody(clause)
		if ok {
			walkStatementsDepth(pass, body, depth, max)
		}
	}
}

func walkStatementsDepth(pass *analysis.Pass, statements []ast.Stmt, depth int, max int) {
	for _, statement := range statements {
		walkStatementDepth(pass, statement, depth, max)
	}
}

func reportDepth(pass *analysis.Pass, node ast.Node, depth int, max int) int {
	nextDepth := depth + 1
	if nextDepth > max {
		report(
			pass,
			node,
			"LEG003",
			"max-control-flow-depth",
			"Control-flow depth is too high. Return early or extract a helper.",
		)
	}

	return nextDepth
}

func functionBody(node ast.Node) *ast.BlockStmt {
	switch typed := node.(type) {
	case *ast.FuncDecl:
		return typed.Body
	case *ast.FuncLit:
		return typed.Body
	default:
		return nil
	}
}

func caseClauseBody(statement ast.Stmt) ([]ast.Stmt, bool) {
	switch typed := statement.(type) {
	case *ast.CaseClause:
		return typed.Body, true
	case *ast.CommClause:
		return typed.Body, true
	default:
		return nil, false
	}
}
