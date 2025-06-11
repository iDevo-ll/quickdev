@echo off
echo Starting TypeScript manual test suite for Nehonix File Reloader (NHR)
echo.
echo Setup Steps:
echo 1. Installing TypeScript dependencies...
call npm install
echo.
echo Test Steps:
echo 1. The TypeScript test script will start and show version 1
echo 2. Wait for a heartbeat message
echo 3. Manually edit test_script.ts to change version = 2
echo 4. Watch the reloader detect changes and restart
echo 5. Verify the new version number appears
echo.
echo Press any key to start the test...
pause > nul

REM Get the directory where the batch file is located
set "SCRIPT_DIR=%~dp0"
REM Go up two levels to the project root
cd "%SCRIPT_DIR%..\.."
REM Run the Go program with ts-node
go run main.go -script "npx ts-node %SCRIPT_DIR%test_script.ts" 