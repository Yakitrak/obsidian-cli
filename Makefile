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

# Release automation
# Usage: make release VERSION=v0.2.2
release:
ifndef VERSION
	$(error VERSION is not set. Usage: make release VERSION=v0.2.2)
endif
	@echo "Starting release process for $(VERSION)..."
	@# Update version in root.go
	@sed -i '' 's/Version: "v[0-9]*\.[0-9]*\.[0-9]*"/Version: "$(VERSION)"/' cmd/root.go
	@echo "✓ Updated version in root.go to $(VERSION)"
	@# Create screenshot
	@$(MAKE) update-usage-image
	@echo "✓ Created usage screenshot"
	@# Build all binaries
	@$(MAKE) build-all
	@echo "✓ Built binaries for all platforms"
	@# Git operations
	@git add cmd/root.go docs/usage.png
	@git commit -m "chore: bump version to $(VERSION)"
	@git tag $(VERSION)
	@git push origin main
	@git push origin $(VERSION)
	@echo "✓ Release $(VERSION) complete!"

# Quick release (interactive version bump)
release-patch:
	@$(eval CURRENT_VERSION := $(shell grep 'Version:' cmd/root.go | sed 's/.*"v\([0-9]*\.[0-9]*\.[0-9]*\)".*/\1/'))
	@$(eval NEW_VERSION := $(shell echo $(CURRENT_VERSION) | awk -F. '{print "v" $$1 "." $$2 "." $$3+1}'))
	@$(MAKE) release VERSION=$(NEW_VERSION)

release-minor:
	@$(eval CURRENT_VERSION := $(shell grep 'Version:' cmd/root.go | sed 's/.*"v\([0-9]*\.[0-9]*\.[0-9]*\)".*/\1/'))
	@$(eval NEW_VERSION := $(shell echo $(CURRENT_VERSION) | awk -F. '{print "v" $$1 "." $$2+1 ".0"}'))
	@$(MAKE) release VERSION=$(NEW_VERSION)

release-major:
	@$(eval CURRENT_VERSION := $(shell grep 'Version:' cmd/root.go | sed 's/.*"v\([0-9]*\.[0-9]*\.[0-9]*\)".*/\1/'))
	@$(eval NEW_VERSION := $(shell echo $(CURRENT_VERSION) | awk -F. '{print "v" $$1+1 ".0.0"}'))
	@$(MAKE) release VERSION=$(NEW_VERSION)