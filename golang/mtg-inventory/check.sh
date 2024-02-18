#!/bin/sh
set -e

go build ./...

go test ./...

golangci-lint run ./...

govulncheck github.com/benrm/mtg-inventory/golang/mtg-inventory
