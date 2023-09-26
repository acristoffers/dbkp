{
  description = "I/O Communication Library";
  inputs = {
    flake-utils.url = github:numtide/flake-utils;
    nixpkgs.url = github:NixOS/nixpkgs/nixos-unstable;
  };
  outputs = { self, nixpkgs, flake-utils }:
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      rec {
        dbkp = pythonPkgs: pythonPkgs.buildPythonPackage rec {
          format = "pyproject";
          name = "dbkp";
          src = ./.;
          propagatedBuildInputs = with pythonPkgs; [
            pydantic
            poetry-core
          ];
        };
        packages.default = dbkp pkgs.python311Packages;
        apps = rec {
          dbkp = { type = "app"; program = "${packages.default}/bin/dbkp"; };
          default = dbkp;
        };
      });
}
