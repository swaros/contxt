package yacl

func MapEntrieSetOrCreate(mapHndl map[string]any, key string, value any, initMap func(map[string]any) map[string]any) map[string]any {

	if mapHndl == nil {
		mapHndl = initMap(mapHndl)
	}
	// rechcheck if the callbacks olves the nil issue
	if mapHndl == nil {
		panic("map is stil nil. you have to use make to create the map in then initMap callback, and return them there")
	}
	mapHndl[key] = value
	return mapHndl
}
