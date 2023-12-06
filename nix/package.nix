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
        || lib.hasPrefix (toString ../testdata) (toString path)
      );
  };

  vendorSha256 = "sha256-dlDSa6UT3c/sLzPYgWnBt4PINk7JhTlC6ZFMKXBkUGw=";

  meta = with lib; {
    description = "Restart services without losing connections";
    homepage = "https://github.com/zimbatm/socketmaster";
    license = licenses.mit;
    maintainers = with maintainers; [ zimbatm ];
  };
}
