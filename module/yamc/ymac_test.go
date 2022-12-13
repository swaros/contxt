package yamc_test

import (
	"io/ioutil"
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
