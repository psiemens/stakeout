<#
.Synopsis
    Install the Stakeout CLI on Windows.
.DESCRIPTION
    By default, the latest release will be installed.
.Parameter Directory
    The destination path to install to.
.Parameter AddToPath
    Add the absolute destination path to the 'User' scope environment variable 'Path'.
.EXAMPLE
    Install the current version
    .\install.ps1
#>
param (
    [string] $directory = "$env:APPDATA\Stakeout",
    [bool] $addToPath = $true
)

Set-StrictMode -Version 3.0

# Enable support for ANSI escape sequences
Set-ItemProperty HKCU:\Console VirtualTerminalLevel -Type DWORD 1

$ErrorActionPreference = "Stop"

$baseURL = "https://raw.githubusercontent.com/psiemens/stakeout/main"

Write-Output("Installing Stakeout...")

New-Item -ItemType Directory -Force -Path $directory | Out-Null

$progressPreference = 'silentlyContinue'

Invoke-WebRequest -Uri "$baseURL/stakeout-x86_64-windows" -UseBasicParsing -OutFile "$directory\stakeout.exe"

if ($addToPath) {
    Write-Output "Adding to PATH ..."
    $newPath = $Env:Path + ";$directory"
    [System.Environment]::SetEnvironmentVariable("PATH", $newPath)
    [System.Environment]::SetEnvironmentVariable("PATH", $newPath, [System.EnvironmentVariableTarget]::User)
}

Write-Output "Done."

Start-Sleep -Seconds 1
