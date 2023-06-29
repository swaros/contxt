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
	// if the struct is ignored, we store the reason here
	IgnoredBecauseOf string
	// the ordered index slice, contains keynames in the order they are defined or read.
	// this is used, if the struct is have to be read in context of the previous key.
	// for example if we have a struct like this:
	// connection:
	//   host: localhost
	//   port: 8080
	// the regular way is to read the struct by using the chained keynames: connection.host
	// but if we only have an intented string like "    host", and we don't know the chain keyname, we have to know the previous keyname
	// what is in this case "connection". that have a lower indentation level.
	// so because of this, it is omportant, that this list is ordered.
	orderedIndexSlice []string
	// the indented chars of the struct. like "  " or "    "
	indentChars string
	// the current indentation level. means how many chars are used as indent
	indentLevel int
	// if true, we allow to search by tag. so we can find Email in a struct like this:
	// type User struct {
	//   Email string `json:"email"`
	// }
	// if false, we only search by the field name. so it must be "email" (because of the yaml keyname)
	allowBytagSearch bool
}

type StructField struct {
	// the name of the field
	Name string
	// the path of the field in relation to the struct
	Path string
	// the type of the field
	Type string
	// the tag of the field
	Tag reflect.StructTag
	// reader depending...
	OrginalTag ReflectTagRef
	// if we are a node, we have children
	Children map[string]StructField
	// for faster access store the amount of children
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
		Init:        false,
		Struct:      strct,
		Fields:      make(map[string]StructField),
		indentChars: " ",
		indentLevel: 0,
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
			s.IgnoredBecauseOf = ignore + " is not supported"
			return true
		}
	}
	// we ignore pointers
	if field[0] == '*' {
		s.IgnoredBecauseOf = "pointers are not supported"
		return true
	}
	// we ignore slices
	if field[0] == '[' {
		s.IgnoredBecauseOf = "slices are not supported"
		return true
	}
	// we ignore arrays
	if field[0] == '{' {
		s.IgnoredBecauseOf = "arrays are not supported"
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
		newField := s.createStructField(field, tagparser, "")
		s.Fields[field.Name] = newField

	}

	return nil

}

func (s *StructDef) createStructField(refStr reflect.StructField, tagparser reftagFunc, parent string) StructField {
	// compose the path
	pathStr := refStr.Name
	if parent != "" {

		pathStr = parent + "." + refStr.Name
	}
	// basic information
	newField := StructField{
		Name: refStr.Name,
		Type: refStr.Type.String(),
		Tag:  refStr.Tag,
		Path: pathStr,
	}
	// structs are more complex.
	// we need to read the child fields
	if refStr.Type.Kind() == reflect.Struct {
		nums := refStr.Type.NumField()
		newField.ChildLen = nums
		if nums > 0 {
			// we have children
			// we need to read them
			newField.Children = make(map[string]StructField)
			for i := 0; i < nums; i++ {
				field := refStr.Type.Field(i)
				children := s.createStructField(field, tagparser, newField.Path)
				newField.Children[field.Name] = children
			}
		}

	}
	// the optional tagparser to resolve the tag information.
	// this is done by the reader.
	if tagparser != nil {
		newField.OrginalTag = tagparser(newField)
	}
	return newField
}

// check the index entries and calculate the indent level
// we need to find the indent level to be able to read the struct
func (s *StructDef) DetectIndentCount(withField string) int {

	workIngSlice := s.orderedIndexSlice
	if withField != "" {
		workIngSlice = append(workIngSlice, withField)
	}
	first := true
	levelFound := 0
	for _, ent := range workIngSlice {
		_, level := s.trimAndGetLevel(ent)
		if first && level != 0 {
			levelFound = level
			first = false
		} else {
			if levelFound < level && levelFound != 0 {
				levelFound = level
			}
		}
	}
	if levelFound != 0 {
		s.indentLevel = levelFound
	}
	return s.indentLevel
}

// GetField returns the field information for the given field name
// if the field name contains a dot, we try to find the child
func (s *StructDef) GetField(field string) (StructField, error) {
	if !s.Init {
		return StructField{}, fmt.Errorf("structRead: struct not initialized")
	}
	var returnErr error
	var returnField StructField
	// we have a request to read a child and we only get leading spaces.
	// so we need to find the parent by the orderedIndex
	if s.haveIntend(field) {
		// if we have an intend level of 0, we have to run the auto detect
		if s.indentLevel == 0 {
			// we using the current field to detect the indent level too
			// this could be the only source we have
			s.DetectIndentCount(field)
		}
		returnField, returnErr = s.getField(s.getChainByIndex(field), s.Fields)
	} else {
		returnField, returnErr = s.getField(field, s.Fields)
	}
	return returnField, returnErr
}

