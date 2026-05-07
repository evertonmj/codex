import json
import shutil
import subprocess


class CodexClient:
    def __init__(self, file, cli_path='codex-cli', ledger=False, encryption_key=None):
        if not file:
            raise ValueError('file is required')

        if shutil.which(cli_path) is None:
            raise FileNotFoundError(f"codex-cli not found at '{cli_path}' in PATH")

        self.cli_path = cli_path
        self.file = file
        self.ledger = ledger
        self.env = None

        if encryption_key is not None:
            self.env = dict(**__import__('os').environ)
            self.env['CODEX_KEY'] = encryption_key

    def _base_args(self):
        args = ['--file', self.file]
        if self.ledger:
            args.append('--ledger')
        return args

    def _run(self, *cmd_args):
        args = [self.cli_path, *self._base_args(), *cmd_args]
        cp = subprocess.run(args, capture_output=True, text=True, env=self.env)
        if cp.returncode != 0:
            raise RuntimeError(cp.stderr.strip() or cp.stdout.strip() or f'codex-cli failed ({cp.returncode})')
        return cp.stdout.strip()

    def set(self, key, value):
        payload = json.dumps(value, ensure_ascii=False)
        self._run('set', key, payload)
        return True

    def get(self, key):
        raw = self._run('get', key)
        if raw == 'null' or raw == '':
            return None
        return json.loads(raw)

    def delete(self, key):
        self._run('delete', key)
        return True

    def keys(self):
        raw = self._run('keys')
        if not raw:
            return []
        return [k for k in raw.split('\n') if k]

    def has(self, key):
        raw = self._run('has', key)
        return raw.strip().lower() == 'true'

    def clear(self):
        self._run('clear')
        return True
