@echo off
setlocal

:: Remove existing build if it exists
if exist bin\watchtower.exe (
    echo Removing existing build...
    del /f /q bin\watchtower.exe
)

:: Create bin directory if it doesn't exist
if not exist bin mkdir bin

:: Build the binary
echo Building WatchTower...
go build -o bin/watchtower.exe ./internal

if errorlevel 1 (
    echo Build failed!
    exit /b 1
)

echo Build successful! Binary created at bin/watchtower.exe 