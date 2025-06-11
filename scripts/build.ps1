# Build script for Nehonix WatchTower
param (
    [string]$Version = "1.0.0",
    [string]$BuildType = "release",
    [switch]$Clean = $false
)

# Script variables
$ErrorActionPreference = "Stop"
$BuildDir = "dist"
$BinaryName = "watchtower"
if ($IsWindows) {
    $BinaryName += ".exe"
}

# Print build info
Write-Host "`nBuilding Nehonix WatchTower" -ForegroundColor Cyan
Write-Host "================================" -ForegroundColor Cyan
Write-Host "Version: $Version"
Write-Host "Build Type: $BuildType"
Write-Host "Output: $BuildDir\$BinaryName`n"

# Ensure we're in the project root
$ProjectRoot = $PSScriptRoot
if ($ProjectRoot -match "scripts$") {
    $ProjectRoot = Split-Path $ProjectRoot -Parent
}
Set-Location $ProjectRoot

# Clean build directory if requested
if ($Clean -and (Test-Path $BuildDir)) {
    Write-Host "Cleaning build directory..." -ForegroundColor Yellow
    Remove-Item -Recurse -Force $BuildDir
}

# Create build directory if it doesn't exist
if (-not (Test-Path $BuildDir)) {
    New-Item -ItemType Directory -Path $BuildDir | Out-Null
}

# Build flags
$BuildFlags = @(
    "-ldflags",
    "`"-X 'main.Version=$Version' -s -w`""
)

if ($BuildType -eq "debug") {
    $BuildFlags += "-gcflags"
    $BuildFlags += "all=-N -l"
    $env:CGO_ENABLED = "1"
} else {
    $env:CGO_ENABLED = "0"
}

try {
    # Run tests first
    Write-Host "Running tests..." -ForegroundColor Yellow
    go test ./... -v
    if ($LASTEXITCODE -ne 0) {
        throw "Tests failed"
    }
    Write-Host "Tests passed`n" -ForegroundColor Green

    # Build the binary
    Write-Host "Building binary..." -ForegroundColor Yellow
    $OutputPath = Join-Path $BuildDir $BinaryName
    
    # Build command
    $BuildCmd = "go build"
    $BuildCmd += " -o `"$OutputPath`""
    foreach ($flag in $BuildFlags) {
        $BuildCmd += " $flag"
    }
    $BuildCmd += " ."
    
    Write-Host "Build command: $BuildCmd`n" -ForegroundColor Gray
    
    Invoke-Expression $BuildCmd
    if ($LASTEXITCODE -ne 0) {
        throw "Build failed"
    }

    # Verify the binary exists and is executable
    if (-not (Test-Path $OutputPath)) {
        throw "Binary not found at $OutputPath"
    }

    # Get binary size
    $BinarySize = (Get-Item $OutputPath).Length
    $BinarySizeMB = [math]::Round($BinarySize / 1MB, 2)

    Write-Host "`nBuild successful!" -ForegroundColor Green
    Write-Host "Binary: $OutputPath"
    Write-Host "Size: $BinarySizeMB MB"

    # Run the binary with --version if it supports it
    try {
        $VersionOutput = & $OutputPath --version 2>&1
        Write-Host "Version: $VersionOutput"
    } catch {
        Write-Host "Note: --version flag not supported" -ForegroundColor Yellow
    }

} catch {
    Write-Host "`nBuild failed: $_" -ForegroundColor Red
    exit 1
}

Write-Host "`nBuild complete!`n" -ForegroundColor Green
