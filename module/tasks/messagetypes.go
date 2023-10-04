// Copyright (c) 2023 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// # Licensed under the MIT License
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.
package tasks

// MsgCommand is the command to execute
type MsgCommand string

// MsgTarget is the target to execute the command on
type MsgTarget struct {
	Target  string
	Context string
	Info    string
}

// MsgReason is the reason that is used to set why somethingis triggered. like stopreason
type MsgReason string

// MsgType is the type of the message
type MsgType string

// MsgInfo is the info that is just some additional context
type MsgInfo string

// MsgNumber is some numeric value
type MsgNumber int
type MsgArgs []string

// MsgProcess is the process id that is running
type MsgProcess string
type MsgPid int
type MsgError struct {
	Err       error
	Target    string
	Reference string
}
type MsgExecOutput struct {
	Target string
	Output string
}
type MsgStickCursor bool
