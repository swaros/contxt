package tasks

type PlaceHolder interface {
	SetPH(key, value string)
	AppendToPH(key, value string) bool
	SetIfNotExists(key, value string)
	GetPHExists(key string) (string, bool)
	GetPH(key string) string
	GetPlaceHoldersFnc(inspectFunc func(phKey string, phValue string))
	HandlePlaceHolder(line string) string
	HandlePlaceHolderWithScope(line string, scopeVars map[string]string) string
	ClearAll()
}
