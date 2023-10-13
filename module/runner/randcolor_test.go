package runner_test

import (
	"fmt"
	"testing"

	"github.com/swaros/contxt/module/runner"
)

func TestColorPick(t *testing.T) {
	randColor := runner.NewRandColorStore()

	// create a couple of colors with different target names
	// and lets check if we hit some twice
	colors := []runner.RandColor{}
	for i := 0; i < randColor.GetMaxVariants(); i++ {
		targetName := "test" + fmt.Sprintf("_%v", i)
		colorPicked := randColor.GetOrSetIndexColor(targetName)
		// just using this test for also testing the IsUnused() method
		if !randColor.IsInusage(colorPicked) {
			t.Error("whoops. this color :", colorPicked, "for target:", targetName, " MUST be detected as used")
		}
		for inx, color := range colors {
			if colorPicked == color {
				t.Error("color picked twice:", colorPicked, "for target:", targetName, "already picked for index:", inx)
			}
		}

		colors = append(colors, colorPicked)
	}
}

func TestColorPickRand(t *testing.T) {
	randColor := runner.NewRandColorStore()

	successCount := 0
	failCount := 0

	// create a couple of colors with different target names
	// and lets check if we hit some twice
	colors := []runner.RandColor{}
	for i := 0; i < randColor.GetMaxVariants(); i++ {
		targetName := "test" + fmt.Sprintf("_%v", i)
		colorPicked, wasUnused := randColor.GetOrSetRandomColor(targetName)

		// lets check if we get the same color again for the same task
		colorPicked2, wasUnused2 := randColor.GetOrSetRandomColor(targetName)
		if colorPicked != colorPicked2 {
			t.Error("color picked twice:", colorPicked, "for target:", targetName, "already picked for index:", i)
		}
		if !wasUnused2 {
			t.Error("wasUnused could not be false if we requesting an existing color:", targetName)
		}

		if wasUnused {
			successCount++
			for inx, color := range colors {
				if colorPicked == color {
					t.Error(
						"color picked twice:",
						colorPicked, "for target:",
						targetName, "already picked for index:",
						inx,
					)
				}
			}
		} else {
			failCount++
		}

		colors = append(colors, colorPicked)
	}
	// we accept a 10% fail rate
	// there is always a chance that we hit a color twice
	if failCount > successCount/10 {
		t.Error("fail rate to high:", failCount, "of", randColor.GetMaxVariants())
	}
}

func TestColorPickRandom(t *testing.T) {
	// just test if we ever get an error
	// if we try to pick a color 2000 times.
	// an error would be an go error, because we would hit an index out of range.
	for i := 0; i < 2000; i++ {
		runner.PickRandColor()
	}
}

func TestLastInstance(t *testing.T) {
	randColor := runner.NewRandColorStore()
	if randColor != runner.LastRandColorInstance() {
		t.Error("last instance is not the same as the new instance")
	}
}
