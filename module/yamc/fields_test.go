package yamc_test

import (
	"strings"
	"testing"

	"github.com/swaros/contxt/module/yamc"
)

// jsontag parser. copied from jsonreader. because it is not exported
func parseJsonFn(info yamc.StructField) yamc.ReflectTagRef {
	if info.Tag.Get("json") != "" {
		all := info.Tag.Get("json")
		parts := strings.Split(all, ",")
		adds := []string{}
		if len(parts) > 1 {
			adds = parts[1:]
		}
		return yamc.ReflectTagRef{
			TagRenamed:    parts[0],
			TagAdditional: adds,
		}
	}

	return yamc.ReflectTagRef{}
}

func parseYamlTagfunc(info yamc.StructField) yamc.ReflectTagRef {
	if info.Tag.Get("yaml") != "" {
		all := info.Tag.Get("yaml")
		parts := strings.Split(all, ",")
		adds := []string{}
		if len(parts) > 1 {
			adds = parts[1:]
		}
		return yamc.ReflectTagRef{
			TagRenamed:    parts[0],
			TagAdditional: adds,
		}
	}

	return yamc.ReflectTagRef{}
}

type assertStruct struct {
	Name               string
	Type               string
	Path               string
	OriginalTagRenamed string
}

func assertPathResult(t *testing.T, fields *yamc.StructDef, path string, expected assertStruct) {
	t.Helper()
	field, err := fields.GetField(path)
	if err != nil {
		t.Error(err, "path", path)
	} else {
		// we assert the field if we have one
		// so we don't spam the output with errors
		assertStructField(t, field, expected)
	}
}

func assertStructField(t *testing.T, field yamc.StructField, expected assertStruct) {
	t.Helper()
	if expected.Name != "" && field.Name != expected.Name {
		t.Errorf("expected name [%s], got [%s]", expected.Name, field.Name)
	}

	if expected.Type != "" && field.Type != expected.Type {
		t.Errorf("expected type [%s], got [%s]", expected.Type, field.Type)
	}

	if expected.Path != "" && field.Path != expected.Path {
		t.Errorf("expected path [%s], got [%s]", expected.Path, field.Path)
	}

	if expected.OriginalTagRenamed != "" && field.OrginalTag.TagRenamed != expected.OriginalTagRenamed {
		t.Errorf("expected tag [%s], got [%s]", expected.OriginalTagRenamed, field.OrginalTag.TagRenamed)
	}

}

func TestNotInitialized(t *testing.T) {
	var data interface{}
	fields := yamc.NewStructDef(&data)
	if fields.Init {
		t.Error("expected fields not to be initialized")
	}

	_, err := fields.GetField("test")
	if err == nil {
		t.Error("expected error, because we do not have any fields")
	}
}

func TestNilInterface(t *testing.T) {
	var data interface{}
	fields := yamc.NewStructDef(data)
	if err := fields.ReadStruct(parseJsonFn); err == nil {
		t.Error("expected error, because we do not have a pointer")
	} else {
		expectedError := "structRead: given struct is nil"
		if err.Error() != expectedError {
			t.Errorf("expected error [%s], got [%s]", expectedError, err.Error())
		}
	}

}

func TestNonPointer(t *testing.T) {
	type testData struct {
		Name    string `json:"name"`
		Age     int    `json:"age"`
		Contact struct {
			Email string `json:"email"`
			Phone string `json:"phone"`
		} `json:"contact"`

		Subs []string `json:"subs"`
	}
	var data testData
	fields := yamc.NewStructDef(data)
	if err := fields.ReadStruct(parseJsonFn); err == nil {
		t.Error("expected error, because we do not have a pointer")
	} else {
		expectedError := "structRead: given struct is not a pointer"
		if err.Error() != expectedError {
			t.Errorf("expected error [%s], got [%s]", expectedError, err.Error())
		}
	}
}

