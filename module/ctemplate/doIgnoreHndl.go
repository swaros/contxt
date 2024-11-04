package ctemplate

import "strings"

type IgnorreHndl struct {
	origin    string
	masked    string
	ignoreSet IgnoreSetMap
}

func NewIgnorreHndl(origin string) *IgnorreHndl {
	return &IgnorreHndl{
		origin:    origin,
		ignoreSet: NewIgnoreSetMap(),
	}
}

func (i *IgnorreHndl) AddIgnores(stringToIgnore ...string) {
	for _, toIgnore := range stringToIgnore {
		i.ignoreSet.Add(toIgnore)
	}
}

func (i *IgnorreHndl) CreateMaskedString() string {
	i.masked = i.origin
	for _, ign := range i.ignoreSet {
		key := ign.key
		searchText := ign.origin
		i.masked = strings.ReplaceAll(i.masked, searchText, key)
	}
	return i.masked
}

func (i *IgnorreHndl) RestoreOriginalString(useThisString string) string {
	for _, ign := range i.ignoreSet {
		key := ign.key
		searchText := ign.origin
		useThisString = strings.ReplaceAll(useThisString, key, searchText)
	}
	return useThisString
}
