#!/bin/bash
#

BUILD_DIR="build"
BIN_NAME="ssh-forward"
echo "rm -rf $BUILD_DIR/*"
rm -rf $BUILD_DIR/*

for os in darwin linux;do
  for arch in amd64 arm64;do
    export GOOS=$os
    export GOARCH=$arch
    echo "go build -o $BUILD_DIR/$BIN_NAME"
    go build -o $BUILD_DIR/$BIN_NAME
    cd $BUILD_DIR && tar czf $GOOS-$GOARCH.tar.gz $BIN_NAME && cd - || return
  done
done
