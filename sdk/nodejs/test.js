const path = require('path');
const fs = require('fs');
const { CodexClient } = require('./index');

(async () => {
  const dbPath = path.resolve(__dirname, 'test-db.db');

  // cleanup before test
  try { fs.unlinkSync(dbPath); } catch (e) {}

  const client = new CodexClient({ file: dbPath });

  await client.set('test-key', { hello: 'world' });
  const value = await client.get('test-key');

  if (value.hello !== 'world') throw new Error('value mismatch');

  const hasKey = await client.has('test-key');
  if (!hasKey) throw new Error('expected key to exist');

  const keys = await client.keys();
  if (!keys.includes('test-key')) throw new Error('expected keys to include test-key');

  await client.delete('test-key');
  const hasAfterDelete = await client.has('test-key');
  if (hasAfterDelete) throw new Error('expected key to be deleted');

  await client.clear();

  // cleanup after test
  try { fs.unlinkSync(dbPath); } catch (e) {}

  console.log('Node.js SDK smoke test passed');
})();
