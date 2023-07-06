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

 package main

import (
	"fmt"

	"github.com/swaros/contxt/module/yacl"
	"github.com/swaros/contxt/module/yamc"
)

type Config struct {
	AuthUrl  string `yaml:"authurl"`
	UserName string `yaml:"username"`
	Password string `yaml:"password"`
}

func main() {
	// create a new yacl instance
	config := Config{}
	cfgApp := yacl.New(
		&config,
		yamc.NewYamlReader(),
	)

	// define the subdirectory where the config files are located
	cfgApp.SetSubDirs("config")

	// load all config files from the config directory
	if err := cfgApp.Load(); err != nil {
		panic(err)
	}

	// just print the overwritten values
	fmt.Println(
		" connecting: ", config.AuthUrl,
		", user: ", config.UserName,
		", password: ", config.Password,
	)

}
