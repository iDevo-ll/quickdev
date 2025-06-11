@echo off
echo Starting manual test suite for Nehonix File Reloader (NHR)
echo.
echo Test Steps:
echo 1. The test script will start and show version 1
echo 2. Wait for a heartbeat message
echo 3. Manually edit test_script.js to change version = 2
echo 4. Watch the reloader detect changes and restart
echo 5. Verify the new version number appears
echo.
echo Press any key to start the test...
pause > nul

REM Get the directory where the batch file is located
set "SCRIPT_DIR=%~dp0"
REM Go up two levels to the project root
cd "%SCRIPT_DIR%..\.."
REM Run the Go program
go run main.go -script "%SCRIPT_DIR%test_script.js" 