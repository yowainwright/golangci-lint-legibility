package analyzers

import (
	"go/ast"
	"strings"
	"unicode"

	"golang.org/x/tools/go/analysis"
)

var getterPrefixes = []string{"Get", "get"}

func newNoGetterPrefix() ruleSpec {
	return newAnalyzer(
		"LEG042",
		"no-getter-prefix",
		"Avoid the Get prefix on accessor names.",
		func(pass *analysis.Pass) (any, error) {
			checkGetterPrefixes(pass)
			return nil, nil
		},
	)
}

func checkGetterPrefixes(pass *analysis.Pass) {
	for _, file := range pass.Files {
		if isGeneratedFile(file) {
			continue
		}

		ast.Inspect(file, func(node ast.Node) bool {
			decl, ok := node.(*ast.FuncDecl)
			if ok {
				checkGetterName(pass, decl)
			}

			return true
		})
	}
}

func checkGetterName(pass *analysis.Pass, decl *ast.FuncDecl) {
	if !hasGetterPrefix(decl.Name.Name) {
		return
	}

	if !isGetterSignature(decl.Type) {
		return
	}

	report(
		pass,
		decl.Name,
		"LEG042",
		"no-getter-prefix",
		"Drop the Get prefix; name the accessor after the value it returns.",
	)
}

func hasGetterPrefix(name string) bool {
	for _, prefix := range getterPrefixes {
		remainder, found := strings.CutPrefix(name, prefix)
		if found {
			return startsUppercase(remainder)
		}
	}

	return false
}

func isGetterSignature(funcType *ast.FuncType) bool {
	if fieldCount(funcType.Params) != 0 {
		return false
	}

	return fieldCount(funcType.Results) > 0
}

func fieldCount(fields *ast.FieldList) int {
	if fields == nil {
		return 0
	}

	count := 0
	for _, field := range fields.List {
		count += fieldNameCount(field)
	}

	return count
}

func fieldNameCount(field *ast.Field) int {
	if len(field.Names) == 0 {
		return 1
	}

	return len(field.Names)
}

func startsUppercase(text string) bool {
	if text == "" {
		return false
	}

	return unicode.IsUpper([]rune(text)[0])
}

func isGeneratedFile(file *ast.File) bool {
	for _, group := range file.Comments {
		if hasGeneratedMarker(file, group) {
			return true
		}
	}

	return false
}

func hasGeneratedMarker(file *ast.File, group *ast.CommentGroup) bool {
	isBeforePackage := group.Pos() < file.Package
	if !isBeforePackage {
		return false
	}

	return groupHasGeneratedComment(group)
}

func groupHasGeneratedComment(group *ast.CommentGroup) bool {
	for _, comment := range group.List {
		if generatedCommentPattern.MatchString(comment.Text) {
			return true
		}
	}

	return false
}
