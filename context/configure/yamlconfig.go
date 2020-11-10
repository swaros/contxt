package configure

// StopReasons mapping for RunConfig.Task[].Stopreasons
type StopReasons struct {
	Onerror        bool
	OnoutcountLess int
	OnoutcountMore int
	OnoutContains  []string
}

// IncludePaths are files the defines how variables should be parsed.
// they indludes folders they have to be parsed first so they contents
// can be sued to proceeed with test/template.
// otherwise the yaml file is not readable
type IncludePaths struct {
	Include struct {
		Basedir bool     `yaml:"basedir"`
		Folders []string `yaml:"folders"`
	} `yaml:"include"`
}

// ++++++++++++++++++++ IMPORTANT !!!! ++++++++++++++++++++++++++++++
// after autogenerate todos:
// Variables are map[string]string and contains settings for Placeholders. add yaml:"variables,omitempty
//
// change all variables anywhere
//
// Variables    map[string]string `yaml:"variables,omitempty"`
//
// same for environment in Require
//
// Environment map[string]string `yaml:"environment,omitempty"`
//

// RunConfig defines the structure of the local stored execution files
type RunConfig struct {
	Config Config `yaml:"config"`
	Task   []Task `yaml:"task"`
}

// Autorun defines the targets they have to be executed
// if a special event is triggered
type Autorun struct {
	// this target will be executed if to this workspace was changed
	Onenter string `yaml:"onenter"`

	// this target will be executed if we changing to another workspace
	// so we leaving the current workspace
	// can be used for cleanup ...as a example
	Onleave string `yaml:"onleave"`
}

// Config is the main Configuration part of the Template.
type Config struct {
	Sequencially bool              `yaml:"sequencially"`
	Coloroff     bool              `yaml:"coloroff"`
	Loglevel     string            `yaml:"loglevel"`
	Variables    map[string]string `yaml:"variables,omitempty"`
	Autorun      Autorun           `yaml:"autorun"`
	Imports      []string          `yaml:"imports"`
}

// Require defines what is required to execute the task
type Require struct {
	Exists      []string          `yaml:"exists"`
	NotExists   []string          `yaml:"notExists"`
	Environment map[string]string `yaml:"environment,omitempty"`
	Variables   map[string]string `yaml:"variables,omitempty"`
}

// Stopreasons defines reasons to stop execution of the script
// all of them depends currently on parsing the output
// or just if a error happens by trying to execute a script-line
type Stopreasons struct {
	Onerror        bool     `yaml:"onerror"`
	OnoutcountLess int      `yaml:"onoutcountLess"`
	OnoutcountMore int      `yaml:"onoutcountMore"`
	OnoutContains  []string `yaml:"onoutContains"`
}

// Options are the per-task options
type Options struct {
	IgnoreCmdError bool     `yaml:"ignoreCmdError"`
	Format         string   `yaml:"format"`
	Stickcursor    bool     `yaml:"stickcursor"`
	Colorcode      string   `yaml:"colorcode"`
	Bgcolorcode    string   `yaml:"bgcolorcode"`
	Panelsize      int      `yaml:"panelsize"`
	Displaycmd     bool     `yaml:"displaycmd"`
	Hideout        bool     `yaml:"hideout"`
	Maincmd        string   `yaml:"maincmd"`
	Mainparams     []string `yaml:"mainparams"`
}

// Trigger are part of listener. The defines
// some events they are triggered by executing scripts
// most of them watching the output
type Trigger struct {
	Onerror        bool     `yaml:"onerror"`
	OnoutcountLess int      `yaml:"onoutcountLess"`
	OnoutcountMore int      `yaml:"onoutcountMore"`
	OnoutContains  []string `yaml:"onoutContains"`
}

// Action defines what should happens Next.
type Action struct {
	Target  string   `yaml:"target"`
	Stopall bool     `yaml:"stopall"`
	Script  []string `yaml:"script"`
}

// Listener are used for watching events
// and triggers an action if a event happens
type Listener struct {
	Trigger Trigger `yaml:"trigger"`
	Action  Action  `yaml:"action"`
}

// Task is the main Script
type Task struct {
	ID          string            `yaml:"id"`
	Variables   map[string]string `yaml:"variables,omitempty"`
	Requires    Require           `yaml:"require"`
	Stopreasons Stopreasons       `yaml:"stopreasons"`
	Options     Options           `yaml:"options"`
	Script      []string          `yaml:"script"`
	Listener    []Listener        `yaml:"listener"`
}

/*
// RunConfig defines the structure of the local stored execution files
type RunConfigOld struct {
	Config struct {
		Sequencially bool              `yaml:"sequencially"`
		Coloroff     bool              `yaml:"coloroff"`
		LogLevel     string            `yaml:"loglevel"`
		Variables    map[string]string `yaml:"variables,omitempty"`
		Imports      []string          `yAML:"imports"`
	} `yaml:"config"`
	Task []struct {
		ID          string            `yaml:"id"`
		Variables   map[string]string `yaml:"variables,omitempty"`
		Stopreasons struct {
			Onerror        bool     `yaml:"onerror"`
			OnoutcountLess int      `yaml:"onoutcountLess"`
			OnoutcountMore int      `yaml:"onoutcountMore"`
			OnoutContains  []string `yaml:"onoutContains"`
		} `yaml:"stopreasons"`
		Options struct {
			Format         string   `yaml:"format"`
			Stickcursor    bool     `yaml:"stickcursor"`
			IgnoreCmdError bool     `yaml:"ignoreCmdError"`
			Colorcode      string   `yaml:"colorcode"`
			Bgcolorcode    string   `yaml:"bgcolorcode"`
			Panelsize      int      `yaml:"panelsize"`
			Displaycmd     bool     `yaml:"displaycmd"`
			Hideout        bool     `yaml:"hideout"`
			Maincmd        string   `yaml:"maincmd"`
			Mainparams     []string `yaml:"mainparams"`
		} `yaml:"options"`
		Script   []string `yaml:"script"`
		Listener []struct {
			Trigger struct {
				Onerror        bool     `yaml:"onerror"`
				OnoutcountLess int      `yaml:"onoutcountLess"`
				OnoutcountMore int      `yaml:"onoutcountMore"`
				OnoutContains  []string `yaml:"onoutContains"`
			} `yaml:"trigger"`
			Action struct {
				Target  string   `yaml:"target"`
				Stopall bool     `yaml:"stopall"`
				Script  []string `yaml:"script"`
			} `yaml:"action"`
		} `yaml:"listener"`
	} `yaml:"task"`
}
*/
