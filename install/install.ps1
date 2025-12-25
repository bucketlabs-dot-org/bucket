# Bucket Labs Windows Installer
$ErrorActionPreference = "Stop"

# Colors for output (like Bash's info/error)
function Write-Info { param([string]$Message); Write-Host $Message -ForegroundColor Cyan }
function Write-Warn { param([string]$Message); Write-Host $Message -ForegroundColor Yellow }
function Write-Success { param([string]$Message); Write-Host $Message -ForegroundColor Green }

Write-Info "Installing Bucket CLI..."

# --- Load VERSION (like Bash) ---
$VERSION_URL = "https://raw.githubusercontent.com/bucketlabs-dot-org/bucket/refs/heads/main/install/VERSION"
try {
    $version = (Invoke-WebRequest -Uri $VERSION_URL -UseBasicParsing).Content.Trim()
    if ([string]::IsNullOrWhiteSpace($version)) { throw "Empty response" }
    Write-Info "Using VERSION: $version"
} catch {
    $version = "latest"
    Write-Warn "Couldn't fetch VERSION from $VERSION_URL (using 'latest' fallback)"
}

# Detect architecture
$arch = if ([Environment]::Is64BitOperatingSystem) { "x86_64" } else { "i686" }

# Download URL (dynamic tag prefix)
$releaseTag = if ($version -eq "latest") { "latest" } else { "v$version" }
$url = "https://github.com/bucketlabs-dot-org/bucket/releases/download/$releaseTag/bucket-windows-$arch.exe"

# Install location
$installDir = "$env:LOCALAPPDATA\Bucket"
$exePath = "$installDir\bucket.exe"

# Create directory
New-Item -ItemType Directory -Force -Path $installDir | Out-Null

# Download
Write-Info "Downloading from $url..."
try {
    Invoke-WebRequest -Uri $url -OutFile $exePath
} catch {
    Write-Warn "Download failed: $($_.Exception.Message)"
    exit 1
}

# Add to PATH
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    Write-Success "Added to PATH. Restart terminal to use 'bucket' command."
}

Write-Success "✓ Bucket CLI installed successfully!"
Write-Info "Run 'bucket --version' to verify."
