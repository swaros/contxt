package configure

// Configuration includes all paths for the current workspace
// this is the main Configuration that is needed to eep track which path
// is currently used.
// this is the used format til version 0.4.0
type Configuration struct {
	CurrentSet string
	Paths      []string
	LastIndex  int
	PathInfo   map[string]WorkspaceInfo
}

// new version of the configuration starts here
type ConfigMetaV2 struct {
	CurrentSet string                     `yaml:"currentSet"`
	Configs    map[string]ConfigurationV2 `yaml:"configs"`
}

type WorkspaceInfoV2 struct {
	Path    string `yaml:"path"`
	Project string `yaml:"project"`
	Role    string `yaml:"role"`
	Version string `yaml:"version"`
}

type ConfigurationV2 struct {
	Name         string                     `yaml:"name"`         // the name of the workspace
	CurrentIndex string                     `yaml:"currentIndex"` // what of the workspaces are the current used
	Paths        map[string]WorkspaceInfoV2 `yaml:"paths"`
}

// CommandLine defines a line of commands that can be executed
type CommandLine struct {
	Require            RequireCheck
	Command            string
	Params             string
	Comment            string
	StopOnError        bool
	StopOnOutCountLess int
	StopOnOutCountMore int
	StopOnOutContains  string
	TraceOutput        bool
}

// ExecuteDefinition Defines the structure of a .execute file that defines commands they have to executed
type ExecuteDefinition struct {
	TestScript  []CommandLine
	InitScript  []CommandLine
	CleanScript []CommandLine
	Script      []CommandLine
}

// RequireCheck defines some variables they have to be valid before script runs
type RequireCheck struct {
	FileExists    []string
	FileNotExists []string
}

type GitVersionInfo struct {
	HashUsed    string
	Reference   string
	Repositiory string
	Path        string
	Exists      bool
}
