package taskrun_test

import (
	"strings"
	"testing"

	"github.com/swaros/contxt/module/configure"
	"github.com/swaros/contxt/module/systools"
	"github.com/swaros/contxt/module/taskrun"
)

func clearStrings(compare string) string {
	compare = strings.ReplaceAll(compare, " ", "")
	compare = strings.ReplaceAll(compare, "\n", "")
	compare = strings.ReplaceAll(compare, "\t", "")
	compare = strings.ReplaceAll(compare, "\r", "")
	return compare
}

func TestIssue82(t *testing.T) {
	jsonFileName := "../../docs/test/issue82/docker-inspect.json"
	data, jerr := taskrun.ImportJSONFile(jsonFileName)
	if jerr != nil {
		t.Error(jerr)
	}

	if _, ok := data["0"]; !ok {
		t.Error("could not get the key")
	}

	jsonFileName = "../../docs/test/issue82/mapped-inspect.json"
	data, jerr = taskrun.ImportJSONFile(jsonFileName)
	if jerr != nil {
		t.Error(jerr)
	}

	if _, ok := data["image"]; !ok {
		t.Error("could not get the key")
	}
}

type testingGsonKey struct {
	varName  string
	path     string
	expected string
}

func TestIssue82Usage(t *testing.T) {
	assertTestFolderVarFn(t, "issue82", "issue82", func() {
		assertVarStrNotEquals(t, "IMAGE-INFO", "")

		var alltests []testingGsonKey = []testingGsonKey{
			{
				varName:  "json_b",
				path:     "image.0.Id",
				expected: "sha256:14d13f9624cdefd69d648dec8ec436a08dd50004530199ae8f1a2b88c36755d6",
			},
			{
				varName:  "json_b",
				path:     "image.0.RepoTags.0",
				expected: "golib-local:latest",
			},
			{
				varName:  "json_a",
				path:     "0.Id",
				expected: "sha256:14d13f9624cdefd69d648dec8ec436a08dd50004530199ae8f1a2b88c36755d6",
			},
			{
				varName:  "json_a",
				path:     "0.Config.WorkingDir",
				expected: "/usr/src/golibdb",
			},
			{
				varName:  "json_a",
				path:     "0.Config.Entrypoint.0",
				expected: "docker-php-entrypoint",
			},
		}

		for _, testCase := range alltests {

			eqCheckVarMapValue(
				testCase.varName,
				testCase.path,
				testCase.expected,
				func(result string) { /* nothing to do ..we found it */ },
				func(result string) {
					t.Error("not found expected key: ", testCase.varName, testCase.path, testCase.expected, " -> in -> ", result)
				},
				func(result string) { t.Error("the while data on  not exists. check key ", testCase.varName) },
			)

		}

	})
}

func TestIssue82_2(t *testing.T) {
	jsonFileName := "../../docs/test/issue82/mapped-inspect.json"
	_, jerr := taskrun.ImportJSONFile(jsonFileName)
	if jerr != nil {
		t.Error(jerr)
	}

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
		return false, systools.ExitOk
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
		returnCode := systools.ExitOk
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
			returnCode = systools.ExitByStopReason
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

	if rCode != systools.ExitByStopReason {
		t.Error("unexpected return code")
	}

}

func TestTryParseJsonImport(t *testing.T) {
	var script []string
	script = append(script, "#@import-json json-data {\"hello\":\"world\"}")

	taskrun.TryParse(script, func(line string) (bool, int) {
		return false, systools.ExitOk
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
		return false, systools.ExitOk
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
		return false, systools.ExitOk
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
		return false, systools.ExitOk
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
		return false, systools.ExitOk
	})

	teststr := taskrun.GetJSONPathValueString("case_a", "root.first.name")
	if teststr != "rudolf" {
		t.Error("replace var by path is not working. got [", teststr, "]")
	}

}

func TestExportAsYaml(t *testing.T) {
	fileName := "docs/test/03values/values.yml"
	err := taskrun.ImportDataFromYAMLFile("case_a", "../../"+fileName)
	if err != nil {
		t.Error(err)
	}
	// create script section
	var script []string
	script = append(script, "#@export-to-yaml case_a yamlstring")
	taskrun.TryParse(script, func(line string) (bool, int) {
		return false, systools.ExitOk
	})
	expect := `
	root:
		first:
		  name: charly
	`
	assertVarStrEquals(t, "yamlstring", expect)
}

func TestDynamicVarReplace(t *testing.T) {
	if runError := folderRunner("./../../docs/test/03values", t, func(t *testing.T) {
		// note. they keys are sorted in root. so version must be at the end
		expectedYaml := `
		services:
		   website:
			  host: myhost
			  port: 8080
		version: 5
	`

		taskrun.RunTargets("main", false)
		assertVarStrEquals(t, "host", "myhost")
		assertVarStrEquals(t, "A", "myhost:8080")
		assertVarStrEquals(t, "B", "myhost")
		assertVarStrEquals(t, "C", expectedYaml)

	}); runError != nil {
		t.Error(runError)
	}
}

func TestImportFileAsTemplae(t *testing.T) {
	if err := folderRunner("./../../docs/test/01tplImport", t, func(t *testing.T) {
		taskrun.RunTargets("tpl-test", false)
		mapData := taskrun.GetOriginMap()
		if mapData == nil {
			t.Error("task data should be exists")
		}
		expectedYaml := `
		testing:
           checks:
               check-out-ABC: set
               check-out-CHECK: set
               check-out-something: set
`
		assertVarStrEquals(t, "OUT-TO-YAML", expectedYaml)
	}); err != nil {
		t.Error(err)
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
