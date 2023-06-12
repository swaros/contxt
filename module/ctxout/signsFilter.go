package ctxout

import "strings"

type SignFilter struct {
	behaveInfo PostFilterInfo
	signs      SignSet
	enabled    bool
}

// NewSignFilter returns a new SignFilter
// if signSet is nil, the default signs are used
func NewSignFilter(signSet *SignSet) *SignFilter {
	sf := &SignFilter{}
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

func (sf *SignFilter) Command(cmd string) string {
	didIt := false
	for _, sign := range sf.signs.Signs {
		expectedMarkup := "<sign " + sign.Name + ">"
		if strings.Contains(cmd, expectedMarkup) {
			didIt = true
			cmd = sf.doSignInText(sign, expectedMarkup, cmd)
		}

	}
	// fallback if the sign is not found
	// we just clear the markup istead of <sign whatever> we print <!whatever>
	// as a hint, that this was not working
	if !didIt {
		cmd = strings.Replace(cmd, "<sign ", "<!", -1)
	}
	return cmd
}

func (sf *SignFilter) doSignInText(sign Sign, markup, cmd string) string {
	if sf.enabled {
		return strings.Replace(cmd, markup, sign.Glyph, -1)
	}
	return strings.Replace(cmd, markup, sign.Fallback, -1)
}
