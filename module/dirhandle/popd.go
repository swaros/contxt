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

package dirhandle

import "os"

// Popd is a struct to handle pushd/popd
// it is like the regular pushd/popd from the shell
// but it is not a stack. it is just a single directory
// that can be pushed and poped.
// the directory is saved in the struct.
// the directory is changed with the Popd() function.
// the directory is changed back with the Popd() function in the returned struct.
// so you always have the context, and do not need to remember the directory.
//
//	p := dirhandle.Pushd()
//	defer p.Popd()
//	// do something
//	// the directory is changed back to the original directory
//	// when the function returns
//

type Popd struct {
	dir string
}

// Pushd returns a new popd struct with the current directory
func Pushd() *Popd {
	if dir, err := Current(); err != nil {
		panic(err)
	} else {
		return &Popd{dir}
	}
}

// Popd changes the directory back to the directory that was saved in the struct
func (p *Popd) Popd() error {
	return os.Chdir(p.dir)
}