func TestFieldGetField(t *testing.T) {
	type testData struct {
		Name    string `json:"name"`
		Age     int    `json:"age"`
		Contact struct {
			Email string `json:"email"`
			Phone string `json:"phone"`
		} `json:"contact"`

		Subs []string `json:"subs"`
	}
	var data testData
	fields := yamc.NewStructDef(&data)

	// verify all fields
	if err := fields.ReadStruct(parseJsonFn); err != nil {
		t.Error(err)
	}
	// test to get a field without any children
	if nameField, err := fields.GetField("Name"); err == nil {
		if nameField.Name != "Name" {
			t.Errorf("expected name field, got %s", nameField.Name)
		}

		if nameField.Type != "string" {
			t.Errorf("expected string type, got %s", nameField.Type)
		}

		if nameField.OrginalTag.TagRenamed != "name" {
			t.Errorf("expected tag name, got [%s]", nameField.OrginalTag.TagRenamed)
		}
	} else {
		t.Error(err)
	}
	// test to get a field with children
	if contactField, err := fields.GetField("Contact.Email"); err == nil {
		if contactField.Name != "Email" {
			t.Errorf("expected email field, got %s", contactField.Name)
		}

		if contactField.Type != "string" {
			t.Errorf("expected string type, got %s", contactField.Type)
		}

		if contactField.OrginalTag.TagRenamed != "email" {
			t.Errorf("expected tag email, got [%s]", contactField.OrginalTag.TagRenamed)
		}
	} else {
		t.Error(err)
	}
}

// same as the test above, but with more levels of nesting
func TestFieldGetFieldDeep(t *testing.T) {
	type testData struct {
		Name    string `json:"name"`
		Age     int    `json:"age"`
		Contact struct {
			Email string `json:"email"`
			Phone string `json:"phone"`
			Addr  struct {
				Street string `json:"street"`
				City   string `json:"city"`
			} `json:"addr"`
		} `json:"contact"`

		Subs []string `json:"subs"`
	}
	var data testData
	fields := yamc.NewStructDef(&data)

	// verify all fields
	if err := fields.ReadStruct(parseJsonFn); err != nil {
		t.Error(err)
	}

	// test to get a field with children
	if contactField, err := fields.GetField("Contact.Email"); err == nil {
		if contactField.Name != "Email" {
			t.Errorf("expected email field, got %s", contactField.Name)
		}

		if contactField.Type != "string" {
			t.Errorf("expected string type, got %s", contactField.Type)
		}

		if contactField.OrginalTag.TagRenamed != "email" {
			t.Errorf("expected tag email, got [%s]", contactField.OrginalTag.TagRenamed)
		}
	} else {
		t.Error(err)
	}

	// test to get nested field with children
	if contactField, err := fields.GetField("Contact.Addr.Street"); err == nil {
		if contactField.Name != "Street" {
			t.Errorf("expected street field, got %s", contactField.Name)
		}

		if contactField.Type != "string" {
			t.Errorf("expected string type, got %s", contactField.Type)
		}

		if contactField.OrginalTag.TagRenamed != "street" {
			t.Errorf("expected tag street, got [%s]", contactField.OrginalTag.TagRenamed)
		}
	} else {
		t.Error(err)
	}

	// test to get nested field with children
	if contactField, err := fields.GetField("Contact.Addr.City"); err == nil {
		if contactField.Name != "City" {
			t.Errorf("expected city field, got %s", contactField.Name)
		}

		if contactField.Type != "string" {
			t.Errorf("expected string type, got %s", contactField.Type)
		}

		if contactField.OrginalTag.TagRenamed != "city" {
			t.Errorf("expected tag city, got [%s]", contactField.OrginalTag.TagRenamed)
		}
	} else {
		t.Error(err)
	}

	// test to get a non existing child from a existing field
	if _, err := fields.GetField("Contact.Addr.NonExisting"); err == nil {
		t.Errorf("expected error, got nil")

	} else {
		expectedError := "structRead: field [NonExisting] not found"
		if err.Error() != expectedError {
			t.Errorf("expected error [%s], got [%s]", expectedError, err.Error())
		}
	}

	// test to get a non existing field
	if _, err := fields.GetField("NonExisting"); err == nil {
		t.Errorf("expected error, got nil")
	} else {
		expectedError := "structRead: field [NonExisting] not found"
		if err.Error() != expectedError {
			t.Errorf("expected error [%s], got [%s]", expectedError, err.Error())
		}
	}

	// test to get a child from a non existing field without children
	if _, err := fields.GetField("Name.LastName"); err == nil {
		t.Errorf("expected error, got nil")
	} else {
		expectedError := "structRead: field [Name] has no children"
		if err.Error() != expectedError {
			t.Errorf("expected error [%s], got [%s]", expectedError, err.Error())
		}
	}

	// test an overall invalid path
	if _, err := fields.GetField("Contact."); err == nil {
		t.Errorf("expected error, got nil")
	} else {
		expectedError := "structRead: invalid path [Contact.]"
		if err.Error() != expectedError {
			t.Errorf("expected error [%s], got [%s]", expectedError, err.Error())
		}
	}
}

