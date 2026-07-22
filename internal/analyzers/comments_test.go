package analyzers

import (
	"testing"

	"golang.org/x/tools/go/analysis"
)

func TestNoUnmatchedCommentsReportsContiguousCommentOnce(t *testing.T) {
	source := readTestSource(t, "unmatched_comments.go")

	analyzer := explicitNoUnmatchedCommentsAnalyzer(t)
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 1)
	requireDiagnostic(t, diagnostics, "LEG039 no-unmatched-comments")
}

func TestNoUnmatchedCommentsAcceptsConfiguredAllowPaths(t *testing.T) {
	settings := Settings{
		CommentMatchers:          []string{`\beng-\d+\b`},
		CommentPrefixIdentifiers: []string{"human"},
		CommentSuffixIdentifiers: []string{"@OWNED"},
	}

	analyzer := analyzerByRuleWithSettings(t, "no-unmatched-comments", settings)
	source := readTestSource(t, "allowed_comments.go")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 0)
}

func TestNoUnmatchedCommentsRequiresIdentifierBoundaries(t *testing.T) {
	settings := Settings{
		CommentPrefixIdentifiers: []string{"HUMAN"},
		CommentSuffixIdentifiers: []string{"@owned"},
	}
	source := readTestSource(t, "identifier_boundaries.go")

	analyzer := analyzerByRuleWithSettings(t, "no-unmatched-comments", settings)
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 2)
}

func TestNoUnmatchedCommentsIgnoresGoMetadata(t *testing.T) {
	source := readTestSource(t, "go_metadata.go")

	analyzer := explicitNoUnmatchedCommentsAnalyzer(t)
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 0)
}

func TestNoUnmatchedCommentsChecksMalformedBlockLineDirective(t *testing.T) {
	source := readTestSource(t, "malformed_block_line_directive.go")

	analyzer := explicitNoUnmatchedCommentsAnalyzer(t)
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 1)
}

func TestNoUnmatchedCommentsChecksUnknownToolSyntax(t *testing.T) {
	source := readTestSource(t, "unknown_tool_syntax.go")

	analyzer := explicitNoUnmatchedCommentsAnalyzer(t)
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 1)
}

func TestNoUnmatchedCommentsChecksMixedMetadataGroups(t *testing.T) {
	source := readTestSource(t, "mixed_metadata.go")

	analyzer := explicitNoUnmatchedCommentsAnalyzer(t)
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 1)
}

func TestNoUnmatchedCommentsChecksMisplacedGeneratedMarker(t *testing.T) {
	source := readTestSource(t, "misplaced_generated_marker.go")

	analyzer := explicitNoUnmatchedCommentsAnalyzer(t)
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 1)
}

func TestNoUnmatchedCommentsChecksMisplacedCgoMarker(t *testing.T) {
	source := readTestSource(t, "misplaced_cgo_generated_marker.go")

	analyzer := explicitNoUnmatchedCommentsAnalyzer(t)
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 1)
}

func TestNoUnmatchedCommentsIgnoresCgoPreamble(t *testing.T) {
	source := readTestSource(t, "cgo_preamble.go")

	analyzer := explicitNoUnmatchedCommentsAnalyzer(t)
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 0)
}

func TestNoUnmatchedCommentsIgnoresCgoGeneratedUnsafeImport(t *testing.T) {
	source := readTestSource(t, "cgo_generated_unsafe.go")

	analyzer := explicitNoUnmatchedCommentsAnalyzer(t)
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 0)
}

func TestNoUnmatchedCommentsChecksDocumentedBlankUnsafeImport(t *testing.T) {
	source := readTestSource(t, "blank_unsafe_comment.go")

	analyzer := explicitNoUnmatchedCommentsAnalyzer(t)
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 1)
}

func TestNoUnmatchedCommentsIgnoresInvalidMatchers(t *testing.T) {
	matchers := []string{"["}
	settings := Settings{CommentMatchers: matchers}
	source := readTestSource(t, "unmatched_comments.go")

	analyzer := analyzerByRuleWithSettings(t, "no-unmatched-comments", settings)
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 1)
}

