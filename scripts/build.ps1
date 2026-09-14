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

# Bake the CLI version into the binary. $env:PAPER_CLI_VERSION wins (the
# release workflow sets it to the tag being released, so no network call
# happens there); otherwise ask GitHub what the latest release is (local
# git tags can be stale or unpushed, so they are only a fallback);
# otherwise "dev" (plain `go build` without this flag does that).
$Version = $env:PAPER_CLI_VERSION
if ([string]::IsNullOrWhiteSpace($Version)) {
  try {
    $Latest = Invoke-RestMethod -Uri "https://api.github.com/repos/CHE3MZ/paper-cli/releases/latest" -TimeoutSec 10
    if ($null -ne $Latest.tag_name -and -not [string]::IsNullOrWhiteSpace($Latest.tag_name)) { $Version = $Latest.tag_name }
  } catch { $Version = $null }
}
if ([string]::IsNullOrWhiteSpace($Version)) {
  try { $Version = (git describe --tags --abbrev=0 2>$null) } catch { $Version = $null }
  if ([string]::IsNullOrWhiteSpace($Version)) { $Version = "dev" }
}

Write-Host "==> building $Out (Paper CLI version $Version)"
go build -ldflags "-X github.com/CHE3MZ/paper-cli/src.CLIVersion=$Version" -o $Out ./cmd/paper
Write-Host "done: $Out"
