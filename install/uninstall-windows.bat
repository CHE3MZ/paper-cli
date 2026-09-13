@echo off
setlocal
set "TARGET=%USERPROFILE%\.local\bin\paper.exe"

echo Uninstalling paper...
echo   from: %TARGET%

if exist "%TARGET%" (
  del "%TARGET%"
  echo Removed %TARGET%
) else (
  echo Nothing to do: %TARGET% not found.
)
