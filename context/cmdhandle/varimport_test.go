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
	compare = strings.ReplaceAll(compare, "\r", "")
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

func TestTryParse(t *testing.T) {
	err := cmdhandle.ImportDataFromYAMLFile("test1", "../../docs/test/foreach/importFile.yaml")
	if err != nil {
		t.Error(err)
	}
	var script []string
	script = append(script, "not changed")
	script = append(script, "#@foreach test1 testData.section.simple")
	script = append(script, "#@- output:[__LINE__]")
	script = append(script, "#@end")
	script = append(script, "not changed too")
	hitCounter := 0
	_, _, newScript := cmdhandle.TryParse(script, func(line string) (bool, int) {
		hitCounter++
		fails := false
		switch hitCounter {
		case 1:
			if line != "not changed" {
				fails = true
			}
		case 2:
			if line != "output:[firstLine]" {
				fails = true
			}
		case 3:
			if line != "output:[secondLine]" {
				fails = true
			}
		case 4:
			if line != "output:[thirdLine]" {
				fails = true
			}
		case 5:
			if line != "not changed too" {
				fails = true
			}
		}
		if fails {
			t.Error("failing because line", hitCounter, "have unexpected content [", line, "]")
		}
		return false, cmdhandle.ExitOk
	})

	if len(newScript) < 1 {
		t.Error("generated script should not be empty")
	}
	if len(newScript) != 5 {
		t.Error("unexpected result length ", len(newScript))
	}

}

func TestTryParseError(t *testing.T) {
	err := cmdhandle.ImportDataFromYAMLFile("test1", "../../docs/test/foreach/importFile.yaml")
	if err != nil {
		t.Error(err)
	}
	var script []string
	script = append(script, "not changed")
	script = append(script, "#@foreach test1 testData.section.simple")
	script = append(script, "#@- output:[__LINE__]")
	script = append(script, "#@end")
	script = append(script, "not changed too")
	hitCounter := 0
	rAbort, rCode, newScript := cmdhandle.TryParse(script, func(line string) (bool, int) {
		hitCounter++
		fails := false
		abort := false
		returnCode := cmdhandle.ExitOk
		switch hitCounter {
		case 1:
			if line != "not changed" {
				fails = true
			}
		case 2:
			if line != "output:[firstLine]" {
				fails = true
			}
		case 3:
			if line != "output:[secondLine]" {
				fails = true
			}
			abort = true
			returnCode = cmdhandle.ExitByStopReason
		}
		if fails {
			t.Error("failing because line", hitCounter, "have unexpected content [", line, "]")
		}
		return abort, returnCode
	})

	if len(newScript) < 1 {
		t.Error("generated script should not be empty")
	}
	if len(newScript) != 3 {
		t.Error("unexpected result length ", len(newScript))
	}

	if rAbort == false {
		t.Error("unexpected abort result ")
	}

	if rCode != cmdhandle.ExitByStopReason {
		t.Error("unexpected return code")
	}

}

func TestTryParseJsonImport(t *testing.T) {
	var script []string
	script = append(script, "#@import-json json-data {\"hello\":\"world\"}")

	cmdhandle.TryParse(script, func(line string) (bool, int) {
		return false, cmdhandle.ExitOk
	})

	have, data := cmdhandle.GetData("json-data")
	if have == false {
		t.Error("json was not imported")
	}

	if data["hello"] != "world" {
		t.Error("import was not working as expected")
	}

}

func TestTryParseJsonImportByExec(t *testing.T) {

	var script []string
	script = append(script, "#@import-json-exec exec-import-data cat ../../docs/test/foreach/forcat.json")

	cmdhandle.TryParse(script, func(line string) (bool, int) {
		return false, cmdhandle.ExitOk
	})

	have, data := cmdhandle.GetData("exec-import-data")
	if have == false {
		t.Error("json was not imported")
	}

	if data["main"] == nil {
		t.Error("import was not working as expected")
	}

}

func TestTryParseVar(t *testing.T) {

	var script []string
	cmdhandle.SetPH("test-var-out", "first")
	script = append(script, "#@var check-replace-out echo test-${test-var-out}-case")

	cmdhandle.TryParse(script, func(line string) (bool, int) {
		return false, cmdhandle.ExitOk
	})

	teststr := cmdhandle.GetPH("check-replace-out")
	if teststr != "test-first-case" {
		t.Error("set var by command is not working. got [", teststr, "]")
	}
}

func TestSetVar(t *testing.T) {
	var script []string
	script = append(script, "#@set test-var-set hello")
	cmdhandle.TryParse(script, func(line string) (bool, int) {
		return false, cmdhandle.ExitOk
	})
	teststr := cmdhandle.GetPH("test-var-set")
	if teststr != "hello" {
		t.Error("set var by command is not working. got [", teststr, "]")
	}
}

/*
func TestTryParseWithKeys(t *testing.T) {
	err := cmdhandle.ImportDataFromYAMLFile("test-with-key", "../../docs/test/foreach/importFile.yaml")
	if err != nil {
		t.Error(err)
	}
	var script []string
	script = append(script, "not changed")
	script = append(script, "#@foreach test-with-key testData.section.import.keyValue")
	script = append(script, "#@- output:[__LINE__]")
	script = append(script, "#@end")
	script = append(script, "not changed too")
	hitCounter := 0
	_, _, newScript := cmdhandle.TryParse(script, func(line string) (bool, int) {
		hitCounter++
		fails := false
		switch hitCounter {
		case 1:
			if line != "not changed" {
				fails = true
			}
		case 2:
			if line != "output:[firstLine]" {
				fails = true
			}
		case 3:
			if line != "output:[secondLine]" {
				fails = true
			}
		case 4:
			if line != "output:[thirdLine]" {
				fails = true
			}
		case 5:
			if line != "not changed too" {
				fails = true
			}
		}
		if fails {
			t.Error("failing because line", hitCounter, "have unexpected content [", line, "]")
		}
		return false, cmdhandle.ExitOk
	})

	if len(newScript) < 1 {
		t.Error("generated script should not be empty")
	}
	if len(newScript) != 5 {
		t.Error("unexpected result length ", len(newScript))
	}

}
*/
