package kvmap

import (
	"fmt"
	"reflect"
	"testing"

	"golang.org/x/exp/constraints"
)

type Embedded struct{ a uint }

func (e Embedded) Get() uint {
	return e.a
}

type Embedded2 struct{ a uint8 }

func (e Embedded2) Get() uint {
	return uint(e.a)
}

type EmbeddedInterface interface {
	Get() uint
}

type FixedSizeStruct struct {
	Embedded
	a [4]uint16
	b int64
	//lint:ignore U1000 accessed via reflection
	c *string
	_ chan uint
	//lint:ignore U1000 accessed via reflection
	d complex128
	//lint:ignore U1000 accessed via reflection
	e struct{ a int }
}

type FixedSizeEmptyStruct struct{}

type NonFixedSizeStructWithInterface struct {
	//lint:ignore U1000 accessed via reflection
	a interface{ Get() uint }
}

type NonFixedSizeStructWithString struct {
	//lint:ignore U1000 accessed via reflection
	a string
}

type NonFixedSizeStructWithNonFixedSizeStruct struct {
	//lint:ignore U1000 accessed via reflection
	a NonFixedSizeStructWithString
}

type NonFixedSizeStructWithNonFixedSizeArray struct {
	//lint:ignore U1000 accessed via reflection
	a [4]NonFixedSizeStructWithInterface
}

type NonFixedSizeStructWithEmbeddedInterface struct {
	EmbeddedInterface
}

type NonFixedSizeStructWithEmbeddedNonFixedSizedStruct struct {
	NonFixedSizeStructWithInterface
}

func TestIsFixedSize(t *testing.T) {
	var k any = FixedSizeStruct{}
	if !isFixedSize(reflect.TypeOf(k)) {
		t.Errorf("Expected type %T to be considered fixed size", k)
	}

	k = FixedSizeEmptyStruct{}
	if !isFixedSize(reflect.TypeOf(k)) {
		t.Errorf("Expected type %T to be considered fixed size", k)
	}

	k = [4]int{}
	if !isFixedSize(reflect.TypeOf(k)) {
		t.Errorf("Expected type %T to be considered fixed size", k)
	}

	k = (*[]int)(nil)
	if !isFixedSize(reflect.TypeOf(k)) {
		t.Errorf("Expected type %T to be considered fixed size", k)
	}

	k = [4]string{}
	if isFixedSize(reflect.TypeOf(k)) {
		t.Errorf("Expected type %T to be considered non-fixed size", k)
	}

	k = NonFixedSizeStructWithInterface{}
	if isFixedSize(reflect.TypeOf(k)) {
		t.Errorf("Expected type %T to be considered non-fixed size", k)
	}

	k = NonFixedSizeStructWithString{}
	if isFixedSize(reflect.TypeOf(k)) {
		t.Errorf("Expected type %T to be considered non-fixed size", k)
	}

	k = NonFixedSizeStructWithNonFixedSizeStruct{}
	if isFixedSize(reflect.TypeOf(k)) {
		t.Errorf("Expected type %T to be considered non-fixed size", k)
	}

	k = NonFixedSizeStructWithEmbeddedInterface{}
	if isFixedSize(reflect.TypeOf(k)) {
		t.Errorf("Expected type %T to be considered non-fixed size", k)
	}

	k = NonFixedSizeStructWithEmbeddedNonFixedSizedStruct{}
	if isFixedSize(reflect.TypeOf(k)) {
		t.Errorf("Expected type %T to be considered non-fixed size", k)
	}

	k = FixedSizeStruct{}
	if isFixedSize(reflect.TypeOf(&k).Elem()) {
		// reflect.TypeOf(&k).Elem() is type any (an Interface), which is
		// non-fixed size even though we know the dynamic type is fixed size.
		t.Errorf("Expected type %T to be considered non-fixed size", k)
	}
}

func ComparableMapHasherTest[K comparable](v1, v2 K) func(t *testing.T) {
	return func(t *testing.T) {
		mh := ComparableMapHasher[K]()
		i, j, k := v1, v1, v2
		if v1 == v2 {
			t.Errorf("Expected v1 != v2; Got v1 == v2 (v1: %v, v2: %v)", v1, v2)
		}
		if h1, h2 := mh.Hash(&i), mh.Hash(&j); h1 != h2 {
			t.Errorf("Expected Hash(%v) == Hash(%v); Got Hash(%[1]v) == %[3]v, Hash(%[2]v) == %[4]v", i, j, h1, h2)
		}
		if h1, h2 := mh.Hash(&j), mh.Hash(&k); h1 == h2 {
			t.Errorf("Expected Hash(%v) != Hash(%v); Got Hash(%[1]v) == Hash(%[2]v) == %v", j, k, h1)
		}
	}
}

