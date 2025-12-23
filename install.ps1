# Bucket Labs Windows Installer
$ErrorActionPreference = "Stop"

Write-Host "Installing Bucket CLI..." -ForegroundColor Cyan

# Detect architecture
$arch = if ([Environment]::Is64BitOperatingSystem) { "x86_64" } else { "i686" }

# Download URL
$version = "latest"
$url = "https://github.com/bucketlabs-dot-org/bucket/releases/download/$version/bucket-windows-$arch.exe"

# Install location
$installDir = "$env:LOCALAPPDATA\Bucket"
$exePath = "$installDir\bucket.exe"

# Create directory
New-Item -ItemType Directory -Force -Path $installDir | Out-Null

# Download
Write-Host "Downloading from $url..." -ForegroundColor Yellow
Invoke-WebRequest -Uri $url -OutFile $exePath

# Add to PATH
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$installDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$installDir", "User")
    Write-Host "Added to PATH. Restart terminal to use 'bucket' command." -ForegroundColor Green
}

Write-Host "✓ Bucket CLI installed successfully!" -ForegroundColor Green
Write-Host "Run 'bucket --version' to verify." -ForegroundColor Cyan
