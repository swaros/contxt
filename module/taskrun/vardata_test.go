package taskrun_test

import (
	"encoding/json"
	"testing"

	"github.com/swaros/contxt/taskrun"
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

	taskrun.AddData("test_1", data)

	ok, gotData := taskrun.GetData("test_1")
	if !ok {
		t.Error("error by getting data from key test_1.")
	}
	if gotData == nil {
		t.Error("data from text_1 key is nil.")
	}

	numberOne := gotData["testnumber"]
	var checkNumber float64 = 77
	if numberOne != checkNumber {
		t.Error("expected number was 77 but got ", numberOne)
	}

}

func TestGetFileContent(t *testing.T) {
	fileName := "../../docs/test/varimport/test1.json"

	err := taskrun.ImportDataFromJSONFile("fileimport", fileName)
	if err != nil {
		t.Error(err)
	}

	_, jsonMap := taskrun.GetData("fileimport")

	if jsonMap["msg"] != "hello world" {
		t.Error("msg should be 'hello world'")
	}
}

func TestGetFileYamlContent(t *testing.T) {
	fileName := "../../docs/test/varimport/values.yaml"

	err := taskrun.ImportDataFromYAMLFile("yamltest", fileName)
	if err != nil {
		t.Error(err)
	}

	_, jsonMap := taskrun.GetData("yamltest")

	if jsonMap["Names"] != "test-run" {
		t.Error("msg should be 'test-run'")
	}

	pathData := taskrun.GetJSONPathValueString("yamltest", "data.instance")
	if pathData != "1258" {
		t.Error("unexpected instant data ", pathData)
	}
}
