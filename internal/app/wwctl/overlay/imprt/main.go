package imprt

import (
	"archive/tar"
	"compress/gzip"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/overlaydiff"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	if ArchiveImport {
		return importArchive(args[0], args[1])
	}

	var dest string

	source := args[1]

	if len(args) == 3 {
		dest = args[2]
	} else {
		dest = source
	}
	overlay_, err := overlay.Get(args[0])
	if err != nil {
		return err
	}
	if !overlay_.IsSiteOverlay() {
		overlay_, err = overlay_.CloneToSite()
		if err != nil {
			return err
		}
	}

	if util.IsDir(overlay_.File(dest)) {
		dest = path.Join(dest, path.Base(source))
	}

	if !OverwriteFile && util.IsFile(overlay_.File(dest)) {
		return fmt.Errorf("a file with that name already exists in the overlay")
	}

	if CreateDirs {
		parent := filepath.Dir(overlay_.File(dest))
		if _, err = os.Stat(parent); os.IsNotExist(err) {
			wwlog.Debug("Create dir: %s", parent)
			srcInfo, err := os.Stat(source)
			if err != nil {
				return fmt.Errorf("could not retrieve the stat for file: %w", err)
			}
			mode := srcInfo.Mode()
			mode |= ((mode & 0444) >> 2) // add execute permission wherever srcInfo has read
			err = os.MkdirAll(parent, mode)
			if err != nil {
				return fmt.Errorf("could not create parent dir: %s: %w", parent, err)
			}
		}
	}

	err = util.CopyFile(source, overlay_.File(dest))
	if err != nil {
		return fmt.Errorf("could not copy file into overlay: %w", err)
	}

	return nil
}

func importArchive(overlayName string, archivePath string) error {
	if err := overlaydiff.ValidateOverlayName(overlayName); err != nil {
		return err
	}

	siteDir := config.Get().Paths.SiteOverlaydir()
	if err := os.MkdirAll(siteDir, 0o755); err != nil {
		return fmt.Errorf("could not create site overlay directory: %w", err)
	}

	targetOverlay := filepath.Join(siteDir, overlayName)
	if _, err := os.Lstat(targetOverlay); err == nil {
		if !OverwriteFile {
			return fmt.Errorf("overlay already exists: %s", overlayName)
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("could not inspect target overlay: %w", err)
	}

	stageRoot, err := os.MkdirTemp(siteDir, ".ww-overlay-import-")
	if err != nil {
		return fmt.Errorf("could not create archive import staging directory: %w", err)
	}
	cleanupStage := true
	defer func() {
		if cleanupStage {
			_ = os.RemoveAll(stageRoot)
		}
	}()

	if err := extractArtifactArchive(archivePath, stageRoot); err != nil {
		return err
	}
	if err := overlaydiff.ValidateArtifact(stageRoot); err != nil {
		return err
	}
	manifest, err := overlaydiff.LoadArtifactManifest(filepath.Join(stageRoot, overlaydiff.ArtifactManifestFileName))
	if err != nil {
		return err
	}
	if manifest.OverlayName != overlayName {
		return fmt.Errorf("artifact overlay name %q does not match requested overlay %q", manifest.OverlayName, overlayName)
	}

	if OverwriteFile {
		if err := os.RemoveAll(targetOverlay); err != nil {
			return fmt.Errorf("could not remove existing overlay: %w", err)
		}
	}
	if err := os.Rename(stageRoot, targetOverlay); err != nil {
		return fmt.Errorf("could not install imported overlay: %w", err)
	}
	cleanupStage = false

	return nil
}

func extractArtifactArchive(archivePath string, destRoot string) error {
	archiveFile, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("could not open artifact archive: %w", err)
	}
	defer archiveFile.Close()

	gzipReader, err := gzip.NewReader(archiveFile)
	if err != nil {
		return fmt.Errorf("artifact archive is not gzip-compressed tar: %w", err)
	}
	defer gzipReader.Close()

	tarReader := tar.NewReader(gzipReader)
	for {
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return fmt.Errorf("could not read artifact archive: %w", err)
		}
		if err := extractArtifactEntry(tarReader, header, destRoot); err != nil {
			return err
		}
	}

	return nil
}

