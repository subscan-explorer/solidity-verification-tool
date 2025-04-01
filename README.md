# Solidity Smart Contract Verification Tool

This tool allows you to verify Solidity smart contracts by comparing the bytecode on the blockchain with the recompiled
bytecode.

## Features

- Recompiles Solidity smart contracts using a specified compiler version.
- Compares the on-chain bytecode with the recompiled bytecode.
- Provides detailed error messages if verification fails.
- Polkadot revive(https://github.com/paritytech/revive) support.
- No dependency on external services or third-party libraries.

## Installation

1. Clone the repository:
```sh
git clone https://github.com/subscan-explorer/solidity-verification-tool.git
```

2. Navigate to the project directory:

```sh
cd solidity-verification-tool
```

3. Install dependencies:

```sh
go mod tidy
```

## Usage

1. Start the server:

```sh
go run main.go
```

2. Send a POST request to `/verify` with the contract metadata and compiler version.

```sh
curl -X POST -H "Content-Type: application/json" -d '{"metadata": {...}, "compilerVersion": "v0.8.26+commit.8a97fa7a","chain":46,"address":"xxxx}' http://localhost:8081/verify
```

## Revive support

Building Solidity contracts for PolkaVM requires installing extra dependencies. To install revive, run the following command:

```sh
go run . download # if will auto download resolc binary in static folder
```

## License

This project is licensed under the MIT License.
