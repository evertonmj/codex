#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Get script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SDK_DIR="$(dirname "$SCRIPT_DIR")"
CODEX_DIR="$(dirname "$(dirname "$SDK_DIR")")"

echo -e "${YELLOW}📦 CodexDB SDK - NPM Publish Script${NC}"
echo ""

# Check if we're in the right directory
if [ ! -f "$SDK_DIR/package.json" ]; then
  echo -e "${RED}❌ Error: package.json not found in $SDK_DIR${NC}"
  exit 1
fi

cd "$SDK_DIR"

# Get current version from package.json
CURRENT_VERSION=$(node -e "console.log(require('./package.json').version)")
echo -e "${YELLOW}Current version: ${GREEN}$CURRENT_VERSION${NC}"
echo ""

# Step 1: Ask for version bump type
echo -e "${YELLOW}1️⃣  Select version increment:${NC}"
echo "   1) patch  (increment patch version)"
echo "   2) minor  (increment minor version)"
echo "   3) major  (increment major version)"
echo "   4) skip   (keep current version)"
echo ""

read -p "Choose (1-4): " -n 1 version_choice
echo ""

case $version_choice in
  1)
    VERSION_TYPE="patch"
    ;;
  2)
    VERSION_TYPE="minor"
    ;;
  3)
    VERSION_TYPE="major"
    ;;
  4)
    VERSION_TYPE="skip"
    ;;
  *)
    echo -e "${RED}Invalid choice${NC}"
    exit 1
    ;;
esac

# Update version if not skipped
if [ "$VERSION_TYPE" != "skip" ]; then
  echo -e "${YELLOW}Updating version ($VERSION_TYPE)...${NC}"
  npm version $VERSION_TYPE
  NEW_VERSION=$(node -e "console.log(require('./package.json').version)")
  echo -e "${GREEN}✓ Version updated to $NEW_VERSION${NC}"
  echo ""
else
  NEW_VERSION=$CURRENT_VERSION
fi

# Step 2: Install dependencies
echo -e "${YELLOW}2️⃣  Installing dependencies...${NC}"
npm install --ignore-scripts
echo -e "${GREEN}✓ Dependencies installed${NC}"
echo ""

# Step 3: Run tests
echo -e "${YELLOW}3️⃣  Running tests...${NC}"
npm test
echo -e "${GREEN}✓ Tests passed${NC}"
echo ""

# Step 4: Check if codex-cli binary exists (optional)
echo -e "${YELLOW}4️⃣  Checking for codex-cli binary (optional)...${NC}"
if [ -f "$SDK_DIR/codex-cli" ] && [ -x "$SDK_DIR/codex-cli" ]; then
  echo -e "${GREEN}✓ codex-cli binary found and executable${NC}"
else
  echo -e "${YELLOW}⚠ codex-cli binary not found (will be downloaded on npm install)${NC}"
fi
echo ""

# Step 5: Verify npm credentials
echo -e "${YELLOW}5️⃣  Verifying npm credentials...${NC}"
if ! npm whoami > /dev/null 2>&1; then
  echo -e "${RED}❌ Error: Not authenticated with npm${NC}"
  echo "   Please login first: npm login"
  exit 1
fi
echo -e "${GREEN}✓ npm credentials valid${NC}"
echo ""

# Step 6: Dry run to check for issues
echo -e "${YELLOW}6️⃣  Running npm publish --dry-run...${NC}"
npm publish --dry-run
echo -e "${GREEN}✓ Dry run successful${NC}"
echo ""

# Step 7: Confirmation before publishing
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${YELLOW}Ready to publish codexdb-sdk@$NEW_VERSION to npm${NC}"
echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
read -p "Continue? (y/N) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
  echo "Publish cancelled."
  if [ "$VERSION_TYPE" != "skip" ]; then
    echo -e "${YELLOW}Reverting version change...${NC}"
    git checkout package.json package-lock.json 2>/dev/null || true
  fi
  exit 0
fi

# Step 8: Publish to npm
echo -e "${YELLOW}7️⃣  Publishing to npm...${NC}"
npm publish
echo -e "${GREEN}✓ Published successfully!${NC}"
echo ""

echo -e "${GREEN}🎉 CodexDB SDK@$NEW_VERSION published to npm!${NC}"
echo ""
echo "Next steps:"
echo "  - Verify at: https://www.npmjs.com/package/codexdb-sdk/v/$NEW_VERSION"
echo "  - Test installation: npm install codexdb-sdk@$NEW_VERSION"
echo "  - Create GitHub release: https://github.com/evertonmj/codex/releases/new"


