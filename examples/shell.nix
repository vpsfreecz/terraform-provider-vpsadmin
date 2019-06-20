let
  pkgs = import <nixpkgs> {};
  stdenv = pkgs.stdenv;
  get-token = pkgs.buildGoModule {
    name = "get-token";
    src = ../get-token;
    modSha256 = "1m840qx0h876gm1pars3bfyh18xrsadcmp54pibkpqd2n7ib2v7d";
  };

in stdenv.mkDerivation rec {
  name = "terraform-provider-vpsadmin-example";

  buildInputs = with pkgs; [
    get-token
    terraform
  ];
}
