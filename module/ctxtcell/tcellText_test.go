package ctxtcell_test

import (
	"sync"
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/swaros/contxt/module/ctxtcell"
)

func TestAddTextElement(t *testing.T) {
	main := ctxtcell.NewTcell()
	// we need to make sure any previous elements are removed
	main.ClearElements()

	// adding a active Text Element
	text := main.ActiveText("I am active")
	text.SetPos(10, 10)
	if id, err := main.AddElement(text); id != 1 || err != nil {
		if err != nil {
			t.Error(err)
		}
		if id != 1 {
			t.Errorf("Expected id to be 1, but got %v", id)
		}
	}

	// adding a inactive Text Element
	text = main.Text("I am inactive")
	text.SetPos(10, 11)
	if id, err := main.AddElement(text); id != 2 || err != nil {
		if err != nil {
			t.Error(err)
		}
		if id != 2 {
			t.Errorf("Expected id to be 2, but got %v", id)
		}
	}

	// adding a active Text Element
	text = main.ActiveText("I am active")
	text.SetPos(10, 12)
	if id, err := main.AddElement(text); id != 3 || err != nil {
		if err != nil {
			t.Error(err)
		}
		if id != 3 {
			t.Errorf("Expected id to be 3, but got %v", id)
		}
	}

	// adding a inactive Text Element
	text = main.Text("I am inactive")
	text.SetPos(10, 13)
	if id, err := main.AddElement(text); id != 4 || err != nil {
		if err != nil {
			t.Error(err)
		}
		if id != 4 {
			t.Errorf("Expected id to be 4, but got %v", id)
		}
	}

	// adding a active Text Element
	text = main.ActiveText("I am active")
	text.SetPos(10, 14)
	if id, err := main.AddElement(text); id != 5 || err != nil {
		if err != nil {
			t.Error(err)
		}
		if id != 5 {
			t.Errorf("Expected id to be 5, but got %v", id)
		}
	}

	// adding a inactive Text Element
	text = main.Text("I am inactive")
	text.SetPos(10, 15)
	if id, err := main.AddElement(text); id != 6 || err != nil {
		if err != nil {
			t.Error(err)
		}
		if id != 6 {
			t.Errorf("Expected id to be 6, but got %v", id)
		}
	}
	// test if the elements are in the list
	if len(main.GetSortedElements()) != 6 {
		t.Errorf("Expected 6 elements, but got %v", len(main.GetSortedElements()))
	}

	// cycle through the elements should be 1,3,5
	// not active elements should be skipped
	main.CycleFocus()
	if main.GetFocusedElement().GetID() != 1 {
		t.Errorf("Expected id to be 1, but got %v", main.GetFocusedElement().GetID())
	}
	main.CycleFocus()
	if main.GetFocusedElement().GetID() != 3 {
		t.Errorf("Expected id to be 3, but got %v", main.GetFocusedElement().GetID())
	}
	main.CycleFocus()
	if main.GetFocusedElement().GetID() != 5 {
		t.Errorf("Expected id to be 5, but got %v", main.GetFocusedElement().GetID())
	}

	// ids always increasing even if we remove elements
	main.RemoveElementByID(2)
	main.RemoveElementByID(4)
	main.RemoveElementByID(6)

	// test if the elements are in the list
	if len(main.GetSortedElements()) != 3 {
		t.Errorf("Expected 3 elements, but got %v", len(main.GetSortedElements()))
	}

}

func TestMouseHitOnTextWithMargin(t *testing.T) {
	app := GetTestScreen(t)
	app.SetMouse(true)

	// using a waitgroup to wait for the app to be ready
	wg := sync.WaitGroup{}
	wg.Add(1)

	getHit := false
	// adding a active Text Element
	clickableText := app.ActiveText("I am clickable")
	clickableText.SetPos(30, 10)
	clickableText.GetPos().SetMargin(-20, 0)
	clickableText.OnSelect = func(selectedt bool) {
		if selectedt {
			getHit = true
			wg.Done()
			app.Stop()
		}
	}
	if id, err := app.AddElement(clickableText); id != 1 || err != nil {
		if err != nil {
			t.Error(err)
		}
		if id != 1 {
			t.Errorf("Expected id to be 1, but got %v", id)
		}
	}

	// running the app in a goroutine
	go app.Loop()

	// send mouse event that simulates the click on the text
	event := tcell.NewEventMouse(10, 10, tcell.Button1, 1)
	app.SendEvent(event)

	// wait for the app to stop
	wg.Wait()
	if !getHit {
		t.Errorf("Expected to get a hit, but didn't")
	}

}

func TestHits(t *testing.T) {
	app := GetTestScreen(t)
	// this screen is 100 x 100
	screen := app.GetScreen()
	clickableText := app.ActiveText("I am clickable")
	clickableText.SetPos(30, 10)
	clickableText.GetPos().SetMargin(-20, 0)

	// test if the text is hit
	fakeHitPosition := ctxtcell.CreatePosition(10, 10, false)
	if !clickableText.Hit(fakeHitPosition, screen) {
		t.Errorf("Expected to hit the text, but didn't")
	}

	// simualte a click on the text on the right top corner. allign right
	exitTopMenu := app.ActiveText("exit")
	exitTopMenu.SetPosProcentage(100, 0).
		SetStyle(tcell.StyleDefault.Foreground(tcell.ColorGoldenrod).Background(tcell.ColorBlack))

	exitTopMenu.GetPos().SetMargin(-5, 0) // minus length of the text + 1

	fakeHitPosition = ctxtcell.CreatePosition(95, 0, false) // click on the right top corner
	if !exitTopMenu.Hit(fakeHitPosition, screen) {
		t.Errorf("Expected to hit the text, but didn't")
	}

	// now test some positions they should not hit
	fakeHitPosition = ctxtcell.CreatePosition(0, 0, false) // click on the left top corner
	if exitTopMenu.Hit(fakeHitPosition, screen) {
		t.Errorf("Expected to not hit the text, but did")
	}

	fakeHitPosition = ctxtcell.CreatePosition(0, 10, false) // click on the left side
	if exitTopMenu.Hit(fakeHitPosition, screen) {
		t.Errorf("Expected to not hit the text, but did")
	}

	fakeHitPosition = ctxtcell.CreatePosition(95, 10, false) // click on the right side
	if exitTopMenu.Hit(fakeHitPosition, screen) {
		t.Errorf("Expected to not hit the text, but did")
	}
}
