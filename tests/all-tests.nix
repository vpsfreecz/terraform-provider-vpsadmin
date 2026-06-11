{
  pkgs ? <nixpkgs>,
  system ? builtins.currentSystem,
  configuration ? null,
  testConfig ? { },
  suiteArgs ? { },
  testFramework,
}:
let
  vpsadminPath = suiteArgs.vpsadminPath or (throw "suiteArgs.vpsadminPath is required");
  providerPackage = suiteArgs.providerPackage or (throw "suiteArgs.providerPackage is required");
  suiteArgs' = suiteArgs // {
    inherit
      vpsadminPath
      providerPackage
      ;
  };

  nixpkgs = import pkgs { inherit system; };
  lib = nixpkgs.lib;
  testLib = testFramework.makeTestLib {
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
  "workflows"
]
