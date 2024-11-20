package runner

type OutputHandler interface {
	GetOutHandler(c *CmdExecutorImpl) func(msg ...interface{})
	GetName() string
}
