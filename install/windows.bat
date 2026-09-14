@echo off
setlocal
set "REPO=CHE3MZ/paper-cli"
set "ASSET=paper-windows.exe"
set "URL=https://github.com/%REPO%/releases/latest/download/%ASSET%"
set "DEST=%USERPROFILE%\.local\bin\paper.exe"
set "TMP=%DEST%.tmp"

echo Installing paper for Windows...
echo   from: %URL%
echo   to:   %DEST%

if not exist "%USERPROFILE%\.local\bin" mkdir "%USERPROFILE%\.local\bin"

where curl >nul 2>nul
if %ERRORLEVEL%==0 (
  curl -fL -o "%TMP%" "%URL%"
) else (
  powershell -NoProfile -ExecutionPolicy Bypass -Command "$ProgressPreference='SilentlyContinue'; Invoke-WebRequest -UseBasicParsing -Uri '%URL%' -OutFile '%TMP%'"
)
if %ERRORLEVEL% neq 0 (
  del /F /Q "%TMP%" 2>nul
  echo error: download failed. Check your network and that a release exists at https://github.com/%REPO%/releases >&2
  exit /b 1
)

rem Swap into place so an interrupted download can never leave a corrupt paper.exe behind.
move /Y "%TMP%" "%DEST%" >nul
if %ERRORLEVEL% neq 0 (
  echo error: could not move "%TMP%" to "%DEST%". Copy it over by hand. >&2
  exit /b 1
)

echo Installed to %DEST%
echo Make sure %%USERPROFILE%%\.local\bin is on your PATH. Test with: paper help
