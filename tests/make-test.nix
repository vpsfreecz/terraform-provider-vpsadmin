testFn:
{
  vpsadminosPath,
  vpsadminPath,
  providerPackage,
  ...
}@args:
let
  upstream = import (vpsadminosPath + "/tests/make-test.nix") testFn;
  mergedExtraArgs = {
    vpsadminos = vpsadminosPath;
    vpsadmin = vpsadminPath;
    inherit providerPackage;
  }
  // (args.extraArgs or { });
  argsWithExtra = args // {
    extraArgs = mergedExtraArgs;
  };
in
upstream argsWithExtra
