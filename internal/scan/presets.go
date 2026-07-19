package scan

var DefaultNames = []string{
	"node_modules",
	".venv",
	"venv",
	"env",
	"__pycache__",
	".pytest_cache",
	".mypy_cache",
	".ruff_cache",
	".tox",
	"dist",
	"build",
	".next",
	".nuxt",
	".svelte-kit",
	".turbo",
	".cache",
	".parcel-cache",
	"Pods",
	"DerivedData",
	".gradle",
}

var projectMarkers = []string{
	".git",
	"package.json",
	"pnpm-workspace.yaml",
	"yarn.lock",
	"package-lock.json",
	"bun.lock",
	"bun.lockb",
	"pyproject.toml",
	"requirements.txt",
	"Pipfile",
	"poetry.lock",
	"go.mod",
	"Cargo.toml",
	"Gemfile",
	"Podfile",
	"Makefile",
}

var alwaysSkipNames = map[string]struct{}{
	".git":                    {},
	".hg":                     {},
	".svn":                    {},
	".Trash":                  {},
	".Spotlight-V100":         {},
	".fseventsd":              {},
	".DocumentRevisions-V100": {},
	".TemporaryItems":         {},
}
