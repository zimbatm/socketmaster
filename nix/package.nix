{ buildGoModule, lib }:

buildGoModule {
  name = "socketmaster";
  src = lib.cleanSourceWith {
    src = lib.cleanSource ../.;
    filter = path: type:
      baseNameOf path != "nix" && (
        lib.hasSuffix ".go" path
        || lib.hasSuffix "/go.mod" path
        || lib.hasSuffix "/go.sum" path
      );
  };

  vendorSha256 = "sha256-bPIzk+g8lxntEREAp1ZJiHTGxXFRybtZjhR23uGIFR4=";

  meta = with lib; {
    description = "Restart services without losing connections";
    homepage = "https://github.com/zimbatm/socketmaster";
    license = licenses.mit;
    maintainers = with maintainers; [ zimbatm ];
  };
}
