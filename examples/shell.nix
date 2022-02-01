let
  pkgs = import <nixpkgs> {};
  stdenv = pkgs.stdenv;
  get-token = pkgs.buildGoModule {
    name = "get-token";
    src = ../get-token;
    vendorSha256 = "sha256:0dpjll0qhkgq6yidz2p7451l35z92c60a2dggk2vq7x54w0ksi3l";
  };

in stdenv.mkDerivation rec {
  name = "terraform-provider-vpsadmin-example";

  buildInputs = with pkgs; [
    get-token
    terraform
  ];
}
