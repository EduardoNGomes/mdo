#Requires -Version 5.1
#Requires -RunAsAdministrator

$ErrorActionPreference = "Stop"
$Installer = Join-Path (Split-Path $PSScriptRoot -Parent) "install-windows.ps1"
$RunningOnWindows = $env:OS -eq "Windows_NT"
$TestRoot = Join-Path ([IO.Path]::GetTempPath()) ("mdo-installer-test-" + [guid]::NewGuid())
$OriginalProgramFiles = $env:ProgramFiles
$OriginalArchitecture = $env:PROCESSOR_ARCHITECTURE
$OriginalProcessPath = [Environment]::GetEnvironmentVariable("Path", "Process")
$OriginalUserPath = [Environment]::GetEnvironmentVariable("Path", "User")
$OriginalMachinePath = [Environment]::GetEnvironmentVariable("Path", "Machine")
$InstallerTestState = @{
    DownloadShouldFail = $false
    Downloads = New-Object 'System.Collections.Generic.List[string]'
    FixtureArchive = $null
}

function Assert-True {
    param([bool]$Condition, [string]$Message)
    if (-not $Condition) {
        throw $Message
    }
}

# Keep downloads local while exercising real archive extraction, copying and PATH writes.
function Invoke-WebRequest {
    [CmdletBinding()]
    param([string]$Uri, [string]$OutFile, [switch]$UseBasicParsing)

    $InstallerTestState.Downloads.Add($OutFile)
    if ($InstallerTestState.DownloadShouldFail) {
        Write-Error "Simulated download failure"
        return
    }
    Copy-Item -LiteralPath $InstallerTestState.FixtureArchive -Destination $OutFile
}

function Invoke-TestInstallation {
    $Messages = New-Object 'System.Collections.Generic.List[object]'
    $Failure = $null
    $PreviousPreference = $ErrorActionPreference
    try {
        # The installer must stop on failures even under PowerShell's default preference.
        $ErrorActionPreference = "Continue"
        & $Installer | ForEach-Object { $Messages.Add($_) }
    } catch {
        $Failure = $_
    } finally {
        $ErrorActionPreference = $PreviousPreference
    }

    foreach ($Download in $InstallerTestState.Downloads) {
        Assert-True (-not (Test-Path -LiteralPath (Split-Path $Download -Parent))) "Temporary download directory was not removed"
    }
    $InstallerTestState.Downloads.Clear()

    [PSCustomObject]@{ Messages = $Messages; Failure = $Failure }
}

