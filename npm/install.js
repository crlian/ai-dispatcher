#!/usr/bin/env node

const fs = require('fs');
const path = require('path');
const https = require('https');
const { spawnSync } = require('child_process');

// Get package version
const packageJson = require('./package.json');
const version = packageJson.version;

// Detect platform and architecture
const platform = process.platform; // darwin, linux, win32
const arch = process.arch; // x64, arm64

// Map Node.js platform/arch to Go build names
const platformMap = {
  darwin: 'darwin',
  linux: 'linux',
  win32: 'windows'
};

const archMap = {
  x64: 'amd64',
  arm64: 'arm64'
};

const goPlatform = platformMap[platform];
const goArch = archMap[arch];

if (!goPlatform || !goArch) {
  console.error(`Unsupported platform: ${platform}-${arch}`);
  console.error('ai-router supports: darwin/linux/windows on amd64/arm64');
  process.exit(1);
}

// Build binary name
const binaryName = platform === 'win32'
  ? `ai-router-${goPlatform}-${goArch}.exe`
  : `ai-router-${goPlatform}-${goArch}`;

const binaryPath = path.join(__dirname, 'bin', platform === 'win32' ? 'ai-router.exe' : 'ai-router');

// GitHub release URL
const downloadUrl = `https://github.com/crlian/ai-dispatcher/releases/download/v${version}/${binaryName}`;

console.log('📦 Installing ai-router...');
console.log(`   Platform: ${platform}-${arch}`);
console.log(`   Version: v${version}`);
console.log(`   Downloading: ${binaryName}`);
console.log();

// Create bin directory if it doesn't exist
const binDir = path.join(__dirname, 'bin');
if (!fs.existsSync(binDir)) {
  fs.mkdirSync(binDir, { recursive: true });
}

// Download binary
downloadBinary(downloadUrl, binaryPath)
  .then(() => {
    // Make executable on Unix systems
    if (platform !== 'win32') {
      try {
        fs.chmodSync(binaryPath, 0o755);
      } catch (err) {
        console.error('Failed to make binary executable:', err.message);
        process.exit(1);
      }
    }

    console.log('✅ ai-router installed successfully!');
    console.log();
    console.log('Try it out:');
    console.log('  ai-router --help');
    console.log('  ai-router status');
    console.log();
  })
  .catch((err) => {
    console.error('❌ Installation failed:', err.message);
    console.error();
    console.error('Alternative installation methods:');
    console.error('  1. Install from source:');
    console.error('     git clone https://github.com/crlian/ai-dispatcher.git');
    console.error('     cd ai-dispatcher');
    console.error('     make install');
    console.error();
    console.error('  2. Download binary manually from:');
    console.error(`     ${downloadUrl}`);
    console.error();
    process.exit(1);
  });

function downloadBinary(url, destination) {
  return new Promise((resolve, reject) => {
    const file = fs.createWriteStream(destination);

    https.get(url, (response) => {
      // Handle redirects
      if (response.statusCode === 302 || response.statusCode === 301) {
        file.close();
        fs.unlinkSync(destination);
        return downloadBinary(response.headers.location, destination)
          .then(resolve)
          .catch(reject);
      }

      if (response.statusCode !== 200) {
        file.close();
        fs.unlinkSync(destination);
        reject(new Error(`Download failed with status ${response.statusCode}`));
        return;
      }

      const totalSize = parseInt(response.headers['content-length'], 10);
      let downloadedSize = 0;
      let lastPercent = 0;

      response.on('data', (chunk) => {
        downloadedSize += chunk.length;
        const percent = Math.floor((downloadedSize / totalSize) * 100);

        if (percent !== lastPercent && percent % 10 === 0) {
          process.stdout.write(`\r   Progress: ${percent}%`);
          lastPercent = percent;
        }
      });

      response.pipe(file);

      file.on('finish', () => {
        file.close();
        process.stdout.write('\r   Progress: 100%\n');
        resolve();
      });

      file.on('error', (err) => {
        file.close();
        fs.unlinkSync(destination);
        reject(err);
      });
    }).on('error', (err) => {
      file.close();
      fs.unlinkSync(destination);
      reject(err);
    });
  });
}
