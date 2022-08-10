package tviewapp_test

import (
	"testing"

	"github.com/swaros/contxt/tviewapp"
)

func TestHit(t *testing.T) {
	App := tviewapp.New()

	box1 := tviewapp.NewBox()
	box1.SetDim(10, 10, 100, 100)

	box2 := tviewapp.NewBox()
	box2.SetDim(20, 20, 100, 100)
	App.AddElement(box1, box2)

}
