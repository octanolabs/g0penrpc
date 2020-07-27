package openrpc

import (
	jptr "github.com/qri-io/jsonpointer"
	"testing"
)

type treeTestStruct struct{ t1, t2 *PointerTree }

var treeTestData = struct {
	first, second, third, fourth treeTestStruct
}{
	treeTestStruct{
		t1: &PointerTree{
			ptr: newPointerFromRefs([]string{}),
			nodes: map[string]*PointerTree{
				"field1": nil,
				"field2": nil,
				"field3": nil,
			},
		},
		t2: &PointerTree{
			ptr: newPointerFromRefs([]string{}),
			nodes: map[string]*PointerTree{
				"field1": nil,
				"field2": nil,
				"field3": nil,
			},
		}},
	treeTestStruct{
		t1: &PointerTree{
			ptr: newPointerFromRefs([]string{"root", "child"}),
			nodes: map[string]*PointerTree{
				"field1": nil,
				"field2": nil,
				"field3": nil,
			},
		},
		t2: &PointerTree{
			ptr: newPointerFromRefs([]string{"root", "child"}),
			nodes: map[string]*PointerTree{
				"field1": nil,
				"field2": nil,
				"field3": nil,
			},
		}},
	treeTestStruct{
		t1: &PointerTree{
			ptr: newPointerFromRefs([]string{"root"}),
			nodes: map[string]*PointerTree{
				"field1": nil,
				"field2": nil,
				"field3": nil,
			},
		},
		t2: &PointerTree{
			ptr: newPointerFromRefs([]string{"root", "child"}),
			nodes: map[string]*PointerTree{
				"field1": nil,
				"field2": nil,
				"field3": nil,
			},
		}},
	treeTestStruct{
		t1: &PointerTree{
			ptr: newPointerFromRefs([]string{"root", "child"}),
			nodes: map[string]*PointerTree{
				"field1": nil,
				"field3": nil,
			},
		},
		t2: &PointerTree{
			ptr: newPointerFromRefs([]string{"root", "child"}),
			nodes: map[string]*PointerTree{
				"field1": nil,
				"field2": nil,
			},
		}},
}

func TestEquality(t *testing.T) {

	t.Run("equalsEmptyPointer", func(t *testing.T) {

		data := treeTestData.first

		if !data.t1.equals(data.t2) {
			t.Errorf("error, should be equal, but is not: \n %v \n %v", data.t1, data.t2)
		}
	})

	t.Run("equalsWithPointer", func(t *testing.T) {

		data := treeTestData.second

		if !data.t1.equals(data.t2) {
			t.Errorf("error, should be equal, but is not: \n %v \n %v", data.t1, data.t2)
		}
	})

	t.Run("notEqualsDiffPointer", func(t *testing.T) {

		data := treeTestData.third

		if data.t1.equals(data.t2) {
			t.Errorf("error, should not be equal: \n %v \n %v", data.t1, data.t2)
		}
	})

	t.Run("notEqualsDiffChildren", func(t *testing.T) {

		data := treeTestData.fourth

		if data.t1.equals(data.t2) {
			t.Errorf("error, should not be equal: \n %v \n %v", data.t1, data.t2)
		}
	})

}

func TestFind(t *testing.T) {

	root := newPointerFromRefs([]string{"root"})
	parent := newPointerFromRefs([]string{"root", "parent"})
	child := newPointerFromRefs([]string{"root", "parent", "child"})
	subchild := newPointerFromRefs([]string{"root", "parent", "child", "subchild"})

	tree := NewPointerTree(newPointerFromRefs(nil))

	tree.Insert(root).
		Insert(parent).
		Insert(child)

	first := tree.Find(child)
	second := tree.Find(parent)
	third := tree.Find(root)
	fourth := tree.Find(subchild)

	if first.ptr.String() != child.String() {
		t.Errorf("error, got %v instead of %v", first, child)
	}

	if second.ptr.String() != parent.String() {
		t.Errorf("error, got %v instead of %v", second, second)
	}
	if third.ptr.String() != root.String() {
		t.Errorf("error, got %v instead of %v", third, root)
	}

	if fourth != nil {
		t.Errorf("error, got %v instead of nil", fourth)
	}

}

func TestInsert(t *testing.T) {

	data := treeTestData

	t.Run("insertionEmptyPointer", func(t *testing.T) {

		toTest := NewPointerTree(newPointerFromRefs(jptr.Pointer{})).
			Insert(newPointerFromRefs([]string{"field1"})).
			Insert(newPointerFromRefs([]string{"field2"})).
			Insert(newPointerFromRefs([]string{"field3"}))

		if !toTest.equals(data.first.t1) {
			t.Errorf("error, insertion failed: \n %v \n %v", toTest, data.first.t1)
		}
	})

	t.Run("insertionWithPointer", func(t *testing.T) {

		toTest := NewPointerTree(newPointerFromRefs([]string{"root", "child"})).
			Insert(newPointerFromRefs([]string{"field1"})).
			Insert(newPointerFromRefs([]string{"field2"})).
			Insert(newPointerFromRefs([]string{"field3"}))

		if !toTest.equals(data.second.t1) {
			t.Errorf("error, insertion failed: \n %v \n %v", toTest, data.second.t1)
		}
	})

	t.Run("insertionSameElement", func(t *testing.T) {

		toTest := NewPointerTree(newPointerFromRefs([]string{"root", "child"})).
			Insert(newPointerFromRefs([]string{"field1"})).
			Insert(newPointerFromRefs([]string{"field2"})).
			Insert(newPointerFromRefs([]string{"field3"}))

		toTest.Insert(newPointerFromRefs([]string{"field1"})).
			Insert(newPointerFromRefs([]string{"field2"})).
			Insert(newPointerFromRefs([]string{"field3"}))

		if !toTest.equals(data.second.t1) {
			t.Errorf("error, insertion failed: \n %v \n %v", toTest, data.second.t1)
		}
	})

}
