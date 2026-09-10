package util

import (
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
)

var (
	versionPattern *regexp.Regexp
	imageSuffixPattern *regexp.Regexp
)

func init() {
	versionPattern = regexp.MustCompile(`\d+\.\d+\.\d+(-[\w+\.-]+)?`)
	imageSuffixPattern = regexp.MustCompile(`\.gz|\.bz2|\.xz|\.lz4|\.zst|\.lzma|\.lzo`)
}

func ParseVersion(versionString string) *version.Version {
	matches := versionPattern.FindAllString(versionString, -1)
	for _, match := range matches {
		match = imageSuffixPattern.ReplaceAllString(match, "")
		if version_, err := version.NewVersion(strings.ReplaceAll(strings.TrimSuffix(match, "."), "_", "-")); err == nil {
			return version_
		}
	}
	return nil
}
