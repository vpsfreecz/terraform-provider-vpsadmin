let
  pkgs = import <nixpkgs> {};
  stdenv = pkgs.stdenv;
  get-token = pkgs.buildGoModule {
    name = "get-token";
    src = ../get-token;
    modSha256 = "049pbqxpzcgz37h11446n16h1592iymn3ni7rw9lc230qmxbf7ss";
  };

in stdenv.mkDerivation rec {
  name = "terraform-provider-vpsadmin-example";

  buildInputs = with pkgs; [
    get-token
    terraform
  ];
}
