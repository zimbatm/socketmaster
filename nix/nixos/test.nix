{ nixosTest, pkgs, extraModules ? [ ] }:

let
  socket-server = pkgs.callPackage ./test/socket-server { };
in

nixosTest {
  nodes.machine = { config, lib, ... }: {
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
      };

      socketmaster.services.socket-server = {
        settings.command = lib.mkDefault "${socket-server}/bin/socket-server";
        settings.environment.ECHO_VALUE = lib.mkDefault "v1";
      };

      specialisation.system-v2 = {
        inheritParentConfig = true;
        configuration = { config, lib, pkgs, ... }: {

          socketmaster.services.socket-server = {
            # Require both environment and command replacement to set it to "v2"
            settings.environment.ECHO_VALUE = "v2-please";
            settings.command = "${pkgs.writeScript "socket-server-wrapped" ''
              #!${pkgs.runtimeShell}
              ECHO_VALUE=''${ECHO_VALUE/-please/}
              exec ${socket-server}/bin/socket-server
            ''}";
          };

        };
      };

      # The activation script runs after outdated services were stopped, but
      # before the updated ones start.
      system.activationScripts.etc.deps = [ "check-socket" ]; # run before etc
      system.activationScripts.check-socket.text = ''
        # # check that the service has been stopped
        # ${config.systemd.package}/bin/systemctl status socket-server.service >/dev/console
        # r=$?
        # if [[ $r != 3 ]]; then
        #   echo >&2 systemctl returned unexpected exit code $r >/dev/console
        #   exit 1
        # fi

        # # The above trips the activation script error checker. Undo that.
        # _localstatus=0
        # _status=0

        # detach a client
        ( (echo '{"cmd":"echo"}'; while ! [[ -e /tmp/client.stop ]]; do sleep 0.1; done; echo '{"cmd":"echo"}') \
          | ${pkgs.socat}/bin/socat - TCP4:localhost:2022 \
          | tee /tmp/client.out \
          | while read ln; do
              echo "client got response: $ln"
            done >/dev/console
          ) </dev/null >/tmp/client.out 2>/tmp/client.err &
        clientpid=$!
        echo $clientpid >/tmp/client.pid
        disown %%

        # delay to make sure client starts while service was down
        sleep 1

        # Fail fast if something is wrong with the client
        if ! [[ -d /proc/$clientpid ]]; then
          echo >/dev/console "client exited unexpectedly. It shouldn't be able to complete while the service is not running."
          exit 1
        fi
      '';

    };
  };
  testScript = ''
    machine.wait_for_unit("sockets.target")

    with subtest("Service v1 works"):
      machine.succeed("""
        echo '{"cmd":"echo"}' \
          | socat - TCP4:localhost:2022 \
          | grep -E '^"v1"$'
      """)

    machine.succeed("""
      /run/booted-system/specialisation/system-v2/bin/switch-to-configuration test
    """)

    # Wait for the configured startup time to elapse
    machine.succeed("sleep 4")

    with subtest("Service v2 works"):
      machine.succeed("""
        echo '{"cmd":"echo"}' \
          | socat - TCP4:localhost:2022 \
          | tee /dev/console \
          | grep -E '^"v2"$'
      """)

    with subtest("Service v1 still serves a client that connected during NixOS switching"):
      machine.succeed("""
        (
        touch /tmp/client.stop
        while kill -0 "$(cat /tmp/client.pid)"; do 
          echo Waiting for client to exit
          sleep 0.1;
        done
        echo out:
        cat /tmp/client.out
        echo err:
        cat /tmp/client.err
        ) >/dev/console
      """)

  '';
}
