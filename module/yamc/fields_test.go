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
	} else {
		t.Error(err)
	}
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
	} else {
		t.Error(err)
	}

	// test to get a field with children
	if subInfoField, err := fields.GetField("SubInfo"); err == nil {
		if subInfoField.Name != "SubInfo" {
			t.Errorf("expected SubInfo field, got %s", subInfoField.Name)
		}

		if subInfoField.Type != "yamc_test.subData" {
			t.Errorf("expected yamc_test.subData type, got %s", subInfoField.Type)
		}

		if subInfoField.OrginalTag.TagRenamed != "subinfo" {
			t.Errorf("expected tag subinfo, got [%s]", subInfoField.OrginalTag.TagRenamed)
		}
	} else {
		t.Error(err)
	}

	// test to get a child from a existing field
	if subInfoField, err := fields.GetField("SubInfo.Header"); err == nil {
		if subInfoField.Name != "Header" {
			t.Errorf("expected Header field, got %s", subInfoField.Name)
		}

		if subInfoField.Type != "string" {
			t.Errorf("expected string type, got %s", subInfoField.Type)
		}

		if subInfoField.OrginalTag.TagRenamed != "header" {
			t.Errorf("expected tag header, got [%s]", subInfoField.OrginalTag.TagRenamed)
		}
	} else {
		t.Error(err)
	}
}
