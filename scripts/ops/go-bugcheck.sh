#!/usr/bin/env bash

set -o pipefail

# a simple bugcheck script to assign the execution of to a jenkins pipeline etc.
# checks project for bugs and logs it into the logs/ folder.

# Always run from the repo root regardless of where the script is invoked.
cd "$(dirname "$0")/../.."

mkdir -p logs

golangci-lint run ./... 2>&1 | tee logs/go-bugcheck-golangci-lint.txt

# gopls check cannot take directory patterns on Windows ("Incorrect
# function" reading the dir), so pass explicit files instead. Relative
# paths keep this working with spaces in the repo path and on Linux CI.
gopls check $(find . -name '*.go') 2>&1 | tee logs/go-bugcheck-gopls.txt

govulncheck ./... 2>&1 | tee logs/go-bugcheck-govulncheck.txt

staticcheck ./... 2>&1 | tee logs/go-bugcheck-staticcheck.txt

go vet ./... 2>&1 | tee logs/go-bugcheck-vet.txt

go test ./... 2>&1 | tee logs/go-bugcheck-test.txt
