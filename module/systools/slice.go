package systools

import "sort"

func SliceContains(slice []string, search string) bool {
	for _, str := range slice {
		if str == search {
			return true
		}
	}
	return false
}

func MapRangeSortedFn(m map[string]any, fn func(key string, value any)) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fn(k, m[k])
	}
}
