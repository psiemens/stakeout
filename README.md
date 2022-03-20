# Stakeout

A command line tool thats print a history of your [FLOW](https://www.onflow.org/) staking rewards (e.g. for tax purposes).

![stakeout](https://user-images.githubusercontent.com/2547035/159144265-e385a9d7-2aca-4bd5-9cc6-7b22a6343119.gif)

## Install

> :warning: This installation method only works for macOS and Linux.

Paste this command in your [macOS Terminal](https://support.apple.com/en-ca/guide/terminal/apd5265185d-f365-44cb-8b09-71a064a42125/mac) or Linux shell and hit enter:

```sh
sh -ci "$(curl -fsSL https://raw.githubusercontent.com/psiemens/stakeout/main/install.sh)"
```

## Usage

```sh
stakeout <address>
```

## Current Limitations

This tool is in beta and has some limitations. Open an issue if you want me to fix any of these! :smile:

- It only prints rewards from delegating, not staking.
- It only searches the epochs from **October 12 to December 29, 2021**.
- It may break for accounts with more than 100 transactions.
- The pre-built binaries are only compatible with Linux and macOS.

## Development

### Run with Go

```sh
go run main.go <address>
```

### Build

```sh
make binaries
```