func TestGetFieldDeepByIntentList(t *testing.T) {
	type testData struct {
		Name    string `json:"name"`
		Age     int    `json:"age"`
		Contact struct {
			Email string `json:"email"`
			Phone string `json:"phone"`
			Addr  struct {
				Street string `json:"street"`
				City   string `json:"city"`
			} `json:"addr"`
		} `json:"contact"`

		Subs []string `json:"subs"`
	}
	var data testData
	fields := yamc.NewStructDef(&data)

	// verify all fields
	if err := fields.ReadStruct(parseJsonFn); err != nil {
		t.Error(err)
	}

	// add fields to index
	fields.AddToIndex("prevoius", "    data", "contact", "    Email", "    Phone")

	fields.SetAllowedTagSearch(true)

	expectedLevel := 4 // chars of leading spaces
	if fields.DetectIndentCount("") != expectedLevel {
		t.Errorf("expected indent level [%d], got [%d]", expectedLevel, fields.DetectIndentCount(""))
	}

	if contactField, err := fields.GetField("    Email"); err == nil {
		if contactField.Name != "Email" {
			t.Errorf("expected email field, got %s", contactField.Name)
		}

		if contactField.Type != "string" {
			t.Errorf("expected string type, got %s", contactField.Type)
		}

		if contactField.OrginalTag.TagRenamed != "email" {
			t.Errorf("expected tag email, got [%s]", contactField.OrginalTag.TagRenamed)
		}
	} else {
		t.Error(err)
	}

	if contactField, err := fields.GetField("    Phone"); err == nil {
		if contactField.Name != "Phone" {
			t.Errorf("expected phone field, got %s", contactField.Name)
		}

		if contactField.Type != "string" {
			t.Errorf("expected string type, got %s", contactField.Type)
		}

		if contactField.OrginalTag.TagRenamed != "phone" {
			t.Errorf("expected tag phone, got [%s]", contactField.OrginalTag.TagRenamed)
		}
	} else {
		t.Error(err)
	}
}

