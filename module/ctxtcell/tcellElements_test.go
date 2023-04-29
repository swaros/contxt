package ctxtcell_test

import (
	"context"
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/swaros/contxt/module/awaitgroup"
	"github.com/swaros/contxt/module/ctxtcell"
)

func TestAddElements(t *testing.T) {
	main := ctxtcell.NewTcell()
	main.ClearElements()
	// adding a active Text Element
	activeText := main.ActiveText("I am active")
	if id, _ := main.AddElement(activeText); id != 1 {
		t.Errorf("Expected id to be 1, but got %v", id)
	}

	// adding a passive Text Element
	passiveText := main.Text("I am passive")
	if id, _ := main.AddElement(passiveText); id != 2 {
		t.Errorf("Expected id to be 2, but got %v", id)
	}

	// testing if the elements are in the list
	if len(main.GetSortedElements()) != 2 {
		t.Errorf("Expected 2 elements, but got %v", len(main.GetSortedElements()))
	}

	// ids always increasing even if we remove elements
	main.RemoveElementByID(1)
	activeText2 := main.ActiveText("I am active")
	if id, _ := main.AddElement(activeText2); id != 3 {
		t.Errorf("Expected id to be 3, but got %v", id)
	}

	// adding a element again should end in an error
	if _, err := main.AddElement(activeText2); err == nil {
		t.Errorf("Expected an error, but got none")
	}
}

func helpAddTextElementsByCount(main *ctxtcell.CtCell, preString string, count int) []int {
	ids := make([]int, count)
	for i := 0; i < count; i++ {
		elem := main.ActiveText(fmt.Sprintf("%v %v", preString, i))
		id, _ := main.AddElement(elem)
		ids[i] = id
	}

	return ids
}

// TestAddAndRemoveAsync tests if we can add and remove elements in parallel
func TestAddAndRemoveAsync(t *testing.T) {
	main := ctxtcell.NewTcell()

	// we need to make sure any previous elements are removed
	main.ClearElements()

	// adding a active Text Element
	var addTasks []awaitgroup.FutureStack
	for i := 0; i < 20; i++ {
		addTasks = append(addTasks, awaitgroup.FutureStack{
			AwaitFunc: func(ctx context.Context) interface{} {
				helpAddTextElementsByCount(main, "activeText", 100)
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

	// now we remove elements between 1000 and 1500
	// and we try this in 20 goroutines for the same ids
	var removeTasks []awaitgroup.FutureStack
	for i := 0; i < 20; i++ {
		removeTasks = append(removeTasks, awaitgroup.FutureStack{
			AwaitFunc: func(ctx context.Context) interface{} {
				rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
				for i := 1000; i < 1500; i++ {
					main.RemoveElementByID(i)
					// wait random time so we can have some concurrency
					min := 10
					max := 30
					tm := rnd.Intn(max-min+1) + min
					time.Sleep(time.Nanosecond * time.Duration(tm))
				}
				return 1
			},
			Argument: nil,
		})

	}

	futures = awaitgroup.ExecFutureGroup(removeTasks)

	results = awaitgroup.WaitAtGroup(futures)
	// results should have 20 elements
	if len(results) != 20 {
		t.Errorf("Expected 2 results, but got %v", len(results))
	}

	// we should have 1500 elements
	if len(main.GetSortedElements()) != 1500 {
		t.Errorf("Expected 1500 elements, but got %v", len(main.GetSortedElements()))
	}

	// there should now have an gap in the ids. so anything between 1000 and 1500 should be missing
	ids = main.GetSortedKeys()
	for i := 0; i < len(ids); i++ {
		if ids[i] >= 1000 && ids[i] < 1500 {
			t.Errorf("Expected %v to be missing, but got %v", ids[i], ids[i])
		}
	}

}

func TestSorted(t *testing.T) {
	main := ctxtcell.NewTcell()
	// we need to make sure any previous elements are removed
	main.ClearElements()

	// adding a active Text Element
	ids := helpAddTextElementsByCount(main, "i am text NO", 10)

	// we should have 10 elements
	if len(main.GetSortedElements()) != 10 {
		t.Errorf("Expected 10 elements, but got %v", len(main.GetSortedElements()))
	}

	// we should have 10 elements
	if len(main.GetSortedKeys()) != len(ids) {
		t.Errorf("Expected %v elements, but got %v", len(ids), len(main.GetSortedKeys()))
	}

	sorted := main.GetSortedKeys()
	for i := 0; i < len(sorted); i++ {
		if sorted[i] != ids[i] {
			t.Errorf("Expected %v, but got %v", ids[i], sorted[i])
		}
	}

	// testing callBack doing the same sorting
	expectedId := 1
	main.SortedCallBack(func(elm ctxtcell.TcElement) bool {
		if elm.GetID() != expectedId {
			t.Errorf("Expected %v, but got %v", expectedId, elm.GetID())
		}
		expectedId++
		return true
	})

}

func TestCycleFocus(t *testing.T) {
	main := ctxtcell.NewTcell()
	// we need to make sure any previous elements are removed
	main.ClearElements()

	// adding a active Text Element

	ids := helpAddTextElementsByCount(main, "activeText", 10)

	// we should have 10 elements
	if len(main.GetSortedElements()) != 10 {
		t.Errorf("Expected 10 elements, but got %v", len(main.GetSortedElements()))
	}

	// we should have 10 elements
	if len(main.GetSortedKeys()) != len(ids) {
		t.Errorf("Expected %v elements, but got %v", len(ids), len(main.GetSortedKeys()))
	}

	focused := main.GetFocusedElement()
	if focused != nil {
		t.Errorf("Expected focused is nil, but got %v", focused)
	}

	main.CycleFocus()
	focused = main.GetFocusedElement()
	if focused == nil {
		t.Errorf("Expected focused is not nil, but got %v", focused)
	} else if focused.GetID() != ids[0] {
		// we expect the first element to be focused
		t.Errorf("Expected focused is %v, but got %v", ids[0], focused.GetID())
	}

	main.CycleFocus()
	focused = main.GetFocusedElement()
	if focused == nil {
		t.Errorf("Expected focused is not nil, but got %v", focused)
	} else if focused.GetID() != ids[1] {
		// we expect the first element to be focused
		t.Errorf("Expected focused is %v, but got %v", ids[1], focused.GetID())
	}
	// focus the last element
	main.SetFocusById(ids[len(ids)-1])

	main.CycleFocus()
	focused = main.GetFocusedElement()
	if focused == nil {
		t.Errorf("Expected focused is not nil, but got %v", focused)
	} else if focused.GetID() != ids[0] {
		// we expect the first element to be focused
		t.Errorf("Expected focused is %v, but got %v", ids[0], focused.GetID())
	}
}
