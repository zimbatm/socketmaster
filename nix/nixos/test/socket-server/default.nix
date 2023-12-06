{ buildGoModule }:

buildGoModule {
  pname = "socket-server";
  version = "1.0.0";
  src = ./.;
  vendorSha256 = "sha256-wivLZzDVpLPi+RgVti03Se0TQ0kco7CBnPFOITLFwx8=";
}
