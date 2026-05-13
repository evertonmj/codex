const { CodexClient } = require('./index');

// Example 1: Using default data directory (.codex-data/default.db)
(async () => {
  console.log('Example 1: Using default data directory\n');

  // No file option needed - uses .codex-data/default.db automatically
  const client = new CodexClient();

  console.log('Database file:', client.file);
  console.log('Data directory:', client.dataDir);
  console.log('Initialized:', client._initialized);

  // Initialize creates the directory if it doesn't exist
  await client.initialize();
  console.log('After initialize - Initialized:', client._initialized);

  try {
    // Set a value
    await client.set('message', { text: 'Hello from CodexDB!' });
    console.log('✓ Set message');

    // Get the value
    const value = await client.get('message');
    console.log('✓ Got message:', value.text);

    // List all keys
    const keys = await client.keys();
    console.log('✓ All keys:', keys);

    // Check if key exists
    const exists = await client.has('message');
    console.log('✓ Key exists:', exists);

    // Delete the key
    await client.delete('message');
    console.log('✓ Deleted message');
  } catch (error) {
    console.error('Error:', error.message);
  }
})();

// Example 2: Using custom data directory
console.log('\n---\n');
console.log('Example 2: Using custom data directory\n');

(async () => {
  const path = require('path');

  // Custom file path - resolved relative to current working directory
  const client = new CodexClient({ file: 'custom-data.db' });

  console.log('Database file:', client.file);
  console.log('Data directory:', client.dataDir);

  try {
    // Operations automatically initialize the directory
    await client.set('config', { theme: 'dark', language: 'pt-BR' });
    console.log('✓ Set config');

    const config = await client.get('config');
    console.log('✓ Got config:', config);
  } catch (error) {
    console.error('Error:', error.message);
  }
})();
