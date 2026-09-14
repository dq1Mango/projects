{
  pkgs ? import <nixpkgs> { },
}:

let
  # Baked in at build time — reproducible, no host mount needed.
  fishConfig = pkgs.runCommand "fish-config" { } ''
    mkdir -p $out/root/.config/
    cp -r ${~/.config/dotfiles/shared/fish} $out/root/.config/
  '';

  nixConf = pkgs.writeTextDir "etc/nix/nix.conf" ''
    experimental-features = nix-command flakes
    sandbox = false
    build-users-group =
  '';
in
pkgs.dockerTools.buildImageWithNixDb {
  name = "sandbox";
  tag = "latest";

  copyToRoot = [
    pkgs.fish
    pkgs.bashInteractive
    pkgs.coreutils
    pkgs.busybox
    pkgs.nix
    pkgs.git
    pkgs.cacert # nix + fetching deps
    pkgs.ncurses # terminfo, for TUIs
    pkgs.dockerTools.fakeNss # /etc/passwd so fish/nix have a user
    pkgs.dockerTools.binSh # /bin/sh
    pkgs.dockerTools.usrBinEnv # /usr/bin/env for shebangs
    fishConfig
    nixConf
  ];

  config = {
    Cmd = [ "/bin/fish" ];
    WorkingDir = "/workspace";
    Env = [
      "HOME=/root"
      "USER=root"
      "TERM=xterm-256color"
      "LANG=C.UTF-8"
      # lets nix reach cache.nixos.org over TLS
      "NIX_SSL_CERT_FILE=${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt"
      "SSL_CERT_FILE=${pkgs.cacert}/etc/ssl/certs/ca-bundle.crt"
      # so classic nix-shell / <nixpkgs> resolve to a pinned tree
      "NIX_PATH=nixpkgs=${pkgs.path}"
    ];
  };
}
