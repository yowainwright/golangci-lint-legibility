package analyzers

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func newNoTrivialWrapperFunctions() ruleSpec {
	return newAnalyzer(
		"LEG008",
		"no-trivial-wrapper-functions",
		"Avoid functions that only forward parameters to another call.",
		func(pass *analysis.Pass) (any, error) {
			checkTrivialWrappers(pass)
			return nil, nil
		},
	)
}

func checkTrivialWrappers(pass *analysis.Pass) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			decl, ok := node.(*ast.FuncDecl)
			if ok {
				checkTrivialWrapper(pass, decl)
			}

			return true
		})
	}
}

func checkTrivialWrapper(pass *analysis.Pass, decl *ast.FuncDecl) {
	call := singleReturnCall(decl.Body)
	if call == nil {
		return
	}

	params := fieldNames(decl.Type.Params)
	args := callArgNames(call)
	if !sameNames(params, args) {
		return
	}

	report(
		pass,
		decl.Name,
		"LEG008",
		"no-trivial-wrapper-functions",
		"This function only forwards its parameters to another call.",
	)
}

func singleReturnCall(body *ast.BlockStmt) *ast.CallExpr {
	if body == nil {
		return nil
	}

	if len(body.List) != 1 {
		return nil
	}

	returnStmt, ok := body.List[0].(*ast.ReturnStmt)
	if !ok {
		return nil
	}

	if len(returnStmt.Results) != 1 {
		return nil
	}

	call, _ := returnStmt.Results[0].(*ast.CallExpr)
	return call
}

func fieldNames(fields *ast.FieldList) []string {
	if fields == nil {
		return nil
	}

	names := make([]string, 0, len(fields.List))
	for _, field := range fields.List {
		names = appendFieldNames(names, field)
	}

	return names
}

func appendFieldNames(names []string, field *ast.Field) []string {
	for _, name := range field.Names {
		names = append(names, name.Name)
	}

	return names
}

func callArgNames(call *ast.CallExpr) []string {
	names := make([]string, 0, len(call.Args))
	for _, arg := range call.Args {
		identifier, ok := arg.(*ast.Ident)
		if !ok {
			return nil
		}

		names = append(names, identifier.Name)
	}

	return names
}

func sameNames(left []string, right []string) bool {
	if len(left) == 0 {
		return false
	}

	if len(left) != len(right) {
		return false
	}

	for index, name := range left {
		if right[index] != name {
			return false
		}
	}

	return true
}
