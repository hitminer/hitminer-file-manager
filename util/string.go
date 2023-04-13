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

func CommonSuffix(a, b string) string {
	p := Min(len(a), len(b))
	offsetA, offsetB := len(a)-p, len(b)-p
	p--
	for p >= 0 {
		if a[p+offsetA] != b[p+offsetB] {
			return a[p+offsetA+1:]
		}
		p--
	}
	return a[p+offsetA+1:]
}
