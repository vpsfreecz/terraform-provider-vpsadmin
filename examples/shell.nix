let
  pkgs = import <nixpkgs> {};
  stdenv = pkgs.stdenv;
  get-token = pkgs.buildGoModule {
    name = "get-token";
    src = ../get-token;
    vendorHash = "sha256-TOqys10Q1BEEGhBHIppemjMP5iT0HFQbcAEgqh0AEyw=";
  };

in stdenv.mkDerivation rec {
  name = "terraform-provider-vpsadmin-example";

  buildInputs = with pkgs; [
    get-token
    opentofu
  ];
}
