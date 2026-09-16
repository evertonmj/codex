'use strict';
const { test } = require('node:test');
const assert = require('node:assert/strict');
const fs = require('node:fs');
const os = require('node:os');
const path = require('node:path');
const { CodexDatabase } = require('./');
const binding = require('./build/codex.node');

for (const ledger of [false, true]) {
  test(`embedded persistence and CRUD (ledger=${ledger})`, () => {
    const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'codex-native-'));
    const file = path.join(dir, 'data.json');
    let db;
    try {
      db = new CodexDatabase({ file, ledger });
      assert.throws(() => new CodexDatabase({ file, ledger }), { code: 'LOCKED' });
      const value = { name: 'João', nested: [true, null, { age: 6 }] };
      assert.equal(db.set('child', value), true);
      assert.deepEqual(db.get('child'), value);
      assert.equal(db.has('child'), true);
      assert.deepEqual(db.keys(), ['child']);
      db.set('null', null);
      assert.equal(db.get('null'), null);
      assert.throws(() => db.set('bad', undefined));
      assert.throws(() => db.set('bad', 1n));
      db.close(); db.close();
      assert.throws(() => db.get('child'));
      db = new CodexDatabase({ file, ledger });
      assert.deepEqual(db.get('child'), value);
      if (!ledger) {
        const disk = JSON.parse(fs.readFileSync(file, 'utf8'));
        assert.deepEqual(disk.data.values.child, value);
      }
      assert.equal(db.delete('child'), true);
      assert.equal(db.has('child'), false);
      assert.throws(() => db.get('child'), { code: 'NOT_FOUND' });
      assert.equal(db.clear(), true);
      assert.deepEqual(db.keys(), []);
    } finally { db?.close(); fs.rmSync(dir, { recursive: true, force: true }); }
  });
}
for (const ledger of [false, true]) {
  test(`encryption is explicit, wrong/missing keys fail (ledger=${ledger})`, () => {
    const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'codex-encrypted-'));
    const file = path.join(dir, 'db');
    const encryptionKey = '01234567890123456789012345678901';
    let db;
    try {
      assert.throws(() => new CodexDatabase({ file, ledger, encryptionKey: 'short' }), { code: 'INVALID_KEY' });
      db = new CodexDatabase({ file, ledger, encryptionKey });
      db.set('secret', 'sensitive-data'); db.close();
      assert.equal(fs.readFileSync(file).includes(Buffer.from('sensitive-data')), false);
      assert.throws(() => new CodexDatabase({ file, ledger }));
      assert.throws(() => new CodexDatabase({ file, ledger, encryptionKey: '11234567890123456789012345678901' }));
      db = new CodexDatabase({ file, ledger, encryptionKey });
      assert.equal(db.get('secret'), 'sensitive-data');
    } finally { db?.close(); fs.rmSync(dir, { recursive: true, force: true }); }
  });
}
test('default path and native boundary validation', () => {
  const dir = fs.mkdtempSync(path.join(os.tmpdir(), 'codex-default-'));
  const cwd = process.cwd();
  let db;
  try {
    process.chdir(dir);
    db = new CodexDatabase(); db.set('', [1, 'x']);
    assert.deepEqual(db.get(''), [1, 'x']);
    assert.equal(fs.existsSync('.codex-data/default.db'), true);
    assert.throws(() => binding.call());
    assert.throws(() => binding.call(42));
    assert.equal(JSON.parse(binding.call('{')).ok, false);
  } finally { db?.close(); process.chdir(cwd); fs.rmSync(dir, { recursive: true, force: true }); }
});
