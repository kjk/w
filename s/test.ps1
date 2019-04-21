#!/usr/local/bin/pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

# go build -o gen.exe
# exitIfFailed
# Remove-Item .\gen.exe

go test -v github.com\kjk\winapigen\w
exitIfFailed
# Remove-Item .\testw.exe
