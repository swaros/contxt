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

package configure

// Configuration includes all paths for the current workspace
// this is the main Configuration that is needed to eep track which path
// is currently used.
// this is the used format til version 0.4.0
type Configuration struct {
	CurrentSet string
	Paths      []string
	LastIndex  int
	PathInfo   map[string]WorkspaceInfo
}

// new version of the configuration starts here
type ConfigMetaV2 struct {
	CurrentSet string                     `yaml:"CurrentSet"`
	Configs    map[string]ConfigurationV2 `yaml:"Configs"`
}

type WorkspaceInfoV2 struct {
	Path    string `yaml:"Path"`
	Project string `yaml:"Project"`
	Role    string `yaml:"Role"`
	Version string `yaml:"Version"`
}

type ConfigurationV2 struct {
	Name         string                     `yaml:"Name"`         // the name of the workspace
	CurrentIndex string                     `yaml:"CurrentIndex"` // what of the workspaces are the current used
	Paths        map[string]WorkspaceInfoV2 `yaml:"Paths"`
}

// CommandLine defines a line of commands that can be executed
type CommandLine struct {
	Require            RequireCheck
	Command            string
	Params             string
	Comment            string
	StopOnError        bool
	StopOnOutCountLess int
	StopOnOutCountMore int
	StopOnOutContains  string
	TraceOutput        bool
}

// ExecuteDefinition Defines the structure of a .execute file that defines commands they have to executed
type ExecuteDefinition struct {
	TestScript  []CommandLine
	InitScript  []CommandLine
	CleanScript []CommandLine
	Script      []CommandLine
}

// RequireCheck defines some variables they have to be valid before script runs
type RequireCheck struct {
	FileExists    []string
	FileNotExists []string
}

type GitVersionInfo struct {
	HashUsed    string
	Reference   string
	Repositiory string
	Path        string
	Exists      bool
}
