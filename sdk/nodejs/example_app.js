const path = require('path');
const CodexClient = require('./index');

const dbPath = path.resolve(__dirname, 'example-db.db');
const client = new CodexClient({ file: dbPath });

(async () => {
  // Set a value
  await client.set('greeting', { msg: 'Hello from Node.js!' });

  // Get the value
  const value = await client.get('greeting');
  console.log('Greeting:', value.msg);

  // List all keys
  const keys = await client.keys();
  console.log('All keys:', keys);

  // Delete the key
  const deleted = await client.delete('greeting');
  console.log('Deleted:', deleted);

  // Cleanup
  const fs = require('fs');
  try { fs.unlinkSync(dbPath); } catch (e) {}
})();
