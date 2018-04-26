#!/usr/local/bin/pwsh

go build -o winapigen.exe
.\winapigen.exe
Remove-Item winapigen.exe
