package tasks

type TaskOutCtrl struct {
	IgnoreCase bool
}

type TaskOutLabel struct {
	Message interface{}
	FColor  string
}

type TaskTargetOut struct {
	ForeCol     string
	BackCol     string
	SplitLabel  string
	Target      string
	Alternative string
	PanelSize   int
}
