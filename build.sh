#!/bin/bash
#
# Build zmon for multiple platforms.

set -eu

mkdir -p ~/www/Linux/i686
export GOARCH=386
go build
cp zmon ~/www/Linux/i686

mkdir -p ~/www/Linux/x86_64
export GOARCH=amd64
go build
cp zmon ~/www/Linux/x86_64

chmod a+rx ~/www/Linux -R
