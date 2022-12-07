package configure

// Configuration includes all paths for the current workspace
// this is the main Configuration that is needed to eep track which path
// is currently used.
// this is the used format til version 0.4.0
type Configuration struct {
	CurrentSet string
	Paths      []string
	LastIndex  int
	LastPath   string // since 0.5.0
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
