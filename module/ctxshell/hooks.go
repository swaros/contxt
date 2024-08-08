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

package ctxshell

import "regexp"

type Hook struct {
	Before  func() error
	After   func() error
	Pattern string
}

func NewHook(pattern string, before, after func() error) Hook {
	return Hook{
		Before:  before,
		After:   after,
		Pattern: pattern,
	}
}

// Match returns true if the pattern matches the hook's pattern by regexp.
func (h Hook) Match(pattern string) bool {
	re := regexp.MustCompile(h.Pattern)
	return re.MatchString(pattern)
}

func (t *Cshell) AddHook(hook Hook) {
	t.hooks = append(t.hooks, hook)
}

func (t *Cshell) GetHooksByPattern(pattern string) []Hook {
	hooks := []Hook{}
	for _, hook := range t.hooks {
		if hook.Match(pattern) {
			hooks = append(hooks, hook)
		}
	}
	return hooks
}

func (t *Cshell) executeHooksBefore(hooks []Hook) error {
	for _, hook := range hooks {
		if hook.Before != nil {
			err := hook.Before()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *Cshell) executeHooksAfter(hooks []Hook) error {
	for _, hook := range hooks {
		if hook.After != nil {
			err := hook.After()
			if err != nil {
				return err
			}
		}
	}
	return nil
}
