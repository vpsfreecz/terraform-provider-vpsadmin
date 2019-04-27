let
  pkgs = import <nixpkgs> {};
  stdenv = pkgs.stdenv;

in stdenv.mkDerivation rec {
  name = "terraform-provider-vpsadmin";

  buildInputs = with pkgs;[
    git
    go
    gotools
    terraform
  ];
}
