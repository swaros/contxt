package ctxtcell_test

import (
	"testing"

	"github.com/swaros/contxt/module/ctxout"
	"github.com/swaros/contxt/module/ctxtcell"
)

func TestStreamUsage(t *testing.T) {
	outStr := ctxout.ToString(ctxtcell.NewCtOutputNoTty(), "hello", "test")
	if outStr != "hellotest" {
		t.Errorf("Expected 'hellotest', got '%v'", outStr)
	}
}
