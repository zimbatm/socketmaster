{ nixosTest, pkgs, extraModules ? [ ] }:

let
  socket-server = pkgs.callPackage ./test/socket-server { };
in

nixosTest {
  nodes.machine = { ... }: {
    imports = extraModules;
    config = {
      environment.systemPackages = [
        # Used by testScript
        pkgs.socat
      ];
      systemd.sockets.socket-server = {
        wantedBy = [ "sockets.target" ];
        listenStreams = [ "0.0.0.0:2022" ];
      };
      systemd.services.socket-server = {
        wantedBy = [ "multi-user.target" ];
        # TODO: this will have to move to the socketmaster config
        environment.ECHO_VALUE = "v1";
        # serviceConfig.ExecStart = "${socket-server}/bin/socket-server";
      };

      socketmaster.services.socket-server = {
        settings.command = "${socket-server}/bin/socket-server";
      };
    };
  };
  testScript = ''
    machine.wait_for_unit("sockets.target")
    machine.succeed("""
      echo '{"cmd":"echo"}' \
        | socat - TCP4:localhost:2022 \
        | grep -E '^"v1"$'
    """)

    machine.wait_for_unit("socket-server.service")
    machine.succeed("sleep 3")
  '';
}
