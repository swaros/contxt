package yamc_test

import (
	"testing"

	"github.com/swaros/contxt/module/yamc"
)

func TestYamlFile(t *testing.T) {
	if conv, err := yamc.NewYmacByYaml("testdata/test003.yml"); err != nil {
		t.Error(err)
	} else {
		LazyAssertGjsonPathEq(t, conv, "name", "Martin D'vloper")
		LazyAssertGjsonPathEq(t, conv, "foods.2", "Strawberry")
		LazyAssertGjsonPathEq(t, conv, "languages.perl", "Elite")
	}
}

func TestJsonFile(t *testing.T) {
	if conv, err := yamc.NewYmacByJson("testdata/test002.json"); err != nil {
		t.Error(err)
	} else {
		LazyAssertGjsonPathEq(t, conv, "_id", "5973782bdb9a930533b05cb2")
		LazyAssertGjsonPathEq(t, conv, "isActive", true)
		LazyAssertGjsonPathEq(t, conv, "age", float64(32))
		LazyAssertGjsonPathEq(t, conv, "friends.1.id", float64(1))
		LazyAssertGjsonPathEq(t, conv, "friends.2.name", "Carol Martin")
	}
}
