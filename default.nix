{
  pkgs ? import <nixpkgs> { },
}:

{

  jpd = pkgs.callPackage dist/nix/pkgs/javascript-package-delegator { };

}
