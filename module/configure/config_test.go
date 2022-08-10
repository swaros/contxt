package configure

import (
	"fmt"
	"testing"
)

func TestAddPath(t *testing.T) {
	UsedConfig.Paths = nil
	origin := len(UsedConfig.Paths)
	fmt.Println(origin)
	if origin != 0 {
		t.Error("there should no paths stored. but have: ", origin)
	}

	AddPath("testpath")
	newCnt := len(UsedConfig.Paths)
	if newCnt != 1 {
		t.Error("ther should be one path added. but have: ", newCnt)
	}

	for _, path := range UsedConfig.Paths {
		if path != "testpath" {
			t.Error("ther should only 'testpath in, but got ': ", path)
		}
	}

	removedFail := RemovePath("xxx")
	if removedFail != false {
		t.Error("path xxx is not added so it can not be removed", UsedConfig.Paths)
	}

	removal := RemovePath("testpath")
	if removal == false {
		t.Error("path should removed successfully", UsedConfig.Paths)
	}

	cnt := len(UsedConfig.Paths)
	if cnt != 0 {
		t.Error("there should no paths stored. but have: ", origin, UsedConfig.Paths)
	}

}
