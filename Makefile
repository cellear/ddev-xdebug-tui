# Makefile for ddev-xdebug-tui
#
# Development Reinstall Workflow:
#   1. make install                               # Build and install binary to ~/go/bin/
#   2. ddev add-on get /absolute/path/to/ddev-xdebug-tui # Install add-on into DDEV project
#   3. ddev xdebug-tui                            # Launch the debugger
#
# Note: ddev get will reinstall without requiring uninstall between iterations.

.PHONY: build install clean

build:
	@mkdir -p bin
	go build -o bin/ddev-xdebug-tui ./cmd/ddev-xdebug-tui

install: build
	@mkdir -p ~/go/bin
	cp bin/ddev-xdebug-tui ~/go/bin/ddev-xdebug-tui
	@echo "Binary installed to ~/go/bin/ddev-xdebug-tui"
	@echo "Ensure ~/go/bin is in your PATH for 'ddev xdebug-tui' to work"

clean:
	rm -rf bin/
