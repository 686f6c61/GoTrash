package cmd

import "strings"

func candidateTypeIcon(name string) string {
	switch strings.ToLower(name) {
	case "node_modules", ".next", ".nuxt", ".svelte-kit", ".turbo", ".parcel-cache":
		return "⬢"
	case ".venv", "venv", "env", "__pycache__", ".pytest_cache", ".mypy_cache", ".ruff_cache", ".tox":
		return "◌"
	case "deriveddata", "pods":
		return "⌘"
	case ".cache", ".gradle", ".npm":
		return "◔"
	case "dist", "build":
		return "▣"
	default:
		return "•"
	}
}