try {
    New-Item -ItemType Directory -Path $TestRoot | Out-Null
    $env:ProgramFiles = Join-Path $TestRoot "Program Files"
    $env:PROCESSOR_ARCHITECTURE = "AMD64"
    New-Item -ItemType Directory -Path $env:ProgramFiles | Out-Null
    $ExpectedInstallPath = Join-Path $env:ProgramFiles "mdo"
    $ExpectedTarget = Join-Path $ExpectedInstallPath "mdo.exe"
    $FixtureBinary = Join-Path $TestRoot "mdo.exe"
    [IO.File]::WriteAllText($FixtureBinary, "installer test fixture")
    $InstallerTestState.FixtureArchive = Join-Path $TestRoot "package.zip"
    Compress-Archive -LiteralPath $FixtureBinary -DestinationPath $InstallerTestState.FixtureArchive

    $ExistingPath = Join-Path $TestRoot "Existing tools"
    New-Item -ItemType Directory -Path $ExistingPath | Out-Null
    [IO.File]::WriteAllText((Join-Path $ExistingPath "mdo.exe"), "older installation")
    $ExpectedMachinePath = $ExistingPath + ";" + $ExpectedInstallPath
    $ExpectedProcessPath = $ExpectedInstallPath + ";" + $ExistingPath
    [Environment]::SetEnvironmentVariable("Path", $ExistingPath, "Process")
    if ($RunningOnWindows) {
        [Environment]::SetEnvironmentVariable("Path", $ExistingPath, "Machine")
    }

    $Result = Invoke-TestInstallation
    Assert-True ($null -eq $Result.Failure) "First installation failed: $($Result.Failure)"
    Assert-True (Test-Path -LiteralPath $ExpectedTarget -PathType Leaf) "Executable was not installed"
    Assert-True ([Environment]::GetEnvironmentVariable("Path", "Process") -eq $ExpectedProcessPath) "Current session PATH was not updated or existing entries changed"
    if ($RunningOnWindows) {
        Assert-True ([Environment]::GetEnvironmentVariable("Path", "Machine") -eq $ExpectedMachinePath) "System PATH was not persisted"
        Assert-True ([Environment]::GetEnvironmentVariable("Path", "User") -ceq $OriginalUserPath) "Installation changed user PATH"
        Assert-True ((Get-Command mdo -CommandType Application).Source -eq $ExpectedTarget) "mdo is not available in the current terminal"
    }
    Write-Output "PASS: first installation preserves PATH entries and updates the current session"

    $Result = Invoke-TestInstallation
    Assert-True ($null -eq $Result.Failure) "Reinstallation failed: $($Result.Failure)"
    Assert-True ([Environment]::GetEnvironmentVariable("Path", "Process") -eq $ExpectedProcessPath) "Reinstallation duplicated the current session PATH entry"
    if ($RunningOnWindows) {
        Assert-True ([Environment]::GetEnvironmentVariable("Path", "Machine") -eq $ExpectedMachinePath) "Reinstallation duplicated the system PATH entry"
    }
    Write-Output "PASS: reinstallation does not duplicate PATH entries"

    # A terminal opened before installation still needs updating even if Machine PATH is correct.
    [Environment]::SetEnvironmentVariable("Path", $ExistingPath, "Process")
    $Result = Invoke-TestInstallation
    Assert-True ($null -eq $Result.Failure) "Installation in a stale terminal failed: $($Result.Failure)"
    Assert-True ([Environment]::GetEnvironmentVariable("Path", "Process") -eq $ExpectedProcessPath) "An existing system PATH entry prevented the current session from being updated"
    Write-Output "PASS: an already registered system PATH does not leave the current session stale"

    # These forms refer to the same directory and must not create duplicate system entries.
    foreach ($EquivalentEntry in @(($ExpectedInstallPath.ToUpperInvariant() + '\'), ('"' + $ExpectedInstallPath + '"'), '%ProgramFiles%\mdo')) {
        if ($RunningOnWindows) {
            [Environment]::SetEnvironmentVariable("Path", $ExistingPath + ";" + $EquivalentEntry, "Machine")
            $PathBefore = [Environment]::GetEnvironmentVariable("Path", "Machine")
            [Environment]::SetEnvironmentVariable("Path", $ExistingPath, "Process")
            $Result = Invoke-TestInstallation
            Assert-True ($null -eq $Result.Failure) "Installation with an equivalent PATH entry failed: $($Result.Failure)"
            Assert-True ([Environment]::GetEnvironmentVariable("Path", "Machine") -ceq $PathBefore) "Equivalent system PATH entry was duplicated or rewritten: $EquivalentEntry"
            Assert-True ([Environment]::GetEnvironmentVariable("Path", "Process") -eq $ExpectedProcessPath) "Equivalent system PATH entry left the current session stale"
        }
    }
    if ($RunningOnWindows) {
        Write-Output "PASS: system PATH recognizes case, trailing separators, quotes and environment variables"
    } else {
        Write-Output "SKIP: persistent system PATH and Windows command lookup require Windows"
    }

    [Environment]::SetEnvironmentVariable("Path", $null, "Process")
    if ($RunningOnWindows) {
        [Environment]::SetEnvironmentVariable("Path", $null, "Machine")
    }
    $Result = Invoke-TestInstallation
    Assert-True ($null -eq $Result.Failure) "Installation with an empty PATH failed: $($Result.Failure)"
    Assert-True ([Environment]::GetEnvironmentVariable("Path", "Process") -eq $ExpectedInstallPath) "Empty current session PATH was not initialized"
    if ($RunningOnWindows) {
        Assert-True ([Environment]::GetEnvironmentVariable("Path", "Machine") -eq $ExpectedInstallPath) "Empty system PATH was not initialized"
    }
    Write-Output "PASS: empty PATH is initialized without an empty entry"

    $InstallerTestState.DownloadShouldFail = $true
    [Environment]::SetEnvironmentVariable("Path", $ExistingPath, "Process")
    $Result = Invoke-TestInstallation
    Assert-True ($null -ne $Result.Failure) "A download failure did not stop installation"
    Assert-True (-not ($Result.Messages -match "Installation complete")) "Failed installation reported success"
    Assert-True ([Environment]::GetEnvironmentVariable("Path", "Process") -eq $ExistingPath) "Failed installation changed PATH"
    if ($RunningOnWindows) {
        Assert-True ([Environment]::GetEnvironmentVariable("Path", "Machine") -eq $ExpectedInstallPath) "Failed installation changed system PATH"
        Assert-True ([Environment]::GetEnvironmentVariable("Path", "User") -ceq $OriginalUserPath) "Installation changed user PATH"
    }
    Assert-True ([IO.File]::ReadAllText($ExpectedTarget) -eq "installer test fixture") "Failed update changed the existing executable"
    Write-Output "PASS: failures stop installation before updating PATH or reporting success"
} finally {
    [Environment]::SetEnvironmentVariable("Path", $OriginalProcessPath, "Process")
    if ($RunningOnWindows) {
        [Environment]::SetEnvironmentVariable("Path", $OriginalMachinePath, "Machine")
    }
    $env:ProgramFiles = $OriginalProgramFiles
    $env:PROCESSOR_ARCHITECTURE = $OriginalArchitecture
    if (Test-Path -LiteralPath $TestRoot) {
        Remove-Item -LiteralPath $TestRoot -Recurse -Force
    }
}
