package yamc

import (
	"fmt"
	"reflect"
	"strings"
)

type StructDef struct {
	Init bool
	// the struct we want to read
	Struct interface{}
	// the fields of the struct
	Fields map[string]StructField
}

type StructField struct {
	// the name of the field
	Name string
	// the type of the field
	Type string
	// the tag of the field
	Tag reflect.StructTag
	// reader depending...
	OrginalTag ReflectTagRef
	// if we are a node we have children
	Children map[string]StructField
	// for faster access store the length of the children
	ChildLen int
}

type ReflectTagRef struct {
	// if the field is renamed by a tag, we put this name here
	TagRenamed string

	// if we have additional tags, we put them here
	TagAdditional []string
}

// reftagFunc is a function that can be used to parse the tag by an specialized data reader, so the reader can solve the tag.
type reftagFunc func(StructField) ReflectTagRef

// NewStructDef returns a new struct reader
func NewStructDef(strct interface{}) *StructDef {
	return &StructDef{
		Init:   false,
		Struct: strct,
		Fields: make(map[string]StructField),
	}
}

// ReadStruct reads the given struct and returns a map with all values
// the tagparser is a function that can be used to parse the tag
// by an specialized data reader, so the reader can solve the tag.
// the tagparser can be nil, if you don't need it.
func (s *StructDef) ReadStruct(tagparser reftagFunc) error {
	return s.readStruct(s.Struct, tagparser)
}

// we ignore some type of base types.
// we wan't to the read simple structs.
// the field information is sugar for some validation or value converting.
// but not an requirement.
func (s *StructDef) ignoreField(field string) bool {
	ignores := []string{
		"map[string]interface {}",
		"[]interface {}",
		"interface {}",
	}
	for _, ignore := range ignores {
		if field == ignore {
			return true
		}
	}
	// we ignore pointers
	if field[0] == '*' {
		return true
	}
	// we ignore slices
	if field[0] == '[' {
		return true
	}
	// we ignore arrays
	if field[0] == '{' {
		return true
	}
	return false
}

// readStruct reads the given struct and returns a map with all values
// the tagparser is a function that can be used to parse the tag
// by an specialized data reader, so the reader can solve the tag.
// the tagparser can be nil, if you don't need it.
func (s *StructDef) readStruct(strct interface{}, tagparser reftagFunc) error {
	if strct == nil {
		return fmt.Errorf("structRead: given struct is nil")
	}
	refMap := reflect.TypeOf(strct)

	if refMap.Kind() != reflect.Ptr {
		return fmt.Errorf("structRead: given struct is not a pointer")
	}
	// get the value of the pointer
	refMap = refMap.Elem()
	rfname := refMap.String()
	// we go the happy path and ignore some types
	if s.ignoreField(rfname) {
		return nil
	}
	// we pass anything that we don't want to read
	// time to mark us as ready
	s.Init = true

	// build the field information map
	for i := 0; i < refMap.NumField(); i++ {
		field := refMap.Field(i)
		newField := s.createStructField(field, tagparser)
		s.Fields[field.Name] = newField

	}

	return nil

}

func (s *StructDef) createStructField(refStr reflect.StructField, tagparser reftagFunc) StructField {
	newField := StructField{
		Name: refStr.Name,
		Type: refStr.Type.String(),
		Tag:  refStr.Tag,
	}
	if refStr.Type.Kind() == reflect.Struct {
		nums := refStr.Type.NumField()
		newField.ChildLen = nums
		if nums > 0 {
			// we have children
			// we need to read them
			newField.Children = make(map[string]StructField)
			for i := 0; i < nums; i++ {
				field := refStr.Type.Field(i)
				children := s.createStructField(field, tagparser)
				newField.Children[field.Name] = children
			}
		}

	}
	if tagparser != nil {
		newField.OrginalTag = tagparser(newField)
	}
	return newField
}

// GetField returns the field information for the given field name
// if the field name contains a dot, we try to find the child
func (s *StructDef) GetField(field string) (StructField, error) {
	if !s.Init {
		return StructField{}, fmt.Errorf("structRead: struct not initialized")
	}
	return s.getField(field, s.Fields)
}

func (s *StructDef) getField(fieldPath string, from map[string]StructField) (StructField, error) {
	if field, ok := from[fieldPath]; !ok {
		// did not find the field. we need to search deeper
		// the path is a dot separated string
		if strings.Contains(fieldPath, ".") { // so do we have a dot?
			rootname := strings.Split(fieldPath, ".")[0] // we need to get the root name
			childs := strings.Split(fieldPath, ".")[1:]  // we need to remove the first element to get the childs path
			if field, ok := from[rootname]; ok {         // we found the root field
				if len(childs) > 0 && childs[0] != "" { // make sure we have a valid child read path
					if field.Children == nil { // and of course make sure we have children to read
						return StructField{}, fmt.Errorf("structRead: field [%s] has no children", rootname)
					}
					return s.getField(strings.Join(childs, "."), field.Children) // we have children, so we can read them
				} else {
					return StructField{}, fmt.Errorf("structRead: invalid path [%s]", fieldPath) // we have no valid child path
				}
			} else {
				return StructField{}, fmt.Errorf("structRead: root field [%s] exists, but not having [%s]", rootname, fieldPath)
			}
		}
		return StructField{}, fmt.Errorf("structRead: field [%s] not found", fieldPath)

	} else {
		return field, nil
	}

}
