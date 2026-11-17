# Safe Go project prerequisite check. Never downloads installers or edits PATH/GOPATH.
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"
Set-Location -LiteralPath $PSScriptRoot
$go = Get-Command go -ErrorAction SilentlyContinue
if (-not $go) {
    throw "Go is missing. Install it with: winget install --id GoLang.Go --exact ; reopen PowerShell and rerun."
}
& go version
if ($LASTEXITCODE -ne 0) { throw "Go executable is not healthy." }
if (-not (Test-Path -LiteralPath (Join-Path $PSScriptRoot "go.mod"))) {
    throw "This checkout has no go.mod. Re-clone rather than running go mod init."
}
& go mod download
if ($LASTEXITCODE -ne 0) { throw "Module download failed. Check the first Go error." }
Write-Host "Go and repository dependencies are ready. Start with: go run hello.go"
