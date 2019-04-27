#!/usr/local/bin/pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

go build -o gengo.exe ./cmd/gengo
exitIfFailed
.\gengo.exe
Remove-Item .\gengo.exe
