#!/usr/bin/env pwsh
# Build paper from the path in current-build-version.json.
# 1. Runs the embed generator (reads current-build-path).
# 2. Builds ./cmd/paper into build/paper(.exe).
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$Root = Split-Path -Parent $PSScriptRoot
Set-Location $Root

Write-Host "==> regenerating embed from current-build-version.json"
go run ./src/genembed
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

New-Item -ItemType Directory -Force -Path "$Root/build" | Out-Null

$Out = "$Root/build/paper"
if ($env:OS -match "Windows") { $Out += ".exe" }

Write-Host "==> building $Out"
go build -o $Out ./cmd/paper
Write-Host "done: $Out"
