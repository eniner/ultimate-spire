@echo off
REM Starts Spire (dev build from source) against the UltEQTest server.
REM Backend must run from the server folder so it finds eqemu_config.json.

set "SPIRE_SRC=%~dp0"
set "EQ_SERVER=C:\Users\E9ine\Desktop\Emul Stuff\ultimate_eq_original_final_backup\eqemu_installer_files"

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
