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

package ctxout

import (
	"strings"
)

type SignFilter struct {
	behaveInfo PostFilterInfo
	signs      SignSet
	enabled    bool
	markup     Markup
	forceEmpty bool
}

// NewSignFilter returns a new SignFilter
// if signSet is nil, the default signs are used
func NewSignFilter(signSet *SignSet) *SignFilter {
	sf := &SignFilter{
		markup: *NewMarkup().SetAccepptedTags([]string{"sign"}),
	}
	if signSet == nil {
		sf.signs = *NewBaseSignSet()
	} else {
		sf.signs = *signSet
	}
	return sf
}

// Update updates the filter with the new info
func (sf *SignFilter) Update(info PostFilterInfo) {
	sf.behaveInfo = info
	// curremtly it is hard to check if the terminal is able to display the utf-8 signs
	// so we disable the filter if we do not have an terminal, or the terminal is not able to display colors
	// and of course if the plugin is in general disabled
	if info.Disabled || !info.IsTerminal || !info.Colored {
		sf.enabled = false
	}
}

func (sf *SignFilter) ForceEmpty(forceEmpty bool) {
	sf.forceEmpty = forceEmpty
}

// Enabled returns true if the filter is enabled
func (sf *SignFilter) Enabled() bool {
	return sf.enabled
}

// Enable enables the filter
func (sf *SignFilter) Enable() {
	sf.enabled = true
}

// Disable disables the filter
func (sf *SignFilter) Disable() {
	sf.enabled = false
}

func (sf *SignFilter) GetSign(name string) Sign {
	return sf.signs.GetSign(name)
}

func (sf *SignFilter) AddSign(sign Sign) *SignFilter {
	sf.signs.AddSign(sign)
	return sf
}

// the format for an sign is:
//
//	anything before <sign info> and anything afterwards <sign warning> and again anything afterwards
func (sf *SignFilter) CanHandleThis(text string) bool {
	// check if the text contains an sign
	return strings.Contains(text, "<sign")
}

func (sf *SignFilter) doSign(signSource string) string {
	if sf.forceEmpty {
		return ""
	}
	for _, sign := range sf.signs.Signs {
		expectedMarkup := "<sign " + sign.Name + ">"
		if expectedMarkup == signSource {
			if sf.enabled {
				return sign.Glyph
			} else {
				return sign.Fallback
			}
		}

	}
	return ""
}

func (sf *SignFilter) Command(cmd string) string {
	parsed := sf.markup.Parse(cmd)
	output := ""
	for _, p := range parsed {
		if p.IsMarkup {
			output += sf.doSign(p.Text)
		} else {
			output += p.Text
		}
	}
	return output
}
