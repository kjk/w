#!/usr/local/bin/pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

go test github.com/kjk/w -benchmem -bench .
exitIfFailed
