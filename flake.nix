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
          sha256 = "sha256-UeI5PrryZzluxWX/kfsUAwdYFC81x5f5vPE7/9GaK6I="; # Confirmed hash
        };
        # The npmDepsHash for this package is still empty. Nix will tell you the correct value.
        npmDepsHash = "sha256-xP0FfQtpzzgv8gE1emsbibobS3Mn1cp+YwVAFGr2H+w=";

        # Removed the problematic preConfigure hook as package-lock.json exists upstream.

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
          echo "Setting up .kilocode symlink..."
          if [ -z "$XDG_DATA_HOME" ]; then
            export XDG_DATA_HOME="$HOME/.local/share"
          fi
          mkdir -p "$XDG_DATA_HOME" # Ensure the parent directory for the symlink exists
          rm -rf "$XDG_DATA_HOME/.kilocode" # Remove any existing conflicting file or directory/symlink
          ln -sf "$PWD/.kilocode" "$XDG_DATA_HOME/.kilocode" # Create the symlink
        '';
      };
    };
}
