package main

import (
	"fmt"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/swaros/contxt/module/process"
)

func main() {
	// we will start a process and keep it running
	// for this we use the runner and using pure go
	// we will use the process package to start the process
	// and the tasks package to watch the process
	// and kill it if needed

	if err := os.Chdir("runner"); err != nil {
		panic(err)
	}

	printMyProcInfo()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		startRunner(&wg)
	}()

	printMyProcInfo()

	wg.Add(1)
	go func() {
		startRunnerInBash(&wg)
	}()

	wg.Wait()
	printMyProcInfo()
}

func startRunnerInBash(wg *sync.WaitGroup) {
	defer wg.Done()
	if runtime.GOOS == "windows" {
		return
	}
	runnerProc := process.NewProcess("bash")
	runnerProc.SetOnOutput(func(msg string, err error) bool {
		fmt.Println("bash proc got output:", msg)
		if err != nil {
			fmt.Println("bash proc got error:", err)
		}
		return true
	})
	runnerProc.SetOnInit(func(proc *os.Process) {
		if proc == nil {
			panic("Process is nil")
		}
		fmt.Println("bash runner Process started with pid", proc.Pid)
	})
	runnerProc.SetKeepRunning(true)
	runnerProc.SetTimeout(30 * time.Second)
	runnerProc.Exec()
	if err := runnerProc.Command("go run runner.go"); err != nil {
		panic(
			fmt.Sprintf("failed to run command: %s", err),
		)

	}
	// now just wait for the user to press enter
	runnerProc.BlockWait(60 * time.Second)
	// stop the process
	fmt.Println("stopping bash runner")
	runnerProc.Stop()

}

func startRunner(wg *sync.WaitGroup) {
	defer wg.Done()

	runnerProc := process.NewProcess("go", "run", "runner.go")
	runnerProc.SetOnOutput(func(msg string, err error) bool {
		fmt.Println("\tgot output:", msg)

		procWatch, _ := runnerProc.GetProcessWatcher()
		if procWatch == nil {
			panic("ProcessWatcher is nil")
		}

		//fmt.Println("\trunner process pid:", procWatch.GetPid())
		//fmt.Println("\trunner process cmd:", procWatch.GetCmd())
		threads := procWatch.GetThreads()
		for _, thread := range threads {
			fmt.Println("\tsub Thread:", thread)
		}
		printMyProcInfo()
		if err != nil {
			panic(err)
		}
		return true
	})
	runnerProc.SetOnInit(func(proc *os.Process) {
		if proc == nil {
			panic("Process is nil")
		}
		fmt.Println("runner Process started with pid", proc.Pid)
	})
	runnerProc.SetKeepRunning(true)
	runnerProc.SetTimeout(60 * time.Second)

	runnerProc.Exec()

	// now just wait for the user to press enter
	runnerProc.BlockWait(60 * time.Second)

	runnerProc.Stop()
}

func printMyProcInfo() {
	// check the pid
	myPid := os.Getpid()
	if myPid == 0 {
		panic("os.Getpid() returns 0. wtf")
	}
	fmt.Println("the procId of this Process:", myPid)
	procWatch, err := process.NewProcessWatcherByPid(myPid)
	if err != nil {
		panic(err)
	}
	if procWatch == nil {
		panic("ProcessWatcher is nil")
	}
	fmt.Println("my own pid:", procWatch.GetPid())
	//fmt.Println("cmd:", procWatch.GetCmd())
	//threads := procWatch.GetThreads()
	//for i, thread := range threads {
	//	fmt.Println("my Threads:", thread, " ", i)
	//}
	childs := procWatch.GetChilds()
	for i, child := range childs {
		fmt.Println("sub-child: [", child, "] ", i)
	}

	procWatch.WalkChildProcs(0, func(p *process.ProcData, parentPid int, level int) bool {
		fmt.Println("-->: [", p.Pid, "] ", parentPid, " ", level, " cmd:", p.Cmd)
		return true
	})
}
