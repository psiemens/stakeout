<#
.Synopsis
    Install the Stakeout CLI on Windows.
.DESCRIPTION
    By default, the latest release will be installed.
    If '-Version' is specified, then the given version is installed.
.Parameter Directory
    The destination path to install to.
.Parameter Version
    The version to install.
.Parameter AddToPath
    Add the absolute destination path to the 'User' scope environment variable 'Path'.
.EXAMPLE
    Install the current version
    .\install.ps1
.EXAMPLE
    Install version v0.5.2
    .\install.ps1 -Version v0.5.2
#>
param (
    [string] $version="",
    [string] $directory = "$env:APPDATA\Stakeout",
    [bool] $addToPath = $true
)

Set-StrictMode -Version 3.0

# Enable support for ANSI escape sequences
Set-ItemProperty HKCU:\Console VirtualTerminalLevel -Type DWORD 1

$ErrorActionPreference = "Stop"

$repo = "psiemens/stakeout"
$versionURL = "https://api.github.com/repos/$repo/releases/latest"
$assetsURL = "https://github.com/$repo/releases/download"

if (!$version) {
    $q = (Invoke-WebRequest -Uri "$versionURL" -UseBasicParsing) | ConvertFrom-Json
    $version = $q.tag_name
}

Write-Output("Installing version {0} ..." -f $version)

New-Item -ItemType Directory -Force -Path $directory | Out-Null

$progressPreference = 'silentlyContinue'

Invoke-WebRequest -Uri "$assetsURL/$version/stakeout-x86_64-windows.zip" -UseBasicParsing -OutFile "$directory\stakeout.zip"

Expand-Archive -Path "$directory\stakeout.zip" -DestinationPath "$directory"

Move-Item -Path "$directory\stakeout.exe" -Destination "$directory\stakeout.exe" -Force

if ($addToPath) {
    Write-Output "Adding to PATH ..."
    $newPath = $Env:Path + ";$directory"
    [System.Environment]::SetEnvironmentVariable("PATH", $newPath)
    [System.Environment]::SetEnvironmentVariable("PATH", $newPath, [System.EnvironmentVariableTarget]::User)
}

Write-Output "Done."

Start-Sleep -Seconds 1
