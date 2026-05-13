const path = require('path');
const fs = require('fs/promises');
const fsSync = require('fs');
const { CodexClient } = require('./index');

// Test utilities
const TEST_DIR = path.join(__dirname, '.test-temp');
const delay = (ms) => new Promise((resolve) => setTimeout(resolve, ms));

async function setupTestDir() {
  try {
    await fs.rm(TEST_DIR, { recursive: true, force: true });
  } catch (e) {
    // ignore
  }
  await fs.mkdir(TEST_DIR, { recursive: true });
}

async function cleanupTestDir() {
  try {
    await fs.rm(TEST_DIR, { recursive: true, force: true });
  } catch (e) {
    // ignore
  }
}

async function dirExists(dirPath) {
  try {
    const stats = await fs.stat(dirPath);
    return stats.isDirectory();
  } catch (e) {
    return false;
  }
}

// Color console output
const colors = {
  reset: '\x1b[0m',
  green: '\x1b[32m',
  red: '\x1b[31m',
  yellow: '\x1b[33m',
  blue: '\x1b[34m',
};

function log(message, color = 'reset') {
  console.log(`${colors[color]}${message}${colors.reset}`);
}

function logTest(name) {
  log(`\n  ✓ ${name}`, 'blue');
}

function logPass(message) {
  log(`    ✓ ${message}`, 'green');
}

function logFail(message) {
  log(`    ✗ ${message}`, 'red');
  process.exitCode = 1;
}

// ============================================
// TESTS
// ============================================

async function testDefaultDataDirectory() {
  logTest('Default data directory behavior');

  await cleanupTestDir();
  await setupTestDir();

  const originalCwd = process.cwd();

  try {
    process.chdir(TEST_DIR);

    // Create client without file option
    const client = new CodexClient();

    // Check that file path includes .codex-data
    const expectedPath = path.join(TEST_DIR, '.codex-data', 'default.db');
    if (client.file === expectedPath) {
      logPass(`File path defaults to ${expectedPath}`);
    } else {
      logFail(
        `File path should be ${expectedPath}, got ${client.file}`
      );
    }

    // Check dataDir
    const expectedDataDir = path.join(TEST_DIR, '.codex-data');
    if (client.dataDir === expectedDataDir) {
      logPass(`Data directory defaults to ${expectedDataDir}`);
    } else {
      logFail(
        `Data directory should be ${expectedDataDir}, got ${client.dataDir}`
      );
    }

    // Check that directory is created on initialize
    const existsBefore = await dirExists(expectedDataDir);
    if (!existsBefore) {
      logPass('Directory does not exist before initialize()');
    } else {
      logFail('Directory should not exist before initialize()');
    }

    await client.initialize();

    const existsAfter = await dirExists(expectedDataDir);
    if (existsAfter) {
      logPass('Directory created after initialize()');
    } else {
      logFail('Directory should be created after initialize()');
    }
  } finally {
    process.chdir(originalCwd);
  }
}

async function testCustomFileOption() {
  logTest('Custom file option handling');

  await cleanupTestDir();
  await setupTestDir();

  const originalCwd = process.cwd();

  try {
    process.chdir(TEST_DIR);

    // Test with full path
    const customPath = path.join(TEST_DIR, 'custom', 'my-db.db');
    const client = new CodexClient({ file: customPath });

    if (client.file === customPath) {
      logPass(`Full path option preserved: ${customPath}`);
    } else {
      logFail(
        `File path should be ${customPath}, got ${client.file}`
      );
    }

    // Test with relative filename
    const client2 = new CodexClient({ file: 'my-db.db' });
    const expectedPath = path.resolve(process.cwd(), 'my-db.db');

    if (client2.file === expectedPath) {
      logPass(`Relative filename resolved to: ${expectedPath}`);
    } else {
      logFail(
        `File path should be ${expectedPath}, got ${client2.file}`
      );
    }
  } finally {
    process.chdir(originalCwd);
  }
}

