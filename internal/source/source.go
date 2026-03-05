package source

import (
	"fmt"
	"os"
	"strings"
)

const containerRoot = "/var/www/html"

// MapPath converts a container-side file URI (as reported by Xdebug) to a
// host filesystem path.
//
// It strips the "file://" prefix, then replaces the container root
// (/var/www/html) with the host project root.
//
// Host project root is determined from environment variables:
//   - XDEBUG_TUI_PROJECT_ROOT (if set, used directly)
//   - Otherwise: $DDEV_APPROOT (the DDEV project root, set by DDEV for host commands)
func MapPath(fileURI string) (string, error) {
	// Strip file:// prefix
	path := strings.TrimPrefix(fileURI, "file://")

	// Determine host project root
	hostRoot := os.Getenv("XDEBUG_TUI_PROJECT_ROOT")
	if hostRoot == "" {
		ddevApproot := os.Getenv("DDEV_APPROOT")
		if ddevApproot == "" {
			return "", fmt.Errorf("neither XDEBUG_TUI_PROJECT_ROOT nor DDEV_APPROOT is set")
		}
		hostRoot = ddevApproot
	}

	// Replace container root with host root
	if !strings.HasPrefix(path, containerRoot) {
		return "", fmt.Errorf("path %q does not start with container root %q", path, containerRoot)
	}
	hostPath := hostRoot + strings.TrimPrefix(path, containerRoot)
	return hostPath, nil
}

// ContainerPath converts a host filesystem path back to a container-side path.
// This is the inverse of MapPath, used when sending breakpoint URIs to Xdebug.
func ContainerPath(hostPath string) (string, error) {
	hostRoot := os.Getenv("XDEBUG_TUI_PROJECT_ROOT")
	if hostRoot == "" {
		ddevApproot := os.Getenv("DDEV_APPROOT")
		if ddevApproot == "" {
			return "", fmt.Errorf("neither XDEBUG_TUI_PROJECT_ROOT nor DDEV_APPROOT is set")
		}
		hostRoot = ddevApproot
	}

	if !strings.HasPrefix(hostPath, hostRoot) {
		return "", fmt.Errorf("path %q does not start with host root %q", hostPath, hostRoot)
	}
	containerPath := containerRoot + strings.TrimPrefix(hostPath, hostRoot)
	return "file://" + containerPath, nil
}

// Format reads the source file at hostPath and returns a tview-formatted string
// with line numbers and the current line highlighted using a region tag.
//
// The returned string uses tview region syntax: ["N"]text[""] wraps the current
// line so it can be highlighted and scrolled to with Highlight("N") on the TextView.
//
// If the file cannot be read, a human-readable error message is returned (no error
// value) so the caller can display it directly in the Source panel.
func Format(hostPath string, currentLine int) (string, error) {
	data, err := os.ReadFile(hostPath)
	if err != nil {
		return fmt.Sprintf("source not found: %s", hostPath), nil
	}

	lines := strings.Split(string(data), "\n")
	// Remove trailing empty element from final newline
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	var sb strings.Builder
	for i, line := range lines {
		lineNum := i + 1

		// Escape tview color/region tag brackets in source content so PHP code
		// like arrays ($a[0]) or annotations doesn't get misinterpreted.
		escaped := strings.ReplaceAll(line, "[", "[[")

		if lineNum == currentLine {
			// Wrap in a named region for Highlight() + ScrollToHighlight().
			// [black:yellow] = black text on yellow background (current line marker).
			// [-:-:-] resets all colour/style attributes.
			fmt.Fprintf(&sb, "[\"%d\"][black:yellow]%4d \u2502 %s[-:-:-][\"\"]\n", lineNum, lineNum, escaped)
		} else {
			fmt.Fprintf(&sb, "%4d \u2502 %s\n", lineNum, escaped)
		}
	}

	return sb.String(), nil
}
