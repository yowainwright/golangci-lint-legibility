class GolangciLintLegibility < Formula
  source_url = "https://github.com/yowainwright/golangci-lint-legibility/releases/download/v0.1.0/" \
               "golangci-lint-legibility_0.1.0_source.tar.gz"
  source_sha = "51051b184ec0c1a7117629fc2851858c" \
               "e57e4fdd1bc9424bf1947b8cd6dfa307"

  desc "Syntax-only Go readability rules for golangci-lint"
  homepage "https://github.com/yowainwright/golangci-lint-legibility"
  url source_url
  sha256 source_sha
  license "MIT"

  livecheck do
    url :stable
    strategy :github_latest
  end

  depends_on "golangci-lint" => :build
  depends_on "go"

  def install
    system "golangci-lint", "custom"
    bin.install "bin/legibility-golangci-lint"
  end

  test do
    (testpath/"go.mod").write <<~GOMOD
      module example.com/legibility-test

      go 1.25
    GOMOD

    (testpath/"main.go").write <<~GO
      package main

      func main() {
        if true && false {
          println("unreachable")
        }
      }
    GO

    (testpath/".golangci.yml").write <<~YAML
      version: "2"

      linters:
        default: none
        enable:
          - legibility
        settings:
          custom:
            legibility:
              type: module
              description: Syntax-only Go legibility rules.
              original-url: github.com/yowainwright/golangci-lint-legibility
    YAML

    command = "#{bin}/legibility-golangci-lint run ./... 2>&1"
    output = shell_output(command, 1)
    assert_match "LEG002 hoist-if-operators", output
  end
end
