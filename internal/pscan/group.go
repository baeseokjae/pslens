package pscan

import (
	"path/filepath"
	"strings"
)

// AppKey returns a human-friendly application name for a process.
func AppKey(p Process) string {
	c := p.Comm
	if strings.Contains(c, ".app/") {
		parts := strings.Split(c, ".app/")
		if len(parts) >= 1 {
			base := filepath.Base(parts[0])
			return strings.TrimSuffix(base, ".app")
		}
	}
	if strings.Contains(c, ".local/bin/claude") {
		return "Claude"
	}
	if strings.Contains(c, "/codex") || strings.Contains(c, "codex-darwin") {
		return "Codex"
	}
	if strings.Contains(c, "/node") || strings.HasSuffix(c, "/node") {
		return "node"
	}
	short := filepath.Base(c)
	if short == "" || short == "." || short == "/" {
		return c
	}
	return short
}