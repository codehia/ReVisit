{ pkgs, ... }: {
  packages = with pkgs; [ git gofumpt ];
  languages = {
    go.enable = true;
    nix.enable = true;
  };
}
