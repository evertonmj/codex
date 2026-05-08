const https = require('https');
const fs = require('fs');
const os = require('os');
const fs = require('fs');
const os = require('os');
const path = require('path');
const { https } = require('follow-redirects');

// Map OS/platform to binary URLs
const BINARIES = {
  darwin: 'https://github.com/evertonmj/codex/releases/latest/download/codex-cli-macos',
  linux: 'https://github.com/evertonmj/codex/releases/latest/download/codex-cli-linux',
  win32: 'https://github.com/evertonmj/codex/releases/latest/download/codex-cli-win.exe'
};

const platform = os.platform();
const url = BINARIES[platform];
if (!url) {
  console.error(`No binary available for platform: ${platform}`);
  process.exit(1);
}

const dest = path.join(__dirname, '..', 'codex-cli' + (platform === 'win32' ? '.exe' : ''));

console.log(`Downloading codex-cli for ${platform}...`);

https.get(url, (res) => {
  if (res.statusCode !== 200) {
    console.error(`Failed to download binary: ${res.statusCode}`);
    process.exit(1);
  }
  const file = fs.createWriteStream(dest, { mode: 0o755 });
  res.pipe(file);
  file.on('finish', () => {
    file.close(() => {
      console.log('codex-cli downloaded to', dest);
    });
  });
}).on('error', (err) => {
  console.error('Download error:', err);
  process.exit(1);
});
