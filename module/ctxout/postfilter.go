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

 package ctxout

import (
	"fmt"

	"github.com/swaros/contxt/module/systools"
)

var (
	postFilters map[string]PostFilter = make(map[string]PostFilter) // all post filters that are registered
	filterOrder []string              = []string{}                  // the order of the filters
)

// AddPostFilter adds a post filter to the list of post filters
// these filters are called after the markup filter
// they works only on strings
func AddPostFilter(filter PostFilter) {
	if filter == nil {
		return
	}
	initCtxOut()
	nameOfFilter := fmt.Sprintf("%T", filter)
	postFilters[nameOfFilter] = filter
	// if we have the filter already in the list we do not add it again
	// in the order list
	if !systools.SliceContains(filterOrder, nameOfFilter) {
		filterOrder = append(filterOrder, nameOfFilter)
	}
	// but we update the filter
	filter.Update(termInfo)
}

// GetPostFilters returns the list of post filters
func GetPostFilters() []PostFilter {
	var filters []PostFilter
	for _, nameOfFilter := range filterOrder {
		if postFilters[nameOfFilter] == nil {
			continue
		}
		filters = append(filters, postFilters[nameOfFilter])
	}
	return filters
}

func GetPostFilter(nameOfFilter string) PostFilter {
	return postFilters[nameOfFilter]
}

func GetPostFilterbyRef(filter PostFilter) PostFilter {
	return GetPostFilter(fmt.Sprintf("%T", filter))
}

// ClearPostFilters clears the list of post filters
func ClearPostFilters() {
	postFilters = make(map[string]PostFilter)
	filterOrder = []string{}
}

func UpdateFilterByName(nameOfFilter string, info PostFilterInfo) error {
	if postFilters[nameOfFilter] == nil {
		return fmt.Errorf("filter [%s] not found", nameOfFilter)
	}
	postFilters[nameOfFilter].Update(info)
	return nil
}

func UpdateFilterByRef(filter PostFilter, info PostFilterInfo) error {
	return UpdateFilterByName(fmt.Sprintf("%T", filter), info)
}

// Updates all registered post filters with the new terminal information
func ForceFilterUpdate(info PostFilterInfo) {
	if info.Id == "" {
		info.Id = FilterId()
	}

	for _, filter := range GetPostFilters() {
		filter.Update(info)
	}

}
