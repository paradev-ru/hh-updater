#!/usr/bin/env bash
#
# This script builds the application from source.
set -e

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/.." && pwd )"

# Change into that directory
cd $DIR

# Build
echo "--> Building..."
go build \
    -o bin/hh-updater \
    -v