func extractArtifactEntry(reader io.Reader, header *tar.Header, destRoot string) error {
	name, err := cleanArchiveEntryName(header.Name)
	if err != nil {
		return err
	}
	if name != overlaydiff.ArtifactManifestFileName && name != "rootfs" && !strings.HasPrefix(name, "rootfs"+string(filepath.Separator)) {
		return fmt.Errorf("artifact archive entry is outside allowed layout: %s", header.Name)
	}

	target := filepath.Join(destRoot, name)
	if err := ensureImportPath(destRoot, target); err != nil {
		return err
	}

	switch header.Typeflag {
	case tar.TypeDir:
		if err := ensureImportDir(destRoot, target, os.FileMode(header.Mode)); err != nil {
			return err
		}
	case tar.TypeReg, tar.TypeRegA:
		if err := ensureImportDir(destRoot, filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := writeArchiveFile(reader, target, os.FileMode(header.Mode)); err != nil {
			return err
		}
	case tar.TypeSymlink:
		if err := validateArchiveSymlink(name, header.Linkname); err != nil {
			return err
		}
		if err := ensureImportDir(destRoot, filepath.Dir(target), 0o755); err != nil {
			return err
		}
		if err := os.Symlink(header.Linkname, target); err != nil {
			return fmt.Errorf("could not create artifact symlink %s: %w", name, err)
		}
	case tar.TypeLink:
		return fmt.Errorf("artifact archive hardlinks are not supported: %s", header.Name)
	default:
		return fmt.Errorf("unsupported artifact archive entry type for %s", header.Name)
	}

	return nil
}

func cleanArchiveEntryName(name string) (string, error) {
	if strings.TrimSpace(name) == "" {
		return "", fmt.Errorf("artifact archive contains an empty path")
	}
	if filepath.IsAbs(name) || strings.HasPrefix(name, "/") {
		return "", fmt.Errorf("artifact archive contains absolute path: %s", name)
	}
	cleaned := filepath.Clean(filepath.FromSlash(name))
	if cleaned == "." || cleaned == ".." || strings.HasPrefix(cleaned, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("artifact archive path escapes root: %s", name)
	}
	return cleaned, nil
}

func ensureImportPath(root string, target string) error {
	rel, err := filepath.Rel(root, target)
	if err != nil {
		return fmt.Errorf("could not resolve artifact path: %w", err)
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("artifact archive path escapes target root: %s", target)
	}
	return nil
}

func ensureImportDir(root string, dir string, mode os.FileMode) error {
	if err := ensureImportPath(root, dir); err != nil {
		return err
	}
	if mode == 0 {
		mode = 0o755
	}

	rel, err := filepath.Rel(root, dir)
	if err != nil {
		return fmt.Errorf("could not resolve artifact directory: %w", err)
	}
	current := root
	if rel == "." {
		return nil
	}
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		if part == "" || part == "." {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err == nil {
			if info.Mode()&os.ModeSymlink != 0 {
				return fmt.Errorf("refusing to extract through symlinked directory: %s", current)
			}
			if !info.IsDir() {
				return fmt.Errorf("artifact extraction path is not a directory: %s", current)
			}
			continue
		}
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("could not inspect artifact extraction directory %s: %w", current, err)
		}
		if err := os.Mkdir(current, mode); err != nil {
			return fmt.Errorf("could not create artifact extraction directory %s: %w", current, err)
		}
	}
	return nil
}

func writeArchiveFile(reader io.Reader, target string, mode os.FileMode) error {
	output, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return fmt.Errorf("could not create artifact file %s: %w", target, err)
	}
	if _, err := io.Copy(output, reader); err != nil {
		_ = output.Close()
		return fmt.Errorf("could not extract artifact file %s: %w", target, err)
	}
	if err := output.Close(); err != nil {
		return fmt.Errorf("could not close artifact file %s: %w", target, err)
	}
	if err := os.Chmod(target, mode); err != nil {
		return fmt.Errorf("could not set artifact file mode %s: %w", target, err)
	}
	return nil
}

func validateArchiveSymlink(entryName string, linkName string) error {
	if strings.TrimSpace(linkName) == "" {
		return fmt.Errorf("artifact symlink has empty target: %s", entryName)
	}
	if filepath.IsAbs(linkName) || strings.HasPrefix(linkName, "/") {
		return fmt.Errorf("artifact symlink target must be relative: %s -> %s", entryName, linkName)
	}
	resolved := filepath.Clean(filepath.Join(filepath.Dir(entryName), filepath.FromSlash(linkName)))
	if resolved == ".." || strings.HasPrefix(resolved, ".."+string(filepath.Separator)) {
		return fmt.Errorf("artifact symlink target escapes archive root: %s -> %s", entryName, linkName)
	}
	if resolved != "rootfs" && !strings.HasPrefix(resolved, "rootfs"+string(filepath.Separator)) {
		return fmt.Errorf("artifact symlink target escapes rootfs: %s -> %s", entryName, linkName)
	}
	return nil
}
