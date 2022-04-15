# shell.nix is still required for vscode Nix Env Selector (2022-04)
(builtins.getFlake ("git+file://" + toString ./.)).devShells.${builtins.currentSystem}.default
