{
  inputs =
    {
      nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";

      flake-utils.url = "github:numtide/flake-utils";

      gitignore.url = "github:hercules-ci/gitignore.nix";
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
          version = (builtins.readFile ./pkg/dbkp/version);
          src = gitignoreSource ./.;
          vendorHash = "sha256-XAs9ucifQJLKxu5sb8BovQmkxcNWnkomX2eARCfgGoY=";
          buildInputs = with pkgs; [ glibc.static ];
          CFLAGS = "-I${pkgs.glibc.dev}/include";
          LDFLAGS = "-L${pkgs.glibc}/lib";
          ldflags = [ "-s" "-w" "-linkmode external" "-extldflags '-static'" ];
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