func TestGetFieldDeepByIntentListOnlyRootKnow(t *testing.T) {
	type testData struct {
		Name    string `json:"name"`
		Age     int    `json:"age"`
		Contact struct {
			Email string `json:"email"`
			Phone string `json:"phone"`
			Addr  struct {
				Street string `json:"street"`
				City   string `json:"city"`
			} `json:"addr"`
		} `json:"contact"`

		Subs []string `json:"subs"`
	}
	var data testData
	fields := yamc.NewStructDef(&data)

	// verify all fields
	if err := fields.ReadStruct(parseJsonFn); err != nil {
		t.Error(err)
	}

	// add fields to index
	fields.AddToIndex("contact")

	fields.SetAllowedTagSearch(true)

	expectedLevel := 4 // chars of leading spaces
	if fields.DetectIndentCount("    Email") != expectedLevel {
		t.Errorf("expected indent level [%d], got [%d]", expectedLevel, fields.DetectIndentCount("    Email"))
	}

	if contactField, err := fields.GetField("    Email"); err == nil {
		if contactField.Name != "Email" {
			t.Errorf("expected email field, got %s", contactField.Name)
		}

		if contactField.Type != "string" {
			t.Errorf("expected string type, got %s", contactField.Type)
		}

		if contactField.OrginalTag.TagRenamed != "email" {
			t.Errorf("expected tag email, got [%s]", contactField.OrginalTag.TagRenamed)
		}

		if contactField.Path != "Contact.Email" {
			t.Errorf("expected parent Contact, got [%s]", contactField.Path)
		}

	} else {
		t.Error(err)
	}

	if contactField, err := fields.GetField("    Phone"); err == nil {
		if contactField.Name != "Phone" {
			t.Errorf("expected phone field, got %s", contactField.Name)
		}

		if contactField.Type != "string" {
			t.Errorf("expected string type, got %s", contactField.Type)
		}

		if contactField.OrginalTag.TagRenamed != "phone" {
			t.Errorf("expected tag phone, got [%s]", contactField.OrginalTag.TagRenamed)
		}
		if contactField.Path != "Contact.Phone" {
			t.Errorf("expected parent Contact, got [%s]", contactField.Path)
		}
	} else {
		t.Error(err)
	}
}

func TestYamlTest0003Load(t *testing.T) {
	type testData struct {
		Name      string            `yaml:"name"`
		Job       string            `yaml:"job"`
		Skill     string            `yaml:"skill"`
		Foods     []string          `yaml:"foods"`
		Languages map[string]string `yaml:"languages"`
		Education string            `yaml:"education"`
	}
	var data testData
	fields := yamc.NewStructDef(&data)

	// verify all fields
	if err := fields.ReadStruct(parseYamlTagfunc); err != nil {
		t.Error(err)
	}

	// test to get a field with children
	if languageField, err := fields.GetField("Languages"); err == nil {
		if languageField.Name != "Languages" {
			t.Errorf("expected languages field, got %s", languageField.Name)
		}

		if languageField.Type != "map[string]string" {
			t.Errorf("expected map[string]string type, got %s", languageField.Type)
		}

		if languageField.OrginalTag.TagRenamed != "languages" {
			t.Errorf("expected tag languages, got [%s]", languageField.OrginalTag.TagRenamed)
		}

		if languageField.Path != "Languages" {
			t.Errorf("expected parent Languages, got [%s]", languageField.Path)
		}
	} else {
		t.Error(err)
	}
	assertPathResult(t, fields, "Languages", assertStruct{
		Name:               "Languages",
		Type:               "map[string]string",
		OriginalTagRenamed: "languages",
		Path:               "Languages",
	})
}

func TestYamlTest0004Load(t *testing.T) {
	type subData struct {
		Header  string `yaml:"header"`
		Content string `yaml:"content"`
	}
	type testData struct {
		Name      string            `yaml:"name"`
		Job       string            `yaml:"job"`
		Skill     string            `yaml:"skill"`
		Foods     []string          `yaml:"foods"`
		Languages map[string]string `yaml:"languages"`
		Education string            `yaml:"education"`
		SubInfo   subData           `yaml:"subinfo"`
	}
	var data testData
	fields := yamc.NewStructDef(&data)

	// verify all fields
	if err := fields.ReadStruct(parseYamlTagfunc); err != nil {
		t.Error(err)
	}
	assertPathResult(t, fields, "Languages", assertStruct{
		Name:               "Languages",
		Type:               "map[string]string",
		OriginalTagRenamed: "languages",
		Path:               "Languages",
	})

	assertPathResult(t, fields, "SubInfo", assertStruct{
		Name:               "SubInfo",
		Type:               "yamc_test.subData",
		Path:               "SubInfo",
		OriginalTagRenamed: "subinfo",
	})

	assertPathResult(t, fields, "SubInfo.Header", assertStruct{
		Name:               "Header",
		Type:               "string",
		Path:               "SubInfo.Header",
		OriginalTagRenamed: "header",
	})

}

