#!/bin/bash

# build and replace linux binary
env GOOS=linux GOARCH=amd64 go build .
mv azure-update-mgmt bin/linux

# build and replace windows binary
env GOOS=windows GOARCH=amd64 go build .
mv azure-update-mgmt.exe bin/windows
