'use strict';
const path = require('node:path');
const binding = require('./build/codex.node');

class CodexDatabase {
  constructor(options = {}) {
    this.handle = this._call({ op: 'open', file: path.resolve(options.file || '.codex-data/default.db'),
      encryptionKey: options.encryptionKey || '', ledger: !!options.ledger });
  }
  _call(request) {
    const response = JSON.parse(binding.call(JSON.stringify(request)));
    if (!response.ok) {
      const error = new Error(response.error);
      error.code = response.code;
      throw error;
    }
    return response.result;
  }
  _request(op, extra = {}) {
    return this._call({ op, handle: this.handle, ...extra });
  }
  set(key, value) { return this._request('set', { key, value }); }
  get(key) { return this._request('get', { key }); }
  delete(key) { return this._request('delete', { key }); }
  has(key) { return this._request('has', { key }); }
  keys() { return this._request('keys'); }
  clear() { return this._request('clear'); }
  close() {
    if (this.handle === null) return;
    this._request('close');
    this.handle = null;
  }
}
module.exports = { CodexDatabase };
