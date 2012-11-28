# -*- mode: ruby -*-
# vi: set ft=ruby :

Vagrant::Config.run do |config|
  config.vm.box = "ec2-precise64"
  config.vm.box_url = "https://s3.amazonaws.com/mediacore-public/boxes/ec2-precise64.box"
  config.vm.provision :shell, :path => "Vagrantfile.sh"
end
