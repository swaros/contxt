package cmdhandle_test

import (
	"encoding/json"
	"testing"

	"github.com/swaros/contxt/context/cmdhandle"
)

func TestAddAndGet(t *testing.T) {
	var testData = `{
		"testnumber" : 77,
		"testData":[
			{
				"first": 45,
				"second": 88
			}
		]
	}`
	data := make(map[string]interface{})
	json.Unmarshal([]byte(testData), &data)

	cmdhandle.AddData("test_1", data)

	ok, gotData := cmdhandle.GetData("test_1")
	if !ok {
		t.Error("error by getting data from key test_1.")
	}
	if gotData == nil {
		t.Error("data from text_1 key is nil.")
	}

	numberOne := gotData["testnumber"]
	var checkNumber float64
	checkNumber = 77
	if numberOne != checkNumber {
		t.Error("expected number was 77 but got ", numberOne)
	}

}

func TestGetFileContent(t *testing.T) {
	fileName := "../../docs/test/varimport/test1.json"

	err := cmdhandle.ImportDataFromJSONFile("fileimport", fileName)
	if err != nil {
		t.Error(err)
	}

	_, jsonMap := cmdhandle.GetData("fileimport")

	if jsonMap["msg"] != "hello world" {
		t.Error("msg should be 'hello world'")
	}
}

func TestGetFileYamlContent(t *testing.T) {
	fileName := "../../docs/test/varimport/values.yaml"

	err := cmdhandle.ImportDataFromYAMLFile("yamltest", fileName)
	if err != nil {
		t.Error(err)
	}

	_, jsonMap := cmdhandle.GetData("yamltest")

	if jsonMap["Names"] != "test-run" {
		t.Error("msg should be 'test-run'")
	}

	pathData := cmdhandle.GetJSONPathValueString("yamltest", "data.instance")
	if pathData != "1258" {
		t.Error("unexpected instant data ", pathData)
	}
}
