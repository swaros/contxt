package yamc_test

import (
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"testing"

	"github.com/swaros/contxt/module/yamc"
)

func helpFileLoad(filename string, dataHdl func(data []byte)) error {

	if data, err := os.ReadFile(filename); err == nil {
		dataHdl(data)
	} else {
		return err
	}
	return nil
}

// Assert helper to handle equal check.
func CallBackAssertPath(ymap *yamc.Yamc, path string, expected any, ifNotEquals func(val any), ifErr func(error)) {
	value, err := ymap.FindValue(path)
	if err != nil {
		ifErr(err)
	}
	if value != expected {
		ifNotEquals(value)
	}
}

// LazyAssertPath wraps CallBackAssertPath for fast testing values by path.
// we loosing the context here, becasue any triggered error will have this funtion as source
func LazyAssertPath(t *testing.T, ymap *yamc.Yamc, path string, expected any) {
	t.Helper()
	CallBackAssertPath(ymap, path, expected, func(val any) {
		_, fnmane, lineNo, _ := runtime.Caller(3)
		errmsgWithLine := "ERROR: " + fnmane + ":" + strconv.Itoa(lineNo)
		if reflect.TypeOf(val) != reflect.TypeOf(expected) {

			t.Error("ERROR: ", fnmane+":"+strconv.Itoa(lineNo), " types not equal. we got ", reflect.TypeOf(val), " we expect ", reflect.TypeOf(expected))
		}
		t.Error(errmsgWithLine, "expected the value (", expected, ") got [", val, "] instead")
	}, func(err error) {
		_, fnmane, lineNo, _ := runtime.Caller(3)
		t.Error("ERROR: ", fnmane+":"+strconv.Itoa(lineNo), err)
	})
}

// Testing simple Parsing of json content
func TestJsonParse(t *testing.T) {
	if err := helpFileLoad("testdata/test001.json", func(data []byte) {

		// init reader
		jsonReader := yamc.NewJsonReader()

		// init yamc
		conv := yamc.New()

		// handle the data
		if err := conv.Parse(jsonReader, data); err != nil {
			t.Error("error after get data", err)
		} else {
			// test the basics
			if !conv.IsLoaded() {
				t.Error("isLoaded should be true")
			}
			// source data was in form of []interface{}
			if conv.GetSourceDataType() != yamc.TYPE_ARRAY {
				t.Error("reported type should be array")
			}

			LazyAssertPath(t, conv, "0.id", float64(1))
			LazyAssertPath(t, conv, "0.first_name", "Jeanette")
			LazyAssertPath(t, conv, "2.first_name", "Noell")
			LazyAssertPath(t, conv, "3.email", `wvalek3@vk.com`)
			LazyAssertPath(t, conv, "2.id", float64(3))
		}

	}); err != nil {
		t.Error(err)
	}
}

func Test002(t *testing.T) {
	if err := helpFileLoad("testdata/test002.json", func(data []byte) {
		conv := yamc.New()
		if err := conv.Parse(yamc.NewJsonReader(), data); err != nil {
			t.Error(err)
		} else {
			if !conv.IsLoaded() {
				t.Error("isLoaded should be true")
			}
			// source data was in form of map[string]interface{}
			if conv.GetSourceDataType() != yamc.TYPE_STRING_MAP {
				t.Error("reported type should be a string map")

			}
			LazyAssertPath(t, conv, "_id", "5973782bdb9a930533b05cb2")
			LazyAssertPath(t, conv, "isActive", true)
			LazyAssertPath(t, conv, "age", float64(32))
			LazyAssertPath(t, conv, "friends.1.id", float64(1))
			LazyAssertPath(t, conv, "friends.2.name", "Carol Martin")
		}
	}); err != nil {
		t.Error(err)
	}
}

func TestOfficialYaml(t *testing.T) {
	if err := helpFileLoad("testdata/official.yaml", func(data []byte) {
		conv := yamc.New()
		if err := conv.Parse(yamc.NewYamlReader(), data); err != nil {
			t.Error(err)
		} else {
			if !conv.IsLoaded() {
				t.Error("isLoaded should be true")
			}
			// source data was in form of map[string]interface{}
			if conv.GetSourceDataType() != yamc.TYPE_STRING_MAP {
				t.Error("reported type should be a string map")
			}

			CallBackAssertPath(conv, "YAML", "YAML Ain't Markup Language™", func(val any) {
				t.Error("value test. got [", val, "] instead of expected")
			}, func(err error) {
				t.Error(err)
			})

			LazyAssertPath(t, conv, "YAML", "YAML Ain't Markup Language™")
			LazyAssertPath(t, conv, "YAML Resources.YAML Specifications.1", "YAML 1.1")
		}
	}); err != nil {
		t.Error(err)
	}
}

