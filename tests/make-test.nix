testFn:
{
  testFramework,
  vpsadminPath,
  providerPackage,
  ...
}@args:
let
  upstream = testFramework.makeTest testFn;
  mergedExtraArgs = {
    vpsadminos = testFramework.sourcePath;
    vpsadmin = vpsadminPath;
    inherit providerPackage;
  }
  // (args.extraArgs or { });
  argsWithExtra = args // {
    extraArgs = mergedExtraArgs;
  };
in
upstream argsWithExtra
