package taskrun_test

import (
	"testing"

	"github.com/swaros/contxt/taskrun"
)

func TestSplitArgsString(t *testing.T) {
	line, args := taskrun.StringSplitArgs("function hello world", "arg")
	if line != "function" {
		t.Error("string shoul be function and not ", line)
	}
	if len(args) != 3 {
		t.Error("unexpected count of args. expected 3", len(args))
	}

	if entr, ok := args["arg0"]; ok {
		if entr != "function" {
			t.Error("index 0 should be function. but not ", entr)
		}
	} else {
		t.Error("expected mp entrie not exists arg0")
	}

	if entr, ok := args["arg1"]; ok {
		if entr != "hello" {
			t.Error("index 1 should be hello. but not ", entr)
		}
	} else {
		t.Error("expected mp entrie not exists arg1")
	}

	if entr, ok := args["arg2"]; ok {
		if entr != "world" {
			t.Error("index 1 should be world. but not ", entr)
		}
	} else {
		t.Error("expected mp entrie not exists arg2")
	}
}

func TestParseArgs(t *testing.T) {
	var args []string = []string{"check abc def", "mama check", "line", "line"}
	result := taskrun.SplitArgs(args, "test", func(s string, m map[string]string) {
		switch s {
		case "check":

			if entr, ok := m["test0"]; ok {
				if entr != "check" {
					t.Error("index 0 should be check. but not ", entr)
				}
			} else {
				t.Error("expected mp entrie not exists test0")
			}

			if entr, ok := m["test1"]; ok {
				if entr != "abc" {
					t.Error("index 1 should be abc. but not ", entr)
				}
			} else {
				t.Error("expected mp entrie not exists test1")
			}

			if entr, ok := m["test2"]; ok {
				if entr != "def" {
					t.Error("index 1 should be def. but not ", entr)
				}
			} else {
				t.Error("expected mp entrie not exists test2")
			}

		case "mama":
			if entr, ok := m["test0"]; ok {
				if entr != "mama" {
					t.Error("index 0 should be mama. but not ", entr)
				}
			} else {
				t.Error("expected mp entrie not exists test0")
			}

			if entr, ok := m["test1"]; ok {
				if entr != "check" {
					t.Error("index 1 should be check. but not ", entr)
				}
			} else {
				t.Error("expected mp entrie not exists test1")
			}
		}
	})

	if len(result) != 4 {
		t.Error("unexpected length of result.", result)
	}

	if result[0] != "check" {
		t.Error("unexpected cleared result for index 0")
	}
	if result[1] != "mama" {
		t.Error("unexpected cleared result for index 1")
	}
	if result[2] != "line" {
		t.Error("unexpected cleared result for index 2")
	}
	if result[3] != "line" {
		t.Error("unexpected cleared result for index 3")
	}
}
