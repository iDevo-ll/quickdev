# Publishing Nehonix QuickDev to npm

This document explains how to publish the QuickDev package to npm.

## Prerequisites

1. **npm account**: Make sure you have an npm account and are logged in
   ```bash
   npm login
   ```

2. **Package name availability**: The name "quickdev" should be available on npm
   ```bash
   npm view quickdev
   ```
   If this returns an error, the name is available.

## Publishing Steps

### 1. Verify Everything is Ready

Run the test script to ensure the package is properly configured:
```bash
node test-npm-package.js
```

### 2. Run Dry Run

Test the publishing process without actually publishing:
```bash
npm run publish:dry
```
or
```bash
node scripts/publish.js
```

This will:
- Check that all required binaries exist
- Validate package.json
- Show what files will be published
- Run `npm publish --dry-run`

### 3. Publish to npm

When you're ready to publish for real:
```bash
npm run publish:real
```
or
```bash
node scripts/publish.js --publish
```

This will ask for confirmation before publishing.

## What Gets Published

The npm package includes:
- Pre-built binaries for all supported platforms:
  - Windows (x64, arm64)
  - macOS (x64, arm64) 
  - Linux (x64, arm64)
- Node.js wrapper scripts
- Installation scripts
- Documentation

## After Publishing

Once published, users can install and use QuickDev with:

```bash
# Install globally
npm install -g quickdev

# Use it
quickdev -script your-script.js
```

## Version Management

To publish a new version:

1. Update the version in `package.json`
2. Rebuild binaries: `npm run build`
3. Test: `npm run publish:dry`
4. Publish: `npm run publish:real`

## Package Structure

```
quickdev/
├── bin/                          # Pre-built binaries
│   ├── quickdev-windows-amd64.exe
│   ├── quickdev-windows-arm64.exe
│   ├── quickdev-linux-amd64
│   ├── quickdev-linux-arm64
│   ├── quickdev-darwin-amd64
│   ├── quickdev-darwin-arm64
│   └── quickdev.js              # CLI wrapper
├── scripts/
│   ├── install.js               # Post-install script
│   └── publish.js               # Publishing helper
├── index.js                     # Main entry point
├── package.json
└── README.md
```

## Troubleshooting

### Binary Not Found Error
If users get "Binary not found" errors:
1. Check that all platform binaries are included in the package
2. Verify the binary naming matches the detection logic in `index.js`
3. Ensure the post-install script runs correctly

### Permission Errors
On Unix systems, if binaries aren't executable:
1. The post-install script should handle this automatically
2. Users can manually fix with: `chmod +x $(which quickdev)`

### Platform Support
Currently supported platforms:
- Windows: x64, arm64
- macOS: x64, arm64  
- Linux: x64, arm64

To add support for new platforms, update:
1. `scripts/build-all.js` - Add new platform to build targets
2. `index.js` - Add platform detection logic
3. `scripts/install.js` - Update platform validation
