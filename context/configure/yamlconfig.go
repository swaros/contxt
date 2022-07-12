// Copyright (c) 2020 Thomas Ziegler <thomas.zglr@googlemail.com>. All rights reserved.
//
// Licensed under the MIT License
//
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

package configure

// StopReasons mapping for RunConfig.Task[].Stopreasons
type StopReasons struct {
	Onerror        bool
	OnoutcountLess int
	OnoutcountMore int
	OnoutContains  []string
}

// IncludePaths are files the defines how variables should be parsed.
// they indludes folders they have to be parsed first so they contents
// can be sued to proceeed with test/template.
// otherwise the yaml file is not readable
type IncludePaths struct {
	Include struct {
		Basedir bool     `yaml:"basedir"`
		Folders []string `yaml:"folders"`
	} `yaml:"include"`
}

// ++++++++++++++++++++ IMPORTANT !!!! ++++++++++++++++++++++++++++++
// after autogenerate todos:
// Variables are map[string]string and contains settings for Placeholders. add yaml:"variables,omitempty
//
// change all variables anywhere
//
// Variables    map[string]string `yaml:"variables,omitempty"`
//
// same for environment in Require
//
// Environment map[string]string `yaml:"environment,omitempty"`
//

// RunConfig defines the structure of the local stored execution files
type RunConfig struct {
	Version string `yaml:"version"`
	Config  Config `yaml:"config"`
	Task    []Task `yaml:"task"`
}

// Autorun defines the targets they have to be executed
// if a special event is triggered
type Autorun struct {
	// this target will be executed if to this workspace was changed
	Onenter string `yaml:"onenter"`

	// this target will be executed if we changing to another workspace
	// so we leaving the current workspace
	// can be used for cleanup ...as a example
	Onleave string `yaml:"onleave"`
}

// Config is the main Configuration part of the Template.
type Config struct {
	Sequencially  bool              `yaml:"sequencially"`
	Coloroff      bool              `yaml:"coloroff"`
	Loglevel      string            `yaml:"loglevel"`
	Variables     map[string]string `yaml:"variables,omitempty"`
	Autorun       Autorun           `yaml:"autorun"`
	Imports       []string          `yaml:"imports"`
	Use           []string          `yaml:"use"`
	Require       []string          `yaml:"require"`
	MergeTasks    bool              `yaml:"mergetasks"`
	AllowMutliRun bool              `yaml:"allowmultiplerun"`
}

// Require defines what is required to execute the task
type Require struct {
	Exists      []string          `yaml:"exists"`
	NotExists   []string          `yaml:"notExists"`
	Environment map[string]string `yaml:"environment,omitempty"`
	Variables   map[string]string `yaml:"variables,omitempty"`
	System      string            `yaml:"system,omitempty"`
}

// Options are the per-task options
type Options struct {
	IgnoreCmdError bool     `yaml:"ignoreCmdError"`
	Format         string   `yaml:"format"`
	Stickcursor    bool     `yaml:"stickcursor"`
	Colorcode      string   `yaml:"colorcode"`
	Bgcolorcode    string   `yaml:"bgcolorcode"`
	Panelsize      int      `yaml:"panelsize"`
	Displaycmd     bool     `yaml:"displaycmd"`
	Hideout        bool     `yaml:"hideout"`
	Invisible      bool     `yaml:"invisible"`
	Maincmd        string   `yaml:"maincmd"`
	Mainparams     []string `yaml:"mainparams"`
	NoAutoRunNeeds bool     `yaml:"noAutoRunNeeds"`
	TimeoutNeeds   int      `yaml:"timeoutNeeds"`
	TickTimeNeeds  int      `yaml:"tickTimeNeeds"`
	WorkingDir     string   `yaml:"workingdir"`
}

// Trigger are part of listener. The defines
// some events they are triggered by executing scripts
// most of them watching the output
type Trigger struct {
	Onerror        bool     `yaml:"onerror"`
	OnoutcountLess int      `yaml:"onoutcountLess"`
	OnoutcountMore int      `yaml:"onoutcountMore"`
	OnoutContains  []string `yaml:"onoutContains"`
	Now            bool     `yaml:"now"`
}

// Action defines what should happens Next.
type Action struct {
	Target  string   `yaml:"target"`
	Stopall bool     `yaml:"stopall"`
	Script  []string `yaml:"script"`
}

// Listener are used for watching events
// and triggers an action if a event happens
type Listener struct {
	Trigger Trigger `yaml:"trigger"`
	Action  Action  `yaml:"action"`
}

// Task is the main Script
type Task struct {
	ID          string            `yaml:"id"`
	Variables   map[string]string `yaml:"variables,omitempty"`
	Requires    Require           `yaml:"require"`
	Stopreasons Trigger           `yaml:"stopreasons"`
	Options     Options           `yaml:"options"`
	Script      []string          `yaml:"script"`
	Listener    []Listener        `yaml:"listener"`
	Next        []string          `yaml:"next"`
	RunTargets  []string          `yaml:"runTargets"`
	Needs       []string          `yaml:"needs"`
}
