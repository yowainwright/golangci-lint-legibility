package analyzers

import "testing"

func TestNoGetterPrefixReportsGetAccessor(t *testing.T) {
	source := readTestSource(t, "no_getter_prefix.go")

	diagnostics := runAnalyzer(t, analyzerByRule(t, "no-getter-prefix"), "p.go", source)
	requireDiagnostic(t, diagnostics, "LEG042 no-getter-prefix")
}
