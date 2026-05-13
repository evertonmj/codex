# Publishing to npm

## Quick Start (Automated)

```bash
cd /Users/everton/workspace/codex/sdk/nodejs
npm run publish
```

This runs a complete automated workflow with validation and confirmation.

## Prerequisites

1. **npm account**: https://www.npmjs.com/signup
2. **Login locally**:
   ```bash
   npm login
   ```
3. **Verify credentials**:
   ```bash
   npm whoami
   ```

## Publishing Methods

### Method 1: Automated Script (Recommended)

```bash
npm run publish
```

This script:
- ✓ Installs dependencies
- ✓ Runs tests
- ✓ Verifies codex-cli binary
- ✓ Validates npm credentials
- ✓ Performs dry-run
- ✓ Asks for confirmation
- ✓ Publishes to npm

### Method 2: Pre-Publish Checklist

Validate everything before publishing:

```bash
npm run publish-check
```

Then publish manually:

```bash
npm publish
```

### Method 3: Manual Steps

```bash
# Install dependencies
npm install

# Run tests
npm test

# Publish
npm publish
```

## Version Management

### Automated versioning

```bash
# Patch (0.1.6 → 0.1.7)
npm version patch

# Minor (0.1.6 → 0.2.0)
npm version minor

# Major (0.1.6 → 1.0.0)
npm version major
```

Then publish:
```bash
npm publish
```

### Manual versioning

Edit `package.json` and update the `version` field, then publish:
```bash
npm publish
```

## Troubleshooting

### "Not authenticated"
```bash
npm login
npm whoami  # verify
```

### "Already published at this version"
```bash
npm version patch
npm publish
```

### "Binary download fails"

The postinstall script downloads codex-cli from GitHub releases. Verify:
- Version in package.json matches a GitHub release tag
- GitHub release contains the binary files

### "Tests fail"
```bash
npm test  # See detailed errors
```

## Verification After Publishing

```bash
# Check latest version
npm view codexdb-sdk@latest

# Test installation
npm install codexdb-sdk@latest

# View on npm
open https://www.npmjs.com/package/codexdb-sdk
```

## Files Included in npm Package

Included (via package.json + files field or .npmignore):
- index.js
- package.json
- README.md
- LICENSE
- codex-cli (binary)
- node_modules/

Excluded:
- test.js
- example_app.js
- scripts/
- .npmignore
- IDE config
- OS files

## Automated Publishing (CI/CD)

Example GitHub Actions workflow:

```yaml
name: Publish to npm
on:
  push:
    tags:
      - 'v*'

jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-node@v2
        with:
          node-version: '18'
          registry-url: 'https://registry.npmjs.org'
      - run: cd sdk/nodejs && npm install
      - run: cd sdk/nodejs && npm test
      - run: cd sdk/nodejs && npm publish
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```

Setup:
1. Create npm token: https://www.npmjs.com/settings/~/tokens
2. Add to GitHub: Settings → Secrets → `NPM_TOKEN`

## Additional Resources

- npm docs: https://docs.npmjs.com/cli/
- GitHub releases: https://github.com/evertonmj/codex/releases
- Main README: ../../README.md

