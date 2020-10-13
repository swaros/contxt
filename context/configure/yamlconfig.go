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

// RunConfig defines the structure of the local stored execution files
type RunConfig struct {
	Config struct {
		Sequencially bool `yaml:"sequencially"`
	} `yaml:"config"`
	Task []struct {
		ID          string `yaml:"id"`
		Stopreasons struct {
			Onerror        bool     `yaml:"onerror"`
			OnoutcountLess int      `yaml:"onoutcountLess"`
			OnoutcountMore int      `yaml:"onoutcountMore"`
			OnoutContains  []string `yaml:"onoutContains"`
		} `yaml:"stopreasons"`
		Options struct {
			Format     string   `yaml:"format"`
			Displaycmd bool     `yaml:"displaycmd"`
			Hideout    bool     `yaml:"hideout"`
			Maincmd    string   `yaml:"maincmd"`
			Mainparams []string `yaml:"mainparams"`
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
