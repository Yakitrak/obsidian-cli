BINARY_NAME=obsidian-cli

build-all:
	GOOS=darwin GOARCH=amd64 go build -o bin/darwin/${BINARY_NAME}
	GOOS=linux GOARCH=amd64 go build -o bin/linux/${BINARY_NAME}
	GOOS=windows GOARCH=amd64 go build -o bin/windows/${BINARY_NAME}.exe

clean-all:
	go clean
	rm bin/darwin/${BINARY_NAME}
	rm bin/linux/${BINARY_NAME}
	rm bin/windows/${BINARY_NAME}.exe

test:
	go test ./...

test-coverage:
	go test ./... -coverprofile=coverage.out

update-usage-image:
	freeze --execute "go run main.go --help" --theme dracula  --output docs/usage.png