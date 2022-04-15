{ config, lib, ... }:
let
  inherit (lib)
    mkOption
    mkIf
    types
    ;

  cfg = config.socketmaster;

in
{
  options.socketmaster = {
    # TODO
  };
  config = mkIf (cfg != { }) {
    # TODO
    systemd.services = { };
  };
}
