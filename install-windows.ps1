#Requires -Version 5.1
#Requires -RunAsAdministrator

$ErrorActionPreference = "Stop"

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
$InstallPath = Join-Path $env:ProgramFiles "mdo"
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

    # Persist PATH for all users and update this PowerShell session too.
    # Check each scope independently: Machine may already be correct in a stale terminal.
    foreach ($Scope in @([EnvironmentVariableTarget]::Machine, [EnvironmentVariableTarget]::Process)) {
        $CurrentPath = [Environment]::GetEnvironmentVariable("Path", $Scope)
        $PathEntries = @($CurrentPath -split ";" | ForEach-Object {
            [Environment]::ExpandEnvironmentVariables($_.Trim().Trim('"')).TrimEnd('\', '/')
        })

        if ($PathEntries -contains $InstallPath.TrimEnd('\', '/')) {
            continue
        }

        if ([string]::IsNullOrEmpty($CurrentPath)) {
            $NewPath = $InstallPath
        } elseif ($Scope -eq [EnvironmentVariableTarget]::Process) {
            # Prefer this installation over an older copy in the user's bin directory.
            $NewPath = $InstallPath + ";" + $CurrentPath.TrimStart(";")
        } else {
            $NewPath = $CurrentPath.TrimEnd(";") + ";" + $InstallPath
        }

        [Environment]::SetEnvironmentVariable("Path", $NewPath, $Scope)
        Write-Output "Added $InstallPath to $Scope PATH."
    }

    Write-Output "Installation complete! Run: mdo <file.md>"
} finally {
    if (Test-Path $WorkDir) {
        Remove-Item $WorkDir -Recurse -Force
    }
}
