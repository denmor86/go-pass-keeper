echo Building for macOS...
set GOOS=darwin
set GOARCH=amd64
go build -o dist/keeper-client-macos-amd64.exe cmd/client/main.go
go build -o dist/keeper-server-macos-amd64.exe cmd/server/main.go
echo Build completed!