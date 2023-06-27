package yamc_test

import (
	"testing"

	"github.com/swaros/contxt/module/yamc"
)

func TestJsonRead(t *testing.T) {
	type test001 struct {
		Id        int    `json:"id"`
		FirstName string `json:"first_name"`
		LastName  string `json:"last_name"`
		Email     string `json:"email"`
		Ip        string `json:"ip_address"`
	}

	var test001s []test001

	rdr := yamc.NewJsonReader()
	err := rdr.FileDecode("testdata/test001.json", &test001s)
	if err != nil {
		t.Error(err)
	}

	if len(test001s) != 4 {
		t.Errorf("expected 4 records, got %d", len(test001s))
	}

	if test001s[0].Id != 1 {
		t.Errorf("expected id 1, got %d", test001s[0].Id)
	}

	if test001s[0].FirstName != "Jeanette" {
		t.Errorf("expected first name Jeanette, got %s", test001s[0].FirstName)
	}

	if test001s[0].LastName != "Penddreth" {
		t.Errorf("expected last name Penddreth, got %s", test001s[0].LastName)
	}

	if test001s[0].Email != "jpenddreth0@census.gov" {
		t.Errorf("expected email jpenddreth0@census.gov, got %s", test001s[0].Email)
	}

	if test001s[0].Ip != "26.58.193.2" {
		t.Errorf("expected ip 26.58.193.2, got %s", test001s[0].Ip)
	}

	// because we read a slice, we can also check the last record
	if test001s[3].Id != 4 {
		t.Errorf("expected id 4, got %d", test001s[3].Id)
	}

	if test001s[3].FirstName != "Willard" {
		t.Errorf("expected first name Willard, got %s", test001s[3].FirstName)
	}

	// we have an load an slice. this should ignored by reading the additional records
	haveRecord := rdr.HaveFields()
	if haveRecord {
		t.Error("expected no records that could be read")
	}
}

func TestDecode01(t *testing.T) {

	type testData struct {
		Name    string `json:"name"`
		Age     int    `json:"age"`
		Contact struct {
			Email string `json:"email"`
			Phone string `json:"phone"`
		} `json:"contact"`

		Subs []string `json:"subs"`
	}
	source := `{
		"name": "test",
		"age": 42,
		"contact": {
			"email": "walkingtest@demo.gov",
			"phone": "1234567890"
		}
	}`

	rdr := yamc.NewJsonReader()
	var data testData
	err := rdr.Unmarshal([]byte(source), &data)
	if err != nil {
		t.Error(err)
	}

	// verify all fields
	if data.Name != "test" {
		t.Errorf("expected name test, got %s", data.Name)
	}

	if data.Age != 42 {
		t.Errorf("expected age 42, got %d", data.Age)
	}

	if data.Contact.Email != "walkingtest@demo.gov" {
		t.Errorf("expected email walkingtest@demo.gov, got %s", data.Contact.Email)
	}

	if data.Contact.Phone != "1234567890" {
		t.Errorf("expected phone 1234567890, got %s", data.Contact.Phone)
	}

	if len(data.Subs) != 0 {
		t.Errorf("expected 0 subs, got %d", len(data.Subs))
	}

	// do we have a struct reported?
	haveRecord := rdr.HaveFields()
	if !haveRecord {
		t.Error("expected records that could be read")
	}

	// verify the struct of the Name field
	fieldName, fErr := rdr.GetFields().GetField("Name")
	if fErr != nil {
		t.Error(fErr)
	} else {
		if fieldName.Name != "Name" {
			t.Errorf("expected field name, got %s", fieldName.Name)
		}

		if fieldName.OrginalTag.TagRenamed != "name" {
			t.Errorf("expected field tag name, got %s", fieldName.OrginalTag.TagRenamed)
		}

		if fieldName.Type != "string" {
			t.Errorf("expected field type string, got %s", fieldName.Type)
		}
	}
	// verify the struct of the Age field
	fieldAge, fErr := rdr.GetFields().GetField("Age")
	if fErr != nil {
		t.Error(fErr)
	} else {
		if fieldAge.Name != "Age" {
			t.Errorf("expected field name, got %s", fieldAge.Name)
		}

		if fieldAge.OrginalTag.TagRenamed != "age" {
			t.Errorf("expected field tag name, got %s", fieldAge.OrginalTag.TagRenamed)
		}

		if fieldAge.Type != "int" {
			t.Errorf("expected field type int, got %s", fieldAge.Type)
		}
	}

	// verify the struct of the Contact field
	fieldContact, fErr := rdr.GetFields().GetField("Contact")
	if fErr != nil {
		t.Error(fErr)
	} else {
		if fieldContact.Name != "Contact" {
			t.Errorf("expected field name, got %s", fieldContact.Name)
		}

		if fieldContact.OrginalTag.TagRenamed != "contact" {
			t.Errorf("expected field tag name, got %s", fieldContact.OrginalTag.TagRenamed)
		}
	}
}
