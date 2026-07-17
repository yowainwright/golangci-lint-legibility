class GolangciLintLegibility < Formula
  source_url = "https://github.com/yowainwright/golangci-lint-legibility/releases/download/v0.2.0/" \
               "golangci-lint-legibility_0.2.0_source.tar.gz"
  source_sha = "02dc8216777c6504d9240329be5c4b9d" \
               "ab2dce41a929fc03a1a2d0682c85e273"

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
    ENV["GOFLAGS"] = "-buildvcs=false"
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
