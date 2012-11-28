#!/bin/sh
# Vagrant bootstrap script

export DEBIAN_FRONTEND='noninteractive'
apt-get install -qy git golang ruby1.9.3 build-essential
gem install fpm -q --no-ri --no-rdoc
