{
  inputs =
    {
      nixpkgs.url = github:NixOS/nixpkgs/nixpkgs-unstable;

      flake-utils.url = github:numtide/flake-utils;

      gitignore.url = github:hercules-ci/gitignore.nix;
      gitignore.inputs.nixpkgs.follows = "nixpkgs";
    };

  outputs = inputs:
    let
      inherit (inputs) nixpkgs gitignore flake-utils;
      inherit (gitignore.lib) gitignoreSource;
    in
    flake-utils.lib.eachDefaultSystem (system:
      let
        pkgs = import nixpkgs { inherit system; };
      in
      rec {
        formatter = pkgs.nixpkgs-fmt;
        packages.default = packages.dbkp;
        packages.dbkp = pkgs.buildGoModule {
          pname = "dbkp";
          version = "2.0.0";
          src = gitignoreSource ./.;
          vendorHash = "sha256-XStQdswKY/SieK6dxu4xxHA3AP43XOIiJLJjwgaBUeg=";
          installPhase = ''
            runHook preInstall
            mkdir -p $out/bin
            mkdir -p build
            $GOPATH/bin/docgen
            cp -r build/share $out/share
            cp $GOPATH/bin/dbkp $out/bin/dbkp
            runHook postInstall
          '';
        };
        apps = rec {
          dbkp = { type = "app"; program = "${packages.dbkp}/bin/dbkp"; };
          default = dbkp;
        };
        devShell = pkgs.mkShell {
          packages = with pkgs;[ packages.dbkp go git man busybox ];
          shellHook = ''
            export fish_complete_path=${packages.dbkp}/share/fish/completions
          '';
        };
      }
    );
}
