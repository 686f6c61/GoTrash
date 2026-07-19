package scan

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"basura/internal/safety"
)

type Options struct {
	Roots    []string
	Names    []string
	MinSize  int64
	Progress func(Progress)
}

type Progress struct {
	Current      string
	DirsVisited  int
	Matches      int
	BytesMatched int64
}

type Result struct {
	Roots       []string
	Candidates  []Candidate
	DirsVisited int
	TotalBytes  int64
	Errors      []ScanError
}

type Candidate struct {
	Name      string
	Path      string
	Project   string
	SizeBytes int64
}

type ScanError struct {
	Path    string
	Message string
}

func Scan(options Options) (Result, error) {
	if len(options.Roots) == 0 {
		return Result{}, errors.New("no hay rutas para escanear")
	}

	names := normalizeNames(options.Names)
	result := Result{
		Roots: append([]string(nil), options.Roots...),
	}

	lastUpdate := time.Time{}
	emit := func(current string, force bool) {
		if options.Progress == nil {
			return
		}
		if !force && time.Since(lastUpdate) < 120*time.Millisecond {
			return
		}
		lastUpdate = time.Now()
		options.Progress(Progress{
			Current:      current,
			DirsVisited:  result.DirsVisited,
			Matches:      len(result.Candidates),
			BytesMatched: result.TotalBytes,
		})
	}

	for _, root := range options.Roots {
		root = filepath.Clean(root)
		info, err := os.Stat(root)
		if err != nil {
			return Result{}, err
		}
		if !info.IsDir() {
			return Result{}, errors.New("una de las rutas no es una carpeta")
		}

		err = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				result.Errors = append(result.Errors, ScanError{
					Path:    path,
					Message: walkErr.Error(),
				})
				if entry != nil && entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if entry == nil || !entry.IsDir() {
				return nil
			}

			if shouldSkipDir(path, entry.Name(), root) {
				return filepath.SkipDir
			}

			result.DirsVisited++
			name := strings.ToLower(entry.Name())
			emit(path, false)

			if _, ok := names[name]; !ok {
				return nil
			}

			size, nestedErrors := directorySize(path)
			result.Errors = append(result.Errors, nestedErrors...)

			if size >= options.MinSize {
				candidate := Candidate{
					Name:      entry.Name(),
					Path:      path,
					Project:   detectProjectRoot(path, root),
					SizeBytes: size,
				}
				result.Candidates = append(result.Candidates, candidate)
				result.TotalBytes += size
				emit(path, true)
			}

			return filepath.SkipDir
		})
		if err != nil {
			return Result{}, err
		}
	}

	sort.Slice(result.Candidates, func(i, j int) bool {
		if result.Candidates[i].SizeBytes == result.Candidates[j].SizeBytes {
			return result.Candidates[i].Path < result.Candidates[j].Path
		}
		return result.Candidates[i].SizeBytes > result.Candidates[j].SizeBytes
	})

	emit("", true)
	return result, nil
}

func normalizeNames(raw []string) map[string]struct{} {
	source := raw
	if len(source) == 0 {
		source = DefaultNames
	}

	names := make(map[string]struct{}, len(source))
	for _, item := range source {
		name := strings.ToLower(strings.TrimSpace(item))
		if name == "" {
			continue
		}
		names[name] = struct{}{}
	}
	return names
}

func shouldSkipDir(path string, base string, root string) bool {
	if safety.IsProtectedPath(path) {
		return true
	}
	if path == root {
		return false
	}
	if _, ok := alwaysSkipNames[base]; ok {
		return true
	}

	return false
}

func directorySize(root string) (int64, []ScanError) {
	var size int64
	errorsFound := make([]ScanError, 0)

	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			errorsFound = append(errorsFound, ScanError{
				Path:    path,
				Message: walkErr.Error(),
			})
			if entry != nil && entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry == nil {
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return nil
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			errorsFound = append(errorsFound, ScanError{
				Path:    path,
				Message: err.Error(),
			})
			return nil
		}
		size += info.Size()
		return nil
	})

	return size, errorsFound
}

func detectProjectRoot(target string, scanRoot string) string {
	current := filepath.Dir(target)
	fallback := current
	limit := filepath.Clean(scanRoot)

	for {
		if hasProjectMarker(current) {
			return current
		}
		if current == limit || current == filepath.Dir(current) {
			break
		}
		current = filepath.Dir(current)
	}

	return fallback
}

func hasProjectMarker(path string) bool {
	for _, marker := range projectMarkers {
		if _, err := os.Stat(filepath.Join(path, marker)); err == nil {
			return true
		}
	}
	return false
}
