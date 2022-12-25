package configure

// Creating the interface for handling contxt configuration
// til ow, this was done pure function based.
// that ends in loss off overview and a mix of responsibilities.
// first stepp is to build the whole configuration procedure
// in a clear context

// ContxtConfigV1 the needs for V1
type ContxtConfigV1 interface {
	InitConfig() error // InitConfig initilize the configuration files
	ClearPaths()       // ClearPaths removes all paths from current Workspace
	CheckSetup()
	ChangeWorkspace(workspace string, oldspace func(string) bool, newspace func(string)) error
	RemoveWorkspace(name string) error
	SaveDefaultConfiguration(workSpaceConfigUpdate bool) error
	SaveActualPathByIndex(useIndex int) error
	SaveActualPathByPath(pathToSave string) error
	PathWorker(callbackInDirextory func(int, string), callbackBackToOrigin func(origin string)) error
	PathWorkerNoCd(callback func(int, string)) error
	LoadExtConfiguration(path string) (Configuration, error)
	SaveConfiguration(config Configuration, path string) error
	GetConfigPath(fileName string) (string, error)
	AddPath(path string)
	RemovePath(path string) bool
	PathExists(pathSearch string) bool
	PathMeightPartOfWs(pathSearch string) bool
}
