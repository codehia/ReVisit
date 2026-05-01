{ pkgs, ... }: {
  packages = with pkgs; [ git gofumpt sqlite golangci-lint goose go-jet ];
  languages = {
    go.enable = true;
    nix.enable = true;
  };
}
