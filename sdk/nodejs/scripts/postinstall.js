const fs = require('fs');
const os = require('os');
const path = require('path');
const { https } = require('follow-redirects');
const { execSync } = require('child_process');

// Get package.json path relative to this script
const packageJsonPath = path.join(__dirname, '..', 'package.json');
const packageJson = require(packageJsonPath);
const version = packageJson.version;

// Map platform to OS name in release
const OS_MAP = {
  darwin: 'darwin',
  linux: 'linux',
  win32: 'windows'
};

// Get system info
const platform = os.platform();
const osName = OS_MAP[platform];
const arch = os.arch() === 'arm64' ? 'arm64' : 'amd64';

if (!osName) {
  console.log(`⚠️  No binary available for platform: ${platform}`);
  process.exit(0); // Don't fail, just skip
}

const dest = path.join(__dirname, '..', 'codex-cli' + (platform === 'win32' ? '.exe' : ''));
const tempDir = path.join(__dirname, '..', '.temp-extract');

console.log(`Fetching release info for v${version}...`);

// Fetch release information from GitHub API
const apiUrl = `https://api.github.com/repos/evertonmj/codex/releases/tags/v${version}`;

https.get(apiUrl, {
  headers: { 'User-Agent': 'codex-postinstall' }
}, (res) => {
  let data = '';
  res.on('data', chunk => { data += chunk; });
  res.on('end', () => {
    try {
      const release = JSON.parse(data);
      
      // Check if the API returned an error
      if (release.message) {
        console.log(`⚠️  Release v${version} not found on GitHub`);
        console.log(`   This is normal during development before publishing`);
        console.log(`   The codex-cli binary will be available after creating a GitHub release`);
        process.exit(0); // Don't fail, just skip
      }
      
      if (!release.assets || !Array.isArray(release.assets)) {
        console.log(`⚠️  Invalid release data: no assets found`);
        process.exit(0); // Don't fail, just skip
      }
      
      // Find the correct binary for this platform/arch
      const pattern = `codexdb_${version}_${osName}_${arch}`;
      const asset = release.assets.find(a => a.name.includes(pattern));
      
      if (!asset) {
        console.log(`⚠️  No binary found for ${osName}/${arch}`);
        console.log(`   Expected pattern: ${pattern}`);
        console.log(`   Available assets: ${release.assets.map(a => a.name).join(', ')}`);
        process.exit(0); // Don't fail, just skip
      }

      console.log(`Downloading ${asset.name}...`);
      
      // Create temp directory
      if (!fs.existsSync(tempDir)) {
        fs.mkdirSync(tempDir, { recursive: true });
      }

      const tempFile = path.join(tempDir, asset.name);
      
      https.get(asset.browser_download_url, (res) => {
        if (res.statusCode !== 200) {
          console.log(`⚠️  Failed to download binary: ${res.statusCode}`);
          process.exit(0); // Don't fail, just skip
        }
        const file = fs.createWriteStream(tempFile);
        res.pipe(file);
        file.on('finish', () => {
          file.close(() => {
            extractAndInstall(tempFile, asset.name, dest, tempDir);
          });
        });
      }).on('error', (err) => {
        console.log(`⚠️  Download error: ${err.message}`);
        process.exit(0); // Don't fail, just skip
      });
    } catch (err) {
      console.log(`⚠️  Failed to parse release info: ${err.message}`);
      process.exit(0); // Don't fail, just skip
    }
  });
}).on('error', (err) => {
  console.log(`⚠️  API error: ${err.message}`);
  console.log(`   The codex-cli binary must be downloaded manually or will be installed on first use`);
  process.exit(0); // Don't fail, just skip
});

function extractAndInstall(archivePath, archiveName, destPath, tempDir) {
  try {
    console.log(`Extracting ${archiveName}...`);
    
    if (archiveName.endsWith('.tar.gz')) {
      execSync(`tar -xzf ${archivePath} -C ${tempDir}`);
    } else if (archiveName.endsWith('.zip')) {
      execSync(`unzip -q ${archivePath} -d ${tempDir}`);
    } else if (archiveName.endsWith('.deb')) {
      // For .deb files, extract the data
      execSync(`ar x ${archivePath} data.tar.xz -o ${tempDir}/data.tar.xz`);
      execSync(`tar -xf ${tempDir}/data.tar.xz -C ${tempDir}`);
    }
    
    // Find the binary in the extracted files
    const files = execSync(`find ${tempDir} -type f -name 'codex-cli*' -o -name 'codexdb' 2>/dev/null`, { encoding: 'utf-8' }).trim().split('\n').filter(f => f);
    
    if (files.length === 0) {
      console.log('⚠️  Binary not found in archive');
      process.exit(0); // Don't fail, just skip
    }
    
    const binaryPath = files[0];
    fs.copyFileSync(binaryPath, destPath);
    fs.chmodSync(destPath, 0o755);
    
    // Cleanup
    execSync(`rm -rf ${tempDir}`);
    
    console.log('✓ codex-cli installed to', destPath);
    process.exit(0);
  } catch (err) {
    console.log(`⚠️  Extraction error: ${err.message}`);
    process.exit(0); // Don't fail, just skip
  }
}
