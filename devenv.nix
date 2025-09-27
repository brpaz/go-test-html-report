{
  pkgs,
  lib,
  config,
  inputs,
  ...
}: {
  # https://devenv.sh/packages/
  packages = with pkgs; [
    go
    go-task
    gotestsum
    gomarkdoc
    goreleaser
    golangci-lint
    lefthook
    commitlint
  ];

  enterShell = ''
    lefthook install
  '';
}
