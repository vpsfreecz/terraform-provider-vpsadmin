let
  pkgs = import <nixpkgs> {};
  stdenv = pkgs.stdenv;
  get-token = pkgs.buildGoModule {
    name = "get-token";
    src = ../get-token;
    modSha256 = "sha256:1fqm78dkbqzwky6dd85baw9b4lix2qf3yr460jzfmcv722dq82zh";
  };

in stdenv.mkDerivation rec {
  name = "terraform-provider-vpsadmin-example";

  buildInputs = with pkgs; [
    get-token
    terraform
  ];
}
