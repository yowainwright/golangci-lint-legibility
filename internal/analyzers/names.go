package analyzers

import (
	"go/ast"
	"regexp"

	"golang.org/x/tools/go/analysis"
)

var defaultNegativeConditionNamePattern = regexp.MustCompile(
	`^(is|are|was|were|has|have|had|can|could|should|will|would|did|does)(Not|No)[A-Z]`,
)

func newPreferPositiveConditionNames(settings Settings) ruleSpec {
	pattern := negativeConditionNamePattern(settings)
	return newAnalyzer(
		"LEG007",
		"prefer-positive-condition-names",
		"Prefer positive condition names.",
		func(pass *analysis.Pass) (any, error) {
			checkConditionNames(pass, pattern)
			return nil, nil
		},
	)
}

func negativeConditionNamePattern(settings Settings) *regexp.Regexp {
	if settings.NegativeConditionNamePattern == "" {
		return defaultNegativeConditionNamePattern
	}

	pattern, err := regexp.Compile(settings.NegativeConditionNamePattern)
	if err != nil {
		return defaultNegativeConditionNamePattern
	}

	return pattern
}

func checkConditionNames(pass *analysis.Pass, pattern *regexp.Regexp) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(node ast.Node) bool {
			checkNegativeNameNode(pass, node, pattern)
			return true
		})
	}
}

func checkNegativeNameNode(pass *analysis.Pass, node ast.Node, pattern *regexp.Regexp) {
	switch typed := node.(type) {
	case *ast.AssignStmt:
		checkAssignedNames(pass, typed.Lhs, pattern)
	case *ast.ValueSpec:
		checkDeclaredNames(pass, typed.Names, pattern)
	}
}

func checkAssignedNames(pass *analysis.Pass, expressions []ast.Expr, pattern *regexp.Regexp) {
	for _, expression := range expressions {
		identifier, ok := expression.(*ast.Ident)
		if ok {
			checkConditionName(pass, identifier, pattern)
		}
	}
}

func checkDeclaredNames(pass *analysis.Pass, names []*ast.Ident, pattern *regexp.Regexp) {
	for _, name := range names {
		checkConditionName(pass, name, pattern)
	}
}

func checkConditionName(pass *analysis.Pass, identifier *ast.Ident, pattern *regexp.Regexp) {
	if !pattern.MatchString(identifier.Name) {
		return
	}

	report(
		pass,
		identifier,
		"LEG007",
		"prefer-positive-condition-names",
		"Prefer a positive boolean name and negate at use sites.",
	)
}
