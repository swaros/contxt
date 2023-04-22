package runner_test

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	"github.com/swaros/contxt/module/runner"
)

func TestPosition(t *testing.T) {
	position := runner.CreatePosition(10, 10)

	position.SetProcentage()

	screen := tcell.NewSimulationScreen("UTF-8")
	screen.Init()
	screen.SetSize(1000, 100)

	if position.IsAbsolute() {
		t.Errorf("Expected position to be procentage")
	}

	if !position.IsProcentage() {
		t.Errorf("Expected position to be procentage")
	}

	if position.GetX(screen) == 10 {
		t.Errorf("Expected position to be different to 10 (relative to screen)")
	}

	if position.GetY(screen) != 10 {
		t.Errorf("Expected position to be 10 (relative to screen). got %v", position.GetY(screen))
	}

	if position.GetX(screen) != 100 {
		t.Errorf("Expected position to be 100 (relative to screen). got %v", position.GetX(screen))
	}

	// now set margin. margin will be added absolute to the procentage value
	position.SetMargin(10, 2)

	if position.GetX(screen) != 110 {
		t.Errorf("Expected position to be 110 (relative to screen). got %v", position.GetX(screen))
	}

	if position.GetY(screen) != 12 {
		t.Errorf("Expected position to be 12 (relative to screen). got %v", position.GetY(screen))
	}

	// now set absolute
	position.SetAbsolute()

	if !position.IsAbsolute() {
		t.Errorf("Expected position to be absolute")
	}

	if position.IsProcentage() {
		t.Errorf("Expected position to be absolute")
	}

	tx, ty := position.GetXY(screen)
	// do not forget to add the margin
	if tx != 10+10 {
		t.Errorf("Expected position to be 10 (relative to screen). got %v", tx)
	}

	if ty != 10+2 {
		t.Errorf("Expected position to be 10 (relative to screen). got %v", ty)
	}

	position.SetProcentage()

	tx, ty = position.GetXY(screen)

	if tx != 100+10 {
		t.Errorf("Expected position to be 100 (relative to screen). got %v", tx)
	}

	if ty != 10+2 {
		t.Errorf("Expected position to be 10 (relative to screen). got %v", ty)
	}
}

func TestCoordString(t *testing.T) {
	coord := runner.CreatePosition(10, 10)

	expected := "x:10px y:10px"
	if coord.String() != expected {
		t.Errorf("Expected coord string to be [%s]. got %v", expected, coord.String())
	}

	coord.SetMargin(5, 5)
	expected = "x:10px y:10px margin: 5,5"
	if coord.String() != expected {
		t.Errorf("Expected coord string to be [%s]. got %v", expected, coord.String())
	}

	coord.SetProcentage()
	expected = "x:10% y:10% margin: 5,5"
	if coord.String() != expected {
		t.Errorf("Expected coord string to be [%s]. got %v", expected, coord.String())
	}

}

func TestGetRealPosition(t *testing.T) {
	testPos := runner.CreatePosition(10, 10)
	testPos.SetProcentage()

	screen := tcell.NewSimulationScreen("UTF-8")
	screen.Init()
	screen.SetSize(1000, 100)

	// 10% of 1000 is 100
	if testPos.GetX(screen) != 100 {
		t.Errorf("Expected position to be 100 (relative to screen). got %v", testPos.GetX(screen))
	}

	// 10% of 100 is 10
	if testPos.GetY(screen) != 10 {
		t.Errorf("Expected position to be 10 (relative to screen). got %v", testPos.GetY(screen))
	}

	testPos.SetAbsolute()

	// 10px of 1000 is 10
	if testPos.GetX(screen) != 10 {
		t.Errorf("Expected position to be 10 (relative to screen). got %v", testPos.GetX(screen))
	}

	// 10px of 100 is 10
	if testPos.GetY(screen) != 10 {
		t.Errorf("Expected position to be 10 (relative to screen). got %v", testPos.GetY(screen))
	}

	screen.SetSize(100, 100)

	// 10% of 100 is 10
	if testPos.GetX(screen) != 10 {
		t.Errorf("Expected position to be 10 (relative to screen). got %v", testPos.GetX(screen))
	}

	// 10% of 100 is 10
	if testPos.GetY(screen) != 10 {
		t.Errorf("Expected position to be 10 (relative to screen). got %v", testPos.GetY(screen))
	}

	testPos.SetMargin(10, 10)

	// 10% of 100 is 10 + 10 margin is 20
	if testPos.GetX(screen) != 20 {
		t.Errorf("Expected position to be 20 (relative to screen). got %v", testPos.GetX(screen))
	}

	// 10% of 100 is 10 + 10 margin is 20
	if testPos.GetY(screen) != 20 {
		t.Errorf("Expected position to be 20 (relative to screen). got %v", testPos.GetY(screen))
	}

	readPos := testPos.GetReal(screen)

	// simpele check if the position is the same
	if readPos.GetX(screen) != testPos.GetX(screen) || readPos.GetY(screen) != testPos.GetY(screen) {
		t.Errorf("Expected position to be the same. got %v", readPos.GetX(screen))
	}

}

