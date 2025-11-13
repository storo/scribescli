# ScribesAI - PulseAudio Installation Script for Windows
# This script must be run as Administrator

#Requires -RunAsAdministrator

param(
    [string]$InstallPath = "C:\Program Files\pulse",
    [switch]$AutoStart = $true,
    [switch]$ConfigureFirewall = $true
)

$ErrorActionPreference = "Stop"

Write-Host "════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "  ScribesAI - PulseAudio Setup for WSL2" -ForegroundColor Cyan
Write-Host "════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""

# Step 1: Download PulseAudio
Write-Host "[1/6] Downloading PulseAudio..." -ForegroundColor Yellow

$pulseUrl = "https://www.freedesktop.org/software/pulseaudio/releases/pulseaudio-1.1.zip"
$zipPath = "$env:TEMP\pulseaudio.zip"

try {
    Invoke-WebRequest -Uri $pulseUrl -OutFile $zipPath -UseBasicParsing
    Write-Host "  ✓ Downloaded successfully" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Download failed: $_" -ForegroundColor Red
    Write-Host ""
    Write-Host "Alternative: Download manually from:" -ForegroundColor Yellow
    Write-Host "  $pulseUrl" -ForegroundColor White
    exit 1
}

# Step 2: Extract PulseAudio
Write-Host "[2/6] Extracting PulseAudio..." -ForegroundColor Yellow

# Create installation directory
if (Test-Path $InstallPath) {
    Write-Host "  ! Installation directory already exists" -ForegroundColor Yellow
    $response = Read-Host "  Overwrite? (y/N)"
    if ($response -ne "y") {
        Write-Host "Installation cancelled" -ForegroundColor Red
        exit 1
    }
    Remove-Item -Path $InstallPath -Recurse -Force
}

New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null

try {
    Expand-Archive -Path $zipPath -DestinationPath $InstallPath -Force

    # PulseAudio 1.1 extracts to a subdirectory, move contents up
    $subDir = Get-ChildItem -Path $InstallPath -Directory | Select-Object -First 1
    if ($subDir) {
        Get-ChildItem -Path $subDir.FullName -Recurse | Move-Item -Destination $InstallPath -Force
        Remove-Item -Path $subDir.FullName -Recurse -Force
    }

    Write-Host "  ✓ Extracted to $InstallPath" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Extraction failed: $_" -ForegroundColor Red
    exit 1
}

# Step 3: Configure PulseAudio
Write-Host "[3/6] Configuring PulseAudio..." -ForegroundColor Yellow

$configPath = Join-Path $InstallPath "etc\pulse"

# Backup existing configs
if (Test-Path "$configPath\default.pa") {
    Copy-Item "$configPath\default.pa" "$configPath\default.pa.backup" -Force
}
if (Test-Path "$configPath\daemon.conf") {
    Copy-Item "$configPath\daemon.conf" "$configPath\daemon.conf.backup" -Force
}

# Modify default.pa
$defaultPa = Get-Content "$configPath\default.pa" -Raw
if ($defaultPa -notmatch "module-native-protocol-tcp") {
    $tcpConfig = @"

# ScribesAI - WSL2 Configuration
load-module module-native-protocol-tcp port=4713 auth-ip-acl=172.16.0.0/12
load-module module-esound-protocol-tcp port=4714 auth-ip-acl=172.16.0.0/12

"@
    $defaultPa = $tcpConfig + $defaultPa
}

# Fix module-waveout line
$defaultPa = $defaultPa -replace "load-module module-waveout sink_name=output source_name=input$", "load-module module-waveout sink_name=output source_name=input record=0"

Set-Content -Path "$configPath\default.pa" -Value $defaultPa -Force

# Modify daemon.conf
$daemonConf = Get-Content "$configPath\daemon.conf" -Raw
$daemonConf = $daemonConf -replace "; exit-idle-time = 20", "exit-idle-time = -1"
$daemonConf = $daemonConf -replace "exit-idle-time = 20", "exit-idle-time = -1"

# Add if not present
if ($daemonConf -notmatch "exit-idle-time") {
    $daemonConf += "`nexit-idle-time = -1`n"
}

Set-Content -Path "$configPath\daemon.conf" -Value $daemonConf -Force

