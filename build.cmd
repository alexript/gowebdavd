@echo off
REM Windows build script for gowebdavd

if "%1"=="" goto :build
if "%1"=="build" goto :build
if "%1"=="build-release" goto :build-release
if "%1"=="test" goto :test
if "%1"=="cover" goto :cover
if "%1"=="clean" goto :clean
if "%1"=="fmt" goto :fmt
if "%1"=="vet" goto :vet
if "%1"=="tidy" goto :tidy
if "%1"=="run" goto :run
if "%1"=="help" goto :help
goto :help

:build
if not exist bin mkdir bin
go build -o bin\gowebdavd.exe .\cmd\gowebdavd
echo Build complete: bin\gowebdavd.exe
goto :eof

:build-release
if not exist bin mkdir bin
go build -ldflags="-s -w" -o bin\gowebdavd.exe .\cmd\gowebdavd
echo Release build complete: bin\gowebdavd.exe
goto :eof

:test
go test .\...
goto :eof

:cover
go test -coverprofile=coverage.out .\...
go tool cover -func=coverage.out
goto :eof

:clean
go clean
go clean -testcache
if exist bin rmdir /s /q bin
if exist coverage.out del /f coverage.out
if exist coverage.html del /f coverage.html
if exist coverage rmdir /s /q coverage
if exist gowebdavd.exe del /f gowebdavd.exe
echo Clean complete
goto :eof

:fmt
go fmt .\...
goto :eof

:vet
go vet .\...
goto :eof

:tidy
go mod tidy
goto :eof

:run
call :build
bin\gowebdavd.exe run -dir . -port 8080 -bind 127.0.0.1
goto :eof

:help
echo Usage: build.cmd [command]
echo.
echo Commands:
echo   build         - Build the project (debug)
echo   build-release - Build the project (release with -s -w flags)
echo   test          - Run tests
echo   cover         - Run tests with coverage
echo   clean         - Clean build artifacts
echo   fmt           - Format Go code
echo   vet           - Vet Go code
echo   tidy          - Tidy go.mod
echo   run           - Build and run server
echo   help          - Show this help
goto :eof
