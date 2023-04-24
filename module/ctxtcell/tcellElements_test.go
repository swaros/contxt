package ctxtcell_test

import (
	"context"
	"testing"
	"time"

	"github.com/swaros/contxt/module/awaitgroup"
	"github.com/swaros/contxt/module/ctxtcell"
)

func TestAddElements(t *testing.T) {
	main := ctxtcell.NewTcell()

	// adding a active Text Element
	activeText := main.ActiveText("I am active")
	if id := main.AddElement(activeText); id != 1 {
		t.Errorf("Expected id to be 1, but got %v", id)
	}

	// adding a passive Text Element
	passiveText := main.Text("I am passive")
	if id := main.AddElement(passiveText); id != 2 {
		t.Errorf("Expected id to be 2, but got %v", id)
	}

	// testing if the elements are in the list
	if len(main.GetSortedElements()) != 2 {
		t.Errorf("Expected 2 elements, but got %v", len(main.GetSortedElements()))
	}

	// ids always increasing even if we remove elements
	main.RemoveElementByID(1)
	if id := main.AddElement(activeText); id != 3 {
		t.Errorf("Expected id to be 3, but got %v", id)
	}
}

func helpAddElementsByCount(main *ctxtcell.CtCell, elem ctxtcell.TcElement, count int) []int {
	ids := make([]int, count)
	for i := 0; i < count; i++ {
		id := main.AddElement(elem)
		ids[i] = id
	}

	return ids
}

func TestAddAndRemoveAsync(t *testing.T) {
	main := ctxtcell.NewTcell()

	// we need to make sure any previous elements are removed
	main.ClearElements()

	// adding a active Text Element
	activeText := main.ActiveText("I am active")
	var addTasks []awaitgroup.FutureStack
	for i := 0; i < 20; i++ {
		addTasks = append(addTasks, awaitgroup.FutureStack{
			AwaitFunc: func(ctx context.Context) interface{} {
				helpAddElementsByCount(main, activeText, 100)
				// wait for 1 millisecond
				time.Sleep(time.Millisecond * 1)
				return 1
			},
			Argument: nil,
		})

	}

	futures := awaitgroup.ExecFutureGroup(addTasks)

	results := awaitgroup.WaitAtGroup(futures)
	// results should have 20 elements
	if len(results) != 20 {
		t.Errorf("Expected 2 results, but got %v", len(results))
	}

	// we should have 2000 elements
	if len(main.GetSortedElements()) != 2000 {
		t.Errorf("Expected 2000 elements, but got %v", len(main.GetSortedElements()))
	}

	// there should not have any gaps in the ids. using GetSortedKeys to get the ids
	ids := main.GetSortedKeys()
	for i := 0; i < len(ids); i++ {
		if ids[i] != i+1 {
			t.Errorf("Expected %v, but got %v", i+1, ids[i])
		}
	}

}
