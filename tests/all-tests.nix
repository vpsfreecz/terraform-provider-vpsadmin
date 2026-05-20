{
  pkgs ? <nixpkgs>,
  system ? builtins.currentSystem,
  configuration ? null,
  testConfig ? { },
  suiteArgs ? { },
}:
let
  vpsadminosPath = suiteArgs.vpsadminosPath or (throw "suiteArgs.vpsadminosPath is required");
  vpsadminPath = suiteArgs.vpsadminPath or (throw "suiteArgs.vpsadminPath is required");
  providerPackage = suiteArgs.providerPackage or (throw "suiteArgs.providerPackage is required");
  suiteArgs' = suiteArgs // {
    inherit
      vpsadminosPath
      vpsadminPath
      providerPackage
      ;
  };

  nixpkgs = import pkgs { inherit system; };
  lib = nixpkgs.lib;
  testLib = import (vpsadminosPath + "/test-runner/nix/lib.nix") {
    inherit
      pkgs
      system
      lib
      configuration
      testConfig
      ;
    suiteArgs = suiteArgs';
    suitePath = ./suite;
  };
in
testLib.makeTests [
  "provider/basic"
]
