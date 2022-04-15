{
  description = "A very basic flake";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    let
      inherit (flake-utils.lib) eachDefaultSystem;
      inherit (nixpkgs) lib;

      flakeAttrs = {
        nixosModules.default = { imports = [ ./nix/nixos/module.nix ]; };
      };

      perSystem = system:
        let
          pkgs = nixpkgs.legacyPackages.${system};
        in
        {
          devShell = self.devShells.${system}.default;
          devShells.default = pkgs.mkShell {
            nativeBuildInputs = [
              pkgs.go
              pkgs.gopls
              pkgs.go-outline
              pkgs.nixpkgs-fmt
            ];
          };

          defaultPackage = self.packages.${system}.default;
          packages.default = pkgs.callPackage ./nix/package.nix { };
          checks =
            lib.optionalAttrs pkgs.stdenv.isLinux {
              nixos = pkgs.callPackage ./nix/nixos/test.nix {
                socketmaster = self.packages.${system}.default;
              };
            };
        };

      systemAttrs = eachDefaultSystem perSystem;

    in
    systemAttrs // flakeAttrs;

}
