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
	"errors"
	"strings"
)

// defining an sign that can be used to display information as an sign.
// this sign is an utf-8 character. and a fallback string that is used
// if the device is not able to display the sign.
type Sign struct {
	Name     string // name of the sign. used also as identifier
	Glyph    string // the utf-8 character
	Fallback string // the fallback string
}

type SignSet struct {
	Signs []Sign
}

var (
	// NotFoundSign is the sign that is returned if the sign is not found
	NotFoundSign = Sign{
		Name:     "notfound",
		Glyph:    "?",
		Fallback: "[?]",
	}
)

// NewBaseSignSet returns a new SignSet with the basic signs
func NewBaseSignSet() *SignSet {
	set := &SignSet{}
	set.Signs = []Sign{
		{
			Name:     "info",
			Glyph:    "🗩",
			Fallback: "[i]",
		},
		{
			Name:     "warning",
			Glyph:    "⚠",
			Fallback: "[!]",
		},
		{
			Name:     "error",
			Glyph:    "⛔",
			Fallback: "[×]",
		},
		{
			Name:     "success",
			Glyph:    "✔",
			Fallback: "[ok]",
		},
		{
			Name:     "debug",
			Glyph:    "👓",
			Fallback: "[•…]",
		},
		{
			Name:     "screen",
			Glyph:    "🖵",
			Fallback: "▕░▒░▏",
		},
	}
	return set
}

// Default Constants based on the basic signs

const (
	BaseSignInfo    = "<sign info>"
	BaseSignWarning = "<sign warning>"
	BaseSignError   = "<sign error>"
	BaseSignSuccess = "<sign success>"
	BaseSignDebug   = "<sign debug>"
	BaseSignScreen  = "<sign screen>"
)

// GetSign returns the sign with the given name
func (s *SignSet) GetSign(name string) Sign {
	for _, sign := range s.Signs {
		if sign.Name == name {
			return sign
		}
	}
	return NotFoundSign
}

// AddSign adds an sign to the set
func (s *SignSet) AddSign(sign Sign) error {
	var err error
	// remove any markup chars from fallback
	if s.cleanUpUnwanted(&sign) {
		err = errors.New("fallback contains unwanted markup chars")
	}
	s.Signs = append(s.Signs, sign)
	return err
}

func (s *SignSet) cleanUpUnwanted(sign *Sign) bool {
	signUnwanted := []string{"<", ">"}
	orig := sign.Fallback
	for _, signUnwanted := range signUnwanted {
		sign.Fallback = strings.ReplaceAll(sign.Fallback, signUnwanted, "invalid-fallback-char-removed")
	}
	return orig != sign.Fallback

}
