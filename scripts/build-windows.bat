echo Building for Windows...
set GOOS=windows
set GOARCH=amd64
go build -o dist/keeper-client-windows-amd64.exe cmd/client/main.go
go build -o dist/keeper-server-windows-amd64.exe cmd/server/main.go
echo Build completed!