# This file was generated by GoReleaser. DO NOT EDIT.
cask "javascript-package-delegator" do
  desc "Software to create fast and easy drum rolls."
  homepage ""
  version "1.0.0"

  livecheck do
    skip "Auto-generated on release."
  end

  binary "jpd"

  on_macos do
    on_intel do
      url "https://github.com/louiss0/javascript-package-delegator/releases/download/v1.0.0/javascript-package-delegator_1.0.0_darwin_amd64.tar.gz"
      sha256 "568d5481c3417bf63081382036b22c6079d4d7a36c9d72b307e71d5c5498de6e"
    end
    on_arm do
      url "https://github.com/louiss0/javascript-package-delegator/releases/download/v1.0.0/javascript-package-delegator_1.0.0_darwin_arm64.tar.gz"
      sha256 "3b63a0edbbe8328247bcec488466db59e7dcad94cd3c7dc9ee64f9458054ec8a"
    end
  end

  on_linux do
    on_intel do
      url "https://github.com/louiss0/javascript-package-delegator/releases/download/v1.0.0/javascript-package-delegator_1.0.0_linux_amd64.tar.gz"
      sha256 "97c020674f1fde55f9cecc45d750dd7bd81ab8b471c81704f241e885c2fcba0e"
    end
    on_arm do
      url "https://github.com/louiss0/javascript-package-delegator/releases/download/v1.0.0/javascript-package-delegator_1.0.0_linux_arm64.tar.gz"
      sha256 "96514fe3e0e52d1d6fe01ab6bc0bbc66a5aeb0d7ac6858cacc01db1b9dbf8144"
    end
  end

  # No zap stanza required
end
