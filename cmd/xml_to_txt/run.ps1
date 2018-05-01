#!/usr/local/bin/pwsh
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
function exitIfFailed { if ($LASTEXITCODE -ne 0) { exit } }

Set-Location -Path cmd\xml_to_txt
go build -o xml_to_txt.exe
Set-Location -Path ..\..
exitIfFailed
.\cmd\xml_to_txt\xml_to_txt.exe
Remove-Item .\cmd\xml_to_txt\xml_to_txt.exe
