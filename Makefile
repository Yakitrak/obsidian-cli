BINARY_NAME=obs

build-all:
	GOOS=darwin GOARCH=amd64 go build -o bin/${BINARY_NAME}-darwin
	GOOS=linux GOARCH=amd64 go build -o bin/${BINARY_NAME}-linux
	GOOS=windows GOARCH=amd64 go build -o bin/${BINARY_NAME}-windows.exe

clean-all:
	go clean
	rm bin/${BINARY_NAME}-darwin
	rm bin/${BINARY_NAME}-linux
	rm bin/${BINARY_NAME}-windows.exe

test:
	go test ./...

test_coverage:
	go test ./... -coverprofile=coverage.out

