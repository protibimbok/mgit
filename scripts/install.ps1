# Install mgit from GitHub releases.
#
# Usage:
#   irm https://raw.githubusercontent.com/protibimbok/mgit/master/scripts/install.ps1 | iex
#   irm ... | iex; install-mgit -Version v1.0.0
#   irm ... | iex; install-mgit -InstallDir "$env:LOCALAPPDATA\Programs\mgit"

[CmdletBinding()]
param(
    [string]$InstallDir = "",
    [string]$Version = "latest",
    [switch]$SkipPathHint,
    [switch]$SkipDepCheck
)

$ErrorActionPreference = "Stop"

$Repo = "protibimbok/mgit"
$Binary = "mgit"

function Write-Info([string]$Message) {
    Write-Host "[mgit] $Message" -ForegroundColor Green
}

function Write-Warn([string]$Message) {
    Write-Host "[mgit] $Message" -ForegroundColor Yellow
}

function Write-Err([string]$Message) {
    Write-Host "[mgit] error: $Message" -ForegroundColor Red
    exit 1
}

function Get-Arch {
    if ([System.Runtime.InteropServices.RuntimeInformation]::ProcessArchitecture -eq [System.Runtime.InteropServices.Architecture]::Arm64) {
        return "arm64"
    }
    return "amd64"
}

function Get-DefaultInstallDir {
    if ($InstallDir) { return $InstallDir }
    if ($env:INSTALL_DIR) { return $env:INSTALL_DIR }
    return Join-Path $env:LOCALAPPDATA "Programs\mgit"
}

function Get-GitMissingMessage {
@'
mgit: git is not installed or not on your PATH.

  mgit wraps git — install Git first, then restart your terminal.

  Windows:
    winget install Git.Git
    — or — https://git-scm.com/download/win
    Choose "Git from the command line and also from 3rd-party software" during setup.
'@
}

function Get-SSHKeygenMissingMessage {
@'
mgit: ssh-keygen is not installed or not on your PATH.

  mgit gen creates SSH keys using ssh-keygen.

  Windows:
    Settings → Apps → Optional features → Add "OpenSSH Client"
    — or — install Git for Windows (includes ssh-keygen)
    — or — winget install Microsoft.OpenSSH.Beta
    Then restart your terminal.
'@
}

function Test-Dependencies {
    $missing = @()
    if (-not (Get-Command git -ErrorAction SilentlyContinue)) {
        $missing += "git"
    }
    if (-not (Get-Command ssh-keygen -ErrorAction SilentlyContinue)) {
        $missing += "ssh-keygen"
    }
    if ($missing.Count -eq 0) { return }

    Write-Warn "Some prerequisites are missing (mgit needs them for most commands):"
    if ($missing -contains "git") {
        Write-Host ""
        Write-Host (Get-GitMissingMessage)
    }
    if ($missing -contains "ssh-keygen") {
        Write-Host ""
        Write-Host (Get-SSHKeygenMissingMessage)
    }
    Write-Host ""
}

function Resolve-ReleaseTag {
    $ver = $Version
    if ($env:MGIT_VERSION) { $ver = $env:MGIT_VERSION }
    if ($ver -ne "latest") {
        if ($ver -notmatch '^v') { return "v$ver" }
        return $ver
    }
    $release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
    return $release.tag_name
}

function Verify-Checksum([string]$ArchivePath, [string]$ChecksumsPath) {
    $archiveName = Split-Path $ArchivePath -Leaf
    $expected = (Get-Content $ChecksumsPath | Where-Object { $_ -match [regex]::Escape($archiveName) } | ForEach-Object { ($_ -split '\s+')[0] } | Select-Object -First 1)
    if (-not $expected) {
        Write-Err "checksum not found for $archiveName"
    }
    $actual = (Get-FileHash -Algorithm SHA256 -Path $ArchivePath).Hash.ToLower()
    if ($actual -ne $expected.ToLower()) {
        Write-Err "checksum mismatch for $archiveName"
    }
}

function Show-PathHint([string]$Dir) {
    if ($SkipPathHint) { return }

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $parts = $userPath -split ';' | Where-Object { $_ }
    if ($parts -contains $Dir) { return }

    Write-Warn "$Dir is not on your PATH"
    Write-Host ""
    Write-Host "Add it to your user PATH:"
    Write-Host ""
    Write-Host "  [Environment]::SetEnvironmentVariable(""Path"", ""$Dir;"" + [Environment]::GetEnvironmentVariable(""Path"", ""User""), ""User"")"
    Write-Host ""
    Write-Host "Then open a new terminal."
    Write-Host ""
}

$Arch = Get-Arch
$InstallDir = Get-DefaultInstallDir
$Tag = Resolve-ReleaseTag
$Archive = "${Binary}_windows_${Arch}.zip"
$Url = "https://github.com/$Repo/releases/download/$Tag/$Archive"

Write-Info "Installing mgit $Tag (windows/$Arch) to $InstallDir..."

if (-not $SkipDepCheck) {
    Test-Dependencies
}

$TempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("mgit-install-" + [guid]::NewGuid().ToString())
New-Item -ItemType Directory -Path $TempDir -Force | Out-Null

try {
    $ArchivePath = Join-Path $TempDir $Archive
    $ChecksumsPath = Join-Path $TempDir "checksums.txt"

    Write-Info "Downloading $Url"
    Invoke-WebRequest -Uri $Url -OutFile $ArchivePath -UseBasicParsing

    $ChecksumUrl = "https://github.com/$Repo/releases/download/$Tag/checksums.txt"
    Invoke-WebRequest -Uri $ChecksumUrl -OutFile $ChecksumsPath -UseBasicParsing
    Verify-Checksum $ArchivePath $ChecksumsPath
    Write-Info "Checksum verified"

    Expand-Archive -Path $ArchivePath -DestinationPath $TempDir -Force
    $ExePath = Join-Path $TempDir "$Binary.exe"
    if (-not (Test-Path $ExePath)) {
        Write-Err "expected $Binary.exe in archive"
    }

    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
    Copy-Item -Path $ExePath -Destination (Join-Path $InstallDir "$Binary.exe") -Force
}
finally {
    Remove-Item -Recurse -Force $TempDir -ErrorAction SilentlyContinue
}

$Installed = Join-Path $InstallDir "$Binary.exe"
Write-Info "Installed: $Installed"
& $Installed --version

Show-PathHint $InstallDir
Write-Info "Run 'mgit gen' to create your first profile."