// testing if the SetAllowedTagSearch respects the the level of search.
// that means if we are looking for the field "Email" but the field is inside a struct
// it should not be found even if the SetAllowedTagSearch is true, because this search would only
// match the field "Email" if it is a direct child of the struct.
// so the tag search only find the nodes at the same level.
func TestPairPaths(t *testing.T) {
	type AdressDemo struct {
		Name    string `yaml:"name"`
		Contact struct {
			Email string `yaml:"email"`
			Phone string `yaml:"phone"`
		} `yaml:"contact"`
		LastName string `yaml:"lastname"`
		Age      int    `yaml:"age"`
	}

	var data AdressDemo
	fields := yamc.NewStructDef(&data)
	if err := fields.ReadStruct(parseYamlTagfunc); err != nil {
		t.Error(err)
	}
	// we enable the possibility to search for tags
	fields.SetAllowedTagSearch(true)

	fields.AddToIndex("Age", "Contact", "    Email", "    Phone")
	// here we search for the field "Email", by using the tag "email" (`yaml:"email"`)  but it is inside a struct.
	// so this should not work
	_, error := fields.GetField("email")
	if error == nil {
		t.Error("that should not work. looking for a field inside a struct by tags needs dot notation, or leading spaces")
	}

	// we need to search for the field "Email" by using the dot notation
	// so we need to use the path "Contact.Email"
	// this should work. here no need for tag search, but for the sake of the test we enable it
	var prop yamc.StructField
	prop, error = fields.GetField("Contact.Email")
	if error != nil {
		t.Error("that should work. looking for a field inside a struct by PropName.PropName")
	} else {
		if prop.Name != "Email" {
			t.Errorf("expected Email field, got %s", prop.Name)
		}
	}

	// here we mix the dot notation with the tag search
	// so we need to use the path "Contact.email"
	// this should not work. because the tag search is only for the same level
	prop, error = fields.GetField("Contact.email")
	if error != nil {
		t.Error("that should work. looking for a field inside a struct PropName.tagname")
	} else {
		if prop.Name != "Email" {
			t.Errorf("expected Email field, got %s", prop.Name)
		}

	}

	assertPathResult(t, fields, "Contact.email", assertStruct{
		Name:               "Email",
		Type:               "string",
		Path:               "Contact.Email",
		OriginalTagRenamed: "email",
	})

	// here we mix the dot notation with the tag search
	// so we need to use the path "Contact.email"
	// this should not work. because the tag search is only for the same level
	prop, error = fields.GetField("    email")
	if error != nil {
		t.Error("that should work. looking for a field inside a struct by using leading spaces '    tagname'")
		t.Error(error)
	}

	assertPathResult(t, fields, "    email", assertStruct{
		Name:               "Email",
		Type:               "string",
		Path:               "Contact.Email",
		OriginalTagRenamed: "email",
	})

	// the same as above but with more leading spaces what should not work
	_, error = fields.GetField("        email")
	if error == nil {
		t.Error("that should not work. looking for a field inside a struct by using leading spaces '    tagname'")
	} else {
		// because we indent the tagname with 8 spaces, the error should be
		expectedError := "structRead: field [Phone] has no children"
		// ... thats because we asume that we looking in the third level (twise of indentlevel 4) for a field
		// and the field "Phone" is in the second level. so we get the error that the field "Phone" has no children
		if error.Error() != expectedError {
			t.Errorf("expected error [%s], got [%s]", expectedError, error.Error())
		}
	}
}