func TestTcellCoordsHitTestRightAndDown(t *testing.T) {
	testPos := runner.CreatePosition(10, 10)

	checkPos := runner.CreatePosition(10, 10)

	// at the same position is an match
	if !testPos.IsMoreOrEvenRightAndDownThen(checkPos) {
		t.Errorf("Expected %v to be more right and down than %v", checkPos, testPos)
	}

	// more right is an match
	checkPos = runner.CreatePosition(11, 10)
	// [10,10] is not more right and down than [11,10]
	if testPos.IsMoreOrEvenRightAndDownThen(checkPos) {
		t.Errorf("Expected %v to be more right and down than %v", checkPos, testPos)
	}

	// more down is an match
	checkPos = runner.CreatePosition(10, 11)
	// [10,10] is not more right and down than [10,11]
	if testPos.IsMoreOrEvenRightAndDownThen(checkPos) {
		t.Errorf("Expected %v to be more right and down than %v", checkPos, testPos)
	}

	// more right and down is an match
	checkPos = runner.CreatePosition(11, 11)
	// [10,10] is not more right and down than [11,11]
	if testPos.IsMoreOrEvenRightAndDownThen(checkPos) {
		t.Errorf("Expected %v to be more right and down than %v", checkPos, testPos)
	}

}

func TestTcellCoordsHitTestLeftAndUp(t *testing.T) {
	testPos := runner.CreatePosition(10, 10)

	checkPos := runner.CreatePosition(10, 10)

	// at the same position is an match
	if !testPos.IsMoreOrEvenRightAndDownThen(checkPos) {
		t.Errorf("Expected %v to be more right and down than %v", checkPos, testPos)
	}

	// more left is not an match
	checkPos = runner.CreatePosition(9, 10)
	// [10,10] is more right and down than [9,10]
	if !testPos.IsMoreOrEvenRightAndDownThen(checkPos) {
		t.Errorf("Expected %v to be more right and down than %v", checkPos, testPos)
	}
	// [10,10] is not more left and up than [9,10]
	if testPos.IsLessOrEvenLeftAndUpThen(checkPos) {
		t.Errorf("Expected %v to be more left and up than %v", checkPos, testPos)
	}

}

func TestHitByCoordsFuncs(t *testing.T) {
	boxTopLeftCoords := runner.CreatePosition(10, 10)
	boxBottomRightCoords := runner.CreatePosition(20, 20)

	checkPos := runner.CreatePosition(15, 15)

	if !checkPos.IsInBox(boxTopLeftCoords, boxBottomRightCoords) {
		t.Errorf("1. Expected %v to be in box %v, %v", checkPos, boxTopLeftCoords, boxBottomRightCoords)
	}

	checkPos = runner.CreatePosition(10, 10)

	if !checkPos.IsInBox(boxTopLeftCoords, boxBottomRightCoords) {
		t.Errorf("2. Expected %v to be in box %v, %v", checkPos, boxTopLeftCoords, boxBottomRightCoords)
	}

}
