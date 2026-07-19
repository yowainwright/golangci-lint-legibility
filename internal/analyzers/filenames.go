package analyzers

import (
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var (
	uppercasePattern = regexp.MustCompile(`[A-Z]`)
	lowercasePattern = regexp.MustCompile(`[a-z]`)
)

var dirnameMatchStandaloneFiles = map[string]bool{
	"constants": true,
	"doc":       true,
	"errors":    true,
	"generate":  true,
	"main":      true,
	"mock":      true,
	"mocks":     true,
	"tools":     true,
	"types":     true,
}

func newRequireFilenameMatchesDirname(settings Settings) ruleSpec {
	min := settings.minDirnameMatchDepth()
	return newOptionalAnalyzer(
		"LEG025",
		"require-filename-matches-dirname",
		"Require files in named subdirectories to match the directory name.",
		func(pass *analysis.Pass) (any, error) {
			checkFilenameMatchesDirname(pass, min)
			return nil, nil
		},
	)
}

func newNoMixedFilenameCasing() ruleSpec {
	return newAnalyzer(
		"LEG026",
		"no-mixed-filename-casing",
		"Avoid filenames that mix casing conventions.",
		func(pass *analysis.Pass) (any, error) {
			checkFilenameCasing(pass)
			return nil, nil
		},
	)
}

func checkFilenameMatchesDirname(pass *analysis.Pass, min int) {
	for _, file := range pass.Files {
		filename := pass.Fset.File(file.Pos()).Name()
		if filenameMatchesDirname(filename, min) {
			continue
		}

		report(
			pass,
			file,
			"LEG025",
			"require-filename-matches-dirname",
			"Filename should include the parent directory name.",
		)
	}
}

func checkFilenameCasing(pass *analysis.Pass) {
	for _, file := range pass.Files {
		filename := pass.Fset.File(file.Pos()).Name()
		if hasMixedFilenameCasing(filename) {
			report(
				pass,
				file,
				"LEG026",
				"no-mixed-filename-casing",
				"Filename mixes casing conventions.",
			)
		}
	}
}

func hasMixedFilenameCasing(path string) bool {
	name := filenameStem(filepath.Base(path))
	characters := filenameCharacters(name)
	return mixesFilenameCasing(name, characters)
}

func filenameCharacters(name string) map[rune]bool {
	characters := map[rune]bool{}
	for _, char := range name {
		characters[char] = true
	}

	return characters
}

func mixesFilenameCasing(name string, characters map[rune]bool) bool {
	hasHyphen := characters['-']
	hasUnderscore := characters['_']
	hasUpper := uppercasePattern.MatchString(name)
	hasLower := lowercasePattern.MatchString(name)

	mixesHyphenWithUpper := hasHyphen && hasUpper
	mixesUnderscoreWithMixedCase := hasUnderscore && hasUpper && hasLower
	mixesSeparators := hasHyphen && hasUnderscore

	if mixesHyphenWithUpper {
		return true
	}

	if mixesUnderscoreWithMixedCase {
		return true
	}

	return mixesSeparators
}

func filenameMatchesDirname(path string, min int) bool {
	if dirnameDepth(path) < min {
		return true
	}

	stem := filenameStem(filepath.Base(path))
	if dirnameMatchStandaloneFiles[stem] {
		return true
	}

	parent := filepath.Base(filepath.Dir(path))
	return normalizedNameContains(stem, parent)
}

func dirnameDepth(path string) int {
	dir := filepath.ToSlash(filepath.Clean(filepath.Dir(path)))
	if dir == "." {
		return 0
	}

	depth := 0
	for _, part := range strings.Split(dir, "/") {
		if part == "" {
			continue
		}

		if part != "." {
			depth++
		}
	}

	return depth
}

func normalizedNameContains(name string, part string) bool {
	normalizedName := normalizeFilenamePart(name)
	normalizedPart := normalizeFilenamePart(part)
	return strings.Contains(normalizedName, normalizedPart)
}

func normalizeFilenamePart(part string) string {
	lower := strings.ToLower(part)
	withoutHyphens := strings.ReplaceAll(lower, "-", "_")
	return strings.ReplaceAll(withoutHyphens, " ", "_")
}

func filenameStem(name string) string {
	trimmed := strings.TrimPrefix(name, ".")
	index := strings.Index(trimmed, ".")
	if index < 0 {
		return trimmed
	}

	return trimmed[:index]
}
