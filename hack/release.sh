#!/bin/bash

# This is i a helper script to generate the release resources

make clean
make build-linux
make build-arm
make build-darwin
make build-windows
