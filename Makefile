BINARY_NAME=obsidian-cli

RELEASE_LDFLAGS=-s -w
RELEASE_FLAGS=-trimpath -ldflags "$(RELEASE_LDFLAGS)" -buildvcs=false
VAULT_BIN_DIR=/Users/colthorp/Obsidian/drews-vault/.bin

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

test_coverage:
	go test ./... -coverprofile=coverage.out

build-stripped:
	CGO_ENABLED=0 go build $(RELEASE_FLAGS) -o bin/${BINARY_NAME}

build-small-vault: build-stripped
	install -d $(VAULT_BIN_DIR)
	install -m 0755 bin/${BINARY_NAME} $(VAULT_BIN_DIR)/${BINARY_NAME}

