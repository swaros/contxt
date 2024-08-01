package tasks

func (t *targetExecuter) SetFunctions(anko *AnkoRunner) {
	// make it possible, to stop the execution of the script, by its own
	anko.Define("exit", func() {
		anko.cancelationFn()
	})
}
