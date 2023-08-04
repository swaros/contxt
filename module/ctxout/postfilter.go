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
