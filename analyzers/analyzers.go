package analyzers

import (
	"strings"

	"golang.org/x/tools/go/analysis"
)

const (
	defaultRuleEnabled = true
	optInRuleEnabled   = false
)

type ruleSpec struct {
	code           string
	name           string
	summary        string
	defaultEnabled bool
	analyzer       *analysis.Analyzer
}

func New(settings Settings) []*analysis.Analyzer {
	specs := []ruleSpec{
		newMaxExpressionOperators(settings),
		newHoistIfOperators(settings),
		newMaxControlFlowDepth(settings),
		newNoQuadraticPatterns(),
		newNoRedundantBooleanLogic(),
		newPreferPositiveConditionNames(settings),
		newNoTrivialWrapperFunctions(),
		newPreferEarlyReturn(),
		newPreferGuardClauses(),
		newMaxArrayChainDepth(settings),
		newNoComputedValues(settings),
		newPreferObjectLookup(settings),
		newRequireFilenameMatchesDirname(settings),
		newNoMixedFilenameCasing(),
		newNoDeepSelectorChain(settings),
		newPreferSwitchOverLongIfChain(settings),
		newNoBoolLiteralArgs(),
		newNoComplexIfInit(settings),
		newNoDeepCompositeLiteralArg(settings),
	}

	analyzers := make([]*analysis.Analyzer, 0, len(specs))
	for _, spec := range specs {
		if settings.RuleEnabled(spec.code, spec.name, spec.defaultEnabled) {
			analyzers = append(analyzers, spec.analyzer)
		}
	}

	return analyzers
}

func analysisName(ruleName string) string {
	identifier := strings.ReplaceAll(ruleName, "-", "_")
	return "legibility_" + identifier
}

func newAnalyzer(
	code string,
	name string,
	summary string,
	run func(*analysis.Pass) (any, error),
) ruleSpec {
	return newRuleSpec(code, name, summary, defaultRuleEnabled, run)
}

func newOptionalAnalyzer(
	code string,
	name string,
	summary string,
	run func(*analysis.Pass) (any, error),
) ruleSpec {
	return newRuleSpec(code, name, summary, optInRuleEnabled, run)
}

func newRuleSpec(
	code string,
	name string,
	summary string,
	defaultEnabled bool,
	run func(*analysis.Pass) (any, error),
) ruleSpec {
	analyzer := &analysis.Analyzer{
		Name: analysisName(name),
		Doc:  summary,
		Run:  run,
	}

	return ruleSpec{
		code:           code,
		name:           name,
		summary:        summary,
		defaultEnabled: defaultEnabled,
		analyzer:       analyzer,
	}
}
