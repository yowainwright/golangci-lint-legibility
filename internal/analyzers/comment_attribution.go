package analyzers

import (
	"go/ast"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

var defaultAutomatedCommentIdentifiers = []string{
	"ai",
	"chatgpt",
	"claude",
	"codex",
	"copilot",
	"gemini",
	"gpt",
	"llm",
	"openai",
}

var commentAuthorPattern = regexp.MustCompile(`(?i)(?:^|\s)@author\b\s*:?\s*(\S.*?)\s*$`)

func newNoAutomatedCommentAttribution(settings Settings) ruleSpec {
	return newAnalyzer(
		"LEG040",
		"no-automated-comment-attribution",
		"Reject explicit automated authorship signatures in comments.",
		func(pass *analysis.Pass) (any, error) {
			checkAutomatedCommentAttribution(pass, settings.automatedCommentIdentifiers())
			return nil, nil
		},
	)
}

func checkAutomatedCommentAttribution(pass *analysis.Pass, identifiers []string) {
	for _, file := range pass.Files {
		checkAutomatedAttributionInFile(pass, file, identifiers)
	}
}

func checkAutomatedAttributionInFile(
	pass *analysis.Pass,
	file *ast.File,
	identifiers []string,
) {
	ignored := cgoCommentGroups(file)
	for _, group := range file.Comments {
		if commentGroupIgnored(file, group, ignored) {
			continue
		}

		identifier := findProhibitedAttribution(commentGroupValue(group), identifiers)
		if identifier != "" {
			reportAutomatedAttribution(pass, group, identifier)
		}
	}
}

func reportAutomatedAttribution(
	pass *analysis.Pass,
	group *ast.CommentGroup,
	identifier string,
) {
	message := "Comment contains the prohibited attribution \"" + identifier + "\"."
	report(pass, group, "LEG040", "no-automated-comment-attribution", message)
}

func findProhibitedAttribution(value string, identifiers []string) string {
	for _, author := range commentAuthorValues(value) {
		if identifier := findAutomatedIdentifier(author, identifiers); identifier != "" {
			return identifier
		}
	}

	for _, identifier := range identifiers {
		if hasGenerationSignature(value, identifier) {
			return identifier
		}
	}

	return ""
}

func commentAuthorValues(value string) []string {
	values := []string{}
	for _, line := range strings.Split(value, "\n") {
		matches := commentAuthorPattern.FindStringSubmatch(normalizeCommentLine(line))
		if len(matches) > 1 {
			values = append(values, matches[1])
		}
	}

	return values
}

func findAutomatedIdentifier(value string, identifiers []string) string {
	normalizedValue := normalizeAttributionText(value)
	for _, identifier := range identifiers {
		normalizedIdentifier := normalizeAttributionText(identifier)
		if normalizedIdentifier == "" {
			continue
		}

		if containsWholePhrase(normalizedValue, normalizedIdentifier) {
			return identifier
		}
	}

	return ""
}

func hasGenerationSignature(value string, identifier string) bool {
	normalizedValue := normalizeAttributionText(value)
	normalizedIdentifier := normalizeAttributionText(identifier)
	if normalizedIdentifier == "" {
		return false
	}

	verbs := []string{"authored", "created", "generated", "produced", "written"}
	for _, verb := range verbs {
		if hasGenerationVerbSignature(normalizedValue, normalizedIdentifier, verb) {
			return true
		}
	}

	return false
}

func hasGenerationVerbSignature(value string, identifier string, verb string) bool {
	leadingPhrase := identifier + " " + verb
	if containsWholePhrase(value, leadingPhrase) {
		return true
	}

	return hasPassiveGenerationSignature(value, identifier, verb)
}

func hasPassiveGenerationSignature(value string, identifier string, verb string) bool {
	if !containsTrailingPhrase(value, identifier) {
		return false
	}

	prefix := strings.TrimSpace(strings.TrimSuffix(value, identifier))
	verbPhrases := []string{verb + " by", verb + " by a", verb + " by an"}
	for _, phrase := range verbPhrases {
		if containsTrailingPhrase(prefix, phrase) {
			return true
		}
	}

	return false
}

func normalizeAttributionText(value string) string {
	normalized := strings.Map(func(character rune) rune {
		isLetter := unicode.IsLetter(character)
		isDigit := unicode.IsDigit(character)
		isAlphaNumeric := isLetter || isDigit
		if isAlphaNumeric {
			return unicode.ToLower(character)
		}

		return ' '
	}, value)

	return strings.Join(strings.Fields(normalized), " ")
}

func containsWholePhrase(value string, phrase string) bool {
	paddedValue := " " + value + " "
	paddedPhrase := " " + phrase + " "
	return strings.Contains(paddedValue, paddedPhrase)
}

func containsTrailingPhrase(value string, phrase string) bool {
	paddedValue := " " + value
	paddedPhrase := " " + phrase
	return strings.HasSuffix(paddedValue, paddedPhrase)
}
