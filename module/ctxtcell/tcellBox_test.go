package ctxtcell_test

import (
	"testing"
)

func TestAddBox(t *testing.T) {
	screen := GetTestScreen(t)

	box := screen.NewBox().SetTopLeft(10, 10).SetBottomRight(20, 20)
	screen.AddElement(box)
	if cnt := screen.DrawAll(); cnt != 1 {
		t.Errorf("Expected 1 box to be drawn but got %v", cnt)
	}

	scr := screen.GetScreen()
	cnt, _, _, _ := scr.GetContent(10, 10)
	if cnt != rune('┌') {
		t.Errorf("Expected '┌' at 10,10 but got %v", cnt)
	}

	cnt, _, _, _ = scr.GetContent(20, 10)
	if cnt != rune('┐') {
		t.Errorf("Expected '┐' at 20,10 but got %v", cnt)
	}

	cnt, _, _, _ = scr.GetContent(10, 20)
	if cnt != rune('└') {
		t.Errorf("Expected '└' at 10,20 but got %v", cnt)
	}

	cnt, _, _, _ = scr.GetContent(20, 20)
	if cnt != rune('┘') {
		t.Errorf("Expected '┘' at 20,20 but got %v", cnt)
	}
}

func TestAddBoxInRealtivePos(t *testing.T) {
	// we have a screen with 100x100
	screen := GetTestScreen(t)
	scr := screen.GetScreen()
	// we have a screen with 100x100 and now we set the size to 1000x100
	scr.SetSize(1000, 100)

	// create a box with the top left at 5% of the screen
	// and the bottom right at 15% of the screen

	box := screen.NewBox().SetTopLeftProcentage(5, 5).SetBottomRightProcentage(15, 15)
	screen.AddElement(box)
	// draw all elements (the box) and make sure only oure textbox is drawn
	if cnt := screen.DrawAll(); cnt != 1 {
		t.Errorf("Expected 1 box to be drawn but got %v", cnt)
	}

	// check the corners of the box
	// with the the positions relative to the screen size
	// 5% of 1000 is 50
	// 5% of 100 is 5
	cnt, _, _, _ := scr.GetContent(50, 5)
	if cnt != rune('┌') {
		t.Errorf("Expected '┌' at 5,5 but got %v", cnt)
	}

	// 15% of 1000 is 150
	// 5% of 100 is 5
	cnt, _, _, _ = scr.GetContent(150, 5)
	if cnt != rune('┐') {
		t.Errorf("Expected '┐' at 5,5 but got %v", cnt)
	}

	// 5% of 1000 is 50
	// 15% of 100 is 15
	cnt, _, _, _ = scr.GetContent(50, 15)
	if cnt != rune('└') {
		t.Errorf("Expected '└' at 5,5 but got %v", cnt)
	}

	// 15% of 1000 is 150
	// 15% of 100 is 15
	cnt, _, _, _ = scr.GetContent(150, 15)
	if cnt != rune('┘') {
		t.Errorf("Expected '┘' at 5,5 but got %v", cnt)
	}

}
