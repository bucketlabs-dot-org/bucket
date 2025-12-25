# Bucket Labs Windows Installer
$ErrorActionPreference = "Stop"

function Write-Info { param([string]$Message); Write-Host $Message -ForegroundColor Cyan }
function Write-Warn { param([string]$Message); Write-Host $Message -ForegroundColor Yellow }
function Write-Success { param([string]$Message); Write-Host $Message -ForegroundColor Green }

Write-Info "Installing Bucket CLI..."

$VERSION_URL = "https://raw.githubusercontent.com/bucketlabs-dot-org/bucket/refs/heads/main/install/VERSION"
try {
    $version = (Invoke-WebRequest -Uri $VERSION_URL -UseBasicParsing).Content.Trim()
    if ([string]::IsNullOrWhiteSpace($version)) { throw "Empty response" }
    Write-Info "Using VERSION: $version"
} catch {
    $version = "latest"
    Write-Warn "Couldn't fetch VERSION from $VERSION_URL (using 'latest' fallback)"
}

$rawArch = $env:PROCESSOR_ARCHITECTURE
switch ($rawArch) {
    "AMD64" { $arch = "amd64" }
    "ARM64" { $arch = "arm64" }
    "X86"   { $arch = "386" }  # 32-bit fallback
    default { throw "Unsupported architecture: $rawArch (only amd64/arm64/386 supported)" }
}

Write-Info "Detected ARCH: $arch"

$releaseTag = if ($version -eq "latest") { "latest" } else { "v$version" }
$url = "https://github.com/bucketlabs-dot-org/bucket/releases/download/$releaseTag/bucket-windows-$arch.exe"

$installDir = "$env:LOCALAPPDATA\Bucket"
$exePath = "$installDir\bucket.exe"

New-Item -ItemType Directory -Force -Path $installDir | Out-Null

Write-Info "Downloading from $url..."
try {
    Invoke-WebRequest -Uri $url -OutFile $exePath
} catch {
    Write-Warn "Download failed: $($_.Exception.Message)"
    exit 1
}

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    Write-Success "Added to PATH. Restart terminal to use 'bucket' command."
}

Write-Success "Bucket CLI installed successfully!"
Write-Info "Run 'bucket --version' to verify."
