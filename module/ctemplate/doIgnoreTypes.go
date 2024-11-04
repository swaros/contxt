package ctemplate

import (
	"github.com/google/uuid"
)

type IgnoreSet struct {
	origin string
	key    string
}

func NewIgnoreSet(origin string) (*IgnoreSet, string) {
	uniqueKey := uuid.New().String()
	return &IgnoreSet{
		origin: origin,
		key:    uniqueKey,
	}, uniqueKey

}

type IgnoreSetMap map[string]*IgnoreSet

func NewIgnoreSetMap() IgnoreSetMap {
	return make(IgnoreSetMap)
}

func (i IgnoreSetMap) Add(origin string) {
	set, key := NewIgnoreSet(origin)
	i[key] = set
}
