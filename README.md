# Stakeout

A command line tool thats print a history of your [FLOW](https://www.onflow.org/) staking rewards (e.g. for tax purposes).

![stakeout](https://user-images.githubusercontent.com/2547035/159144265-e385a9d7-2aca-4bd5-9cc6-7b22a6343119.gif)

## Install

### macOS and Linux

> :warning: This installation method only works for macOS and Linux.

Paste this command in your [macOS Terminal](https://support.apple.com/en-ca/guide/terminal/apd5265185d-f365-44cb-8b09-71a064a42125/mac) or Linux shell and hit enter:

```sh
sh -ci "$(curl -fsSL https://raw.githubusercontent.com/psiemens/stakeout/main/install.sh)"
```

### Windows

_This installation method only works on Windows 10, 8.1, or 7 (SP1, with [PowerShell 3.0](https://www.microsoft.com/en-ca/download/details.aspx?id=34595)), on x86-64._

1. Open PowerShell ([Instructions](https://docs.microsoft.com/en-us/powershell/scripting/install/installing-windows-powershell?view=powershell-7#finding-powershell-in-windows-10-81-80-and-7))
2. In PowerShell, run:

    ```powershell
    iex "& { $(irm 'https://raw.githubusercontent.com/psiemens/stakeout/main/install.ps1') }"
    ```


## Usage

### List rewards from all epochs

```sh
stakeout <address>
```

### Filter rewards by year

```sh
stakeout -year=2022 <address>
```

### Filter rewards by start and end date

```sh
stakeout -start=2021-04-01 -end=2022-04-01 <address>
```

## Current Limitations

This tool has some limitations. Open an issue if you want me to fix either of these! :smile:

- It only prints rewards from delegating, not staking.
- It only searches the epochs up to **December 21, 2022**.

## Development

### Run with Go

```sh
go run main.go <address>
```

### Build

```sh
make binaries
```
