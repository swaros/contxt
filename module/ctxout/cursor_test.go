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
