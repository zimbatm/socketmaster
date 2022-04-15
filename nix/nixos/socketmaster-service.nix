{ config, lib, systemConfig, ... }:
let
  inherit (lib)
    mkOption
    types
    ;
  inherit (types)
    str
    ;
in
{
  options = {

    command = mkOption {
      description = ''
        Program to start
      '';
      type = str;
      # no default
    };

    startMillis = mkOption {
      description = ''
        How long the new process takes to boot in milliseconds.
      '';
      default = 3000;
    };

    ### Internal ###

    systemdServiceModule = mkOption {
      internal = true;
      # https://github.com/NixOS/nixpkgs/pull/163617
      type = types.deferredModule or types.raw or types.unspecified;
    };
  };

  config = {
    systemdServiceModule = { ... }: {
      # Avoid automatic restarts. This will trigger a reload instead.
      reloadIfChanged = true;
      serviceConfig.ExecStart = [
        "${systemConfig.socketmaster.package}/bin/socketmaster -command ${lib.escapeShellArg config.command} -start ${toString config.startMillis} -listen fd://3"
      ];
    };
  };
}
