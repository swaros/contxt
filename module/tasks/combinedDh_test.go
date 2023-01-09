package tasks_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swaros/contxt/module/tasks"
)

// TestCombinedDh tests the CombinedDataHandler
func TestCombinedDhBaseFunction(t *testing.T) {

	cdh := tasks.NewCombinedDataHandler()

	cdh.AddData("key1", "a", "value1")
	cdh.AddData("key2", "a", "value2")
	cdh.AddData("key3", "a", "value3")

	if data, have := cdh.GetDataSub("key1", "a"); !have {
		t.Error("key1 not found")
	} else if data != "value1" {
		t.Error("key1 value not correct. got ", data, " expected value1")
	}

	if data, have := cdh.GetDataSub("key2", "a"); !have {
		t.Error("key2 not found")
	} else if data != "value2" {
		t.Error("key2 value not correct. got ", data, " expected value2")
	}

	if data, have := cdh.GetDataSub("key3", "a"); !have {
		t.Error("key3 not found")
	} else if data != "value3" {
		t.Error("key3 value not correct. got ", data, " expected value3")
	}

	if _, have := cdh.GetDataSub("key4", "a"); have {
		t.Error("key4 found")
	}
}

// Test GsonPathResult
func TestGsonPathResult(t *testing.T) {
	cdh := tasks.NewCombinedDataHandler()
	addErr := cdh.AddJSON("key1", `{"name":"Martin D'vloper","foods":["Apple","Orange","Strawberry"],"languages":{"perl":"Elite","python":"Elite","ruby":"Elite"}}`)
	if addErr != nil {
		t.Error(addErr)
	}

	if key1Data, found := cdh.GetData("key1"); !found {
		t.Error("key1 not found in data")
	} else if key1Data == nil {
		t.Error("key1 data not found")
	}

	if data, have := cdh.GetJSONPathResult("key1", "name"); !have {
		t.Error("key1 not found")
	} else if data.Str != "Martin D'vloper" {
		t.Error("key1 value not correct")
	}

	if ok, jsonStr := cdh.GetDataAsJson("key1"); !ok {
		t.Error("key1 not found")
	} else {
		assert.JSONEq(t,
			`{"name":"Martin D'vloper","foods":["Apple","Orange","Strawberry"],"languages":{"perl":"Elite","python":"Elite","ruby":"Elite"}}`,
			jsonStr, "json string from key1 value not correct")

	}
	// rewrite data by gson path
	cdh.SetJSONValueByPath("key1", "name", "Martin D'vloper2")
	if data, have := cdh.GetJSONPathResult("key1", "name"); !have {
		t.Error("key1 not found")
	} else if data.Str != "Martin D'vloper2" {
		t.Error("key1 value not correct")
	}

	// failtest for getting gson path that does not exist
	if _, have := cdh.GetJSONPathResult("key1", "name2"); have {
		t.Error("key1 - name2 found")
	}

	if _, have := cdh.GetJSONPathResult("keyXX", "name2"); have {
		t.Error("keyXX - name2 found")
	}

	// failtest for using GetDataAsJson
	if have, _ := cdh.GetDataAsJson("key20"); have {
		t.Error("key20 found")
	}

}
