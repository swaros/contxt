package tasks_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/tasks"
)

func TestFitDatahandlerInterface(t *testing.T) {
	fn := func(chk1 tasks.DataMapHandler, chk2 tasks.PlaceHolder) {}

	fn(tasks.NewCombinedDataHandler(), tasks.NewCombinedDataHandler())
}

// TestCombinedDh tests the CombinedDataHandler
func TestCombinedDhBaseFunction(t *testing.T) {

	cdh := tasks.NewCombinedDataHandler()

	// create map[string]interface{} for testing
	testMap := make(map[string]interface{})
	testMap["key1"] = "value1"
	testMap["key2"] = "value2"
	testMap["key3"] = "value3"

	if _, found := cdh.GetData("key1"); found {
		t.Error("key1 should not be present")
	}

	cdh.AddData("key1", testMap)
	if key1Data, found := cdh.GetData("key1"); !found {
		t.Error("key1 not found in data")
	} else if key1Data == nil {
		t.Error("key1 data not found")
	} else {
		if key1Data["key1"] != "value1" {
			t.Error("key1 value not correct")
		}
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

	if jsonStr, ok := cdh.GetDataAsJson("key1"); !ok {
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
	if _, have := cdh.GetDataAsJson("key20"); have {
		t.Error("key20 found")
	}

	// test for using GetDataAsYaml
	if _, have := cdh.GetDataAsYaml("key20"); have {
		t.Error("key20 found")
	}

	// test for using GetDataAsYaml
	if ymlstr, have := cdh.GetDataAsYaml("key1"); !have {
		t.Error("key1 not found")
	} else {
		assert.YAMLEq(t, `name: Martin D'vloper2
foods:
- Apple
- Orange
- Strawberry
languages:
  perl: Elite
  python: Elite
  ruby: Elite
`, ymlstr, "yaml string from key1 value not correct")

	}

}

func TestYamlImport(t *testing.T) {
	cdh := tasks.NewCombinedDataHandler()
	addErr := cdh.AddYaml("key1", `name: Martin D'vloper2
foods:
- Apple
- Orange
- Strawberry
languages:
  perl: Elite
  python: Elite
  ruby: Elite
`)
	if addErr != nil {
		t.Error(addErr)
	}

	if value, found := cdh.GetJSONPathResult("key1", "languages.python"); !found {
		t.Error("key1 not found")
	} else if value.Str != "Elite" {
		t.Error("key1 value not correct")
	}
}

func TestYamlImport2(t *testing.T) {

	yamlFile := `
name: test
path: "root/path"
boolflag: true
subs:
  - first
  - second`

	// create temp dir
	os.MkdirAll("temp", 0777)
	defer os.RemoveAll("temp")
	systools.WriteFileIfNotExists("temp/test.yaml", yamlFile)

	cdh := tasks.NewCombinedDataHandler()
	if err := cdh.ImportDataFromYAMLFile("imported", "temp/test.yaml"); err != nil {
		t.Error(err)
	} else {
		if value, found := cdh.GetJSONPathResult("imported", "name"); !found {
			t.Error("imported not found")
		} else if value.Str != "test" {
			t.Error("imported value not correct")
		}
	}
}

func TestYamlImportInvalidYaml(t *testing.T) {

	yamlFile := `
name: test
   path: "root/path"
boolflag: :: true
subs:
  - first
  - second`

	// create temp dir
	os.MkdirAll("temp", 0777)
	defer os.RemoveAll("temp")
	systools.WriteFileIfNotExists("temp/test2.yaml", yamlFile)

	cdh := tasks.NewCombinedDataHandler()
	if err := cdh.ImportDataFromYAMLFile("fail", "temp/test2.yaml"); err == nil {
		t.Error("this should not work")
	}
}

func TestYamlImportFully(t *testing.T) {
	cdh := tasks.NewCombinedDataHandler()
	addErr := cdh.AddYaml("key1", `name: Martin D'vloper2
foods:
- Apple
- Orange
- Strawberry
languages:
  perl: Hmmm
  python: Overrated
  ruby: Elite
`)
	if addErr != nil {
		t.Error(addErr)
	}

	if value, found := cdh.GetJSONPathResult("key1", "languages.python"); !found {
		t.Error("key1 not found")
	} else if value.Str != "Overrated" {
		t.Error("key1 value not correct")
	}

	shouldBeOverrated := cdh.HandlePlaceHolder("${key1:languages.python} is this: a problem?${key1:languages.perl}")
	assert.Equal(t, "Overrated is this: a problem?Hmmm", shouldBeOverrated)

	shouldBeOverrated = cdh.HandlePlaceHolder("we will say ${key1:languages.python} ... and also ${key1:languages.perl} is overrated")
	assert.Equal(t, "we will say Overrated ... and also Hmmm is overrated", shouldBeOverrated)

	testIter := make(map[string]string)
	testIter["${key1:languages.python}"] = "Overrated"
	testIter["${key1:languages.perl}"] = "Hmmm"
	testIter["${key1:languages.ruby}"] = "Elite"
	testIter["{}is this: a problem?}{"] = "{}is this: a problem?}{"
	testIter["${}is this: a problem?}${"] = "${}is this: a problem?}${"
	testIter["${key1:languages.python} is this: a problem?${key1:languages.perl}"] = "Overrated is this: a problem?Hmmm"
	testIter["we will say ${key1:languages.python} ... and also ${key1:languages.perl} is overrated"] = "we will say Overrated ... and also Hmmm is overrated"
	testIter[" >>> ${key1:something.wrong} <<< "] = " >>> ${key1:something.wrong} <<< "
	testIter[" >>> ${keyX:languages.python} <<< "] = " >>> ${keyX:languages.python} <<< "
	testIter[" >>> ${key1:languages.python} <<< "] = " >>> Overrated <<< "
	testIter[" >>> ${key1:} <<< "] = " >>> ${key1:} <<< "
	testIter[" >>> ${key1} <<< "] = " >>> ${key1} <<< "
	testIter[" >>> ${:} <<< "] = " >>> ${:} <<< "
	testIter[" >>> ${0:} <<< "] = " >>> ${0:} <<< "

	for key, value := range testIter {
		assert.Equal(t, value, cdh.HandlePlaceHolder(key), "failed by testing key: "+key)
	}

}
