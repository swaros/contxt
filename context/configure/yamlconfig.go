package configure

// StopReasons mapping for RunConfig.Task[].Stopreasons
type StopReasons struct {
	Onerror        bool
	OnoutcountLess int
	OnoutcountMore int
	OnoutContains  []string
	Onevents       []string
}

// RunConfig defines the structure of the local stored execution files
type RunConfig struct {
	Task []struct {
		ID      string `yaml:"id"`
		Trigger struct {
			Events []string `yaml:"events"`
		} `yaml:"trigger"`
		Stopreasons struct {
			Onerror        bool     `yaml:"onerror"`
			OnoutcountLess int      `yaml:"onoutcountLess"`
			OnoutcountMore int      `yaml:"onoutcountMore"`
			OnoutContains  []string `yaml:"onoutContains"`
			Onevents       []string `yaml:"onevents"`
		} `yaml:"stopreasons"`
		Options struct {
			Format     string   `yaml:"format"`
			Displaycmd bool     `yaml:"displaycmd"`
			Hideout    bool     `yaml:"hideout"`
			Maincmd    string   `yaml:"maincmd"`
			Mainparams []string `yaml:"mainparams"`
		} `yaml:"options"`
		Script []string `yaml:"script"`
		Watch  []struct {
			Output struct {
				Contains []string `yaml:"contains"`
				Exitcode struct {
					Greater int `yaml:"greater"`
					Equals  int `yaml:"equals"`
					Lower   int `yaml:"lower"`
				} `yaml:"exitcode"`
				Then struct {
					PushEvents []string `yaml:"pushEvents"`
					Stop       bool     `yaml:"stop"`
				} `yaml:"then"`
			} `yaml:"output"`
		} `yaml:"watch"`
	} `yaml:"task"`
}
