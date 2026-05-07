const path = require('path');
const CodexClient = require('codexdb-sdk');

const dbPath = path.resolve(__dirname, 'test-db.db');
const client = new CodexClient({ file: dbPath });

(async () => {
  await client.set('foo', { bar: 123 });
  const value = await client.get('foo');
  console.log('Value for foo:', value);
  await client.delete('foo');
  await client.clear();
  const fs = require('fs');
  try { fs.unlinkSync(dbPath); } catch (e) {}
})();
