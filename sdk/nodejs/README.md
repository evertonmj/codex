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

When you install this package, the postinstall script makes a **best-effort**
attempt to download the correct `codex-cli` binary for your platform from the
GitHub releases page. If the release or a matching asset is not available (for
example before a release is published, on an unsupported platform, or without
network access), the installation **succeeds without installing a binary** —
the script prints a warning and exits cleanly instead of failing.

In that case you must provide the binary yourself:

- Build it from source with `make build` (produces `bin/codex-cli`), or
- Download it manually from the GitHub releases page, and
- Point the client at it with the `cliPath` option when creating the client.
