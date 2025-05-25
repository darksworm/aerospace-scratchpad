{ pkgs, ... }:
  pkgs.buildGoModule rec {
    # name of our derivation
    name = "aerospace-scratchpad";
    # FIXME: once we have the first release, we can use the version
    version = "v0.0.1";
    # version = "64ebad853115b0565f284354efb833bf6c045c33";

    # sources that will be used for our derivation.
    src = pkgs.fetchFromGitHub {
      owner = "cristianoliveira";
      repo = "aerospace-scratchpad";
      rev = version;
      sha256 = "sha256-viyRp0Lxtkr4L4IGFCb4exQ/4+wRKTYi/WljI1b/y9c=";
    };

    vendorHash = "sha256-BVv2GPPZzTUfBASPqOwRfpFaD07XY74EostJ5F6ryfA=";

    ldflags = [
      "-s" "-w"
      "-X main.VERSION=${version}"
    ];

    meta = with pkgs.lib; {
      description = "aerospace-scratchpad: SwayWM like scratchpad for AeroSpaceWM";
      homepage = "https://github.com/cristianoliveira/aerospace-scratchpad";
      license = licenses.mit;
      maintainers = with maintainers; [ cristianoliveira ];
    };
  }
