package runner_test

import (
	"testing"

	"github.com/swaros/contxt/module/runner"
)

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
