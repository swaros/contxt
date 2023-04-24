package ctxtcell

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

type position struct {
	X            int
	Y            int
	isProcentage bool
	margin       margin
}

type dim struct {
	w int
	h int
}

type margin struct {
	top  int
	left int
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

func (p *position) SetMargin(left, top int) {
	p.margin.top = top
	p.margin.left = left
}

func (p *position) GetX(s tcell.Screen) int {
	if p.isProcentage {
		w, _ := s.Size()
		return (w * p.X / 100) + p.margin.left
	}
	return p.X + p.margin.left
}

func (p *position) GetY(s tcell.Screen) int {
	if p.isProcentage {
		_, h := s.Size()
		return (h * p.Y / 100) + p.margin.top
	}
	return p.Y + p.margin.top
}

func (p *position) GetXY(s tcell.Screen) (int, int) {
	if p.isProcentage {
		w, h := s.Size()
		return (w * p.X / 100) + p.margin.left, (h * p.Y / 100) + p.margin.top
	}
	return p.X + p.margin.left, p.Y + p.margin.top
}

func (p *position) GetReal(s tcell.Screen) position {
	x, y := p.GetXY(s)
	return position{X: x, Y: y}
}

// test if a different position more right and down than my self. or at least at same position
// this is not using the margin or relative position
// all positions have to be calculated before, depending on the screen size
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
	metric := "px"
	if p.isProcentage {
		metric = "%"
	}
	marginStr := ""
	if p.margin.top != 0 || p.margin.left != 0 {
		marginStr = fmt.Sprintf(" margin: %d,%d", p.margin.top, p.margin.left)
	}
	return fmt.Sprintf("x:%d%s y:%d%s%s", p.X, metric, p.Y, metric, marginStr)
}