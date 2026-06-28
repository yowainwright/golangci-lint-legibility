package analyzers

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strings"
	"testing"

	"golang.org/x/tools/go/analysis"
)

func TestMaxExpressionOperatorsReportsComplexExpression(t *testing.T) {
	source := `package p
func value(a, b, c, d, e, f bool) bool {
	return a && b && c && d && e && f
}`

	diagnostics := runAnalyzer(t, analyzerByRule(t, "max-expression-operators"), "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG001 max-expression-operators")
}

func TestPreferEarlyReturnReportsElseAfterReturn(t *testing.T) {
	source := `package p
func value(ok bool) int {
	if !ok {
		return 0
	} else {
		return 1
	}
}`

	diagnostics := runAnalyzer(t, analyzerByRule(t, "prefer-early-return"), "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG009 prefer-early-return")
}

func TestPreferObjectLookupReportsLongEqualityChain(t *testing.T) {
	source := `package p
func value(role string) bool {
	return role == "admin" || role == "owner" || role == "staff"
}`

	diagnostics := runAnalyzer(t, analyzerByRule(t, "prefer-object-lookup"), "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG024 prefer-object-lookup")
}

func TestNoDeepSelectorChainReportsDeepChain(t *testing.T) {
	source := `package p
func value(config Config) bool {
	return config.User.Profile.Settings.Email.Enabled
}`

	diagnostics := runAnalyzer(t, analyzerByRule(t, "no-deep-selector-chain"), "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG031 no-deep-selector-chain")
}

func TestPreferSwitchOverLongIfChainReportsRepeatedComparison(t *testing.T) {
	source := `package p
func value(status string) int {
	if status == "new" {
		return 1
	} else if status == "active" {
		return 2
	} else if status == "closed" {
		return 3
	}
	return 0
}`

	analyzer := analyzerByRule(t, "prefer-switch-over-long-if-chain")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG034 prefer-switch-over-long-if-chain")
}

func TestNoBoolLiteralArgsReportsBooleanArgument(t *testing.T) {
	source := `package p
func value() {
	configure("api", true)
}`

	diagnostics := runAnalyzer(t, analyzerByRule(t, "no-bool-literal-args"), "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG035 no-bool-literal-args")
}

func TestNoComplexIfInitReportsOperatorHeavyCondition(t *testing.T) {
	source := `package p
func value(users map[string]User, id string) {
	if user, ok := users[id]; ok && user.Active {
		save(user)
	}
}`

	diagnostics := runAnalyzer(t, analyzerByRule(t, "no-complex-if-init"), "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG036 no-complex-if-init")
}

func TestNoDeepCompositeLiteralArgReportsNestedLiteral(t *testing.T) {
	source := `package p
func value() {
	save(Config{HTTP: HTTPConfig{Timeout: 10}})
}`

	analyzer := analyzerByRule(t, "no-deep-composite-literal-arg")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG037 no-deep-composite-literal-arg")
}

func TestRequireFilenameMatchesDirnameReportsWhenEnabled(t *testing.T) {
	minDepth := 2
	settings := Settings{
		EnabledRules:         []string{"require-filename-matches-dirname"},
		MinDirnameMatchDepth: &minDepth,
	}
	source := `package p
func value() {}
`

	analyzer := analyzerByRuleWithSettings(t, "require-filename-matches-dirname", settings)
	diagnostics := runAnalyzer(t, analyzer, "internal/orders/service.go", source)
	requireDiagnostic(t, diagnostics, "LEG025 require-filename-matches-dirname")
}

func TestSettingsDisableRule(t *testing.T) {
	settings := Settings{DisabledRules: []string{"prefer-early-return"}}
	analyzers := New(settings)

	for _, analyzer := range analyzers {
		if analyzer.Name == analysisName("prefer-early-return") {
			t.Fatal("prefer-early-return should be disabled")
		}
	}
}

func TestRequireFilenameMatchesDirnameIsOptIn(t *testing.T) {
	for _, analyzer := range New(Settings{}) {
		if analyzer.Name == analysisName("require-filename-matches-dirname") {
			t.Fatal("require-filename-matches-dirname should be opt-in")
		}
	}
}

func TestAllAnalyzersUseSyntaxOnlyInputs(t *testing.T) {
	settings := Settings{EnabledRules: []string{"all"}}
	for _, analyzer := range New(settings) {
		if len(analyzer.Requires) != 0 {
			t.Fatalf("%s should not require another analyzer", analyzer.Name)
		}
	}
}

func analyzerByRule(t *testing.T, name string) *analysis.Analyzer {
	t.Helper()

	for _, analyzer := range New(Settings{}) {
		if analyzer.Name == analysisName(name) {
			return analyzer
		}
	}

	t.Fatalf("missing analyzer for %s", name)
	return nil
}

func analyzerByRuleWithSettings(t *testing.T, name string, settings Settings) *analysis.Analyzer {
	t.Helper()

	for _, analyzer := range New(settings) {
		if analyzer.Name == analysisName(name) {
			return analyzer
		}
	}

	t.Fatalf("missing analyzer for %s", name)
	return nil
}

func runAnalyzer(
	t *testing.T,
	analyzer *analysis.Analyzer,
	filename string,
	source string,
) []analysis.Diagnostic {
	t.Helper()

	fileSet := token.NewFileSet()
	file, err := parser.ParseFile(fileSet, filename, source, parser.ParseComments)
	if err != nil {
		t.Fatal(err)
	}

	var diagnostics []analysis.Diagnostic
	pass := &analysis.Pass{
		Fset:  fileSet,
		Files: []*ast.File{file},
		Report: func(diagnostic analysis.Diagnostic) {
			diagnostics = append(diagnostics, diagnostic)
		},
	}

	_, err = analyzer.Run(pass)
	if err != nil {
		t.Fatal(err)
	}

	return diagnostics
}

func requireDiagnostic(t *testing.T, diagnostics []analysis.Diagnostic, text string) {
	t.Helper()

	for _, diagnostic := range diagnostics {
		if strings.Contains(diagnostic.Message, text) {
			return
		}
	}

	t.Fatalf("missing diagnostic %q in %#v", text, diagnostics)
}
