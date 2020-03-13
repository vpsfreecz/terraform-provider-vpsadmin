let
  pkgs = import <nixpkgs> {};
  stdenv = pkgs.stdenv;
  get-token = pkgs.buildGoModule {
    name = "get-token";
    src = ../get-token;
    modSha256 = "sha256:067vdgs0wfk1s8ybvg8swdc2il6wlvsf8w1xgiakcnp0rplbkx0g";
  };

in stdenv.mkDerivation rec {
  name = "terraform-provider-vpsadmin-example";

  buildInputs = with pkgs; [
    get-token
    terraform
  ];
}
