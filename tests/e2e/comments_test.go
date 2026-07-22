//go:build e2e

package e2e_test

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestDefaultPolicyAllowsOrdinaryComments(t *testing.T) {
	requireCommentClean(t, "default", "default")
}

func TestTicketPolicyAllowsMatchingComments(t *testing.T) {
	requireCommentClean(t, "ticket", "ticket/matched")
}

func TestTicketPolicyRejectsUnmatchedComments(t *testing.T) {
	requireCommentDiagnostic(t, "ticket", "ticket/unmatched", "LEG039 no-unmatched-comments")
}

func TestAttributionPolicyAllowsUsefulContext(t *testing.T) {
	requireCommentClean(t, "default", "attribution/context")
}

func TestAttributionPolicyRejectsAutomatedSignatures(t *testing.T) {
	requireCommentDiagnostic(t, "default", "attribution/signature", "LEG040 no-automated-comment-attribution")
}

func TestLineCommentPolicyAllowsLineComments(t *testing.T) {
	requireCommentClean(t, "default", "line-comments/line")
}

func TestLineCommentPolicyRejectsMultilineBlockProse(t *testing.T) {
	requireCommentDiagnostic(t, "default", "line-comments/block", "LEG041 prefer-line-comments")
}

func TestStrictCommentPolicyAllowsGoMetadata(t *testing.T) {
	requireCommentClean(t, "metadata", "metadata/directive")
}

func TestStrictCommentPolicyAllowsCgoPreambles(t *testing.T) {
	requireCommentClean(t, "metadata", "metadata/cgo")
}

func requireCommentClean(t *testing.T, configName string, fixture string) {
	t.Helper()

	output, err := lintCommentFixture(t, configName, fixture)
	if err != nil {
		t.Fatalf("expected clean lint, received %v\n%s", err, output)
	}
}

func requireCommentDiagnostic(t *testing.T, configName string, fixture string, diagnostic string) {
	t.Helper()

	output, err := lintCommentFixture(t, configName, fixture)
	var exitError *exec.ExitError
	if !errors.As(err, &exitError) || exitError.ExitCode() != 1 {
		t.Fatalf("expected lint exit 1, received %v\n%s", err, output)
	}
	if !strings.Contains(output, diagnostic) {
		t.Fatalf("missing diagnostic %q\n%s", diagnostic, output)
	}
}

func lintCommentFixture(t *testing.T, configName string, fixture string) (string, error) {
	t.Helper()

	root := repositoryRoot(t)
	commentsRoot := filepath.Join(root, "tests", "e2e", "testdata", "comments")
	config := filepath.Join(commentsRoot, configName, ".golangci.yml")
	target := filepath.Join(commentsRoot, fixture)
	args := []string{"run", "--color=never", "--show-stats=false", "--config", config, target}
	command := exec.Command(e2eBinary(root), args...)
	command.Dir = root
	command.Env = append(os.Environ(), "GOLANGCI_LINT_CACHE="+t.TempDir())
	output, err := command.CombinedOutput()
	return string(output), err
}

func repositoryRoot(t *testing.T) string {
	t.Helper()

	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}

	return root
}

func e2eBinary(root string) string {
	if binary := os.Getenv("LEGIBILITY_E2E_BINARY"); binary != "" {
		return binary
	}

	return filepath.Join(root, "bin", "legibility-golangci-lint")
}
