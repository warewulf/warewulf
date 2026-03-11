package wwurl

import (
	"net/url"
	"regexp"
)

// sensitiveParams are query parameter keys that should be redacted in logs.
var sensitiveParams = []string{"assetkey"}

// embeddedURLPattern matches an http(s) URL within a larger string, stopping at
// whitespace or a double quote so it works whether the URL is quoted or not.
var embeddedURLPattern = regexp.MustCompile(`https?://[^\s"]+`)

// SanitizeURL returns a copy of the URL string with sensitive query parameters
// replaced by "REDACTED". Useful for logging URLs without leaking secrets.
func SanitizeURL(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	redacted := false
	for _, key := range sensitiveParams {
		if q.Has(key) {
			q.Set(key, "REDACTED")
			redacted = true
		}
	}
	if redacted {
		u.RawQuery = q.Encode()
	}
	return u.String()
}

// SanitizeError returns the error message with sensitive URL query parameters
// redacted. Go's net/http embeds the full request URL (including query params)
// in transport errors, so this prevents secrets from appearing in log output.
func SanitizeError(err error) string {
	return embeddedURLPattern.ReplaceAllStringFunc(err.Error(), SanitizeURL)
}
