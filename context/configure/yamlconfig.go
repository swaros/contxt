package configure

// StopReasons mapping for RunConfig.Task[].Stopreasons
type StopReasons struct {
	Onerror        bool
	OnoutcountLess int
	OnoutcountMore int
	OnoutContains  []string
}

// Action defines a action that can be executed
type Action struct {
	Target  string   `yaml:"target"`
	Stopall bool     `yaml:"stopall"`
	Script  []string `yaml:"script"`
}

// after autogenerate todos:
// Variables are map[string]string and contains settings for Placeholders. add yaml:"variables,omitempty

// RunConfig defines the structure of the local stored execution files
type RunConfig struct {
	Config struct {
		Sequencially bool              `yaml:"sequencially"`
		Coloroff     bool              `yaml:"coloroff"`
		Variables    map[string]string `yaml:"variables,omitempty"`
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
			Format      string   `yaml:"format"`
			Stickcursor bool     `yaml:"stickcursor"`
			Colorcode   string   `yaml:"colorcode"`
			Bgcolorcode string   `yaml:"bgcolorcode"`
			Panelsize   int      `yaml:"panelsize"`
			Displaycmd  bool     `yaml:"displaycmd"`
			Hideout     bool     `yaml:"hideout"`
			Maincmd     string   `yaml:"maincmd"`
			Mainparams  []string `yaml:"mainparams"`
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
