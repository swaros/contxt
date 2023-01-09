package tasks

// MsgCommand is the command to execute
type MsgCommand string

// MsgTarget is the target to execute the command on
type MsgTarget string

// MsgReason is the reason that is used to set why somethingis triggered. like stopreason
type MsgReason string

// MsgType is the type of the message
type MsgType string
type MsgArgs []string

// MsgProcess is the process id that is running
type MsgProcess string
type MsgPid int
type MsgError error
type MsgExecOutput string
type MsgStickCursor bool
