package util

import (
	"regexp"
	"strings"

	"github.com/hashicorp/go-version"
)

var (
	versionPattern *regexp.Regexp
)

func init() {
	versionPattern = regexp.MustCompile(`\d+\.\d+\.\d+(-[\d\.]+|)`)
}

func ParseVersion(versionString string) *version.Version {
	matches := versionPattern.FindAllString(versionString, -1)
	for i := len(matches) - 1; i >= 0; i-- {
		if version_, err := version.NewVersion(strings.TrimSuffix(matches[i], ".")); err == nil {
			return version_
		}
	}
	return nil
}
