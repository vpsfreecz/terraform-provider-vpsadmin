{
  description = "Development shells for terraform-provider-vpsadmin";

  inputs = {
    nixpkgs.url = "github:NixOS/nixpkgs/master";
  };

  outputs =
    { self, nixpkgs, ... }:
    let
      systems = [
        "aarch64-darwin"
        "aarch64-linux"
        "x86_64-darwin"
        "x86_64-linux"
      ];

      forAllSystems = nixpkgs.lib.genAttrs systems;
      pkgsFor =
        system:
        import nixpkgs {
          inherit system;
        };
    in
    {
      packages = forAllSystems (
        system:
        let
          pkgs = pkgsFor system;
          go = pkgs.go_1_26;
          assertGoVersion =
            pkgs.lib.assertMsg (go.version == "1.26.3")
              "terraform-provider-vpsadmin requires Go 1.26.3, got ${go.version}";
          get-token = pkgs.buildGo126Module {
            pname = "get-token";
            version = "0.1.0";
            src = ./get-token;
            vendorHash = "sha256-KSCApDutkD45JIINXOrIbyk4uvDS9DifmiYCzeI0i/4=";
          };
        in
        assert assertGoVersion;
        {
          inherit get-token;
          default = get-token;
        }
      );

      devShells = forAllSystems (
        system:
        let
          pkgs = pkgsFor system;
          go = pkgs.go_1_26;
          assertGoVersion =
            pkgs.lib.assertMsg (go.version == "1.26.3")
              "terraform-provider-vpsadmin requires Go 1.26.3, got ${go.version}";
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
