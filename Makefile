build:
	@go build -o bin/server

run: build
	@./bin/server

test:
	@go test ./... 

win: 
	@GOOS=windows GOARCH=amd64 go build -o bin/server.exe
	