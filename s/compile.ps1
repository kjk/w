#!/usr/local/bin/pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

go build -o gen.exe
exitIfFailed
Remove-Item .\gen.exe

go build -o testw.exe github.com\kjk\winapigen\testw
exitIfFailed
Remove-Item .\testw.exe
