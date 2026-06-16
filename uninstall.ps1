# Teer Windows Uninstaller
# Run: irm https://raw.githubusercontent.com/triadmoko/teer/main/uninstall.ps1 | iex

param(
    [switch]$PurgeConfig,
    [string]$InstallDir = "$env:LOCALAPPDATA\Programs\teer"
)

$ErrorActionPreference = "Stop"

function Write-Info { Write-Host "[teer] $args" -ForegroundColor Green }
function Write-Warn { Write-Host "[teer] $args" -ForegroundColor Yellow }
function Write-Err  { Write-Host "[teer] $args" -ForegroundColor Red; exit 1 }

function Remove-IfExists {
    param([string]$Path)
    if (Test-Path $Path) {
        Remove-Item -Force -Recurse $Path
        Write-Info "Dihapus: $Path"
    }
}

$Dest = Join-Path $InstallDir "teer.exe"
$DesktopShortcut = Join-Path ([Environment]::GetFolderPath("Desktop")) "Teer.lnk"
$StartMenuShortcut = Join-Path $env:APPDATA "Microsoft\Windows\Start Menu\Programs\Teer.lnk"

Remove-IfExists $Dest
Remove-IfExists $InstallDir
Remove-IfExists $DesktopShortcut
Remove-IfExists $StartMenuShortcut

$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -and $UserPath -like "*$InstallDir*") {
    $NewPath = ($UserPath -split ';' | Where-Object { $_ -and $_ -ne $InstallDir }) -join ';'
    [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
    Write-Info "Dihapus $InstallDir dari user PATH"
    Write-Warn "Restart terminal agar perubahan PATH aktif."
}

if ($PurgeConfig -or $env:TEER_PURGE_CONFIG -eq "1") {
    Remove-IfExists (Join-Path $env:APPDATA "teer")
} else {
    Write-Warn "Config di %APPDATA%\teer tidak dihapus. Set `$env:TEER_PURGE_CONFIG=1 atau jalankan uninstall.ps1 -PurgeConfig."
}

Write-Info "Teer berhasil di-uninstall."
