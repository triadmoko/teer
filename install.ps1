# Teer Windows Installer
# Run: irm https://raw.githubusercontent.com/triadmoko/teer/main/install.ps1 | iex

param(
    [string]$Version = "",
    [string]$InstallDir = "$env:LOCALAPPDATA\Programs\teer"
)

$ErrorActionPreference = "Stop"
$Repo = "triadmoko/teer"
$Asset = "teer-windows-amd64.exe"

function Write-Info  { Write-Host "[teer] $args" -ForegroundColor Green }
function Write-Warn  { Write-Host "[teer] $args" -ForegroundColor Yellow }
function Write-Err   { Write-Host "[teer] $args" -ForegroundColor Red; exit 1 }

# --- resolve version ---
if ($Version -eq "") {
    Write-Info "Fetching latest release..."
    try {
        $Release = Invoke-RestMethod "https://api.github.com/repos/$Repo/releases/latest"
        $Version = $Release.tag_name
    } catch {
        Write-Err "Failed to fetch latest version: $_"
    }
    Write-Info "Latest version: $Version"
} else {
    Write-Info "Installing version: $Version"
}

$DownloadUrl = "https://github.com/$Repo/releases/download/$Version/$Asset"

# --- download ---
Write-Info "Downloading $Asset ..."
$TmpFile = [System.IO.Path]::GetTempFileName() + ".exe"
try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $TmpFile -UseBasicParsing
} catch {
    Write-Err "Download failed: $_"
}

# --- install ---
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

$Dest = Join-Path $InstallDir "teer.exe"
Move-Item -Force $TmpFile $Dest
Write-Info "Installed at $Dest"

# --- add to PATH ---
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$UserPath;$InstallDir", "User")
    Write-Info "Added $InstallDir to user PATH"
    Write-Warn "Restart your terminal for PATH changes to take effect."
}

Write-Info "Done! Run: teer"
