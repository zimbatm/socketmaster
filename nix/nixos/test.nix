{ nixosTest, socketmaster }:

nixosTest {
  nodes.machine = { ... }: {
    imports = [ ./module.nix ];
    # TODO
    # .package = socketmaster;
  };
  testScript = ''
    machine.succeed("${socketmaster}/bin/socketmaster -help")
  '';
}
