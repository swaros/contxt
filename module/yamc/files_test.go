package yamc_test

import (
	"testing"

	"github.com/swaros/contxt/module/yamc"
)

func TestYamlFile(t *testing.T) {
	if conv, err := yamc.NewByYaml("testdata/test003.yml"); err != nil {
		t.Error(err)
	} else {
		LazyAssertPath(t, conv, "name", "Martin D'vloper")
		LazyAssertPath(t, conv, "foods.2", "Strawberry")
		LazyAssertPath(t, conv, "languages.perl", "Elite")
	}
}

func TestJsonFile(t *testing.T) {
	if conv, err := yamc.NewByJson("testdata/test002.json"); err != nil {
		t.Error(err)
	} else {
		LazyAssertPath(t, conv, "_id", "5973782bdb9a930533b05cb2")
		LazyAssertPath(t, conv, "isActive", true)
		LazyAssertPath(t, conv, "age", float64(32))
		LazyAssertPath(t, conv, "friends.1.id", float64(1))
		LazyAssertPath(t, conv, "friends.2.name", "Carol Martin")
	}
}
