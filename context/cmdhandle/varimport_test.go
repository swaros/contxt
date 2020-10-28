package cmdhandle_test

import (
	"strings"
	"testing"

	"github.com/swaros/contxt/context/cmdhandle"
)

func clearStrings(compare string) string {
	compare = strings.ReplaceAll(compare, " ", "")
	compare = strings.ReplaceAll(compare, "\n", "")
	compare = strings.ReplaceAll(compare, "\t", "")
	return compare
}

func TestParsingFile(t *testing.T) {
	fileName := "../../docs/test/varimport/test1.json"
	jsonMap, err := cmdhandle.ImportJSONFile(fileName)
	if err != nil {
		t.Error("loading file error:", fileName, err)
	}

	if jsonMap["msg"] != "hello world" {
		t.Error("msg should be 'hello world'")
	}
}

func TestParseTemplateFile2(t *testing.T) {
	jsonFileName := "../../docs/test/varimport/test2.json"
	templateFilename := "../../docs/test/varimport/test2.yaml"

	template, err := cmdhandle.ImportFileContent(templateFilename)
	if err != nil {
		t.Error("loading file error:", templateFilename, err)
	}

	jsonMap, jerr := cmdhandle.ImportJSONFile(jsonFileName)
	if jerr != nil {
		t.Error("loading file error:", jsonFileName, jerr)
	}

	result, herr := cmdhandle.HandleJSONMap(template, jsonMap)
	if herr != nil {
		t.Error("error while parsing:[", herr, "]")
	}

	compare := `testcase:
imports:
	testint: thomas`
	// remove all spaces
	compare2 := clearStrings(compare)
	result2 := clearStrings(result)

	if compare2 != result2 {
		t.Error("result is not matching expected result\n", result, "\n", compare)
	}
}

func TestParseTemplateFile3(t *testing.T) {
	ymlFileName := "../../docs/test/varimport/values.yaml"
	templateFilename := "../../docs/test/varimport/contxt_template.yml"

	template, err := cmdhandle.ImportFileContent(templateFilename)
	if err != nil {
		t.Error("loading file error:", templateFilename, err)
	}

	jsonMap, jerr := cmdhandle.ImportYAMLFile(ymlFileName)
	if jerr != nil {
		t.Error("loading file error:", ymlFileName, jerr)
	}

	result, herr := cmdhandle.HandleJSONMap(template, jsonMap)
	if herr != nil {
		t.Error("error while parsing:[", herr, "]")
	}

	compare := `config:import:-"values.yaml"task:-id:testvarsscript:-echotest-run-echo'next'-echoinstance1258`
	// remove all spaces
	compare2 := clearStrings(compare)
	result2 := clearStrings(result)
	if compare2 != result2 {
		t.Error("result is not matching expected result\n", result, "\n", compare, "\n short:[", result2, "]")
	}
}

func TestFolderCheck(t *testing.T) {
	value, err := cmdhandle.ImportFolders("../../docs/test/01template/contxt.yml", "../../docs/test/01values/")
	if err != nil {
		t.Error(err)
	}

	check1 := clearStrings(value)
	compare := `config:version:1.45.06sequencially:truecoloroff:truevariables:checkApi:context.democheckName:awesometask:-id:scriptscript:-echo'hallowelt'-ls-ga-echo'tagout_1valuehello'-echo'tagout_2valueworld'listener:-trigger:onerror:trueonoutcountLess:0onoutcountMore:0onoutContains:-context.json-fatalaction:target:""stopall:falsescript:-echo'triggeredbyoutputparsing'`
	if check1 != compare {
		t.Error("result is not matching with expected\n", check1, "\n", compare)
	}

}