func TestNoAutomatedCommentAttributionReportsSignatures(t *testing.T) {
	source := readTestSource(t, "automated_signatures.go")
	analyzer := analyzerByRule(t, "no-automated-comment-attribution")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 2)
	requireDiagnostic(t, diagnostics, "LEG040 no-automated-comment-attribution")
}

func TestNoAutomatedCommentAttributionIgnoresTechnologyReferences(t *testing.T) {
	source := readTestSource(t, "technology_references.go")

	analyzer := analyzerByRule(t, "no-automated-comment-attribution")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 0)
}

func TestNoAutomatedCommentAttributionIgnoresMidPhraseIdentifiers(t *testing.T) {
	source := readTestSource(t, "mid_phrase_identifiers.go")

	analyzer := analyzerByRule(t, "no-automated-comment-attribution")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 0)
}

func TestNoAutomatedCommentAttributionReportsArticleSignature(t *testing.T) {
	source := readTestSource(t, "article_signature.go")

	analyzer := analyzerByRule(t, "no-automated-comment-attribution")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 1)
}

func TestNoAutomatedCommentAttributionIgnoresMetadataAndCgo(t *testing.T) {
	identifier := "robot"
	settings := Settings{AutomatedCommentIdentifiers: []string{identifier}}
	source := readTestSource(t, "metadata_and_cgo.go")

	analyzer := analyzerByRuleWithSettings(t, "no-automated-comment-attribution", settings)
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 0)
}

func TestNoAutomatedCommentAttributionSupportsCustomIdentifiers(t *testing.T) {
	identifiers := []string{"robot"}
	settings := Settings{AutomatedCommentIdentifiers: identifiers}
	source := readTestSource(t, "custom_identifier.go")

	analyzer := analyzerByRuleWithSettings(t, "no-automated-comment-attribution", settings)
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 1)
}

func TestNoAutomatedCommentAttributionAllowsEmptyIdentifiers(t *testing.T) {
	identifiers := []string{}
	settings := Settings{AutomatedCommentIdentifiers: identifiers}
	source := readTestSource(t, "empty_identifiers.go")

	analyzer := analyzerByRuleWithSettings(t, "no-automated-comment-attribution", settings)
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 0)
}

func TestPreferLineCommentsReportsMultilineBlockComment(t *testing.T) {
	source := readTestSource(t, "multiline_block_comment.go")

	analyzer := analyzerByRule(t, "prefer-line-comments")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 1)
	requireDiagnostic(t, diagnostics, "LEG041 prefer-line-comments")
}

func TestPreferLineCommentsReportsMalformedMultilineLineDirective(t *testing.T) {
	source := readTestSource(t, "malformed_multiline_line_directive.txt")

	analyzer := analyzerByRule(t, "prefer-line-comments")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 1)
}

func TestPreferLineCommentsAllowsSingleLineBlocksAndCgo(t *testing.T) {
	source := readTestSource(t, "line_comment_exceptions.go")

	analyzer := analyzerByRule(t, "prefer-line-comments")
	diagnostics := runAnalyzer(t, analyzer, "p.go", source)
	requireDiagnosticsCount(t, diagnostics, 0)
}

func TestNormalizeCommentLineTrimsLeadingWhitespace(t *testing.T) {
	got := normalizeCommentLine("  HUMAN: Preserve this.")
	want := "HUMAN: Preserve this."
	if got != want {
		t.Fatalf("normalizeCommentLine() = %q, want %q", got, want)
	}
}

func explicitNoUnmatchedCommentsAnalyzer(t *testing.T) *analysis.Analyzer {
	t.Helper()

	settings := Settings{EnabledRules: []string{"no-unmatched-comments"}}
	return analyzerByRuleWithSettings(t, "no-unmatched-comments", settings)
}
