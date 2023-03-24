package util

import "strings"

func TrimEtag(etag string) string {
	etag = strings.TrimPrefix(etag, "\"")
	return strings.TrimSuffix(etag, "\"")
}

func HTTPEtag(etag string) string {
	if !strings.HasPrefix(etag, "\"") {
		etag = "\"" + etag
	}
	if !strings.HasSuffix(etag, "\"") {
		etag = etag + "\""
	}
	return etag
}
