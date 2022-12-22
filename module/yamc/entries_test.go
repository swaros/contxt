package yamc_test

import (
	"encoding/json"
	"testing"

	"github.com/swaros/contxt/module/yamc"
)

func TestChainFind(t *testing.T) {
	testMap := make(map[string]interface{})

	testMap["test1"] = "hello world"
	if value, err := yamc.FindChain(testMap, "test1"); err != nil {
		t.Error(err)
	} else {
		if value.(string) != "hello world" {
			t.Error("no match")
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
