@echo off
setlocal

:: Remove existing build if it exists
if exist bin\quickdev.exe (
    echo Removing existing build...
    del /f /q bin\quickdev.exe
)

:: Create bin directory if it doesn't exist
if not exist bin mkdir bin

:: Build the binary
echo Building quickdev...
go build -o bin/quickdev.exe ./internal

if errorlevel 1 (
    echo Build failed!
    exit /b 1
)

echo Build successful! Binary created at bin/quickdev.exe 