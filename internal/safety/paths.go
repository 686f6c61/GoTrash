package safety

import (
	"path/filepath"
	"strings"
)

const ProtectedPathReason = "ruta gestionada por sistema o package manager"

var protectedPrefixes = []string{
	"/opt/homebrew",
	"/usr/local",
	"/opt/local",
	"/Library",
	"/System",
	"/Applications",
	"/usr",
	"/bin",
	"/sbin",
	"/dev",
	"/private/preboot",
	"/private/var/db",
	"/private/var/vm",
}

func IsProtectedPath(path string) bool {
	cleanPath := filepath.Clean(path)
	for _, prefix := range protectedPrefixes {
		if cleanPath == prefix || strings.HasPrefix(cleanPath, prefix+string(filepath.Separator)) {
			return true
		}
	}
	return false
}

func GuardReason(path string) string {
	if IsProtectedPath(path) {
		return ProtectedPathReason
	}
	return ""
}
