#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const sdkDir = path.resolve(__dirname, '..');

console.log('📋 CodexDB SDK - Pre-Publish Checklist\n');

const checks = [];

// Check 1: package.json exists and valid
console.log('Checking package.json...');
try {
  const pkg = JSON.parse(fs.readFileSync(path.join(sdkDir, 'package.json'), 'utf8'));
  console.log(`✓ package.json valid (v${pkg.version})`);
  checks.push(true);
} catch (err) {
  console.log(`✗ package.json invalid: ${err.message}`);
  checks.push(false);
}

// Check 2: index.js exists
console.log('Checking index.js...');
if (fs.existsSync(path.join(sdkDir, 'index.js'))) {
  console.log('✓ index.js found');
  checks.push(true);
} else {
  console.log('✗ index.js not found');
  checks.push(false);
}

// Check 3: README.md exists
console.log('Checking README.md...');
if (fs.existsSync(path.join(sdkDir, 'README.md'))) {
  console.log('✓ README.md found');
  checks.push(true);
} else {
  console.log('✗ README.md not found');
  checks.push(false);
}

// Check 4: Test file exists
console.log('Checking test.js...');
if (fs.existsSync(path.join(sdkDir, 'test.js'))) {
  console.log('✓ test.js found');
  checks.push(true);
} else {
  console.log('✗ test.js not found');
  checks.push(false);
}

// Check 5: LICENSE exists
console.log('Checking LICENSE...');
if (fs.existsSync(path.join(sdkDir, 'LICENSE')) || fs.existsSync(path.join(sdkDir, '..', '..', 'LICENSE'))) {
  console.log('✓ LICENSE found');
  checks.push(true);
} else {
  console.log('✗ LICENSE not found');
  checks.push(false);
}

// Check 6: npm login
console.log('Checking npm credentials...');
try {
  execSync('npm whoami', { stdio: 'pipe' });
  console.log('✓ npm credentials valid');
  checks.push(true);
} catch (err) {
  console.log('✗ npm credentials invalid (run: npm login)');
  checks.push(false);
}

// Check 7: .npmignore or files in package.json
console.log('Checking .npmignore or files field...');
const pkg = JSON.parse(fs.readFileSync(path.join(sdkDir, 'package.json'), 'utf8'));
if (fs.existsSync(path.join(sdkDir, '.npmignore')) || pkg.files) {
  console.log('✓ .npmignore or files field present');
  checks.push(true);
} else {
  console.log('⚠ .npmignore not found (optional, but recommended)');
  checks.push(true);
}

console.log('\n' + '='.repeat(40));
const passed = checks.filter(c => c).length;
const total = checks.length;
console.log(`Results: ${passed}/${total} checks passed\n`);

if (passed === total) {
  console.log('✅ Ready to publish!');
  process.exit(0);
} else {
  console.log('❌ Fix the issues above before publishing');
  process.exit(1);
}