func (s *StructDef) getField(fieldPath string, from map[string]StructField) (StructField, error) {
	if field, ok := s.findField(fieldPath, from); !ok {
		// did not find the field. we need to search deeper
		// the path is a dot separated string
		if strings.Contains(fieldPath, ".") { // so do we have a dot?
			rootname := strings.Split(fieldPath, ".")[0]      // we need to get the root name
			childs := strings.Split(fieldPath, ".")[1:]       // we need to remove the first element to get the childs path
			if field, ok := s.findField(rootname, from); ok { // we found the root field
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

func (s *StructDef) findField(name string, from map[string]StructField) (StructField, bool) {
	if field, ok := from[name]; ok {
		return field, true
	}
	if s.allowBytagSearch {
		if structFound, err := s.getFieldByTag(name, from); err == nil {
			return structFound, true
		}
	}

	return StructField{}, false
}

func (s *StructDef) getChainByIndex(str string) string {
	// first we get the trimed word, and the level of the intend
	trimedWord, level := s.trimAndGetLevel(str)

	// iterate over the ordered index from last to first
	for i := len(s.orderedIndexSlice) - 1; i >= 0; i-- {
		// find the next parent by the level that have -1 of the current level
		// first get entry
		entry := s.orderedIndexSlice[i]
		// get the level of the entry
		trimmedEntry, entryLevel := s.trimAndGetLevel(entry)
		// if the entry level is one less than the current level
		if entryLevel == level-s.indentLevel {
			// we found the parent
			// we need to check if the parent is the root
			if entryLevel == 0 {
				// we are at the root
				return trimmedEntry + "." + trimedWord
			} else {
				// we are not at the root
				// we need to get the parent
				// and add the trimed word
				return s.getChainByIndex(entry) + "." + trimedWord
			}
		}
	}
	// we did not find a parent
	// so we are at the root
	return trimedWord
}

func (s *StructDef) trimAndGetLevel(str string) (string, int) {
	trimedWord := strings.TrimLeft(str, s.indentChars)
	if trimedWord == "" {
		return "", 0
	}
	level := (len(str) - len(trimedWord)) / len(s.indentChars)
	return trimedWord, level
}

func (s *StructDef) haveIntend(str string) bool {
	return strings.HasPrefix(str, s.indentChars) && strings.TrimLeft(str, s.indentChars) != ""
}

func (s *StructDef) AddToIndex(field ...string) {
	s.orderedIndexSlice = append(s.orderedIndexSlice, field...)
}

func (s *StructDef) GetOrderedIndexSlice() []string {
	return s.orderedIndexSlice
}

func (s *StructDef) SetIndexSlice(slice []string) {
	s.orderedIndexSlice = slice
}

func (s *StructDef) ResetIndexSlice() {
	s.orderedIndexSlice = nil
}

func (s *StructDef) GetFieldByTag(tag string) (StructField, error) {
	if !s.Init {
		return StructField{}, fmt.Errorf("structRead: struct not initialized")
	}
	return s.getFieldByTag(tag, s.Fields)
}

func (s *StructDef) getFieldByTag(tag string, from map[string]StructField) (StructField, error) {
	for _, field := range from {
		if field.OrginalTag.TagRenamed == tag {
			return field, nil
		}
		if s.haveIntend(tag) {
			if field.Children != nil {
				if f, err := s.getFieldByTag(tag, field.Children); err == nil {
					return f, nil
				}
			}
		}
	}
	return StructField{}, fmt.Errorf("structRead: field with tag [%s] not found", tag)
}

func (s *StructDef) GetFirstFieldByTag(tag string, from map[string]StructField) (StructField, error) {
	for _, field := range from {
		if field.OrginalTag.TagRenamed == tag {
			return field, nil
		}
		if field.Children != nil {
			if f, err := s.GetFirstFieldByTag(tag, field.Children); err == nil {
				return f, nil
			}
		}
	}
	return StructField{}, fmt.Errorf("structRead: field with tag [%s] not found", tag)
}

func (s *StructDef) SetAllowedTagSearch(allow bool) {
	s.allowBytagSearch = allow
}
