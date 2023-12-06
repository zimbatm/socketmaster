{ config, lib, pkgs, ... }:
let
  inherit (lib)
    mapAttrs
    attrValues
    mkIf
    mkOption
    mkMerge
    types
    ;
  inherit (lib.types)
    attrsOf
    submoduleWith
    ;

  cfg = config.socketmaster;

in
{
  options.socketmaster = {
    package = mkOption {
      description = ''
        The socketmaster package to use.

        Note that services are not automatically restarted
        when this value changes.
      '';
      type = types.package;
      default = pkgs.socketmaster;
      defaultText = "socketmaster";
    };
    services = mkOption {
      description = ''
        A collection of socketmaster-driven services.

        Each entry will be mapped to a systemd service
        with the same name.
      '';
      type = attrsOf (submoduleWith {
        modules = [ ./socketmaster-service.nix ];
        specialArgs.systemConfig = config;
        specialArgs.pkgs = pkgs;
      });
      default = { };
    };
  };
  config = mkIf (cfg.services != { }) {
    systemd.services = mapAttrs (k: v: v.systemdServiceModule) cfg.services;
    environment = mkMerge (attrValues (mapAttrs (k: v: v.environmentConfig) cfg.services));
  };
}
