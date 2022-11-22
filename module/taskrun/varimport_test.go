package taskrun_test

import (
	"strings"
	"testing"

	"github.com/swaros/contxt/configure"
	"github.com/swaros/contxt/taskrun"
)

func clearStrings(compare string) string {
	compare = strings.ReplaceAll(compare, " ", "")
	compare = strings.ReplaceAll(compare, "\n", "")
	compare = strings.ReplaceAll(compare, "\t", "")
	compare = strings.ReplaceAll(compare, "\r", "")
	return compare
}

func TestParseArgLine(t *testing.T) {
	param1 := `i sayed 'Hello you' and got the response 'fuck you'`
	params, found := taskrun.GetArgQuotedEntries(param1)
	if !found {
		t.Error("nothing found. that should not happens")
	}
	if len(params) > 2 {
		t.Error("unexcpected amout if entries", params)
	}

	allFounds := taskrun.SplitQuoted(param1, " ")
	if len(allFounds) != 8 {
		t.Error("unexpected amount of strings ", len(allFounds))
	}

}

func TestParsingFile(t *testing.T) {
	fileName := "../../docs/test/varimport/test1.json"
	jsonMap, err := taskrun.ImportJSONFile(fileName)
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

	template, err := taskrun.ImportFileContent(templateFilename)
	if err != nil {
		t.Error("loading file error:", templateFilename, err)
	}

	jsonMap, jerr := taskrun.ImportJSONFile(jsonFileName)
	if jerr != nil {
		t.Error("loading file error:", jsonFileName, jerr)
	}

	result, herr := taskrun.HandleJSONMap(template, jsonMap)
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

	template, err := taskrun.ImportFileContent(templateFilename)
	if err != nil {
		t.Error("loading file error:", templateFilename, err)
	}

	jsonMap, jerr := taskrun.ImportYAMLFile(ymlFileName)
	if jerr != nil {
		t.Error("loading file error:", ymlFileName, jerr)
	}

	result, herr := taskrun.HandleJSONMap(template, jsonMap)
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
	value, err := taskrun.ImportFolders("../../docs/test/01template/contxt.yml", "../../docs/test/01values/")
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
	err := taskrun.ImportDataFromYAMLFile("test1", "../../docs/test/foreach/importFile.yaml")
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
	_, _, newScript := taskrun.TryParse(script, func(line string) (bool, int) {
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
		return false, taskrun.ExitOk
	})

	if len(newScript) < 1 {
		t.Error("generated script should not be empty")
	}
	if len(newScript) != 5 {
		t.Error("unexpected result length ", len(newScript))
	}

}

func TestTryParseError(t *testing.T) {
	err := taskrun.ImportDataFromYAMLFile("test1", "../../docs/test/foreach/importFile.yaml")
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
	rAbort, rCode, newScript := taskrun.TryParse(script, func(line string) (bool, int) {
		hitCounter++
		fails := false
		abort := false
		returnCode := taskrun.ExitOk
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
			returnCode = taskrun.ExitByStopReason
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

	if rCode != taskrun.ExitByStopReason {
		t.Error("unexpected return code")
	}

}

func TestTryParseJsonImport(t *testing.T) {
	var script []string
	script = append(script, "#@import-json json-data {\"hello\":\"world\"}")

	taskrun.TryParse(script, func(line string) (bool, int) {
		return false, taskrun.ExitOk
	})

	have, data := taskrun.GetData("json-data")
	if have == false {
		t.Error("json was not imported")
	}

	if data["hello"] != "world" {
		t.Error("import was not working as expected")
	}

}

func TestTryParseJsonImportByExec(t *testing.T) {

	var script []string
	if configure.GetOs() == "windows" {
		return // skip on windows
	} else {
		script = append(script, "#@import-json-exec exec-import-data cat ../../docs/test/foreach/forcat.json")
	}

	taskrun.TryParse(script, func(line string) (bool, int) {
		return false, taskrun.ExitOk
	})

	have, data := taskrun.GetData("exec-import-data")
	if have == false {
		t.Error("json was not imported")
	}

	if data["main"] == nil {
		t.Error("import was not working as expected")
	}

}

func TestTryParseVar(t *testing.T) {
	if configure.GetOs() == "windows" {
		return // skip on windows
	}
	var script []string
	taskrun.SetPH("test-var-out", "first")
	script = append(script, "#@var check-replace-out echo test-${test-var-out}-case")

	taskrun.TryParse(script, func(line string) (bool, int) {
		return false, taskrun.ExitOk
	})

	teststr := taskrun.GetPH("check-replace-out")
	if teststr != "test-first-case" {
		t.Error("set var by command is not working. got [", teststr, "]")
	}
}

func TestSetVar(t *testing.T) {
	var script []string
	script = append(script, "#@set test-var-set hello")
	taskrun.TryParse(script, func(line string) (bool, int) {
		return false, taskrun.ExitOk
	})
	teststr := taskrun.GetPH("test-var-set")
	if teststr != "hello" {
		t.Error("set var by command is not working. got [", teststr, "]")
	}
}

func TestVariablesSet(t *testing.T) {
	fileName := "docs/test/03values/values.yml"
	err := taskrun.ImportDataFromYAMLFile("case_a", "../../"+fileName)
	if err != nil {
		t.Error(err)
	}

	// create script section
	var script []string
	script = append(script, "#@set-in-map case_a root.first.name rudolf")
	taskrun.TryParse(script, func(line string) (bool, int) {
		return false, taskrun.ExitOk
	})

	teststr := taskrun.GetJSONPathValueString("case_a", "root.first.name")
	if teststr != "rudolf" {
		t.Error("replace var by path is not working. got [", teststr, "]")
	}

}

/*
func TestTryParseWithKeys(t *testing.T) {
	err := taskrun.ImportDataFromYAMLFile("test-with-key", "../../docs/test/foreach/importFile.yaml")
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
	_, _, newScript := taskrun.TryParse(script, func(line string) (bool, int) {
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
		return false, taskrun.ExitOk
	})

	if len(newScript) < 1 {
		t.Error("generated script should not be empty")
	}
	if len(newScript) != 5 {
		t.Error("unexpected result length ", len(newScript))
	}

}
*/
