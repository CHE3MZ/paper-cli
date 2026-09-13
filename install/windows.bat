@echo off
setlocal
set "REPO=CHE3MZ/paper-cli"
set "ASSET=paper-windows.exe"
set "URL=https://github.com/%REPO%/releases/latest/download/%ASSET%"
set "DEST=%USERPROFILE%\.local\bin\paper.exe"

echo Installing paper for Windows...
echo   from: %URL%
echo   to:   %DEST%

if not exist "%USERPROFILE%\.local\bin" mkdir "%USERPROFILE%\.local\bin"

where curl >nul 2>nul
if %ERRORLEVEL%==0 (
  curl -fL -o "%DEST%" "%URL%"
) else (
  powershell -NoProfile -ExecutionPolicy Bypass -Command "Invoke-WebRequest -UseBasicParsing -Uri '%URL%' -OutFile '%DEST%'"
)
if %ERRORLEVEL% neq 0 (
  echo error: download failed. Check your network and that a release exists at https://github.com/%REPO%/releases >&2
  exit /b 1
)

echo Installed to %DEST%
echo Make sure %%USERPROFILE%%\.local\bin is on your PATH. Test with: paper help
