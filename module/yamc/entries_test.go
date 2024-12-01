package yamc_test

import (
	"encoding/json"
	"testing"

	"github.com/swaros/contxt/module/yamc"
)

func TestChainFind(t *testing.T) {
	testMap := make(map[string]interface{})

	testMap["test1"] = "hello world"
	if _, err := yamc.FindChain(nil, "test1"); err != nil {
		if err.Error() != "nil can not being parsed" {
			t.Error("Error is expected (nil can not being parsed). but no this error.", err)
		}
	} else {
		t.Error("this should fail")
	}
}

func TestChainFindError(t *testing.T) {
	testMap := make(map[string]interface{})

	testMap["test1"] = "hello world"
	if _, err := yamc.FindChain(testMap, "test2"); err == nil {
		t.Error("this should fail")
	}
}

func TestChainWrongType(t *testing.T) {

	b := 0.01
	if _, err := yamc.FindChain(b, "test2"); err == nil {
		t.Error("this should fail")
	} else {
		if err.Error() != "unsupported type float64" {
			t.Error("Error is expected (unsupported type float64). but no this error.", err)
		}
	}
}

func TestChainFailUnsupportedValue(t *testing.T) {
	nodes := map[string]interface{}{
		"node1": "a value",
		"node3": struct {
			source string
			price  float64
		}{"chicken", 1.75},
		"steak": true,
	}

	if value, err := yamc.FindChain(nodes, "node1"); err != nil {
		t.Error(err)
	} else {
		if value.(string) != "a value" {
			t.Error("no match")
		}
	}
	// the node3 contains a struct what we can (and will) not handle
	if _, err := yamc.FindChain(nodes, "node3"); err == nil {
		t.Error("this should fail because of struct handling")
	}
}

func TestFindOutOfRange(t *testing.T) {
	nodes := []interface{}{
		"a value",
		"a value",
		"a value",
	}

	if _, err := yamc.FindChain(nodes, "3"); err == nil {
		t.Error("this should fail because of out of range")
	} else {
		expectedError := "index 3 is out of range. have  3 entries in []interface {}. starting at 0 max index is 2"
		if err.Error() != expectedError {
			t.Errorf("expected error [%s] but got [%s]", expectedError, err.Error())
		}
	}
}

func TestChainFailUnsupportedValue2(t *testing.T) {
	nodes := []interface{}{
		"a value",
		struct {
			source string
			price  float64
		}{"chicken", 1.75},
		true,
	}

	if _, err := yamc.FindChain(nodes, "node1"); err == nil {
		t.Error("this should fail because of struct handling")
	}
}

func TestChainFailUnsupportedValue2_2(t *testing.T) {
	nodes := []interface{}{
		"a value",
		[]interface{}{
			struct {
				source string
				price  float64
			}{"chicken", 1.75},
		},
		true,
	}

	if _, err := yamc.FindChain(nodes, "1"); err == nil {
		t.Error("this should fail because of struct handling")
	} else {
		expectedError := "[]interface{} unsupported value type []interface {}"
		if err.Error() != expectedError {
			t.Errorf("expected error [%s] but got [%s]", expectedError, err.Error())
		}
	}
}

func TestChainFailUnsupportedValue3(t *testing.T) {
	nodes := []interface{}{
		"a value",
		map[string]interface{}{
			"source": "chicken",
			"price":  1.75,
			"struct": struct {
				source string
				price  float64
			}{"chicken", 1.75},
		},
		true,
	}

	if _, err := yamc.FindChain(nodes, ""); err == nil {
		t.Error("this should fail because of struct handling")
	} else {
		expectedError := "strconv.Atoi: parsing \"\": invalid syntax"
		if err.Error() != expectedError {
			t.Errorf("expected error %s but got %s", expectedError, err.Error())
		}
	}
}

func TestChainFailUnsupportedValue4(t *testing.T) {
	nodes := map[any]interface{}{
		"a value": "a value",
		"node3": struct {
			source string
			price  float64
		}{"chicken", 1.75},
		"steak": true,
	}

	if val, err := yamc.FindChain(nodes, "a value"); err != nil {
		t.Error(err)
	} else {
		if val.(string) != "a value" {
			t.Error("no match")
		}
	}

	if _, err := yamc.FindChain(nodes, "1"); err == nil {
		t.Error("this should fail because of struct handling")
	} else {
		expectedError := "the map do not contains the key 1 in <nil>"
		if err.Error() != expectedError {
			t.Errorf("expected error [%s] but got [%s]", expectedError, err.Error())
		}
	}
}

func TestChainFailUnsupportedValue5(t *testing.T) {
	nodes := map[any]interface{}{
		"a value": "a value",
		"node3": struct {
			source string
			price  float64
		}{"chicken", 1.75},
		"steak": true,
		"deeper": map[any]interface{}{
			"node1": "a value",
			"node3": struct {
				source string
				price  float64
			}{"chicken", 1.75},
			"steak": true,
		},
	}

	if val, err := yamc.FindChain(nodes, "a value"); err != nil {
		t.Error(err)
	} else {
		if val.(string) != "a value" {
			t.Error("no match")
		}
	}

	if val, err := yamc.FindChain(nodes, "deeper", "node1"); err != nil {
		t.Error(err)
	} else {
		if val.(string) != "a value" {
			t.Error("no match")
		}
	}

	if _, err := yamc.FindChain(nodes, "node3"); err == nil {
		t.Error("this should fail because of struct handling")
	} else {
		expectedError := "map[interface{}]interface{} unsupported value type struct { source string; price float64 }"
		if err.Error() != expectedError {
			t.Errorf("expected error [%s] but got [%s]", expectedError, err.Error())
		}
	}
}
func TestChainFindDeep(t *testing.T) {
	jsonStr := `{
	"node1":"John",
	"node":{
		"sub1": {
			"ent1" : "yes we can"
		}
	},
	"animals": [
		{
			"lion" : {
				"hunter": "yes"
		}
	}
	],	
	"hobbies":[
	   "martial arts",
	   "breakfast foods",
	   "piano"
	]
 }`
	var typed map[string]interface{} = make(map[string]interface{})
	if err := json.Unmarshal([]byte(jsonStr), &typed); err != nil {
		t.Error(err)
	}

	if value, err := yamc.FindChain(typed, "node1"); err != nil {
		t.Error(err)
	} else {
		if value.(string) != "John" {
			t.Error("no match")
		}
	}

	if value, err := yamc.FindChain(typed, "node", "sub1", "ent1"); err != nil {
		t.Error(err)
	} else {
		if value.(string) != "yes we can" {
			t.Error("no match")
		}
	}

	if value, err := yamc.FindChain(typed, "hobbies", "1"); err != nil {
		t.Error(err)
	} else {
		if value.(string) != "breakfast foods" {
			t.Error("no match")
		}
	}

}