func TestComparableMapHasher(t *testing.T) {
	t.Run("int", ComparableMapHasherTest(1, 2))
	t.Run("complex128", ComparableMapHasherTest(2-4i, 3.6+7.9i))
	t.Run("string", ComparableMapHasherTest("abc", "aba"))
	t.Run("fixed-size-array", ComparableMapHasherTest([2]uint16{0, 1}, [2]uint16{1, 2}))
	t.Run("non-fixed-size-array", ComparableMapHasherTest([2]string{"abc", "def"}, [2]string{"abc", "abc"}))
	t.Run("pointer", ComparableMapHasherTest(&struct {
		a int
		b float32
	}{4, 1.0}, &struct {
		a int
		b float32
	}{4, 1.0}))
	t.Run("interface", ComparableMapHasherTest(interface{ Get() uint }(Embedded{a: 2}), interface{ Get() uint }(Embedded2{a: 2})))
	t.Run("nil-interface", ComparableMapHasherTest(interface{ Get() uint }(Embedded{a: 2}), interface{ Get() uint }(nil)))
	t.Run("fixed-size-struct", ComparableMapHasherTest(FixedSizeStruct{Embedded: Embedded{a: 1}, b: 2}, FixedSizeStruct{a: [4]uint16{1}, b: 2}))
	t.Run("non-fixed-size-struct", ComparableMapHasherTest(NonFixedSizeStructWithEmbeddedInterface{Embedded{a: 2}}, NonFixedSizeStructWithEmbeddedInterface{Embedded2{a: 2}}))
	t.Run("struct-with-unexported-interface", ComparableMapHasherTest(struct{ a any }{a: "a"}, struct{ a any }{a: 1.0}))
	t.Run("chan", ComparableMapHasherTest(make(chan int), make(chan int)))
}

func TestComparableMapHasherPanicsForNonComparableDynamicTypes(t *testing.T) {
	defer func() {
		msg := recover()
		expected := "Dynamic type func() is not comparable"
		if msg == nil || fmt.Sprint(msg) != expected {
			t.Errorf(`Expected panic(%q); Got panic("%v")`, expected, msg)
		}
	}()
	mh := ComparableMapHasher[struct{ a any }]()
	mh.Hash(&struct{ a any }{a: func() {}})
}

type SIntWrapper[T constraints.Signed] struct {
	i T
}

func (i SIntWrapper[T]) Int() int64 {
	return int64(i.i)
}

type SInt interface {
	Int() int64
}

type IntKey struct {
	SInt
}

func (k IntKey) Equals(other IntKey) bool {
	return k.Int() == other.Int()
}

func (k IntKey) HashBytes() []byte {
	r := make([]byte, 0, 8)
	const mask = 0xFF
	for i := 0; i < 8; i++ {
		r = append(r, byte((k.Int()>>(i*8))&mask))
	}
	return r
}

func TestHashableKeyMapHasher(t *testing.T) {
	mh := HashableKeyMapHasher[IntKey]()
	v1, v2, v3 := IntKey{SIntWrapper[int16]{i: int16(1023)}}, IntKey{SIntWrapper[int32]{i: int32(1023)}}, IntKey{SIntWrapper[int64]{i: int64(1024)}}

	if !v1.Equals(v2) || !v2.Equals(v1) {
		t.Errorf("Expected v1.Equals(v2) == true; Got false (v1: %v, v2: %v)", v1, v2)
	}
	if h1, h2 := mh.Hash(&v1), mh.Hash(&v2); h1 != h2 {
		t.Errorf("Expected Hash(%v) == Hash(%v); Got Hash(%[1]v) == %[3]v, Hash(%[2]v) == %[4]v", v1, v2, h1, h2)
	}

	if v1.Equals(v3) || v3.Equals(v1) {
		t.Errorf("Expected v1.Equals(v3) == false; Got true (v1: %v, v3: %v)", v1, v3)
	}
	if h1, h2 := mh.Hash(&v1), mh.Hash(&v3); h1 == h2 {
		t.Errorf("Expected Hash(%v) != Hash(%v); Got Hash(%[1]v) == Hash(%[2]v) == %v", v1, v3, h1)
	}

	if v2.Equals(v3) || v3.Equals(v2) {
		t.Errorf("Expected v2.Equals(v3) == false; Got true (v2: %v, v3: %v)", v2, v3)
	}
	if h1, h2 := mh.Hash(&v2), mh.Hash(&v3); h1 == h2 {
		t.Errorf("Expected Hash(%v) != Hash(%v); Got Hash(%[1]v) == Hash(%[2]v) == %v", v2, v3, h1)
	}
}
