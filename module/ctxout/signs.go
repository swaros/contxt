package ctxout

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
			Fallback: "[!!!]",
		},
		{
			Name:     "success",
			Glyph:    "✔",
			Fallback: "[ok]",
		},
		{
			Name:     "debug",
			Glyph:    "👓",
			Fallback: "[¿]",
		},
		{
			Name:     "screen",
			Glyph:    "🖵",
			Fallback: "[«»]",
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
func (s *SignSet) AddSign(sign Sign) {
	s.Signs = append(s.Signs, sign)
}
