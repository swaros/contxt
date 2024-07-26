// MIT License
//
// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the Software), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED AS IS, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// AINC-NOTE-0815

package systools

import (
	"sort"
	"strings"
)

func SliceContains(slice []string, search string) bool {
	for _, str := range slice {
		if str == search {
			return true
		}
	}
	return false
}

// Alias for SliceContains
func StringInSlice(search string, slice []string) bool {
	return SliceContains(slice, search)
}

// like SliceContains but with strings.Contains so we can search for substrings
func SliceContainsSub(slice []string, search string) bool {
	for _, str := range slice {
		if strings.Contains(str, search) {
			return true
		}
	}
	return false
}

// an callback handler to sort the entries in a map by key first
// and then call the callback function with the key and value
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

// RemoveFromSliceOnce removes the first occurence of search in slice
func RemoveFromSliceOnce(slice []string, search string) []string {
	for i, str := range slice {
		if str == search {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}

func StrStr2StrAny(m map[string]string) map[string]any {
	out := make(map[string]any)
	for k, v := range m {
		out[k] = v
	}
	return out
}

func SortByKeyString(m map[string]any, rowExec func(k string, v any)) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		rowExec(k, m[k])
	}
}
