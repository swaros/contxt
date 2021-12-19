package configure

import "testing"

func TestGetOs(t *testing.T) {
	versionStr := GetOs()
	if versionStr == "" {
		t.Error("versionstring should not being empty")
	}
}
