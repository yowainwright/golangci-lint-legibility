package analyzers

import (
	"go/ast"
	"go/token"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/tools/go/analysis"
)

var generatedCommentPattern = regexp.MustCompile(`^// Code generated .* DO NOT EDIT\.$`)

func newNoUnmatchedComments(settings Settings) ruleSpec {
	return newAnalyzer(
		"LEG039",
		"no-unmatched-comments",
		"Reject comments without an allowed matcher or boundary identifier.",
		func(pass *analysis.Pass) (any, error) {
			checkUnmatchedComments(pass, settings)
			return nil, nil
		},
	)
}

func newPreferLineComments() ruleSpec {
	return newAnalyzer(
		"LEG041",
		"prefer-line-comments",
		"Prefer line comments for ordinary multiline prose.",
		func(pass *analysis.Pass) (any, error) {
			checkMultilineBlockComments(pass)
			return nil, nil
		},
	)
}

func checkUnmatchedComments(pass *analysis.Pass, settings Settings) {
	matchers := compileCommentMatchers(settings.CommentMatchers)
	for _, file := range pass.Files {
		checkUnmatchedCommentsInFile(pass, file, matchers, settings)
	}
}

func checkUnmatchedCommentsInFile(
	pass *analysis.Pass,
	file *ast.File,
	matchers []*regexp.Regexp,
	settings Settings,
) {
	ignored := cgoCommentGroups(file)
	for _, group := range file.Comments {
		if commentGroupIgnored(file, group, ignored) {
			continue
		}

		checkAllowedCommentGroup(pass, group, matchers, settings)
	}
}

func checkAllowedCommentGroup(
	pass *analysis.Pass,
	group *ast.CommentGroup,
	matchers []*regexp.Regexp,
	settings Settings,
) {
	value := commentGroupValue(group)
	if isAllowedComment(value, matchers, settings) {
		return
	}

	report(
		pass,
		group,
		"LEG039",
		"no-unmatched-comments",
		"Comment does not match an allowed matcher or boundary identifier.",
	)
}

func checkMultilineBlockComments(pass *analysis.Pass) {
	for _, file := range pass.Files {
		checkMultilineBlockCommentsInFile(pass, file)
	}
}

func checkMultilineBlockCommentsInFile(pass *analysis.Pass, file *ast.File) {
	ignored := cgoCommentGroups(file)
	for _, group := range file.Comments {
		checkCommentGroupStyle(pass, group, ignored[group])
	}
}

func checkCommentGroupStyle(pass *analysis.Pass, group *ast.CommentGroup, ignored bool) {
	if ignored {
		return
	}

	for _, comment := range group.List {
		if isMultilineBlockComment(comment) {
			reportMultilineBlockComment(pass, comment)
		}
	}
}

func reportMultilineBlockComment(pass *analysis.Pass, comment *ast.Comment) {
	report(
		pass,
		comment,
		"LEG041",
		"prefer-line-comments",
		"Use consecutive // comments for multiline prose.",
	)
}

func isMultilineBlockComment(comment *ast.Comment) bool {
	isBlock := strings.HasPrefix(comment.Text, "/*")
	return isBlock && strings.Contains(comment.Text, "\n")
}

func commentGroupIgnored(
	file *ast.File,
	group *ast.CommentGroup,
	cgoGroups map[*ast.CommentGroup]bool,
) bool {
	if cgoGroups[group] {
		return true
	}

	return isMetadataCommentGroup(file, group)
}

func isMetadataCommentGroup(file *ast.File, group *ast.CommentGroup) bool {
	for _, comment := range group.List {
		if !isMetadataComment(file, comment) {
			return false
		}
	}

	return true
}

func isMetadataComment(file *ast.File, comment *ast.Comment) bool {
	body, isLine := strings.CutPrefix(comment.Text, "//")
	if !isLine {
		return false
	}

	isDirective := isGoDirective(body)
	isBuildConstraint := isLegacyBuildConstraint(body)
	isBareNolint := body == "nolint"
	isToolMetadata := isDirective || isBuildConstraint || isBareNolint
	if isToolMetadata {
		return true
	}

	isBeforePackage := comment.Pos() < file.Package
	return isBeforePackage && generatedCommentPattern.MatchString(comment.Text)
}

func isGoDirective(body string) bool {
	if hasLegacyDirectivePrefix(body) {
		return true
	}

	colon := strings.Index(body, ":")
	hasTool := colon > 0
	hasDirective := colon+1 < len(body)
	isWellFormed := hasTool && hasDirective
	if !isWellFormed {
		return false
	}

	return hasValidDirectiveName(body, colon)
}

func hasLegacyDirectivePrefix(body string) bool {
	prefixes := []string{"line ", "extern ", "export "}
	for _, prefix := range prefixes {
		if strings.HasPrefix(body, prefix) {
			return true
		}
	}

	return false
}

func hasValidDirectiveName(body string, colon int) bool {
	for index := 0; index <= colon+1; index++ {
		if index == colon {
			continue
		}

		if !isLowerAlphaNumeric(body[index]) {
			return false
		}
	}

	return true
}

