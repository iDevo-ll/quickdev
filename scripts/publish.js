#!/usr/bin/env node

/**
 * Publish script for Nehonix QuickDev
 *
 * This script prepares and publishes the package to npm
 */

const fs = require('fs');
const path = require('path');
const { execSync } = require('child_process');

/**
 * Check if all required binaries exist
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
    console.error('\nPlease run: node scripts/build-all.js');
    process.exit(1);
  }

  console.log('✅ All required binaries found');
}

/**
 * Validate package.json
 */
function validatePackage() {
  const packagePath = path.join(__dirname, '..', 'package.json');
  const pkg = JSON.parse(fs.readFileSync(packagePath, 'utf8'));

  if (!pkg.name || !pkg.version || !pkg.description) {
    console.error('❌ package.json is missing required fields');
    process.exit(1);
  }

  console.log(`✅ Package: ${pkg.name}@${pkg.version}`);
  return pkg;
}

/**
 * Run npm publish
 */
function publish(dryRun = false) {
  const command = dryRun ? 'npm publish --dry-run --access=public' : 'npm publish --access=public';

  console.log(`\n🚀 Running: ${command}`);

  try {
    execSync(command, { stdio: 'inherit' });

    if (dryRun) {
      console.log('\n✅ Dry run completed successfully!');
      console.log('To actually publish, run: node scripts/publish.js --publish');
    } else {
      console.log('\n🎉 Package published successfully!');
    }
  } catch (error) {
    console.error('\n❌ Publish failed:', error.message);
    process.exit(1);
  }
}

/**
 * Main function
 */
function main() {
  const args = process.argv.slice(2);
  const shouldPublish = args.includes('--publish');

  console.log('📦 Nehonix QuickDev - Publish Script');
  console.log('=====================================\n');

  // Check prerequisites
  checkBinaries();
  const pkg = validatePackage();

  // Show what will be published
  console.log('\n📋 Files to be published:');
  const files = pkg.files || [];
  files.forEach(file => console.log(`  - ${file}`));

  if (!shouldPublish) {
    console.log('\n🔍 Running dry run...');
    publish(true);
  } else {
    console.log('\n⚠️  This will publish to npm registry!');
    console.log('Make sure you are logged in: npm login');
    console.log('Make sure the version is correct in package.json');

    // Confirm publication
    const readline = require('readline');
    const rl = readline.createInterface({
      input: process.stdin,
      output: process.stdout
    });

    rl.question('\nContinue with publication? (y/N): ', (answer) => {
      rl.close();

      if (answer.toLowerCase() === 'y' || answer.toLowerCase() === 'yes') {
        publish(false);
      } else {
        console.log('❌ Publication cancelled');
        process.exit(0);
      }
    });
  }
}

main();
