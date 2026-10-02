@echo off
REM Starts Spire (dev build from source) against a local EQEmu server.
REM Backend must run from the server folder so it finds eqemu_config.json.

set "SPIRE_SRC=%~dp0"

if not defined EQ_SERVER (
  if exist "%SPIRE_SRC%eqemu_config.json" (
    set "EQ_SERVER=%SPIRE_SRC%"
  ) else (
    echo Set EQ_SERVER to your EQEmu server folder ^(the one with eqemu_config.json^).
    pause
    exit /b 1
  )
)

if not exist "%EQ_SERVER%\eqemu_config.json" (
  echo EQ_SERVER does not contain eqemu_config.json: %EQ_SERVER%
  pause
  exit /b 1
)

if not defined SPIRE_QUESTS_ROOT set "SPIRE_QUESTS_ROOT=%EQ_SERVER%\quests"
if not defined SPIRE_WEBSITE_ROOT set "SPIRE_WEBSITE_ROOT=%EQ_SERVER%\web"

taskkill /IM spire-dev.exe /F >nul 2>&1
taskkill /IM spire-windows-amd64.exe /F >nul 2>&1

echo Building backend...
pushd "%SPIRE_SRC%"
go build -o spire-dev.exe . || (echo Backend build failed & pause & exit /b 1)
popd

echo Starting backend on port 3010...
start "Spire Backend" /D "%EQ_SERVER%" "%SPIRE_SRC%spire-dev.exe" http:serve --port=3010

echo Starting frontend on port 8080...
start "Spire Frontend" /D "%SPIRE_SRC%frontend" cmd /k "set NODE_OPTIONS=--openssl-legacy-provider && node --max_old_space_size=4096 --stack-size=10000 node_modules/@vue/cli-service/bin/vue-cli-service.js serve --host=0.0.0.0 --port=8080 --open"

echo Spire will open in your browser at http://localhost:8080 once the frontend finishes compiling.
