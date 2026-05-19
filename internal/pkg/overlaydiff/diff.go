package overlaydiff

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
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
	Path          string    `json:"path"`
	Type          EntryType `json:"type"`
	Mode          uint32    `json:"mode"`
	Size          int64     `json:"size,omitempty"`
	MTimeUnixNano int64     `json:"mtime_unix_nano,omitempty"`
	Inode         uint64    `json:"inode,omitempty"`
	Device        uint64    `json:"device,omitempty"`
	Hash          string    `json:"hash,omitempty"`
	LinkTarget    string    `json:"link_target,omitempty"`
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

// ScanOptions controls scan-time behavior.
type ScanOptions struct {
	Excludes        []string
	IncludeRoots    []string
	ScanWorkers     int
	HashWorkers     int
	BaselineEntries map[string]Entry
}

type hashJob struct {
	absPath string
	relPath string
}

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
	return ScanTreeWithOptions(root, ScanOptions{})
}

// ScanTreeWithOptions walks the filesystem rooted at root and returns a map of
// relative paths to Entry describing each object. Excludes are interpreted as
// normalized path prefixes (e.g. /var/log).
func ScanTreeWithOptions(root string, options ScanOptions) (map[string]Entry, error) {
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

	excludes := NormalizeExcludes(options.Excludes)
	includes := NormalizeExcludes(options.IncludeRoots)
	if len(includes) > 0 {
		return scanIncludedRoots(rootAbs, includes, excludes, options)
	}

	entries := make(map[string]Entry)
	hashJobs := make([]hashJob, 0)
	if err := scanTreeInto(rootAbs, rootAbs, includes, excludes, options, entries, &hashJobs); err != nil {
		return nil, fmt.Errorf("failed scanning tree %s: %w", rootAbs, err)
	}

	if err := populateFileHashes(entries, hashJobs, options.HashWorkers); err != nil {
		return nil, fmt.Errorf("failed hashing files in %s: %w", rootAbs, err)
	}

	return entries, nil
}

func scanIncludedRoots(rootAbs string, includes []string, excludes []string, options ScanOptions) (map[string]Entry, error) {
	includes = dedupeTopLevelIncludes(includes)
	workers := resolveScanWorkers(options.ScanWorkers)
	if workers > len(includes) {
		workers = len(includes)
	}
	if workers < 1 {
		workers = 1
	}

	entries := make(map[string]Entry)
	hashJobs := make([]hashJob, 0)
	var mu sync.Mutex

	includeCh := make(chan string)
	errCh := make(chan error, 1)
	var wg sync.WaitGroup

	workerFn := func() {
		defer wg.Done()
		for include := range includeCh {
			subRoot := filepath.Join(rootAbs, strings.TrimPrefix(include, "/"))
			info, err := os.Lstat(subRoot)
			if err != nil {
				if isSkippableScanError(err) {
					continue
				}
				select {
				case errCh <- fmt.Errorf("failed to inspect include root %s: %w", include, err):
				default:
				}
				return
			}
			if !info.IsDir() {
				continue
			}

			localEntries := make(map[string]Entry)
			localJobs := make([]hashJob, 0)
			if err := scanTreeInto(rootAbs, subRoot, nil, excludes, options, localEntries, &localJobs); err != nil {
				select {
				case errCh <- fmt.Errorf("failed scanning include root %s: %w", include, err):
				default:
				}
				return
			}

			mu.Lock()
			for path, entry := range localEntries {
				entries[path] = entry
			}
			hashJobs = append(hashJobs, localJobs...)
			mu.Unlock()
		}
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go workerFn()
	}

	for _, include := range includes {
		select {
		case err := <-errCh:
			close(includeCh)
			wg.Wait()
			return nil, err
		default:
			includeCh <- include
		}
	}
	close(includeCh)
	wg.Wait()

	select {
	case err := <-errCh:
		return nil, err
	default:
	}

	if err := populateFileHashes(entries, hashJobs, options.HashWorkers); err != nil {
		return nil, fmt.Errorf("failed hashing files in %s: %w", rootAbs, err)
	}

	return entries, nil
}

