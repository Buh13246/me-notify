{ lib, buildGoModule, ... }:
buildGoModule {
  pname = "me-notify";
  version = "master";
  src = ./.;
  vendorSha256 = null;
  modSha256 = lib.fakeSha256;
  buildInputs = [ ];
  nativeBuildInputs = [ ];
}