func isLowerAlphaNumeric(character byte) bool {
	isLower := character >= 'a' && character <= 'z'
	isDigit := character >= '0' && character <= '9'
	return isLower || isDigit
}

func isLegacyBuildConstraint(body string) bool {
	return strings.HasPrefix(body, " +build ")
}

func cgoCommentGroups(file *ast.File) map[*ast.CommentGroup]bool {
	groups := make(map[*ast.CommentGroup]bool)
	for _, declaration := range file.Decls {
		imports, ok := declaration.(*ast.GenDecl)
		isImportDeclaration := ok && imports.Tok == token.IMPORT
		if !isImportDeclaration {
			continue
		}

		addCgoCommentGroups(imports, groups)
	}

	return groups
}

func addCgoCommentGroups(declaration *ast.GenDecl, groups map[*ast.CommentGroup]bool) {
	for _, spec := range declaration.Specs {
		importSpec, ok := spec.(*ast.ImportSpec)
		isCgoSpec := ok && isCgoImport(importSpec)
		if !isCgoSpec {
			continue
		}

		if importSpec.Doc != nil {
			groups[importSpec.Doc] = true
		}
		isSingleImport := len(declaration.Specs) == 1
		hasDeclarationDoc := declaration.Doc != nil
		isDocumentedSingleImport := isSingleImport && hasDeclarationDoc
		if isDocumentedSingleImport {
			groups[declaration.Doc] = true
		}
	}
}

func isCgoImport(spec *ast.ImportSpec) bool {
	path, err := strconv.Unquote(spec.Path.Value)
	if err != nil {
		return false
	}

	return path == "C"
}

func commentGroupValue(group *ast.CommentGroup) string {
	values := make([]string, 0, len(group.List))
	for _, comment := range group.List {
		values = append(values, commentValue(comment.Text))
	}

	return strings.Join(values, "\n")
}

func commentValue(text string) string {
	if value, ok := strings.CutPrefix(text, "//"); ok {
		return value
	}

	value := strings.TrimPrefix(text, "/*")
	return strings.TrimSuffix(value, "*/")
}

func compileCommentMatchers(sources []string) []*regexp.Regexp {
	matchers := make([]*regexp.Regexp, 0, len(sources))
	for _, source := range sources {
		matcher, err := regexp.Compile("(?i)" + source)
		if err == nil {
			matchers = append(matchers, matcher)
		}
	}

	return matchers
}

func isAllowedComment(value string, matchers []*regexp.Regexp, settings Settings) bool {
	for _, matcher := range matchers {
		if matcher.MatchString(value) {
			return true
		}
	}

	normalized := normalizeCommentBoundary(value)
	if matchesPrefixIdentifier(normalized, settings.CommentPrefixIdentifiers) {
		return true
	}

	return matchesSuffixIdentifier(normalized, settings.CommentSuffixIdentifiers)
}

func normalizeCommentBoundary(value string) string {
	lines := strings.Split(value, "\n")
	for index, line := range lines {
		lines[index] = normalizeCommentLine(line)
	}

	return strings.ToLower(strings.TrimSpace(strings.Join(lines, "\n")))
}

func normalizeCommentLine(line string) string {
	normalized := strings.TrimLeftFunc(line, unicode.IsSpace)
	withoutStar, hasStar := strings.CutPrefix(normalized, "*")
	if !hasStar {
		return line
	}

	return strings.TrimPrefix(withoutStar, " ")
}

func matchesPrefixIdentifier(value string, identifiers []string) bool {
	for _, identifier := range identifiers {
		normalized := strings.ToLower(strings.TrimSpace(identifier))
		if normalized == "" {
			continue
		}

		if !strings.HasPrefix(value, normalized) {
			continue
		}

		if hasIdentifierBoundaryAfter(value, len(normalized)) {
			return true
		}
	}

	return false
}

func matchesSuffixIdentifier(value string, identifiers []string) bool {
	for _, identifier := range identifiers {
		normalized := strings.ToLower(strings.TrimSpace(identifier))
		if normalized == "" {
			continue
		}

		if !strings.HasSuffix(value, normalized) {
			continue
		}

		start := len(value) - len(normalized)
		if hasIdentifierBoundaryBefore(value, start) {
			return true
		}
	}

	return false
}

func hasIdentifierBoundaryAfter(value string, index int) bool {
	if index >= len(value) {
		return true
	}

	character, _ := utf8.DecodeRuneInString(value[index:])
	return !isIdentifierCharacter(character)
}

func hasIdentifierBoundaryBefore(value string, index int) bool {
	if index <= 0 {
		return true
	}

	character, _ := utf8.DecodeLastRuneInString(value[:index])
	return !isIdentifierCharacter(character)
}

func isIdentifierCharacter(character rune) bool {
	isUnderscore := character == '_'
	isLetter := unicode.IsLetter(character)
	isDigit := unicode.IsDigit(character)
	isIdentifier := isUnderscore || isLetter || isDigit
	return isIdentifier
}
