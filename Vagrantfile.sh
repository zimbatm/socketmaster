#!/bin/sh
# Vagrant bootstrap script

apt-get update
DEBIAN_FRONTEND='noninteractive' apt-get install -qy git golang ruby1.9.3 build-essential
gem install fpm ronn -q --no-ri --no-rdoc

