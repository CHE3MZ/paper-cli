#!/usr/bin/env bash

set -o pipefail

# a simple bugcheck script to assign the execution of to a jenkins pipeline etc.
# checks project for bugs and logs it into the logs/ folder.

mkdir -p logs

golangci-lint run ./... 2>&1 | tee logs/go-bugcheck-golangci-lint.txt

gopls check ./... 2>&1 | tee logs/go-bugcheck-gopls.txt

govulncheck ./... 2>&1 | tee logs/go-bugcheck-govulncheck.txt

staticcheck ./... 2>&1 | tee logs/go-bugcheck-staticcheck.txt

go vet ./... 2>&1 | tee logs/go-bugcheck-vet.txt

go test ./... 2>&1 | tee logs/go-bugcheck-test.txt

go test -race ./... 2>&1 | tee logs/go-bugcheck-race.txt