Write-Host "  ✓ Configuration files updated" -ForegroundColor Green

# Step 4: Configure Windows Firewall
if ($ConfigureFirewall) {
    Write-Host "[4/6] Configuring Windows Firewall..." -ForegroundColor Yellow

    $pulseExe = Join-Path $InstallPath "bin\pulseaudio.exe"

    try {
        # Remove existing rules if any
        Remove-NetFirewallRule -DisplayName "PulseAudio for WSL2" -ErrorAction SilentlyContinue

        # Add new rule
        New-NetFirewallRule -DisplayName "PulseAudio for WSL2" `
                            -Direction Inbound `
                            -Program $pulseExe `
                            -Action Allow `
                            -Profile Private,Domain `
                            -Protocol TCP `
                            -LocalPort 4713,4714 | Out-Null

        Write-Host "  ✓ Firewall rule created" -ForegroundColor Green
    } catch {
        Write-Host "  ! Firewall configuration failed: $_" -ForegroundColor Yellow
        Write-Host "  You may need to allow pulseaudio.exe manually" -ForegroundColor Yellow
    }
} else {
    Write-Host "[4/6] Skipping firewall configuration" -ForegroundColor Gray
}

# Step 5: Create startup script
if ($AutoStart) {
    Write-Host "[5/6] Configuring auto-start..." -ForegroundColor Yellow

    $startupPath = [Environment]::GetFolderPath('Startup')
    $batchPath = Join-Path $startupPath "pulseaudio-wsl.bat"

    $batchContent = @"
@echo off
REM ScribesAI - Start PulseAudio for WSL2
start /B "" "$InstallPath\bin\pulseaudio.exe"
"@

    Set-Content -Path $batchPath -Value $batchContent -Force
    Write-Host "  ✓ Auto-start configured" -ForegroundColor Green
    Write-Host "    PulseAudio will start automatically on login" -ForegroundColor Gray
} else {
    Write-Host "[5/6] Skipping auto-start configuration" -ForegroundColor Gray
}

# Step 6: Start PulseAudio
Write-Host "[6/6] Starting PulseAudio..." -ForegroundColor Yellow

$pulseExe = Join-Path $InstallPath "bin\pulseaudio.exe"

# Kill any existing pulseaudio processes
Get-Process -Name pulseaudio -ErrorAction SilentlyContinue | Stop-Process -Force

# Start PulseAudio
try {
    Start-Process -FilePath $pulseExe -WindowStyle Hidden
    Start-Sleep -Seconds 2

    # Verify it's running
    $process = Get-Process -Name pulseaudio -ErrorAction SilentlyContinue
    if ($process) {
        Write-Host "  ✓ PulseAudio is running (PID: $($process.Id))" -ForegroundColor Green
    } else {
        Write-Host "  ! PulseAudio may not have started correctly" -ForegroundColor Yellow
    }
} catch {
    Write-Host "  ✗ Failed to start PulseAudio: $_" -ForegroundColor Red
    Write-Host "  Try running manually: $pulseExe" -ForegroundColor Yellow
}

# Cleanup
Remove-Item -Path $zipPath -Force -ErrorAction SilentlyContinue

Write-Host ""
Write-Host "════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host "  Installation Complete!" -ForegroundColor Green
Write-Host "════════════════════════════════════════════════════════" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps in WSL2:" -ForegroundColor Yellow
Write-Host ""
Write-Host "  1. Open your WSL2 terminal" -ForegroundColor White
Write-Host "  2. Run: ./scribescli setup-audio" -ForegroundColor White
Write-Host "     Or manually add to ~/.bashrc:" -ForegroundColor White
Write-Host ""
Write-Host "       export PULSE_SERVER=tcp:`$(ip route show default | awk '{print `$3}')" -ForegroundColor Cyan
Write-Host ""
Write-Host "  3. Reload: source ~/.bashrc" -ForegroundColor White
Write-Host "  4. Test: ./scribescli --test-audio" -ForegroundColor White
Write-Host ""
Write-Host "For troubleshooting, see: WSL_AUDIO_GUIDE.md" -ForegroundColor Gray
Write-Host ""

# Pause if run interactively
if ($Host.Name -eq "ConsoleHost") {
    Write-Host "Press any key to exit..." -ForegroundColor Gray
    $null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")
}
