#!/bin/sh
# Vagrant bootstrap script

apt-get update
DEBIAN_FRONTEND='noninteractive' apt-get install -qy git ruby1.9.3 build-essential
aptitude purge -y golang golang-go golang-doc golang-src
gem install fpm ronn -q --no-ri --no-rdoc

RELEASE=go1.1.1.linux-amd64.tar.gz
if ! which go >/dev/null 2>&1 ; then
  cd /usr/local
  rm -f $RELEASE
  wget -nv http://go.googlecode.com/files/$RELEASE
  tar xzf $RELEASE
  echo 'export PATH="/usr/local/go/bin:$PATH"' > /etc/profile.d/golang.sh
fi
