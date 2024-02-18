#!/bin/sh
set -e

go build ./...

go test ./...

golangci-lint run ./...

ret="$(revive ./...)"
if [ -n "$ret" ] ; then
	revive ./...
	exit 1
fi

govulncheck github.com/benrm/mtg-inventory/golang/mtg-inventory
