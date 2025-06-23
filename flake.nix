{
  description = "Go Enterprise APIs for Circular";

  inputs = {
    claude-code.url = "github:olebedev/claude-code.nix";
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs = { self, nixpkgs, claude-code }:
    let
      system = "x86_64-linux";
      pkgs = import nixpkgs {
        inherit system;
      };
      go-pkg = pkgs.go_1_23;

      # Define claude-task-master as a Nix package
      claude-task-master-pkg = pkgs.buildNpmPackage {
        pname = "task-master-ai";
        version = "0.18.0"; # Using the version we previously found
        src = pkgs.fetchFromGitHub {
          owner = "eyaltoledano";
          repo = "claude-task-master";
          rev = "v0.18.0"; # The tag for the version
          sha256 = "sha256-RnbquGcanpBH5A++MZOVNLXEdn7qVJIVWxUOZEBpF/o="; # IMPORTANT: Replace with the actual hash Nix provides
        };
        npmDepsHash = "sha256-GjRxjafbJ5DqikvO3Z7YPtuUHaG5ezxdrQq9f7WDEi4="; # IMPORTANT: Replace with the actual hash of node_modules Nix provides

        # claude-task-master doesn't have a "build" script, so we skip this phase
        dontNpmBuild = true;

        # Ensure nodejs is available for the npm install phase
        nativeBuildInputs = [ pkgs.nodejs ];
      };
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        buildInputs = [
          go-pkg
          pkgs.gopls
          pkgs.nodejs
          pkgs.gh
          claude-task-master-pkg
          claude-code.packages.${system}.default
          pkgs.python3
          pkgs.python3Packages.pip
          pkgs.python3Packages.python-dotenv
        ];
        shellHook = ''
          echo "Nix shellHook executed. Setting up tm wrapper..."
          mkdir -p $TMPDIR/nix-shell-bin
          ln -sf "$(which task-master)" $TMPDIR/nix-shell-bin/tm
          export PATH=$TMPDIR/nix-shell-bin:$PATH
        '';
      };
    };
}