package openrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// SchemaRegistry is a collection of Schemas
type SchemaRegistry struct {
	reg            *PointerStore
	pTree          *PointerTree
	unmarshalFrom  Pointer
	typeExceptions map[string]reflect.Type
}

var (
	integer = NewSchema()
	number  = NewSchema()
	str     = NewSchema()
	boolean = NewSchema()
	any     = NewSchema()
)

// NewSchemaRegistry returns a new JSON schema registry with 5 basic schemas already registered;
// the Pointer argument is used to select the subtree from which to start marshaling, can be nil
func NewSchemaRegistry(unmarshalFrom Pointer) (*SchemaRegistry, error) {

	reg := &SchemaRegistry{reg: NewPointerRegistry(), pTree: NewPointerTree(nil), unmarshalFrom: unmarshalFrom, typeExceptions: map[string]reflect.Type{}}

	reg.pTree.Insert(unmarshalFrom)

	err := integer.UnmarshalJSON([]byte(integerSchema))
	err = number.UnmarshalJSON([]byte(numberSchema))
	err = str.UnmarshalJSON([]byte(stringSchema))
	err = boolean.UnmarshalJSON([]byte(boolSchema))
	err = any.UnmarshalJSON([]byte(anySchema))

	if err != nil {
		return nil, err
	}

	p, _ := NewPointer(unmarshalFrom.String() + "/integer")
	reg.reg.Set(p, integer)
	reg.pTree.Insert(p)

	p, _ = NewPointer(unmarshalFrom.String() + "/number")
	reg.reg.Set(p, number)
	reg.pTree.Insert(p)

	p, _ = NewPointer(unmarshalFrom.String() + "/string")
	reg.reg.Set(p, str)
	reg.pTree.Insert(p)

	p, _ = NewPointer(unmarshalFrom.String() + "/bool")
	reg.reg.Set(p, boolean)
	reg.pTree.Insert(p)

	p, _ = NewPointer(unmarshalFrom.String() + "/anything")
	reg.reg.Set(p, any)
	reg.pTree.Insert(p)

	return reg, nil
}

func NewRegistry(unmarshalFrom Pointer) (*SchemaRegistry, error) {

	reg := &SchemaRegistry{reg: NewPointerRegistry(), pTree: NewPointerTree(nil), unmarshalFrom: unmarshalFrom, typeExceptions: map[string]reflect.Type{}}

	reg.pTree.Insert(unmarshalFrom)

	return reg, nil
}

//AddTypeException signals that the type t should always be represented as a string regardless what its kind is
func (s *SchemaRegistry) AddTypeException(typ reflect.Type) {

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	s.typeExceptions[formatTypeName(typ)] = typ
}

// Register creates a new schema for the provided type, and returns the Pointer by which it is referenced
func (s *SchemaRegistry) RegisterType(t reflect.Type, registerAsString bool) (Pointer, string, error) {

	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	// Order of if statements is important

	if t.Kind() == reflect.Struct && !s.isTypeException(t) {
		//TODO: refactor this to account for structs pointing to other things

		ptr, sch, name, err := s.createSchema(t)
		if err != nil {
			return nil, "", err
		}

		s.setSchema(ptr, sch)
		return ptr, name, nil
	} else if registerAsString || s.isTypeException(t) {
		ptr, name, err := s.registerAsString(t)
		if err != nil {
			return nil, "", err
		}

		return ptr, name, nil
	} else if t.Kind() == reflect.Slice || t.Kind() == reflect.Array {

		ptr, sch, name, err := s.handleSlice(t)

		if err != nil {
			return nil, "", err
		}

		s.setSchema(ptr, sch)

		return ptr, name, nil
	} else if t.Kind() == reflect.Map {

		ptr, sch, name, err := s.handleMap(t)

		if err != nil {
			return nil, "", err
		}

		s.setSchema(ptr, sch)

		return ptr, name, nil
	} else {

		ptr, sch, name, err := s.createSchema(t)
		if err != nil {
			return nil, "", err
		}

		s.setSchema(ptr, sch)
		return ptr, name, nil
	}
}

func (s *SchemaRegistry) MarshalJSON() ([]byte, error) {

	tree := s.pTree.Find(s.unmarshalFrom)
	if tree == nil {
		return nil, errors.New("unmarshalFrom pointer points to nil tree")
	}

	j, err := tree.ResolvePointers(s.reg)

	if err != nil {
		return nil, err
	}

	return json.Marshal(j)
}

func (s *SchemaRegistry) String() string {

	bytes, _ := json.MarshalIndent(s, "", " ")

	return string(bytes)
}

