#!/usr/bin/env node

/**
 * Create GitHub Release Script for Nehonix QuickDev
 * 
 * This script creates a GitHub release and uploads the binary assets.
 * You'll need to have the GitHub CLI (gh) installed and authenticated.
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

const VERSION = require('../package.json').version;
const RELEASE_TAG = `v${VERSION}`;

/**
 * Check if GitHub CLI is available
 */
function checkGitHubCLI() {
  try {
    execSync('gh --version', { stdio: 'pipe' });
    console.log('✅ GitHub CLI found');
  } catch (error) {
    console.error('❌ GitHub CLI not found. Please install it:');
    console.error('https://cli.github.com/');
    process.exit(1);
  }
}

/**
 * Check if all binaries exist
 */
function checkBinaries() {
  const requiredBinaries = [
    'quickdev-windows-amd64.exe',
    'quickdev-windows-arm64.exe',
    'quickdev-linux-amd64',
    'quickdev-linux-arm64',
    'quickdev-darwin-amd64',
    'quickdev-darwin-arm64'
  ];
  
  const binDir = path.join(__dirname, '..', 'bin');
  const missingBinaries = [];
  
  for (const binary of requiredBinaries) {
    const binaryPath = path.join(binDir, binary);
    if (!fs.existsSync(binaryPath)) {
      missingBinaries.push(binary);
    }
  }
  
  if (missingBinaries.length > 0) {
    console.error('❌ Missing binaries:');
    missingBinaries.forEach(binary => console.error(`  - ${binary}`));
    console.error('\nPlease run: npm run build');
    process.exit(1);
  }
  
  console.log('✅ All required binaries found');
  return requiredBinaries;
}

/**
 * Create GitHub release
 */
function createRelease(binaries) {
  const binDir = path.join(__dirname, '..', 'bin');
  
  console.log(`\n🚀 Creating GitHub release: ${RELEASE_TAG}`);
  
  // Prepare release notes
  const releaseNotes = `# Nehonix QuickDev v${VERSION}

A professional-grade file watcher and development server for TypeScript/JavaScript applications.

## Installation

\`\`\`bash
npm install -g @nehonix/quickdev
\`\`\`

## What's New

- High-performance file watching and reloading
- Intelligent TypeScript/JavaScript detection
- Advanced project configuration detection
- Production-grade process management
- Professional CLI interface with color-coded output

## Supported Platforms

- Windows (x64, arm64)
- macOS (x64, arm64)
- Linux (x64, arm64)

## Usage

\`\`\`bash
quickdev -script your-app.js
\`\`\`

For more information, see the [README](https://github.com/nehonix/quickdev#readme).
`;

  // Write release notes to temporary file
  const notesFile = path.join(__dirname, 'release-notes.md');
  fs.writeFileSync(notesFile, releaseNotes);
  
  try {
    // Create the release
    const createCmd = `gh release create ${RELEASE_TAG} --title "Nehonix QuickDev v${VERSION}" --notes-file "${notesFile}"`;
    console.log('Creating release...');
    execSync(createCmd, { stdio: 'inherit' });
    
    // Upload binaries
    console.log('Uploading binaries...');
    for (const binary of binaries) {
      const binaryPath = path.join(binDir, binary);
      const uploadCmd = `gh release upload ${RELEASE_TAG} "${binaryPath}"`;
      console.log(`Uploading: ${binary}`);
      execSync(uploadCmd, { stdio: 'inherit' });
    }
    
    console.log(`\n🎉 Release created successfully!`);
    console.log(`View at: https://github.com/nehonix/quickdev/releases/tag/${RELEASE_TAG}`);
    
  } catch (error) {
    console.error('❌ Failed to create release:', error.message);
    process.exit(1);
  } finally {
    // Clean up
    if (fs.existsSync(notesFile)) {
      fs.unlinkSync(notesFile);
    }
  }
}

/**
 * Main function
 */
function main() {
  console.log('📦 Nehonix QuickDev - GitHub Release Creator');
  console.log('===========================================\n');
  
  checkGitHubCLI();
  const binaries = checkBinaries();
  
  console.log(`\n📋 Release Information:`);
  console.log(`Version: ${VERSION}`);
  console.log(`Tag: ${RELEASE_TAG}`);
  console.log(`Binaries: ${binaries.length} files`);
  
  createRelease(binaries);
}

main();
