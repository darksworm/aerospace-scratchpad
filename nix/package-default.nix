{ pkgs, ... }:
  pkgs.buildGoModule rec {
    name = "aerospace-scratchpad";
    version = "v0.2.3";

    src = pkgs.fetchFromGitHub {
      owner = "cristianoliveira";
      repo = "aerospace-scratchpad";
      rev = version;
      sha256 = "sha256-N6Ah+Fvyvs7UNpMfh+cP8eOCfdLBSXT4Gc3l69qhueI=";
    };

    vendorHash = "sha256-u2tlLqfcKG8JbzdAA5RnH8m0yzjCGKxQv0FsgxZKDYI=";

    ldflags = [
      "-s" "-w"
      "-X github.com/cristianoliveira/aerospace-scratchpad/cmd.VERSION=${version}"
    ];

    meta = with pkgs.lib; {
      description = "aerospace-scratchpad: Scratchpad for AeroSpaceWM";
      homepage = "https://github.com/cristianoliveira/aerospace-scratchpad";
      license = licenses.mit;
      maintainers = with maintainers; [ cristianoliveira ];
    };
  }
