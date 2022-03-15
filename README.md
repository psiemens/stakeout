# Stakeout

A command line tool thats print a history of your [FLOW](https://www.onflow.org/) staking rewards (e.g. for tax purposes).

## Usage

```sh
go run main.go <address>
```

## Current Limitations

This tool is in beta and has some limitations. Open an issue if you want me to fix any of these! :smile:

- It only prints rewards from delegating, not staking.
- It only searches the epochs from **October 12 to December 29, 2021**.
- It may break for accounts with more than 100 transactions.
