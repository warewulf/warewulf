package overlaydiff

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// EntryType describes the kind of filesystem entry.
type EntryType string

const (
	// EntryFile represents a regular file.
	EntryFile EntryType = "file"
	// EntryDir represents a directory.
	EntryDir EntryType = "dir"
	// EntrySymlink represents a symbolic link.
	EntrySymlink EntryType = "symlink"
)

// ChangeType enumerates the kinds of changes detected between two trees.
type ChangeType string

const (
	// ChangeAdded indicates the entry exists in source but not in baseline.
	ChangeAdded ChangeType = "added"
	// ChangeRemoved indicates the entry existed in baseline but not in source.
	ChangeRemoved ChangeType = "removed"
	// ChangeModified indicates content (or symlink target) differs.
	ChangeModified ChangeType = "modified"
	// ChangeModeChanged indicates only the permission bits/mode changed.
	ChangeModeChanged ChangeType = "mode-changed"
	// ChangeTypeChanged indicates the entry type changed (e.g., file -> dir).
	ChangeTypeChanged ChangeType = "type-changed"
)

// Entry describes a single filesystem object found when scanning a tree.
// Fields such as Size or Hash are populated for regular files; LinkTarget
// is populated for symlinks.
type Entry struct {
	Path       string    `json:"path"`
	Type       EntryType `json:"type"`
	Mode       uint32    `json:"mode"`
	Size       int64     `json:"size,omitempty"`
	Hash       string    `json:"hash,omitempty"`
	LinkTarget string    `json:"link_target,omitempty"`
}

// Change represents a detected difference for a single path between the
// source and baseline trees. `Source` and `Baseline` hold the corresponding
// entries when available.
type Change struct {
	Path     string     `json:"path"`
	Change   ChangeType `json:"change"`
	Type     EntryType  `json:"type"`
	Mode     uint32     `json:"mode"`
	Size     int64      `json:"size,omitempty"`
	Source   *Entry     `json:"source,omitempty"`
	Baseline *Entry     `json:"baseline,omitempty"`
}

// permissionMask filters file mode bits to the permission-related flags.
const permissionMask = os.ModePerm | os.ModeSetuid | os.ModeSetgid | os.ModeSticky

// Diff scans both roots and returns the changes required to make baseline match source.
func Diff(sourceRoot, baselineRoot string) ([]Change, error) {
	source, err := ScanTree(sourceRoot)
	if err != nil {
		return nil, err
	}

	baseline, err := ScanTree(baselineRoot)
	if err != nil {
		return nil, err
	}

	return Compare(source, baseline), nil
}

// ScanTree walks the filesystem rooted at root and returns a map of relative
// paths to Entry describing each object.
func ScanTree(root string) (map[string]Entry, error) {
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve root path %s: %w", root, err)
	}

	rootInfo, err := os.Stat(rootAbs)
	if err != nil {
		return nil, fmt.Errorf("failed to stat root path %s: %w", rootAbs, err)
	}
	if !rootInfo.IsDir() {
		return nil, fmt.Errorf("root path is not a directory: %s", rootAbs)
	}

	entries := make(map[string]Entry)
	err = filepath.WalkDir(rootAbs, func(current string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		if current == rootAbs {
			return nil
		}

		relPath, err := filepath.Rel(rootAbs, current)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		relPath = normalizeRelPath(relPath)

		info, err := os.Lstat(current)
		if err != nil {
			return fmt.Errorf("failed to lstat %s: %w", current, err)
		}

		entry := Entry{
			Path: relPath,
			Mode: uint32(info.Mode() & permissionMask),
		}

		switch {
		case info.Mode()&os.ModeSymlink != 0:
			entry.Type = EntrySymlink
			linkTarget, err := os.Readlink(current)
			if err != nil {
				return fmt.Errorf("failed to read symlink target for %s: %w", current, err)
			}
			entry.LinkTarget = linkTarget
		case d.IsDir():
			entry.Type = EntryDir
		default:
			entry.Type = EntryFile
			entry.Size = info.Size()
			hash, err := hashFile(current)
			if err != nil {
				return fmt.Errorf("failed to hash file %s: %w", current, err)
			}
			entry.Hash = hash
		}

		entries[relPath] = entry
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed scanning tree %s: %w", rootAbs, err)
	}

	return entries, nil
}

// Compare computes a sorted list of Change describing differences between two maps.
func Compare(source map[string]Entry, baseline map[string]Entry) []Change {
	keySet := make(map[string]struct{}, len(source)+len(baseline))
	for key := range source {
		keySet[key] = struct{}{}
	}
	for key := range baseline {
		keySet[key] = struct{}{}
	}

	keys := make([]string, 0, len(keySet))
	for key := range keySet {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	changes := make([]Change, 0)
	for _, key := range keys {
		src, srcExists := source[key]
		base, baseExists := baseline[key]

		switch {
		case srcExists && !baseExists:
			changes = append(changes, newChange(key, ChangeAdded, &src, nil))
		case !srcExists && baseExists:
			changes = append(changes, newChange(key, ChangeRemoved, nil, &base))
		case srcExists && baseExists:
			if src.Type != base.Type {
				changes = append(changes, newChange(key, ChangeTypeChanged, &src, &base))
				continue
			}

			if entryModified(src, base) {
				changes = append(changes, newChange(key, ChangeModified, &src, &base))
				continue
			}

			if src.Mode != base.Mode {
				changes = append(changes, newChange(key, ChangeModeChanged, &src, &base))
			}
		}
	}

	return changes
}

// entryModified reports whether source and baseline differ in content
// (file hash for files or link target for symlinks).
func entryModified(source Entry, baseline Entry) bool {
	switch source.Type {
	case EntryFile:
		return source.Hash != baseline.Hash
	case EntrySymlink:
		return source.LinkTarget != baseline.LinkTarget
	default:
		return false
	}
}

// newChange constructs a Change and populates metadata from source or baseline.
func newChange(path string, changeType ChangeType, source *Entry, baseline *Entry) Change {
	result := Change{
		Path:     path,
		Change:   changeType,
		Source:   source,
		Baseline: baseline,
	}

	if source != nil {
		result.Type = source.Type
		result.Mode = source.Mode
		result.Size = source.Size
	} else if baseline != nil {
		result.Type = baseline.Type
		result.Mode = baseline.Mode
		result.Size = baseline.Size
	}

	return result
}

// normalizeRelPath converts separators to slashes and ensures a leading '/'.
func normalizeRelPath(path string) string {
	path = filepath.ToSlash(path)
	if strings.HasPrefix(path, "/") {
		return path
	}
	return "/" + path
}

// hashFile returns the SHA-256 hex digest of the file at path.
func hashFile(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