func TestSliceProperty(t *testing.T) {
	type targets struct {
		Name     string `yaml:"name"`
		SureName string `yaml:"surename"`
	}

	type slConfig struct {
		Main    string    `yaml:"main"`
		Targets []targets `yaml:"targets"`
	}
	var data slConfig

	fields := yamc.NewStructDef(&data)
	if err := fields.ReadStruct(parseYamlTagfunc); err != nil {
		t.Error(err)
	}
	// we enable the possibility to search for tags
	fields.SetAllowedTagSearch(true)

	assertPathResult(t, fields, "main", assertStruct{
		Name:               "Main",
		Type:               "string",
		Path:               "Main",
		OriginalTagRenamed: "main",
	},
	)

	assertPathResult(t, fields, "targets", assertStruct{
		Name:               "Targets",
		Type:               "[]yamc_test.targets",
		Path:               "Targets",
		OriginalTagRenamed: "targets",
	},
	)

	assertPathResult(t, fields, "targets.name", assertStruct{
		Name:               "Name",
		Type:               "string",
		Path:               "Targets.Name",
		OriginalTagRenamed: "name",
	},
	)

	assertPathResult(t, fields, "targets.surename", assertStruct{
		Name:               "SureName",
		Type:               "string",
		Path:               "Targets.SureName",
		OriginalTagRenamed: "surename",
	},
	)

}

func TestDeepStruct(t *testing.T) {
	type worker struct {
		Name     string `yaml:"name"`
		SureName string `yaml:"surename"`
	}

	type targets struct {
		Worker []worker `yaml:"worker"`
		Labels []string `yaml:"labels"`
	}

	type testConfig struct {
		Main    string  `yaml:"main"`
		Targets targets `yaml:"targets"`
	}
	var testConf testConfig

	fields := yamc.NewStructDef(&testConf)
	if err := fields.ReadStruct(parseYamlTagfunc); err != nil {
		t.Error(err)
	}

	assertPathResult(t, fields, "Main", assertStruct{
		Name:               "Main",
		Type:               "string",
		Path:               "Main",
		OriginalTagRenamed: "main",
	},
	)

	assertPathResult(t, fields, "Targets", assertStruct{
		Name:               "Targets",
		Type:               "yamc_test.targets",
		Path:               "Targets",
		OriginalTagRenamed: "targets",
	},
	)

	assertPathResult(t, fields, "Targets.Worker", assertStruct{
		Name:               "Worker",
		Type:               "[]yamc_test.worker",
		Path:               "Targets.Worker",
		OriginalTagRenamed: "worker",
	},
	)

	assertPathResult(t, fields, "Targets.Worker.Name", assertStruct{
		Name:               "Name",
		Type:               "string",
		Path:               "Targets.Worker.Name",
		OriginalTagRenamed: "name",
	},
	)

	assertPathResult(t, fields, "Targets.Worker.SureName", assertStruct{
		Name:               "SureName",
		Type:               "string",
		Path:               "Targets.Worker.SureName",
		OriginalTagRenamed: "surename",
	},
	)

	assertPathResult(t, fields, "Targets.Labels", assertStruct{
		Name:               "Labels",
		Type:               "[]string",
		Path:               "Targets.Labels",
		OriginalTagRenamed: "labels",
	},
	)

	slice := fields.GetOrderedIndexSlice()
	if len(slice) != 0 {
		t.Errorf("expected 0 fields, got %d", len(slice))
	}

}

