// MIT License
//
// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the Software), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED AS IS, WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

// AINC-NOTE-0815

package systools

const (
	ExitOk               = 0   // Everything is fine
	ErrorExitDefault     = 1   // any exit depending an error that do not need being specific
	ErrorInitApp         = 2   // any application error while setting up
	ErrorWhileLoadCfg    = 3   // any error while loading configuration
	ErrorOnConfigImport  = 5   // a import could not handled
	ErrorTemplate        = 6   // template related error. depending reported issues about version, lint, yaml structure
	ErrorTemplateReading = 7   // errors processing the template
	ErrorCheatMacros     = 8   // errors processing the cheat macros
	ErrorBySystem        = 10  // errors related to the system, like while change dir or reading a file
	ExitByStopReason     = 101 // ExitByStopReason the process stopped because of a defined reason
	ExitNoCode           = 102 // ExitNoCode means there was no code associated
	ExitCmdError         = 103 // ExitCmdError means the execution of the command fails. a error by the command itself
	ExitByRequirement    = 104 // ExitByRequirement means a requirement was not fulfills
	ExitAlreadyRunning   = 105 // ExitAlreadyRunning means the task is not started, because it is already created
	ExitByNoTargetExists = 106 // none of the targets are matching in requirements
	ExitByNothingToDo    = 107 // none of the targets have any usefull work to do or did not match any requirements
)
