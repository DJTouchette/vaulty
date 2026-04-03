class Vaulty < Formula
  desc "CLI tool for managing encrypted secrets"
  homepage "https://github.com/djtouchette/vaulty"
  license "MIT"
  head "https://github.com/djtouchette/vaulty.git", branch: "main"

  # Stable releases are tracked via git tags.
  # Update the url/sha256 when cutting a new release:
  #   url "https://github.com/djtouchette/vaulty/archive/refs/tags/v#{version}.tar.gz"
  #   sha256 "<sha256>"

  depends_on "go" => :build

  def install
    version_str = if build.head?
                    Utils.safe_popen_read("git", "describe", "--tags", "--always", "--dirty").chomp
                  else
                    version.to_s
                  end

    ldflags = %W[
      -s -w
      -X main.version=#{version_str}
    ]

    system "go", "build", *std_go_args(ldflags:), "./cmd/vaulty"
  end

  test do
    assert_match version.to_s, shell_output("#{bin}/vaulty --version")
  end
end
