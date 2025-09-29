echo Building for Linux...
set GOOS=linux
set GOARCH=amd64
go build -o dist/keeper-client-linux-amd64 cmd/client/main.go
go build -o dist/keeper-server-linux-amd64 cmd/server/main.go
echo Build completed!