func TestDeepStructWithTagSearch(t *testing.T) {
	type worker struct {
		Name     string `yaml:"name"`
		SureName string `yaml:"surename"`
	}

	type targets struct {
		Worker []worker `yaml:"worker"`
		Labels []string `yaml:"labels"`
	}

	type testConfig struct {
		Main    string  `yaml:"main"`
		Targets targets `yaml:"targets"`
	}
	var testConf testConfig

	fields := yamc.NewStructDef(&testConf)
	if err := fields.ReadStruct(parseYamlTagfunc); err != nil {
		t.Error(err)
	}
	// we enable the possibility to search for tags
	fields.SetAllowedTagSearch(true)

	assertPathResult(t, fields, "main", assertStruct{
		Name:               "Main",
		Type:               "string",
		Path:               "Main",
		OriginalTagRenamed: "main",
	},
	)

	assertPathResult(t, fields, "targets", assertStruct{
		Name:               "Targets",
		Type:               "yamc_test.targets",
		Path:               "Targets",
		OriginalTagRenamed: "targets",
	},
	)

	assertPathResult(t, fields, "targets.worker", assertStruct{
		Name:               "Worker",
		Type:               "[]yamc_test.worker",
		Path:               "Targets.Worker",
		OriginalTagRenamed: "worker",
	},
	)

	assertPathResult(t, fields, "targets.worker.name", assertStruct{
		Name:               "Name",
		Type:               "string",
		Path:               "Targets.Worker.Name",
		OriginalTagRenamed: "name",
	},
	)

	assertPathResult(t, fields, "targets.worker.surename", assertStruct{
		Name:               "SureName",
		Type:               "string",
		Path:               "Targets.Worker.SureName",
		OriginalTagRenamed: "surename",
	},
	)

	assertPathResult(t, fields, "targets.labels", assertStruct{
		Name:               "Labels",
		Type:               "[]string",
		Path:               "Targets.Labels",
		OriginalTagRenamed: "labels",
	},
	)

}

// testing the indents together with the tag search
// and adding search index.
// here we take care on the right intentation
func TestDeepStructWithTagSearchAndIndents(t *testing.T) {
	type worker struct {
		Name     string `yaml:"name"`
		SureName string `yaml:"surename"`
	}

	type targets struct {
		Worker []worker `yaml:"worker"`
		Labels []string `yaml:"labels"`
	}

	type testConfig struct {
		Main    string  `yaml:"main"`
		Targets targets `yaml:"targets"`
	}
	var testConf testConfig

	fields := yamc.NewStructDef(&testConf)

	if err := fields.ReadStruct(parseYamlTagfunc); err != nil {
		t.Error(err)
	}
	// we enable the possibility to search for tags
	fields.SetAllowedTagSearch(true)

	assertPathResult(t, fields, "Main", assertStruct{
		Name:               "Main",
		Type:               "string",
		Path:               "Main",
		OriginalTagRenamed: "main",
	},
	)

	assertPathResult(t, fields, "Targets", assertStruct{
		Name:               "Targets",
		Type:               "yamc_test.targets",
		Path:               "Targets",
		OriginalTagRenamed: "targets",
	},
	)

	fields.AddToIndex("Main", "Targets", "  Worker")
	assertPathResult(t, fields, "  Worker", assertStruct{
		Name:               "Worker",
		Type:               "[]yamc_test.worker",
		Path:               "Targets.Worker",
		OriginalTagRenamed: "worker", // TODO: tag not exists ...investigate
	},
	)

	fields.AddToIndex("    Name", "    SureName")
	assertPathResult(t, fields, "    Name", assertStruct{
		Name:               "Name",
		Type:               "string",
		Path:               "Targets.Worker.Name",
		OriginalTagRenamed: "name",
	},
	)

	assertPathResult(t, fields, "    SureName", assertStruct{
		Name:               "SureName",
		Type:               "string",
		Path:               "Targets.Worker.SureName",
		OriginalTagRenamed: "surename",
	},
	)

	assertPathResult(t, fields, "  Labels", assertStruct{
		Name:               "Labels",
		Type:               "[]string",
		Path:               "Targets.Labels",
		OriginalTagRenamed: "labels",
	},
	)

}

