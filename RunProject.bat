@echo off
cd /d "%~dp0"
powershell.exe -NoProfile -ExecutionPolicy Bypass -File "%~dp0Setup.ps1"
if errorlevel 1 (
    echo Go prerequisites failed. Nothing was installed or modified.
    exit /b 1
)
go run hello.go
exit /b %errorlevel%
