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
	for _, match := range matches {
		if version_, err := version.NewVersion(strings.TrimSuffix(match, ".")); err == nil {
			return version_
		}
	}
	return nil
}
