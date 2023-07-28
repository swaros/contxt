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

package mimiclog

type NullLogger struct {
}

func NewNullLogger() *NullLogger {
	return &NullLogger{}
}

func (l *NullLogger) Trace(args ...interface{})    {}
func (l *NullLogger) Debug(args ...interface{})    {}
func (l *NullLogger) Info(args ...interface{})     {}
func (l *NullLogger) Error(args ...interface{})    {}
func (l *NullLogger) Warn(args ...interface{})     {}
func (l *NullLogger) Critical(args ...interface{}) {}

func (l *NullLogger) IsLevelEnabled(level string) bool { return false }
func (l *NullLogger) IsTraceEnabled() bool             { return false }
func (l *NullLogger) IsDebugEnabled() bool             { return false }
func (l *NullLogger) IsInfoEnabled() bool              { return false }
func (l *NullLogger) IsWarnEnabled() bool              { return false }
func (l *NullLogger) IsErrorEnabled() bool             { return false }
func (l *NullLogger) IsCriticalEnabled() bool          { return false }

func (l *NullLogger) SetLevelByString(level string) {}
func (l *NullLogger) SetLevel(level interface{})    {}
func (l *NullLogger) GetLevel() string              { return LevelNone }

func (l *NullLogger) GetLogger() interface{} { return nil }

func (l *NullLogger) SetOutput(output interface{}) {}

// uncomment this to make sure NullLogger implements Logger.
// this way we see what we need to implement
/*
func dummyForTest() {
	var _ Logger = NewNullLogger()
}
*/
