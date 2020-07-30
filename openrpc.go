//
package openrpc

import (
	"encoding"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/qri-io/jsonschema"
	"reflect"
)

// NewDocument populates with references the components obj
func NewDocument(methods []Method, info Info) *DocumentSpec1 {
	return &DocumentSpec1{
		OpenRPC:      "1.2",
		Info:         info,
		Servers:      nil,
		Methods:      methods,
		Components:   Components{},
		ExternalDocs: ExternalDocs{},
	}
}

// MakeSchema creates a json schema of t and un-marshals it into schema
func MakeSchema(t reflect.Type, schema Schema) error {
	switch t.Kind() {
	case reflect.Int:
	case reflect.Int8:
	case reflect.Int16:
	case reflect.Int32:
	case reflect.Int64:
	case reflect.Uint:
	case reflect.Uint8:
	case reflect.Uint16:
	case reflect.Uint32:
	case reflect.Uint64:
		return schema.UnmarshalJSON([]byte(`{ "type": "integer" }`))
	case reflect.Float32:
	case reflect.Float64:
		return schema.UnmarshalJSON([]byte(`{ "type": "number" }`))
	case reflect.Slice:
	case reflect.Array:
		return handleArraySchema(t, schema)
	case reflect.String:
		return schema.UnmarshalJSON([]byte(`{ "type": "string" }`))
	case reflect.Bool:
		return schema.UnmarshalJSON([]byte(`{ "type": "boolean" }`))
	case reflect.Map:
		return handleMapSchema(t, schema)
	case reflect.Struct:
		return handleStructSchema(t, schema)
	case reflect.Ptr:
		d := t.Elem()
		return MakeSchema(d, schema)
	case reflect.Interface:
		return schema.UnmarshalJSON([]byte("{}"))
	default:
		return errors.New("Invalid kind: " + t.Kind().String())
	}
	return nil
}

func handleStructSchema(t reflect.Type, schema Schema) error {

	s := &struct {
		T          string            `json:"type"`
		Properties map[string]Schema `json:"properties"`
	}{T: "object", Properties: make(map[string]Schema)}

	txt := reflect.TypeOf(encoding.TextUnmarshaler(nil))
	jsn := reflect.TypeOf(json.Unmarshaler(nil))

	ok := t.Implements(txt)
	if !ok {
		return errors.New(fmt.Sprintf("type %v (%v) does not implement or json.Unmarshaler", t.Name(), t.PkgPath()))
	}

	ok = t.Implements(jsn)
	if !ok {
		return errors.New(fmt.Sprintf("type %v (%v) does not implement json.Unmarshaler", t.Name(), t.PkgPath()))
	}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		sch := NewSchema()

		err := schema.UnmarshalJSON([]byte(`{ "type": "string"}`))
		if err != nil {
			return errors.New("error handling struct type: " + fmt.Sprintf("%v.%v field: %v :", t.PkgPath(), t.Name(), field.Name) + err.Error())
		}
		s.Properties[field.Name] = sch
	}

	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return schema.UnmarshalJSON(data)
}

func handleMapSchema(t reflect.Type, schema Schema) error {
	s := &struct {
		T          string                        `json:"type"`
		Properties map[string]*jsonschema.Schema `json:"properties"`
	}{T: "object", Properties: make(map[string]*jsonschema.Schema)}

	valueType := t.Elem()

	sch := NewSchema()
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

func handleArraySchema(t reflect.Type, schema Schema) error {
	s := &struct {
		T        string `json:"type"`
		Items    Schema `json:"items"`
		MaxItems int    `json:"maxItems,omitempty"`
	}{T: "array"}

	elemType := t.Elem()

	sch := NewSchema()
	err := MakeSchema(elemType, sch)
	if err != nil {
		return errors.New("error handling array type: " + err.Error())
	}

	if t.Kind() == reflect.Array {
		s.MaxItems = t.Len()
	}

	data, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return schema.UnmarshalJSON(data)
}