func scanTreeInto(rootAbs string, walkRoot string, includes []string, excludes []string, options ScanOptions, entries map[string]Entry, hashJobs *[]hashJob) error {
	err := filepath.WalkDir(walkRoot, func(current string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			if isSkippableScanError(walkErr) {
				if d != nil && d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			return walkErr
		}

		if current == walkRoot {
			return nil
		}

		relPath, err := filepath.Rel(rootAbs, current)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %w", err)
		}

		relPath = normalizeRelPath(relPath)

		if !shouldInclude(relPath, includes) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if shouldExclude(relPath, excludes) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		info, err := os.Lstat(current)
		if err != nil {
			if isSkippableScanError(err) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
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
				if isSkippableScanError(err) {
					return nil
				}
				return fmt.Errorf("failed to read symlink target for %s: %w", current, err)
			}
			entry.LinkTarget = linkTarget
		case d.IsDir():
			entry.Type = EntryDir
		default:
			entry.Type = EntryFile
			entry.Size = info.Size()
			entry.MTimeUnixNano = info.ModTime().UnixNano()
			if stat, ok := info.Sys().(*syscall.Stat_t); ok {
				entry.Inode = stat.Ino
				entry.Device = uint64(stat.Dev)
			}

			if baseline, ok := options.BaselineEntries[relPath]; ok && canReuseFileHash(entry, baseline) {
				entry.Hash = baseline.Hash
			} else {
				*hashJobs = append(*hashJobs, hashJob{absPath: current, relPath: relPath})
			}
		}

		entries[relPath] = entry
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func resolveScanWorkers(workers int) int {
	if workers > 0 {
		return workers
	}
	count := runtime.NumCPU() * 2
	if count > 12 {
		count = 12
	}
	if count < 1 {
		count = 1
	}
	return count
}

func dedupeTopLevelIncludes(includes []string) []string {
	if len(includes) <= 1 {
		return includes
	}
	ordered := make([]string, len(includes))
	copy(ordered, includes)
	sort.SliceStable(ordered, func(i, j int) bool {
		if len(ordered[i]) == len(ordered[j]) {
			return ordered[i] < ordered[j]
		}
		return len(ordered[i]) < len(ordered[j])
	})

	result := make([]string, 0, len(ordered))
	for _, include := range ordered {
		skip := false
		for _, existing := range result {
			if include == existing || strings.HasPrefix(include, existing+"/") {
				skip = true
				break
			}
		}
		if !skip {
			result = append(result, include)
		}
	}
	return result
}

func populateFileHashes(entries map[string]Entry, jobs []hashJob, workers int) error {
	if len(jobs) == 0 {
		return nil
	}
	if workers <= 0 {
		workers = runtime.NumCPU() * 3
		if workers > 32 {
			workers = 32
		}
		if workers < 1 {
			workers = 1
		}
	}

	jobCh := make(chan hashJob)
	errCh := make(chan error, 1)
	var wg sync.WaitGroup
	var mu sync.Mutex

	workerFn := func() {
		defer wg.Done()
		for job := range jobCh {
			hash, err := hashFile(job.absPath)
			if err != nil {
				if isSkippableScanError(err) {
					// File vanished or became inaccessible between walk and hash.
					mu.Lock()
					delete(entries, job.relPath)
					mu.Unlock()
					continue
				}
				select {
				case errCh <- fmt.Errorf("failed to hash file %s: %w", job.absPath, err):
				default:
				}
				return
			}

			mu.Lock()
			entry := entries[job.relPath]
			entry.Hash = hash
			entries[job.relPath] = entry
			mu.Unlock()
		}
	}

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go workerFn()
	}

	for _, job := range jobs {
		select {
		case err := <-errCh:
			close(jobCh)
			wg.Wait()
			return err
		default:
			jobCh <- job
		}
	}
	close(jobCh)
	wg.Wait()

	select {
	case err := <-errCh:
		return err
	default:
		return nil
	}
}

func isSkippableScanError(err error) bool {
	return errors.Is(err, fs.ErrPermission) ||
		errors.Is(err, os.ErrPermission) ||
		errors.Is(err, fs.ErrNotExist) ||
		errors.Is(err, os.ErrNotExist) ||
		errors.Is(err, syscall.ENXIO)
}

func canReuseFileHash(current Entry, baseline Entry) bool {
	if current.Type != EntryFile || baseline.Type != EntryFile {
		return false
	}
	if baseline.Hash == "" {
		return false
	}
	if current.Size != baseline.Size || current.Mode != baseline.Mode || current.MTimeUnixNano != baseline.MTimeUnixNano {
		return false
	}

	if current.Inode != 0 && current.Device != 0 && baseline.Inode != 0 && baseline.Device != 0 {
		if current.Inode != baseline.Inode || current.Device != baseline.Device {
			return false
		}
	}

	return true
}

func shouldExclude(path string, excludes []string) bool {
	if hasAutoExcludedSegment(path) {
		return true
	}

	for _, exclude := range excludes {
		if path == exclude || strings.HasPrefix(path, exclude+"/") {
			return true
		}
	}
	return false
}

func shouldInclude(path string, includes []string) bool {
	if len(includes) == 0 {
		// No include roots configured means "scan everything under root".
		return true
	}
	for _, include := range includes {
		if path == include || strings.HasPrefix(path, include+"/") || strings.HasPrefix(include, path+"/") {
			return true
		}
	}
	return false
}

func hasAutoExcludedSegment(path string) bool {
	for _, segment := range strings.Split(path, "/") {
		if segment == "cache" || segment == ".cache" {
			return true
		}
	}
	return false
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
