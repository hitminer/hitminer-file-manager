package util

import "strings"

func SplitFullPath(fp string) []string {
	strs := strings.Split(fp, "/")
	ret := make([]string, 0)
	ret = append(ret, strs[0])
	for _, u := range strs[1:] {
		if u == "" {
			continue
		}
		ret = append(ret, ret[len(ret)-1]+"/"+u)
	}
	return ret
}
