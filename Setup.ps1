# Go installation parameters
$ErrorActionPreference = "Stop"
$goVersion = "1.21.5"
$goMsi = "go$goVersion.windows-amd64.msi"
$goUrl = "https://golang.org/dl/$goMsi"
$goPath = "$env:USERPROFILE\go"
$goRoot = "C:\Program Files\Go"

Write-Host "🚀 Setting up Go development environment..." -ForegroundColor Cyan

try {
    # Download Go installer
    Write-Host "📥 Downloading Go $goVersion..." -ForegroundColor Yellow
    Invoke-WebRequest -Uri $goUrl -OutFile $goMsi

    # Install Go
    Write-Host "⚙️ Installing Go..." -ForegroundColor Yellow
    $process = Start-Process msiexec -ArgumentList "/i `"$goMsi`" /quiet" -Wait -PassThru
    if ($process.ExitCode -ne 0) {
        throw "Go installation failed with exit code $($process.ExitCode)"
    }

    # Set up environment variables
    Write-Host "🔧 Setting up environment variables..." -ForegroundColor Yellow
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if (-not $currentPath.Contains($goRoot)) {
        [Environment]::SetEnvironmentVariable("Path", "$currentPath;$goRoot\bin", "User")
    }
    [Environment]::SetEnvironmentVariable("GOPATH", $goPath, "User")
    
    # Refresh environment variables in current session
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path", "Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path", "User")
    $env:GOPATH = $goPath

    # Clean up installer
    Remove-Item $goMsi -ErrorAction SilentlyContinue

    # Test Go installation
    $goVersion = & "$goRoot\bin\go.exe" version
    Write-Host "✅ Go installed successfully: $goVersion" -ForegroundColor Green

    # Initialize the Go module
    Write-Host "📦 Initializing Go module..." -ForegroundColor Yellow
    & "$goRoot\bin\go.exe" mod init hello
    & "$goRoot\bin\go.exe" mod tidy

    Write-Host "`n✅ Setup complete!" -ForegroundColor Green
    Write-Host "You can now run the project with: go run hello.go" -ForegroundColor Cyan

    # Run the project
    $response = Read-Host "Would you like to run the project now? (y/n)"
    if ($response -eq 'y') {
        Write-Host "`n🚀 Running project..." -ForegroundColor Yellow
        & "$goRoot\bin\go.exe" run hello.go
    }
} catch {
    Write-Host "❌ Error: $_" -ForegroundColor Red
    exit 1
}
