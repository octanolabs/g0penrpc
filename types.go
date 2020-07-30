package openrpc

import (
	"encoding/json"
	"fmt"
)

type JSONRegistry struct {
	reg   *PointerStore
	pTree *PointerTree
}

func (c *JSONRegistry) Register(p Pointer, item Schema) {
	c.reg.Set(p, item)
	c.pTree.Insert(p)
}

func (c *JSONRegistry) MarshalJSON() ([]byte, error) {

	j, err := c.pTree.ResolvePointers(c.reg)

	if err != nil {
		return nil, err
	}

	return json.Marshal(j)
}

func (c *JSONRegistry) String() string {

	bytes, _ := json.MarshalIndent(c, "", " ")

	return string(bytes)
}

var (
	integer = NewSchema()
	number  = NewSchema()
	str     = NewSchema()
	boolean = NewSchema()
)

// NewRegistry returns a new JSON registry with 4 basic schemas already registered
func NewRegistry() (*JSONRegistry, error) {

	reg := &JSONRegistry{reg: NewPointerRegistry(), pTree: NewPointerTree(newPointerFromRefs([]string{}))}

	err := integer.UnmarshalJSON([]byte(`{ "type": "integer" }`))
	err = number.UnmarshalJSON([]byte(`{ "type": "number" }`))
	err = str.UnmarshalJSON([]byte(`{ "type": "string" }`))
	err = boolean.UnmarshalJSON([]byte(`{ "type": "boolean" }`))

	if err != nil {
		return nil, err
	}

	p, err := NewPointer("#/components/schemas/integer")
	reg.Register(p, integer)

	p, err = NewPointer("#/components/schemas/number")
	reg.Register(p, number)

	p, err = NewPointer("#/components/schemas/str")
	reg.Register(p, str)

	p, err = NewPointer("#/components/schemas/boolean")
	reg.Register(p, boolean)

	return reg, nil
}

// PointerTree is used to represent the hierarchy of properties of a json object
type PointerTree struct {
	ptr   Pointer
	nodes map[string]*PointerTree
}

func NewPointerTree(ptr Pointer) *PointerTree {
	return &PointerTree{
		ptr:   ptr,
		nodes: map[string]*PointerTree{},
	}
}

func (pt *PointerTree) equals(opt *PointerTree) bool {

	pt1 := pt.ptr.Refs()
	pt2 := opt.ptr.Refs()

	if len(pt1) == 0 && len(pt2) == 0 {
		return true
	}

	if len(pt1) != len(pt2) {
		return false
	}

	for i := 0; i < len(pt1); i++ {
		if pt1[i] != pt2[i] {
			return false
		}
	}

	for key := range pt.nodes {
		if _, ok := opt.nodes[key]; !ok {
			return false
		}
	}

	//TODO: maybe compare pointers to other trees too?
	return true
}

//ResolvePointers recursively marshals a tree;
//if a tree has no children it is treated as a pointer and used to fetch a Schema from the registry
func (pt *PointerTree) ResolvePointers(reg *PointerStore) (json.RawMessage, error) {

	result := make(map[string]json.RawMessage)

	if len(pt.nodes) == 0 {
		fmt.Println(pt.ptr.String())
		b, err := reg.Get(pt.ptr).MarshalJSON()
		if err != nil {
			return nil, err
		}

		return b, nil
	}

	for prop, tree := range pt.nodes {

		s, err := tree.ResolvePointers(reg)
		if err != nil {
			return nil, err
		}
		result[prop] = s

	}

	b, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (pt *PointerTree) MarshalJSON() ([]byte, error) {
	return json.Marshal(pt.nodes)
}

func (pt *PointerTree) Insert(p Pointer) *PointerTree {

	elems := p.Refs()

	if len(elems) == 1 {
		pt.nodes[elems[0]] = NewPointerTree(p)
		return pt
	}

	var t *PointerTree

	t = pt.Find(newPointerFromRefs(elems[:len(elems)-1]))

	if t == nil {
		pt.Insert(newPointerFromRefs(elems[:len(elems)-1]))
		pt.Insert(newPointerFromRefs(elems))
	} else {
		t.nodes[elems[len(elems)-1]] = NewPointerTree(newPointerFromRefs(elems))
	}

	return pt
}

func (pt *PointerTree) Find(match Pointer) *PointerTree {

	elms := match.Refs()

	for _, item := range elms {
		if subTree, ok := pt.nodes[item]; ok {
			if len(elms) > 1 {
				return subTree.Find(newPointerFromRefs(elms[1:]))
			} else {
				return subTree
			}
		}
	}
	return nil
}

// PointerStore is a simple collection of json pointers
type PointerStore struct {
	m map[string]Schema
}

func (r *PointerStore) Set(p Pointer, item Schema) {
	if _, ok := r.m[p.String()]; !ok {
		r.m[p.String()] = item
	}
}
func (r *PointerStore) Get(p Pointer) Schema {
	return r.m[p.String()]
}

func NewPointerRegistry() *PointerStore {
	return &PointerStore{m: map[string]Schema{}}
}
