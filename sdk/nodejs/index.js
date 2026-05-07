const { execFile } = require('child_process');

function execCodexCli(cliPath, args, env = {}) {
  return new Promise((resolve, reject) => {
    execFile(cliPath, args, { env: { ...process.env, ...env } }, (error, stdout, stderr) => {
      if (error) {
        const message = stderr.trim() || stdout.trim() || error.message;
        return reject(new Error(message));
      }
      resolve(stdout.trim());
    });
  });
}

class CodexClient {
  constructor(options = {}) {
    if (!options.file) {
      throw new Error('options.file is required (path to database file)');
    }

    this.cliPath = options.cliPath || 'codex-cli';
    this.file = options.file;
    this.ledger = options.ledger ? '--ledger' : null;
    this.env = {};

    if (options.encryptionKey) {
      this.env.CODEX_KEY = options.encryptionKey;
    }
  }

  _baseArgs() {
    const args = ['--file', this.file];
    if (this.ledger) args.push(this.ledger);
    return args;
  }

  async set(key, value) {
    const payload = JSON.stringify(value);
    const args = [...this._baseArgs(), 'set', key, payload];
    await execCodexCli(this.cliPath, args, this.env);
    return true;
  }

  async get(key) {
    const args = [...this._baseArgs(), 'get', key];
    const raw = await execCodexCli(this.cliPath, args, this.env);
    if (!raw) return null;
    return JSON.parse(raw);
  }

  async delete(key) {
    const args = [...this._baseArgs(), 'delete', key];
    await execCodexCli(this.cliPath, args, this.env);
    return true;
  }

  async keys() {
    const args = [...this._baseArgs(), 'keys'];
    const raw = await execCodexCli(this.cliPath, args, this.env);
    if (!raw) return [];
    return raw.split('\n').filter((k) => k);
  }

  async has(key) {
    const args = [...this._baseArgs(), 'has', key];
    const raw = await execCodexCli(this.cliPath, args, this.env);
    return raw.trim() === 'true';
  }

  async clear() {
    const args = [...this._baseArgs(), 'clear'];
    await execCodexCli(this.cliPath, args, this.env);
    return true;
  }
}

module.exports = {
  CodexClient
};
