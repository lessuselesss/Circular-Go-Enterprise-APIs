{
  description = "Go Enterprise APIs for Circular";

  inputs = {
    # claude-code.url = "github:olebedev/claude-code.nix";
    nixpkgs.url = "github:NixOS/nixpkgs/nixpkgs-unstable";
  };

  outputs = { self, nixpkgs }:
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

      # Define @anthropic-ai/claude-code as a Nix package
      claude-code-anthropic-pkg = pkgs.buildNpmPackage {
        pname = "@anthropic-ai/claude-code";
        version = "0.0.1"; # IMPORTANT: You might need to find the latest version from npm or their repo
        src = pkgs.fetchFromGitHub {
          owner = "anthropics";
          repo = "claude-code";
          rev = "main"; # Using the main branch
          sha256 = "sha256-xP0FfQtpzzgv8gE1emsbibobS3Mn1cp+YwVAFGr2H+w="; # Confirmed hash
        };
        npmDepsHash = ""; # Keep this empty for now, Nix will tell us the correct one

        # Add a preConfigure hook to generate package-lock.json
        preConfigure = '''
          echo "Generating package-lock.json using bun install..."
          bun install --frozen-lockfile=false --ignore-scripts
        ''';

        # Most npm packages have a 'build' script, if not, add dontNpmBuild = true;
        # dontNpmBuild = true;

        nativeBuildInputs = [ pkgs.nodejs pkgs.bun ];
      };
    in
    {
      devShells.${system}.default = pkgs.mkShell {
        buildInputs = [
          go-pkg
          pkgs.bun
          pkgs.gopls
          pkgs.nodejs
          pkgs.gh
          pkgs.container2wasm
          claude-task-master-pkg
          # Removed the old claude-code input
          claude-code-anthropic-pkg
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
