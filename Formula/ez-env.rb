class EzEnv < Formula
  desc "Git-integrated file encryption using GitHub repository secrets"
  homepage "https://github.com/oliviaBahr/ez-env"
  version "0.1.0"
  
  # These URLs will be updated when you create releases
  # For now, we'll build from source
  head "https://github.com/oliviaBahr/ez-env.git"
  
  # When you have releases, uncomment and update these:
  # url "https://github.com/oliviaBahr/ez-env/releases/download/v#{version}/git-ez-env-darwin-amd64"
  # sha256 "PLACEHOLDER_SHA256"
  
  depends_on "go" => :build
  
  def install
    # Build the binary
    system "go", "build", "-o", "git-ez-env"
    
    # Install the binary
    bin.install "git-ez-env"
  end
  
  test do
    # Test that the binary can be executed and shows help
    output = shell_output("#{bin}/git-ez-env", 1)
    assert_match "Usage: git ez-env", output
  end
end 