func Test003Yaml(t *testing.T) {
	if err := helpFileLoad("testdata/test003.yml", func(data []byte) {
		conv := yamc.New()
		if err := conv.Parse(yamc.NewYamlReader(), data); err != nil {
			t.Error(err)
		} else {
			if !conv.IsLoaded() {
				t.Error("isLoaded should be true")
			}
			// source data was in form of map[string]interface{}
			if conv.GetSourceDataType() != yamc.TYPE_STRING_MAP {
				t.Error("reported type should be a string map")
			}

			LazyAssertPath(t, conv, "name", "Martin D'vloper")
			LazyAssertPath(t, conv, "foods.2", "Strawberry")
			LazyAssertPath(t, conv, "languages.perl", "Elite")
		}
	}); err != nil {
		t.Error(err)
	}
}

func TestJsonInvalid(t *testing.T) {
	data := []byte("[{hello}}]")
	conv := yamc.New()
	if err := conv.Parse(yamc.NewJsonReader(), data); err == nil {
		t.Error("this reading should fail")
	}

}

func TestYamlInvalid(t *testing.T) {
	data := []byte("[uhm]-")
	conv := yamc.New()
	if err := conv.Parse(yamc.NewYamlReader(), data); err == nil {
		t.Error("this reading should fail")
	}

}

func TestJsonYamlToString(t *testing.T) {
	data := []byte(`{"master": 45}`)
	conv := yamc.New()
	if err := conv.Parse(yamc.NewJsonReader(), data); err != nil {
		t.Error("this reading should not fail")
	} else {
		if str, err2 := conv.ToString(yamc.NewYamlReader()); err2 != nil {
			t.Error(err2)
		} else {
			// we do not test the string content because of different line endings on windows
			if str == "" || !strings.Contains(str, "master:") {
				t.Error("empty string?, or master: key missing?", str)
			}
			if _, ok := conv.GetData()["master"]; !ok {
				t.Error("we should have the master node")
			}
		}
	}

}

func TestYamlToJsonString(t *testing.T) {
	yaml := `
hello:
   - world
   - you
`
	data := []byte(yaml)
	conv := yamc.New()
	if err := conv.Parse(yamc.NewYamlReader(), data); err != nil {
		t.Error("this reading should not fail")
	} else {
		if str, err2 := conv.ToString(yamc.NewJsonReader()); err2 != nil {
			t.Error(err2)
		} else {
			// we do not test the string content because of different line endings on windows
			if str != `{"hello":["world","you"]}` {
				t.Error("unexpected string outcome", str)
			}
		}
		if _, ok := conv.GetData()["hello"]; !ok {
			t.Error("we should have the data node")
		}
	}

}

func TestGjson(t *testing.T) {
	data := []byte(`{"name":{"first":"Tom","last":"Anderson"},"age":37,"children":["Sara","Alex","Jack"],"fav.movie":"Deer Hunter","friends":[{"first":"Dale","last":"Murphy","age":44},{"first":"Roger","last":"Craig","age":68},{"first":"Jane","last":"Murphy","age":47}]}`)
	conv := yamc.New()
	if err := conv.Parse(yamc.NewJsonReader(), data); err != nil {
		t.Error("this reading should not fail")
	} else {
		if found, err := conv.Gjson("friends.2.last"); err != nil {
			t.Error(err)
		} else {
			if found.Str != "Murphy" {
				t.Error("unexpected value", found)
			}
		}

		// same test as above but with using GetGsonString
		if found, err := conv.GetGjsonString("friends.2.last"); err != nil {
			t.Error(err)
		} else {
			if found != "Murphy" {
				t.Error("unexpected value", found)
			}
		}

		// test for non existing path
		if found, err := conv.Gjson("friends.2.last2"); err != nil {
			t.Error(err)
		} else {
			if found.Exists() {
				t.Error("unexpected value", found)
			}
		}

		// test for non existing path by using GetGsonString
		if found, err := conv.GetGjsonString("friends.2.last2"); err != nil {
			t.Error(err)
		} else {
			if found != "" {
				t.Error("unexpected value", found)
			}
		}

	}

}