//IsTypeException reports whether type t is an exception
func (s *SchemaRegistry) isTypeException(typ reflect.Type) bool {

	for _, tt := range s.typeExceptions {
		if formatTypeName(typ) == formatTypeName(tt) {
			return true
		}
	}

	return false
}

func (s *SchemaRegistry) handleSlice(t reflect.Type) (pointer Pointer, schema Schema, name string, err error) {
	var (
		tPtr  Pointer
		tSch  Schema
		tName string
		tErr  error
	)

	t = t.Elem()

	tPtr, tSch, tName, tErr = s.createSchema(t)
	if tErr != nil {
		return nil, nil, "", tErr
	}

	if !s.isRegistered(tPtr) {
		s.setSchema(tPtr, tSch)
	}

	m := map[string]interface{}{
		"type":  "array",
		"items": tPtr,
	}

	if t.Kind() == reflect.Array {
		m["maxItems"] = t.Len()
	}

	bytes, err := json.Marshal(m)

	if err != nil {
		return nil, nil, "", err
	}

	sliceSchema := NewSchema()
	err = sliceSchema.UnmarshalJSON(bytes)
	if err != nil {
		return nil, nil, "", err
	}

	slicePtr, err := NewPointer(fmt.Sprintf("%s[]", tPtr))

	return slicePtr, sliceSchema, tName + "[]", nil
}
func (s *SchemaRegistry) handleMap(t reflect.Type) (Pointer, Schema, string, error) {
	var (
		ePtr  Pointer
		eSch  Schema
		eName string
		eErr  error
	)

	e := t.Elem()

	ePtr, eSch, eName, eErr = s.createSchema(e)
	if eErr != nil {
		return nil, nil, "", eErr
	}

	if !s.isRegistered(ePtr) {
		s.setSchema(ePtr, eSch)
	}

	//additionalProperties when marshaled to jsonschema is always the empty schema {}

	m := map[string]interface{}{
		"type": "object",
		"patternProperties": map[string]interface{}{
			"^.+$": ePtr,
		},
	}

	bytes, err := json.Marshal(m)

	if err != nil {
		return nil, nil, "", err
	}

	mapSchema := NewSchema()
	err = mapSchema.UnmarshalJSON(bytes)
	if err != nil {
		return nil, nil, "", err
	}

	mapName := fmt.Sprintf("Object[%s]", eName)
	mapPtr, err := NewPointer(fmt.Sprintf("%s/%s", s.unmarshalFrom.String(), mapName))

	return mapPtr, mapSchema, mapName, nil
}

// TODO
//func (s *SchemaRegistry) handleStruct() {}

func (s *SchemaRegistry) setSchema(ptr Pointer, sch Schema) {
	s.reg.Set(ptr, sch)
	s.pTree.Insert(ptr)
}

// RegisterString registers the provided type as a string, and returns the pointer by which it is referenced
func (s *SchemaRegistry) registerAsString(t reflect.Type) (ptr Pointer, name string, err error) {

	ptr, _, name, err = s.createSchema(t)
	if err != nil {
		return nil, "", err
	}

	sch := NewSchema()

	err = sch.UnmarshalJSON([]byte(stringSchema))

	s.setSchema(ptr, sch)

	return
}

func (s *SchemaRegistry) isRegistered(pointer Pointer) bool {
	_, ok := s.reg.Get(pointer)
	return ok
}

func (s *SchemaRegistry) createSchema(t reflect.Type) (pointer Pointer, schema Schema, name string, err error) {
	sch := NewSchema()

	err = MakeSchema(t, sch)
	if err != nil {
		return nil, nil, "", err
	}

	name = formatTypeName(t)

	p, err := NewPointer(s.unmarshalFrom.String() + "/" + name)
	if err != nil {
		return nil, nil, "", err
	}

	return p, sch, name, nil
}

func formatTypeName(t reflect.Type) string {

	name := t.Name()

	if name == "" {
		idx := strings.LastIndexFunc(t.String(), func(r rune) bool {
			return r == ']' || r == '*'
		})
		name = t.String()[idx+1:]
	}

	switch t.Kind() {
	case reflect.Map:
		e := t.Elem()

		if e.Kind() == reflect.Interface {
			return "Object[anything]"
		}

		return "object[" + name + "]"
	case reflect.Ptr:
		return formatTypeName(t.Elem())
	case reflect.Interface:
		return "anything"
	default:
		s := strings.Split(t.PkgPath(), "/")
		if l := len(s); l > 0 {
			if pkgName := s[l-1]; pkgName != "" {
				return pkgName + "." + name
			}
		}
		return name
	}
}
