# Homebrew formula for mirusync.
# Install from this repo: brew install --build-from-source ./Formula/mirusync.rb
# Or add a tap (see SETUP.md) and: brew install mirusync

class Mirusync < Formula
  desc "Folder sync between two machines over SSH"
  homepage "https://github.com/hanif/mirusync"
  url "https://github.com/hanif/mirusync/archive/refs/heads/main.tar.gz"
  version "0.1.0"
  license "MIT"
  head "https://github.com/hanif/mirusync.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", *std_go_args(ldflags: "-s -w"), "."
  end

  test do
    assert_match "mirusync", shell_output("#{bin}/mirusync --help", 1)
  end
end
