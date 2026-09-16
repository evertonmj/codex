'use strict';
const { execFileSync } = require('node:child_process');
const fs = require('node:fs');
const path = require('node:path');
const root = path.resolve(__dirname, '../../..');
const out = path.resolve(__dirname, '../build');
const platform = process.platform;
if (!['linux', 'darwin'].includes(platform)) {
  throw new Error('Node addon build supports Linux and macOS. Build the C ABI DLL separately on Windows.');
}
fs.mkdirSync(out, { recursive: true });
const library = platform === 'darwin' ? 'libcodex.dylib' : 'libcodex.so';
execFileSync('go', ['build', '-buildmode=c-shared', '-o', path.join(out, library), './cmd/codex-shared'],
  { cwd: root, stdio: 'inherit', env: { ...process.env, CGO_ENABLED: '1' } });
if (platform === 'darwin') {
  execFileSync('install_name_tool', ['-id', '@rpath/libcodex.dylib', path.join(out, library)], { stdio: 'inherit' });
}
const flags = platform === 'darwin' ? ['-dynamiclib', '-undefined', 'dynamic_lookup', '-Wl,-rpath,@loader_path']
  : ['-shared', '-fPIC', '-Wl,-rpath,$ORIGIN'];
execFileSync(process.env.CC || 'cc', [...flags, '-Wall', '-Wextra', '-Werror',
  '-I', process.env.NODE_INCLUDE_DIR || '/usr/include/node', '-I', out,
  path.resolve(__dirname, '../native/addon.c'), '-L', out, '-lcodex',
  '-o', path.join(out, 'codex.node')], { stdio: 'inherit' });
