package yamc_test

import (
	"strings"
	"testing"

	"github.com/swaros/contxt/module/yamc"
)

// jsontag parser. copied from jsonreader.go beaucause it is not exported
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
