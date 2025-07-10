{
  pkgs ? import <nixpkgs> { },
}:

{

  jpd = pkgs.callPackage dist/nix/pkgs/javascript-package-delegator { };

}
# curl -O https://github.com/louiss0/javascript-package-delegator/releases/download/v1.0.0/javascript-package-delegator_1.0.0_linux_amd64.tar.gz
# sha256sum javascript-package-delegator_1.0.0_linux_amd64.tar.gz