// testing the indents together with the tag search
// and adding search index.
// here we have an bigger indent then the regular one
func TestDeepStructWithTagSearchAndBiggerIndents(t *testing.T) {
	type worker struct {
		Name     string `yaml:"name"`
		SureName string `yaml:"surename"`
	}

	type targets struct {
		Worker []worker `yaml:"worker"`
		Labels []string `yaml:"labels"`
	}

	type testConfig struct {
		Main    string  `yaml:"main"`
		Targets targets `yaml:"targets"`
	}
	var testConf testConfig

	fields := yamc.NewStructDef(&testConf)

	if err := fields.ReadStruct(parseYamlTagfunc); err != nil {
		t.Error(err)
	}
	// we enable the possibility to search for tags
	fields.SetAllowedTagSearch(true)

	assertPathResult(t, fields, "Main", assertStruct{
		Name:               "Main",
		Type:               "string",
		Path:               "Main",
		OriginalTagRenamed: "main",
	},
	)

	assertPathResult(t, fields, "Targets", assertStruct{
		Name:               "Targets",
		Type:               "yamc_test.targets",
		Path:               "Targets",
		OriginalTagRenamed: "targets",
	},
	)

	fields.AddToIndex("Main", "Targets", "  Worker")
	assertPathResult(t, fields, "  Worker", assertStruct{
		Name:               "Worker",
		Type:               "[]yamc_test.worker",
		Path:               "Targets.Worker",
		OriginalTagRenamed: "worker", // TODO: tag not exists ...investigate
	},
	)

	assertPathResult(t, fields, "      Name", assertStruct{
		Name:               "Name",
		Type:               "string",
		Path:               "Targets.Worker.Name",
		OriginalTagRenamed: "name",
	},
	)

	assertPathResult(t, fields, "    SureName", assertStruct{
		Name:               "SureName",
		Type:               "string",
		Path:               "Targets.Worker.SureName",
		OriginalTagRenamed: "surename",
	},
	)

	assertPathResult(t, fields, "  Labels", assertStruct{
		Name:               "Labels",
		Type:               "[]string",
		Path:               "Targets.Labels",
		OriginalTagRenamed: "labels",
	},
	)

}

// copied logic for faster testing ...yes it is dirty
func trimAndGetLevel(str string) (string, int) {
	trimedWord := strings.TrimLeft(str, " ")
	if trimedWord == "" {
		return "", 0
	}
	level := (len(str) - len(trimedWord)) / len(" ")
	return trimedWord, level
}

func filterByIndent(strSlice []string) []string {
	max := 0
	// get max needed entries by the indent level
	for _, str := range strSlice {
		_, cur := trimAndGetLevel(str)
		if cur > max {
			max = cur
		}
	}

	var filtered []string
	// create a slice with the max needed entries
	for i := 0; i <= max; i++ {
		filtered = append(filtered, "")
	}
	// fill the slice with the entries depending on the level
	for _, str := range strSlice {
		trimStr, level := trimAndGetLevel(str)
		// do we have an entrie already?
		if filtered[level] != "" {
			// remove the entries they are deeper
			for i := level + 1; i <= max; i++ {
				filtered[i] = ""
			}
		}
		filtered[level] = trimStr

	}
	cleared := []string{}
	for _, str := range filtered {
		if str != "" {
			cleared = append(cleared, str)
		}
	}
	return cleared
}

func TestTemporaryFilterFunction(t *testing.T) {
	testSlice := []string{
		"Main",
		"Targets",
		"  Worker",
		"    Name",
		"    SureName",
		"  CoWorker",
		"    Company",
		"Next",
		"  Labels",
		"    Test",
	}

	filtered := filterByIndent(testSlice)

	expected := "Next.Labels.Test"

	dotted := strings.Join(filtered, ".")

	if dotted != expected {
		t.Errorf("expected %s got %s", expected, dotted)
	}

	testSlice = []string{
		"Main",
		"Targets",
		"  Worker",
		"    Name",
		"    SureName",
		"  CoWorker",
		"    Company",
		"Next",
		"   Labels", // moved one space to right to have a different level, so it should lead to a different result
		"     Test",
	}

	filtered = filterByIndent(testSlice)

	expected = "Next.Labels.Test"

	dotted = strings.Join(filtered, ".")

	if dotted != expected {
		t.Errorf("expected %s got %s", expected, dotted)
	}

}
