package runner

import "github.com/gdamore/tcell/v2"

type position struct {
	X            int
	Y            int
	isProcentage bool
}

type dim struct {
	w int
	h int
}

// Coordinates is a struct that contains the position and the dimensions of an element
type Coordinates struct {
	TopLeft    position
	Dimensions dim
}

func NewCoordinates(topLeft position, w int, h int) *Coordinates {
	return &Coordinates{
		TopLeft: topLeft,
		Dimensions: dim{
			w: w,
			h: h,
		},
	}
}

func CreatePosition(x, y int) position {
	return position{X: x, Y: y}
}

func (p *position) SetProcentage() {
	p.isProcentage = true
}

func (p *position) SetAbsolute() {
	p.isProcentage = false
}

func (p *position) IsProcentage() bool {
	return p.isProcentage
}

func (p *position) IsAbsolute() bool {
	return !p.isProcentage
}

func (p *position) GetXY(s tcell.Screen) (int, int) {
	if p.isProcentage {
		w, h := s.Size()
		return w * p.X / 100, h * p.Y / 100
	}
	return p.X, p.Y
}

func (p *position) GetReal(s tcell.Screen) position {
	if p.isProcentage {
		w, h := s.Size()
		return position{X: w * p.X / 100, Y: h * p.Y / 100}
	}
	return *p
}

// test if a different position more right and down than my self. or at least at same position
func (p *position) IsMoreOrEvenRightAndDownThen(testPosition position) bool {
	if p.X >= testPosition.X && p.Y >= testPosition.Y {
		return true
	}
	return false
}

// test if a different position more left and up than my self. or at least at same position
func (p *position) IsLessOrEvenLeftAndUpThen(testPosition position) bool {
	if p.X <= testPosition.X && p.Y <= testPosition.Y {
		return true
	}
	return false
}

func (p *position) IsInBox(topLeft position, bottomRight position) bool {
	if p.IsMoreOrEvenRightAndDownThen(topLeft) && p.IsLessOrEvenLeftAndUpThen(bottomRight) {
		return true
	}
	return false
}
func (p *position) String() string {
	return "X: " + string(rune(p.X)) + " Y: " + string(rune(p.Y))
}
