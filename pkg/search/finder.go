package search

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

var defaultIgnoredDirs = map[string]bool{
	".git":         true,
	"node_modules": true,
	"vendor":       true,
	".vscode":      true,
	".idea":        true,
	"dist":         true,
	"build":        true,
	"bin":          true,
	"obj":          true,
	"__pycache__":  true,
	".venv":        true,
	"target":       true,
}

// FileEntry describes a discovered file.
type FileEntry struct {
	Path    string `json:"path"`
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	ModTime string `json:"mod_time"`
	IsDir   bool   `json:"is_dir"`
}

// FindOptions defines filtering parameters for searching files.
type FindOptions struct {
	RootPath     string
	IncludeGlobs []string
	ExcludeGlobs []string
	IncludeDirs  bool
	MaxDepth     int
	IgnoreHidden bool
}

// IsBinaryFile checks whether a file contains null bytes in its first 512 bytes.
func IsBinaryFile(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return false
	}
	defer f.Close()

	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return false
	}
	return bytes.IndexByte(buf[:n], 0) != -1
}

// FindFiles searches directories recursively using FindOptions.
func FindFiles(opts FindOptions) ([]FileEntry, error) {
	var entries []FileEntry
	root := opts.RootPath
	if root == "" {
		root = "."
	}

	rootClean := filepath.Clean(root)

	err := filepath.WalkDir(rootClean, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip inaccessible files
		}

		rel, err := filepath.Rel(rootClean, path)
		if err != nil {
			rel = path
		}

		name := d.Name()

		if d.IsDir() {
			if path != rootClean {
				if defaultIgnoredDirs[name] {
					return filepath.SkipDir
				}
				if opts.IgnoreHidden && strings.HasPrefix(name, ".") {
					return filepath.SkipDir
				}
				if opts.MaxDepth > 0 {
					depth := strings.Count(rel, string(filepath.Separator)) + 1
					if depth > opts.MaxDepth {
						return filepath.SkipDir
					}
				}
			}
			if opts.IncludeDirs && path != rootClean {
				info, _ := d.Info()
				entries = append(entries, FileEntry{
					Path:    filepath.ToSlash(path),
					Name:    name,
					IsDir:   true,
					ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
				})
			}
			return nil
		}

		// File filtering
		if opts.IgnoreHidden && strings.HasPrefix(name, ".") {
			return nil
		}

		// Check exclusion globs
		for _, eg := range opts.ExcludeGlobs {
			matched, _ := filepath.Match(eg, name)
			if matched {
				return nil
			}
		}

		// Check inclusion globs
		if len(opts.IncludeGlobs) > 0 {
			matchedAny := false
			for _, ig := range opts.IncludeGlobs {
				matched, _ := filepath.Match(ig, name)
				if matched {
					matchedAny = true
					break
				}
			}
			if !matchedAny {
				return nil
			}
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		entries = append(entries, FileEntry{
			Path:    filepath.ToSlash(path),
			Name:    name,
			Size:    info.Size(),
			IsDir:   false,
			ModTime: info.ModTime().Format("2006-01-02 15:04:05"),
		})

		return nil
	})

	return entries, err
}
