package httputil

import "strings"

// NormalizeHTTPBaseURL ensures the URL has a scheme so net/http can parse it.
// Accepts "host:port" or "host" and prefixes "http://".
func NormalizeHTTPBaseURL(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return ""
	}
	s = strings.TrimRight(s, "/")
	if strings.Contains(s, "://") {
		return s
	}
	return "http://" + s
}
