//
package openrpc

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
)

// NewDocument populates with references the components obj
func NewDocument(methods []*Method, info *Info) *DocumentSpec1 {
	return &DocumentSpec1{
		OpenRPC:      "1.2.0",
		Info:         info,
		Servers:      nil,
		Methods:      methods,
		Components:   &Components{},
		ExternalDocs: nil,
	}
}

//TODO:
// - figure out why integer and number schemas are empty;

func MakeSchema(t reflect.Type, schema Schema) error {
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return schema.UnmarshalJSON([]byte(integerSchema))
	case reflect.Float32, reflect.Float64:
		return schema.UnmarshalJSON([]byte(numberSchema))
	case reflect.Slice, reflect.Array:
		return buildArraySchema(t, schema)
	case reflect.String:
		return schema.UnmarshalJSON([]byte(stringSchema))
	case reflect.Bool:
		return schema.UnmarshalJSON([]byte(boolSchema))
	case reflect.Map:
		return buildMapSchema(t, schema)
	case reflect.Struct:
		return buildStructSchema(t, schema)
	case reflect.Ptr:
		d := t.Elem()
		return MakeSchema(d, schema)
	case reflect.Interface:
		return schema.UnmarshalJSON([]byte(anySchema))
	default:
		return errors.New("Invalid kind: " + t.Kind().String())
	}
}

func buildStructSchema(t reflect.Type, schema Schema) error {

	s := map[string]interface{}{
		"type":       "object",
		"properties": nil,
	}

	//methods with pointer receiver should be tested against pointer to structs/etc

	props := map[string]Schema{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		sch := NewSchema()

		err := sch.UnmarshalJSON([]byte(stringSchema))
		if err != nil {
			return errors.New("error handling struct type: " + fmt.Sprintf("%v.%v field: %v :", t.PkgPath(), t.Name(), field.Name) + err.Error())
		}
		props[field.Name] = sch
	}

	s["properties"] = props

	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return schema.UnmarshalJSON(data)
}

// Map and array schemas should be handled with references inside the registry

func buildMapSchema(t reflect.Type, schema Schema) error {

	sch := NewSchema()

	s := map[string]interface{}{
		"type":                 "object",
		"additionalProperties": sch,
	}
	valueType := t.Elem()

	err := MakeSchema(valueType, sch)
	if err != nil {
		return errors.New("error handling map type: " + err.Error())
	}

	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return schema.UnmarshalJSON(data)
}

func buildArraySchema(t reflect.Type, schema Schema) error {

	sch := NewSchema()

	s := map[string]interface{}{
		"type":  "array",
		"items": sch,
	}

	if t.Kind() == reflect.Array {
		s["maxItems"] = t.Len()
	}

	elemType := t.Elem()

	// TODO: this function does not check if the type in question is marshalable/unmarshalable
	err := MakeSchema(elemType, sch)
	if err != nil {
		return errors.New("error handling array type: " + err.Error())
	}

	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return schema.UnmarshalJSON(data)
}
