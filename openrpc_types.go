package openrpc

const (
	// https://json-schema.org/understanding-json-schema/reference/type.html
	stringSchema string = `{ "type": "string", "pattern": "(.*)" }`
	boolSchema   string = `{ "type": "boolean", "pattern": "(true|false)" }`
	// using [0-9] instead of \d because json returns an error
	integerSchema string = `{ "type": "integer", "pattern": "(^[0-9]*$)" }`
	numberSchema  string = `{ "type": "number", "pattern": "^([0-9]*\.[0-9]+)$|^([0-9]*)$" }`
	anySchema     string = `{}`
	nullSchema    string = `{ "type": "null" }`
)

type DocumentSpec1 struct {
	OpenRPC/* required */ string                  `json:"openrpc"`
	Info/* required */ *Info                      `json:"info"`
	Servers                         []*Server     `json:"servers,omitempty"`
	Methods/* required */ []*Method               `json:"methods"`
	Components                      *Components   `json:"components,omitempty"`
	ExternalDocs                    *ExternalDocs `json:"externalDocs,omitempty"`

	//Objects *ObjectMap `json:"-"`
}

type Info struct {
	Title/* required */ string            `json:"title"`
	Description                  string   `json:"description,omitempty"`
	TermsOfService               string   `json:"termsOfService,omitempty"`
	Contact                      Contact  `json:"contact,omitempty"`
	License                      *License `json:"license,omitempty"`
	Version/* required */ string          `json:"version"`
}

type Server struct {
	Name/* required */ string                            `json:"name"`
	URL/* required */ string                             `json:"url"`
	Summary                   string                     `json:"summary,omitempty"`
	Description               string                     `json:"description,omitempty"`
	Variables                 map[string]*ServerVariable `json:"variables,omitempty"`
}

type ServerVariable struct {
	Enum                         []string `json:"enum,omitempty"`
	Default/* required */ string          `json:"default"`
	Description                  string   `json:"description,omitempty"`
}

type Method struct {
	Name/* required */ string                                   `json:"name"`
	Tags                                      []Tag             `json:"tags,omitempty"`
	Summary                                   string            `json:"summary,omitempty"`
	Description                               string            `json:"description,omitempty"`
	ExternalDocs                              *ExternalDocs     `json:"externalDocs,omitempty"`
	Params/* required */ []*ContentDescriptor                   `json:"params"`
	Result/* required */ *ContentDescriptor                     `json:"result"`
	Deprecated                                bool              `json:"deprecated,omitempty"`
	Servers                                   []Server          `json:"servers,omitempty"`
	Errors                                    []Error           `json:"errors,omitempty"`
	Links                                     []Link            `json:"links,omitempty"`
	ParamStructure                            string            `json:"paramStructure,omitempty"`
	Examples                                  []*ExamplePairing `json:"examples,omitempty"`
}

type Components struct {
	ContentDescriptors    *SchemaRegistry `json:"contentDescriptors,omitempty"`
	Schemas               *SchemaRegistry `json:"schemas,omitempty"`
	Examples              *SchemaRegistry `json:"examples,omitempty"`
	Links                 *SchemaRegistry `json:"links,omitempty"`
	Errors                *SchemaRegistry `json:"errors,omitempty"`
	ExamplePairingObjects *SchemaRegistry `json:"examplePairingObjects,omitempty"`
	Tags                  *SchemaRegistry `json:"tags,omitempty"`
}

type ContentDescriptor struct {
	Name/* required */ string           `json:"name"`
	Summary                      string `json:"summary,omitempty"`
	Description                  string `json:"description,omitempty"`
	Required                     bool   `json:"required,omitempty"`
	Deprecated                   bool   `json:"deprecated,omitempty"`
	Schema/* required */ Pointer        `json:"schema"`
}

type ExternalDocs struct {
	Description              string `json:"description,omitempty"`
	URL/* required */ string        `json:"url"`
}

// Misc

type Contact struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

type License struct {
	Name/* required*/ string        `json:"name"`
	URL                      string `json:"url,omitempty"`
}

type Tag struct {
	Name/* required */ string               `json:"name"`
	Summary                   string        `json:"summary,omitempty"`
	Description               string        `json:"description,omitempty"`
	ExternalDocs              *ExternalDocs `json:"externalDocs,omitempty"`
}

type Error struct {
	Code/* required */ int                   `json:"code"`
	Message/* required */ string             `json:"message"`
	Data                         interface{} `json:"data,omitempty"`
}

type Link struct {
	Name/* required */ string                        `json:"name"`
	Description               string                 `json:"description,omitempty"`
	Summary                   string                 `json:"summary,omitempty"`
	Method                    string                 `json:"method,omitempty"`
	Params                    map[string]interface{} `json:"params,omitempty"`
	Server                    *Server                `json:"server,omitempty"`
}

type Example struct {
	Name          string      `json:"name,omitempty"`
	Summary       string      `json:"summary,omitempty"`
	Description   string      `json:"description,omitempty"`
	Value         interface{} `json:"value,omitempty"`
	ExternalValue string      `json:"externalValue,omitempty"`
}

type ExamplePairing struct {
	Name        string     `json:"name,omitempty"`
	Description string     `json:"description,omitempty"`
	Summary     string     `json:"summary,omitempty"`
	Params      []*Example `json:"params,omitempty"`
	Result      *Example   `json:"result,omitempty"`
}