async function testDirectoryCreationOnDemand() {
  logTest('Directory creation on first operation');

  await cleanupTestDir();
  await setupTestDir();

  const originalCwd = process.cwd();

  try {
    process.chdir(TEST_DIR);

    const client = new CodexClient();
    const dataDir = client.dataDir;

    // Check directory doesn't exist yet
    const existsBefore = await dirExists(dataDir);
    if (!existsBefore) {
      logPass('Directory does not exist before first operation');
    } else {
      logFail('Directory should not exist before first operation');
    }

    // Call initialize (which would be called automatically in real operations)
    await client.initialize();

    // Check directory now exists
    const existsAfter = await dirExists(dataDir);
    if (existsAfter) {
      logPass('Directory created on initialize()');
    } else {
      logFail('Directory should be created on initialize()');
    }
  } finally {
    process.chdir(originalCwd);
  }
}

async function testMultipleLevelsPath() {
  logTest('Multiple levels of directory creation');

  await cleanupTestDir();
  await setupTestDir();

  const originalCwd = process.cwd();

  try {
    process.chdir(TEST_DIR);

    const customPath = path.join(TEST_DIR, 'a', 'b', 'c', 'db.db');
    const client = new CodexClient({ file: customPath });

    const dataDir = client.dataDir;
    const existsBefore = await dirExists(dataDir);

    if (!existsBefore) {
      logPass('Nested directories do not exist before initialize()');
    } else {
      logFail('Nested directories should not exist before initialize()');
    }

    await client.initialize();

    const existsAfter = await dirExists(dataDir);
    if (existsAfter) {
      logPass('Nested directories created: ' + dataDir);
    } else {
      logFail('Nested directories should be created');
    }
  } finally {
    process.chdir(originalCwd);
  }
}

async function testClientProperties() {
  logTest('Client properties and configuration');

  await cleanupTestDir();
  await setupTestDir();

  const originalCwd = process.cwd();

  try {
    process.chdir(TEST_DIR);

    const client = new CodexClient({
      encryptionKey: 'secret-key',
      ledger: true,
    });

    if (client.env.CODEX_KEY === 'secret-key') {
      logPass('Encryption key stored in environment');
    } else {
      logFail('Encryption key not stored correctly');
    }

    if (client.ledger === '--ledger') {
      logPass('Ledger flag set correctly');
    } else {
      logFail('Ledger flag not set correctly');
    }

    const baseArgs = client._baseArgs();
    if (baseArgs.includes('--file') && baseArgs.includes('--ledger')) {
      logPass('Base args include both --file and --ledger flags');
    } else {
      logFail('Base args should include both flags');
    }
  } finally {
    process.chdir(originalCwd);
  }
}

async function testInitializedFlag() {
  logTest('Initialized flag behavior');

  const originalCwd = process.cwd();

  try {
    process.chdir(TEST_DIR);

    const client = new CodexClient();

    if (!client._initialized) {
      logPass('_initialized flag is false before initialize()');
    } else {
      logFail('_initialized flag should be false initially');
    }

    await client.initialize();

    if (client._initialized) {
      logPass('_initialized flag is true after initialize()');
    } else {
      logFail('_initialized flag should be true after initialize()');
    }

    // Calling initialize again should not cause issues
    await client.initialize();
    logPass('Multiple initialize() calls are safe');
  } finally {
    process.chdir(originalCwd);
  }
}

// ============================================
// RUN ALL TESTS
// ============================================

async function runAllTests() {
  log('\n═══════════════════════════════════════════════════════', 'yellow');
  log('  CodexDB SDK - Unit Tests', 'yellow');
  log('═══════════════════════════════════════════════════════\n', 'yellow');

  await setupTestDir();

  try {
    await testDefaultDataDirectory();
    await testCustomFileOption();
    await testDirectoryCreationOnDemand();
    await testMultipleLevelsPath();
    await testClientProperties();
    await testInitializedFlag();

    log('\n═══════════════════════════════════════════════════════', 'yellow');
    if (process.exitCode) {
      log('  Some tests failed!', 'red');
    } else {
      log('  All tests passed! ✓', 'green');
    }
    log('═══════════════════════════════════════════════════════\n', 'yellow');
  } finally {
    await cleanupTestDir();
  }
}

// Run tests
runAllTests().catch((error) => {
  log(`Fatal error: ${error.message}`, 'red');
  process.exit(1);
});
