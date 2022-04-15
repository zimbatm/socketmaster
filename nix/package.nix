{ buildGoModule, lib }:

buildGoModule {
  name = "socketmaster";
  src = ../.; # FIXME cleanSource

  vendorSha256 = "sha256-bPIzk+g8lxntEREAp1ZJiHTGxXFRybtZjhR23uGIFR4=";

  meta = with lib; {
    description = "Restart services without losing connections";
    homepage = "https://github.com/zimbatm/socketmaster";
    license = licenses.mit;
    maintainers = with maintainers; [ zimbatm ];
  };
}
