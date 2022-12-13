package yamc_test

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/swaros/contxt/module/yamc"
)

func helpFileLoad(filename string, dataHdl func(data []byte)) error {
	if data, err := ioutil.ReadFile(filename); err == nil {
		dataHdl(data)
	} else {
		return err
	}
	return nil
}

// Assert helper to handle equal check.
// by using callbacks we still have the line of error in the test output
func assertGjsonStringEquals(ymap *yamc.Yamc, path string, expected string, ifNotEquals func(val any), ifErr func(error)) {
	value, err := ymap.GetGjsonString(path)
	if err != nil {
		ifErr(err)
	}
	if value != expected {
		ifNotEquals(value)
	}
}

// Assert helper to handle equal check.
// by using callbacks we still have the line of error in the test output
func assertGjsonValueEquals(ymap *yamc.Yamc, path string, expected any, ifNotEquals func(val any), ifErr func(error)) {
	value, err := ymap.GetGjsonValue(path)
	if err != nil {
		ifErr(err)
	}
	if value != expected {
		ifNotEquals(value)
	}
}

func LazyAssertGjsonPathEq(t *testing.T, ymap *yamc.Yamc, path string, expected any) {
	assertGjsonValueEquals(ymap, path, expected, func(val any) {
		t.Error("expected the value (", expected, ") got [", val, "] instead")
	}, func(err error) {
		t.Error(err)
	})
}

// Testing simple Parsing of json content
func TestJsonParse(t *testing.T) {
	if err := helpFileLoad("testdata/test001.json", func(data []byte) {

		// init reader
		jsonReader := yamc.NewJsonReader()

		// init yamc
		conv := yamc.NewYmac()

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
			// check the first id. not being lazy now, to get at least
			// once the file and line number if this kinf of test is failing
			// here we want to have the value as string
			assertGjsonStringEquals(conv, "0.id", "1", func(val any) {
				t.Error("string value test. expected is id 1. got [", val, "] instead")
			}, func(err error) {
				t.Error(err)
			})
			// again not beeing lazy by testing the real value
			assertGjsonValueEquals(conv, "0.id", int64(1), func(val any) {
				t.Error("value test. expected is id 1. got [", val, "] instead")
			}, func(err error) {
				t.Error(err)
			})

			// now we are lazy. if oneof these tests fails, we will not point to this line
			// of code by the output
			LazyAssertGjsonPathEq(t, conv, "0.first_name", "Jeanette")
			LazyAssertGjsonPathEq(t, conv, "2.first_name", "Noell")
			LazyAssertGjsonPathEq(t, conv, "3", `{"email":"wvalek3@vk.com","first_name":"Willard","gender":"Male","id":4,"ip_address":"67.76.188.26","last_name":"Valek"}`)
			LazyAssertGjsonPathEq(t, conv, "2.id", int64(3))
		}

	}); err != nil {
		t.Error(err)
	}
}

func Test002(t *testing.T) {
	if err := helpFileLoad("testdata/test002.json", func(data []byte) {
		conv := yamc.NewYmac()
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
			LazyAssertGjsonPathEq(t, conv, "_id", "5973782bdb9a930533b05cb2")
			LazyAssertGjsonPathEq(t, conv, "isActive", true)
			LazyAssertGjsonPathEq(t, conv, "age", int64(32))
			LazyAssertGjsonPathEq(t, conv, "friends.1.id", int64(1))
			LazyAssertGjsonPathEq(t, conv, "friends.2.name", "Carol Martin")
		}
	}); err != nil {
		t.Error(err)
	}
}

func TestOfficialYaml(t *testing.T) {
	if err := helpFileLoad("testdata/official.yaml", func(data []byte) {
		conv := yamc.NewYmac()
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

			LazyAssertGjsonPathEq(t, conv, "YAML", "YAML Ain't Markup Language™")
			LazyAssertGjsonPathEq(t, conv, "YAML Resources.YAML Specifications.1", "YAML 1.1")
		}
	}); err != nil {
		t.Error(err)
	}
}

func Test003Yaml(t *testing.T) {
	if err := helpFileLoad("testdata/test003.yml", func(data []byte) {
		conv := yamc.NewYmac()
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

			LazyAssertGjsonPathEq(t, conv, "name", "Martin D'vloper")
			LazyAssertGjsonPathEq(t, conv, "foods.2", "Strawberry")
			LazyAssertGjsonPathEq(t, conv, "languages.perl", "Elite")
		}
	}); err != nil {
		t.Error(err)
	}
}

func TestJsonInvalid(t *testing.T) {
	data := []byte("[{hello}}]")
	conv := yamc.NewYmac()
	if err := conv.Parse(yamc.NewJsonReader(), data); err == nil {
		t.Error("this reading should fail")
	}

}

func TestYamlInvalid(t *testing.T) {
	data := []byte("[uhm]-")
	conv := yamc.NewYmac()
	if err := conv.Parse(yamc.NewYamlReader(), data); err == nil {
		t.Error("this reading should fail")
	}

}

func TestJsonYamlToString(t *testing.T) {
	data := []byte(`{"master": 45}`)
	conv := yamc.NewYmac()
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
	conv := yamc.NewYmac()
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
