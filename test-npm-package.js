#!/usr/bin/env node

/**
 * Test script to verify the npm package works correctly
 */

const { getBinaryName, getBinaryPath } = require('./index.js');
const fs = require('fs');

console.log('🧪 Testing Nehonix QuickDev npm package...\n');

try {
  // Test binary detection
  const binaryName = getBinaryName();
  console.log(`✅ Detected binary: ${binaryName}`);
  
  // Test binary path
  const binaryPath = getBinaryPath();
  console.log(`✅ Binary path: ${binaryPath}`);
  
  // Test binary exists
  if (fs.existsSync(binaryPath)) {
    console.log('✅ Binary file exists');
  } else {
    throw new Error('Binary file not found');
  }
  
  // Test binary is executable (on Unix systems)
  const stats = fs.statSync(binaryPath);
  if (process.platform !== 'win32') {
    const isExecutable = (stats.mode & parseInt('111', 8)) !== 0;
    if (isExecutable) {
      console.log('✅ Binary is executable');
    } else {
      console.log('⚠️  Binary may not be executable');
    }
  }
  
  console.log('\n🎉 All tests passed! Package is ready for publishing.');
  
} catch (error) {
  console.error('❌ Test failed:', error.message);
  process.exit(1);
}
