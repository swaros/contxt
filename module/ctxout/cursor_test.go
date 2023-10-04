package ctxout_test

import (
	"testing"

	"github.com/swaros/contxt/module/ctxout"
)

func TestCommand(t *testing.T) {
	cursor := ctxout.NewCursorFilter()

	result := cursor.Command("cursor:down,1;")
	if result != "" {
		t.Error("Expected empty string. got ", result)
	}

	result = cursor.Command("cursor:down,1;hello world")
	if result != "hello world" {
		t.Error("Expected hello world. got [", result, "]")
	}

}

func TestIsCursor(t *testing.T) {
	cursor := ctxout.NewCursorFilter()

	result := cursor.IsCursor("cursor:down,1;")
	if !result {
		t.Error("Expected true. got ", result)
	}

	result = cursor.IsCursor("cursor:down,1;hello world")
	if !result {
		t.Error("Expected true. got ", result)
	}
}

func TestCanHandleThis(t *testing.T) {
	cursor := ctxout.NewCursorFilter()

	result := cursor.CanHandleThis("cursor:down,1;")
	if !result {
		t.Error("Expected true. got ", result)
	}

	result = cursor.CanHandleThis("cursor:down,1;hello world")
	if !result {
		t.Error("Expected true. got ", result)
	}
}

func TestFilter(t *testing.T) {
	cursor := ctxout.NewCursorFilter()
	ctxout.AddPostFilter(cursor)

	result := ctxout.ToString("cursor:down,1;")
	if result != "" {
		t.Error("Expected empty string. got ", result)
	}

	result = ctxout.ToString("cursor:down,1;hello world")
	if result != "hello world" {
		t.Error("Expected hello world. got [", result, "]")
	}

	result = ctxout.ToString("cursor:up,1;", "hello world")
	if result != "hello world" {
		t.Error("Expected hello world. got [", result, "]")
	}
}

func TestInvalidCommands(t *testing.T) {
	ctxout.ClearPostFilters()
	cursor := ctxout.NewCursorFilter()
	ctxout.AddPostFilter(cursor)

	result := ctxout.ToString("cursor:down;hello world")
	expected := "hello world"
	if result != expected {
		t.Error("Expected ", expected, " got ", result)
	}

	if cursor.LastError == nil {
		t.Error("Expected error. got ", cursor.LastError)
	} else if cursor.LastError.Error() != "invalid number of params. expected 1 got 0" {
		t.Error("Expected error too few arguments. got ", cursor.LastError)
	}
	// reset error
	cursor.LastError = nil

	result = ctxout.ToString("cursor:down,1,2;hello world")
	expected = "hello world"
	if result != expected {
		t.Error("Expected ", expected, " got ", result)
	}

	if cursor.LastError == nil {
		t.Error("Expected error. got ", cursor.LastError)
	} else if cursor.LastError.Error() != "invalid number of params. expected 1 got 2" {
		t.Error("Expected error too many arguments. got ", cursor.LastError)
	}
}
