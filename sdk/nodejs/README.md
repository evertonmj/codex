# CodexDB Node.js SDK

Node.js SDK for CodexDB. Simple wrapper for `codex-cli` as an external process.

## Installation

Install via npm (from local directory):

```sh
npm install ./
```

Or, for development:

```sh
npm install -g ./
```

## Usage Example

```js
const { CodexClient } = require('codexdb-sdk');

const client = new CodexClient({ file: 'test.db' });
client.set('user:1', { name: 'Alice', age: 30 });
console.log(client.get('user:1'));
console.log(client.keys());
client.delete('user:1');
```

## Requirements

- Node.js 14+

## codex-cli Binary

When you install this package, the correct `codex-cli` binary for your platform will be downloaded automatically from the GitHub releases page. No manual installation is required.

If you want to use a custom binary, set the `cliPath` option when creating the client.
