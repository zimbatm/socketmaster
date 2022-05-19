{ config, lib, name, pkgs, systemConfig, ... }:
let
  inherit (lib)
    mkOption
    types
    ;
  inherit (types)
    str
    ;

  etcPath = "/etc/${etcSubPath}";
  etcSubPath = "socketmaster/services/${name}.yaml";

  format = pkgs.formats.yaml { };

  settingsModule = {
    # Forward compatibility, supporting newer `package`'s config items.
    freeformType = format.type;

    options = {
      command = mkOption {
        description = ''
          Program to start
        '';
        type = str;
        # no default
      };

    };
  };

  configFile = format.generate "socketmaster-service-${name}.yaml" config.settings;

in
{
  options = {

    settings = mkOption {
      type = types.submodule settingsModule;
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

    environmentConfig = mkOption {
      internal = true;
      # https://github.com/NixOS/nixpkgs/pull/163617
      type = types.deferredModule or types.raw or types.unspecified;
    };
  };

  config = {
    systemdServiceModule = { ... }: {
      # Avoid automatic restarts.
      stopIfChanged = false;
      restartIfChanged = false;
      # Reload when the config changes.
      reloadTriggers = [ configFile ];
      reload = "kill -HUP $MAINPID";
      serviceConfig.ExecStart = [
        "${systemConfig.socketmaster.package}/bin/socketmaster -config-file ${lib.escapeShellArg etcPath} -start ${toString config.startMillis} -listen fd://3"
      ];
    };
    environmentConfig = {
      etc."${etcSubPath}" = {
        source = configFile;
      };
    };
  };
}
