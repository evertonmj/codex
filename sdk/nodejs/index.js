const { execFile } = require('child_process');
const path = require('path');
const fs = require('fs/promises');
const fsSync = require('fs');

function defaultCliPath() {
  const localBinary = path.join(__dirname, 'codex-cli' + (process.platform === 'win32' ? '.exe' : ''));
  return fsSync.existsSync(localBinary) ? localBinary : 'codex-cli';
}

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

async function ensureDataDirectory(dataPath) {
  try {
    await fs.mkdir(dataPath, { recursive: true });
  } catch (error) {
    if (error.code !== 'EEXIST') {
      throw error;
    }
  }
}

class CodexClient {
  constructor(options = {}) {
    // If no file is provided, use default: .codex-data/default.db
    let file = options.file;
    if (!file) {
      const dataDir = path.resolve(process.cwd(), '.codex-data');
      file = path.join(dataDir, 'default.db');
    } else {
      // If file is provided but it's just a filename (no path separators),
      // resolve it relative to current working directory
      file = path.resolve(process.cwd(), file);
    }

    this.cliPath = options.cliPath || defaultCliPath();
    this.file = file;
    this.dataDir = path.dirname(this.file);
    this.ledger = options.ledger ? '--ledger' : null;
    this.env = {};

    if (options.encryptionKey) {
      this.env.CODEX_KEY = options.encryptionKey;
    }
  }

  async initialize() {
    // Create data directory if it doesn't exist
    await ensureDataDirectory(this.dataDir);
    this._initialized = true;
  }

  async _ensureInitialized() {
    if (!this._initialized) {
      await this.initialize();
    }
  }

  _baseArgs() {
    const args = ['--file', this.file];
    if (this.ledger) args.push(this.ledger);
    return args;
  }

  async set(key, value) {
    await this._ensureInitialized();
    const payload = JSON.stringify(value);
    const args = [...this._baseArgs(), 'set', key, payload];
    await execCodexCli(this.cliPath, args, this.env);
    return true;
  }

  async get(key) {
    await this._ensureInitialized();
    const args = [...this._baseArgs(), 'get', key];
    const raw = await execCodexCli(this.cliPath, args, this.env);
    if (!raw) return null;
    return JSON.parse(raw);
  }

  async delete(key) {
    await this._ensureInitialized();
    const args = [...this._baseArgs(), 'delete', key];
    await execCodexCli(this.cliPath, args, this.env);
    return true;
  }

  async keys() {
    await this._ensureInitialized();
    const args = [...this._baseArgs(), 'keys'];
    const raw = await execCodexCli(this.cliPath, args, this.env);
    if (!raw) return [];
    return raw.split('\n').filter((k) => k);
  }

  async has(key) {
    await this._ensureInitialized();
    const args = [...this._baseArgs(), 'has', key];
    const raw = await execCodexCli(this.cliPath, args, this.env);
    return raw.trim() === 'true';
  }

  async clear() {
    await this._ensureInitialized();
    const args = [...this._baseArgs(), 'clear'];
    await execCodexCli(this.cliPath, args, this.env);
    return true;
  }
}

module.exports = {
  CodexClient
};
