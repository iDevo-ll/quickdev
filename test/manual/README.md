# Manual Tests for Nehonix File Reloader

This directory contains manual tests for verifying the functionality of the Nehonix File Reloader (NHR).

## Test Files

- `test_script.js`: A simple Node.js script that displays its version number and heartbeat
- `test.bat`: Windows batch script to run the test scenario

## Running the Tests

1. Open a terminal in this directory
2. Run `test.bat`
3. Follow the instructions displayed in the terminal

## Test Scenario

The test verifies these key features:
1. **Basic Execution**: Ability to run a Node.js script
2. **File Watching**: Detection of file changes
3. **Hot Reloading**: Proper process termination and restart
4. **Output Handling**: Proper display of console output

## Manual Test Steps

1. Start the test using `test.bat`
2. Wait for the initial output showing version 1
3. Wait for at least one heartbeat message
4. Edit `test_script.js`:
   - Change `let version = 1` to `let version = 2`
   - Save the file
5. Observe:
   - File change detection message
   - Process restart
   - New version number in the output

## Expected Results

- The script should start and show version 1
- After editing, it should restart automatically
- The new version (2) should be displayed
- Heartbeat messages should continue with the new version

## Troubleshooting

If the test fails:
1. Check if the NHR process is running
2. Verify the file path in test.bat is correct
3. Check if Node.js is installed and accessible
4. Verify the file changes are being saved 