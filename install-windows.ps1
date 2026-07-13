$Repo = "EduardoNGomes/mdo"
$Bin = "mdo.exe"

$Arch = $env:PROCESSOR_ARCHITECTURE
switch ($Arch) {
    "AMD64" { $GoArch = "amd64" }
    "ARM64" { $GoArch = "arm64" }
    default {
        Write-Error "Unsupported Windows architecture: $Arch"
        exit 1
    }
}

$File = "mdo-windows-$GoArch.zip"
$Url = "https://github.com/$Repo/releases/latest/download/$File"
$InstallPath = Join-Path $env:USERPROFILE "bin"
$Target = Join-Path $InstallPath $Bin
$WorkDir = Join-Path ([System.IO.Path]::GetTempPath()) ([System.Guid]::NewGuid().ToString())

try {
    New-Item -ItemType Directory -Path $WorkDir | Out-Null

    if (!(Test-Path $InstallPath)) {
        Write-Output "Creating directory $InstallPath..."
        New-Item -ItemType Directory -Path $InstallPath | Out-Null
    }

    Write-Output "Downloading $File..."
    Invoke-WebRequest -Uri $Url -OutFile (Join-Path $WorkDir $File) -UseBasicParsing

    Write-Output "Extracting $File..."
    Expand-Archive -Path (Join-Path $WorkDir $File) -DestinationPath $WorkDir -Force

    if (Test-Path $Target) {
        Write-Output "Updating mdo..."
    } else {
        Write-Output "Installing mdo in $InstallPath..."
    }

    Copy-Item (Join-Path $WorkDir $Bin) $Target -Force
    Write-Output "Installation complete!"

    $UserPath = [Environment]::GetEnvironmentVariable("Path", [EnvironmentVariableTarget]::User)
    if ($null -eq $UserPath) {
        $UserPath = ""
    }

    if (-not ($UserPath -split ";" | Where-Object { $_ -eq $InstallPath })) {
        if ($UserPath.Length -eq 0) {
            $NewPath = $InstallPath
        } else {
            $NewPath = $UserPath.TrimEnd(";") + ";" + $InstallPath
        }

        [Environment]::SetEnvironmentVariable(
            "Path",
            $NewPath,
            [EnvironmentVariableTarget]::User
        )
        Write-Output "Added $InstallPath to PATH. Restart your terminal."
    }
} finally {
    if (Test-Path $WorkDir) {
        Remove-Item $WorkDir -Recurse -Force
    }
}
