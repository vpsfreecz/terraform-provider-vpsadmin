{
  description = "Terraform provider for vpsAdmin";

  inputs = {
    vpsadmin.url = "github:vpsfreecz/vpsadmin";
    vpsadminos.follows = "vpsadmin/vpsadminos";
    nixpkgs.follows = "vpsadminos/nixpkgs";
  };

  outputs =
    {
      self,
      nixpkgs,
      vpsadmin,
      vpsadminos,
      ...
    }:
    let
      systems = [
        "aarch64-darwin"
        "aarch64-linux"
        "x86_64-darwin"
        "x86_64-linux"
      ];
      testSystems = [ "x86_64-linux" ];

      forAllSystems = nixpkgs.lib.genAttrs systems;
      forTestSystems = nixpkgs.lib.genAttrs testSystems;
      pkgsFor =
        system:
        import nixpkgs {
          inherit system;
        };

      suiteArgsFor = system: {
        vpsadminPath = vpsadmin.outPath;
        providerPackage = self.packages.${system}.terraform-provider-vpsadmin;
      };
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = pkgsFor system;
          go = pkgs.go_1_26;
          assertGoVersion = pkgs.lib.assertMsg (
            go.version == "1.26.3"
          ) "terraform-provider-vpsadmin requires Go 1.26.3, got ${go.version}";
          providerSource = pkgs.lib.cleanSourceWith {
            src = ./.;
            filter =
              path: type:
              let
                root = toString ./.;
                relPath = pkgs.lib.removePrefix "${root}/" (toString path);
              in
              pkgs.lib.cleanSourceFilter path type
              && relPath != "result"
              && relPath != "terraform-provider-vpsadmin";
          };
          terraform-provider-vpsadmin = pkgs.buildGo126Module {
            pname = "terraform-provider-vpsadmin";
            version = "1.2.0";
            src = providerSource;
            vendorHash = "sha256-98lrEH7j4Fl0oSEoL75e0h+8ySLacPK2k0OehY3jqhA=";
            subPackages = [ "." ];
          };
          get-token = pkgs.buildGo126Module {
            pname = "get-token";
            version = "0.1.0";
            src = ./get-token;
            vendorHash = "sha256-KSCApDutkD45JIINXOrIbyk4uvDS9DifmiYCzeI0i/4=";
          };
        in
        assert assertGoVersion;
        {
          inherit get-token terraform-provider-vpsadmin;
          default = terraform-provider-vpsadmin;
        }
        // pkgs.lib.optionalAttrs (system == "x86_64-linux") {
          test-runner = vpsadminos.packages.${system}.test-runner;
        }
      );

      apps = forTestSystems (system: {
        test-runner = {
          type = "app";
          program = "${vpsadminos.packages.${system}.test-runner}/bin/test-runner";
        };
      });

      tests = forTestSystems (
        system:
        vpsadminos.lib.testFramework.mkTests {
          inherit system;
          pkgsPath = nixpkgs.outPath;
          testsRoot = ./tests;
          suiteArgs = suiteArgsFor system;
        }
      );

      testsMeta = forTestSystems (
        system:
        vpsadminos.lib.testFramework.mkTestsMeta {
          inherit system;
          pkgsPath = nixpkgs.outPath;
          testsRoot = ./tests;
          suiteArgs = suiteArgsFor system;
        }
      );

      devShells = forAllSystems (
        system:
        let
          pkgs = pkgsFor system;
          go = pkgs.go_1_26;
          assertGoVersion = pkgs.lib.assertMsg (
            go.version == "1.26.3"
          ) "terraform-provider-vpsadmin requires Go 1.26.3, got ${go.version}";
          get-token = self.packages.${system}.get-token;
        in
        assert assertGoVersion;
        {
          default = pkgs.mkShell {
            packages = [
              pkgs.gcc
              pkgs.git
              pkgs.gnumake
              go
              pkgs.gotools
              pkgs.opentofu
            ];

            CGO_ENABLED = "1";

            shellHook = ''
              export PS1="(vpsadmin-provider) ''${PS1:-}"
            '';
          };

          examples = pkgs.mkShell {
            packages = [
              get-token
              pkgs.opentofu
            ];

            shellHook = ''
              export PS1="(vpsadmin-examples) ''${PS1:-}"
            '';
          };
        }
      );
    };
}
