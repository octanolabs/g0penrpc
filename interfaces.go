package openrpc

import (
	"encoding/json"
	jptr "github.com/qri-io/jsonpointer"
	jsch "github.com/qri-io/jsonschema"
)

// Pointer represents a generic json pointer
type Pointer interface {
	//Refs returns a slice containing all the (ordered) references of a pointer
	Refs() []string
	//String formats the references as a slash-separated string
	String() string

	json.Marshaler
}

type jsonPointer struct {
	p jptr.Pointer
}

func (jp *jsonPointer) Refs() []string {
	return jp.p
}

func (jp *jsonPointer) String() string {
	return jp.p.String()
}

func (jp *jsonPointer) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]string{"$ref": "#" + jp.String()})
}

func NewPointer(path string) (Pointer, error) {
	p, err := jptr.Parse(path)

	return &jsonPointer{p: p}, err
}

func newPointerFromRefs(refs []string) Pointer {

	if refs == nil {
		refs = []string{}
	}

	return &jsonPointer{p: refs}
}

//Schema is a json schema
type Schema interface {
	json.Marshaler
	json.Unmarshaler
}

type jsonSchema struct {
	jsch.Schema
}

func NewSchema() Schema {
	return &jsonSchema{Schema: jsch.Schema{}}
}
