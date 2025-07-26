class EzEnv < Formula
  desc "Git-integrated file encryption using GitHub repository secrets"
  homepage "https://github.com/oliviaBahr/ez-env"
  version "0.1.0"
  
  # Build from source
  head "https://github.com/oliviaBahr/ez-env.git"
  
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