# Makefile for ddev-xdebug-tui
#
# Development Reinstall Workflow:
#   1. make install                               # Build and install binary to ~/go/bin/
#   2. ddev add-on get /absolute/path/to/ddev-xdebug-tui # Install add-on into DDEV project
#   3. ddev xdebug-tui                            # Launch the debugger
#
# Note: ddev get will reinstall without requiring uninstall between iterations.

.PHONY: build install dist clean

PLATFORMS := darwin/amd64 darwin/arm64 linux/amd64 linux/arm64

build:
	@mkdir -p bin
	go build -o bin/ddev-xdebug-tui ./cmd/ddev-xdebug-tui

install: build
	@mkdir -p ~/go/bin
	cp bin/ddev-xdebug-tui ~/go/bin/ddev-xdebug-tui
	@echo "Binary installed to ~/go/bin/ddev-xdebug-tui"
	@echo "Ensure ~/go/bin is in your PATH for 'ddev xdebug-tui' to work"

# Cross-compile binaries for all supported platforms.
# Run before cutting a GitHub release: make dist && gh release create v0.x.0 dist/*
dist:
	@mkdir -p dist
	@for platform in $(PLATFORMS); do \
		os=$${platform%/*}; arch=$${platform#*/}; \
		echo "Building $$os/$$arch..."; \
		GOOS=$$os GOARCH=$$arch go build -o dist/ddev-xdebug-tui-$$os-$$arch ./cmd/ddev-xdebug-tui/; \
	done
	@echo ""
	@echo "Binaries ready in dist/:"
	@ls -lh dist/

clean:
	rm -rf bin/ dist/
