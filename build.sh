#!/bin/sh

set -x
go build -x -ldflags "-X main.timestamp=$(date --rfc-3339=seconds | tr ' ' '_') -X main.version=$(git describe --always --tags --dirty)"
