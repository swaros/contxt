package process

// basic struct to hold the data of a process
type ProcData struct {
	Pid         int    // process id
	Cmd         string // command line
	ThreadCount int    // number of threads
	Threads     []int  // list of threads pids
}
