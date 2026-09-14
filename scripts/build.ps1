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
# happens there, and CI sets it to "dev" so test artifacts never claim a
# release version); otherwise ask GitHub what the latest release is (local
# git tags can be stale or unpushed, so they are only a fallback);
# otherwise "dev" (plain `go build` without this flag does that). A dirty
# tree appends "-dirty" so dev builds can't masquerade as releases.
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
# A binary built from a modified tree is not the release its tag names:
# mark it so `paper version` stays honest and `paper update` doesn't
# wrongly report "already up to date". (Build outputs are gitignored, so a
# clean checkout is unaffected.)
if ($Version -ne "dev") {
  $Dirty = $null
  try {
    # NOTE: do not call this $out — PowerShell variables are
    # case-insensitive, so that would clobber $Out (the binary path) above.
    $porcelain = git status --porcelain 2>$null
    if ($LASTEXITCODE -eq 0) { $Dirty = ($porcelain -join "`n").Trim() }
  } catch { $Dirty = $null }
  if (-not [string]::IsNullOrWhiteSpace($Dirty)) { $Version = "$Version-dirty" }
}

Write-Host "==> building $Out (Paper CLI version $Version)"
go build -ldflags "-X github.com/CHE3MZ/paper-cli/src.CLIVersion=$Version" -o $Out ./cmd/paper
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
Write-Host "done: $Out"
