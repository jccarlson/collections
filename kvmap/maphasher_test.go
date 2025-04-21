package kvmap

import (
	"fmt"
	"math"
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

type NonFixedSizeStructWithNonFixedSizeStructAndBlankField struct {
	//lint:ignore U1000 accessed via reflection
	a NonFixedSizeStructWithString
	_ uint16
}

type FixedSizeStructWithBlankField struct {
	_ uint16
	a uint64
	_ [5]uint16
	b [3]int32
}

func ComparableMapHasherTest[K comparable](v1, v2 K) func(t *testing.T) {
	return func(t *testing.T) {
		mh := ComparableMapHasher[K]()
		i, j, k := v1, v1, v2
		if v1 == v2 {
			t.Errorf("Expected v1 != v2; Got v1 == v2 (v1: %v, v2: %v)", v1, v2)
		}
		if h1, h2 := mh(&i), mh(&j); h1 != h2 {
			t.Errorf("Expected Hash(%v) == Hash(%v); Got Hash(%[1]v) == %[3]v, Hash(%[2]v) == %[4]v", i, j, h1, h2)
		}
		if h1, h2 := mh(&j), mh(&k); h1 == h2 {
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
	t.Run("non-fixed-size-struct-with-different-blank-fields", func(t *testing.T) {
		v1 := NonFixedSizeStructWithNonFixedSizeStructAndBlankField{a: NonFixedSizeStructWithString{a: "a"}}
		v2 := NonFixedSizeStructWithNonFixedSizeStructAndBlankField{a: NonFixedSizeStructWithString{a: "a"}}
		val := reflect.ValueOf(&v1).Elem().FieldByName("_")
		val = reflect.NewAt(val.Type(), val.Addr().UnsafePointer()).Elem()
		val.SetUint(8)
		val = reflect.ValueOf(&v2).Elem().FieldByName("_")
		val = reflect.NewAt(val.Type(), val.Addr().UnsafePointer()).Elem()
		val.SetUint(16)

		mh := ComparableMapHasher[NonFixedSizeStructWithNonFixedSizeStructAndBlankField]()

		if v1 != v2 {
			t.Errorf("Expected v1 == v2; Got v1 != v2 (v1: %v, v2: %v)", v1, v2)
		}
		if h1, h2 := mh(&v1), mh(&v2); h1 != h2 {
			t.Errorf("Expected Hash(%v) == Hash(%v); Got Hash(%[1]v) == %[3]v, Hash(%[2]v) == %[4]v", v1, v2, h1, h2)
		}
	})
	t.Run("fixed-size-struct-with-different-blank-fields", func(t *testing.T) {
		v1 := FixedSizeStructWithBlankField{a: 64, b: [3]int32{1, 2, 3}}
		v2 := FixedSizeStructWithBlankField{a: 64, b: [3]int32{1, 2, 3}}
		val := reflect.ValueOf(&v1).Elem().Field(0)
		val = reflect.NewAt(val.Type(), val.Addr().UnsafePointer()).Elem()
		val.SetUint(8)
		val = reflect.ValueOf(&v2).Elem().Field(0)
		val = reflect.NewAt(val.Type(), val.Addr().UnsafePointer()).Elem()
		val.SetUint(16)

		mh := ComparableMapHasher[FixedSizeStructWithBlankField]()

		if v1 != v2 {
			t.Errorf("Expected v1 == v2; Got v1 != v2 (v1: %v, v2: %v)", v1, v2)
		}
		if h1, h2 := mh(&v1), mh(&v2); h1 != h2 {
			t.Errorf("Expected Hash(%v) == Hash(%v); Got Hash(%[1]v) == %[3]v, Hash(%[2]v) == %[4]v", v1, v2, h1, h2)
		}
	})

	t.Run("float32-positive-negative-zero", func(t *testing.T) {
		v1 := float32(+0.0)
		v2 := float32(math.Copysign(0.0, -1.0))

		mh := ComparableMapHasher[float32]()

		if v1 != v2 {
			t.Errorf("Expected v1 == v2; Got v1 != v2 (v1: %v, v2: %v)", v1, v2)
		}
		if h1, h2 := mh(&v1), mh(&v2); h1 != h2 {
			t.Errorf("Expected Hash(%v) == Hash(%v); Got Hash(%[1]v) == %[3]v, Hash(%[2]v) == %[4]v", v1, v2, h1, h2)
		}

		if !math.Signbit(float64(v2)) {
			t.Errorf("Expected math.Signbit(float64(v2)) == true; Got false")
		}
	})

	t.Run("float64-positive-negative-zero", func(t *testing.T) {
		v1 := float64(+0.0)
		v2 := float64(math.Copysign(0.0, -1.0))

		mh := ComparableMapHasher[float64]()

		if v1 != v2 {
			t.Errorf("Expected v1 == v2; Got v1 != v2 (v1: %v, v2: %v)", v1, v2)
		}
		if h1, h2 := mh(&v1), mh(&v2); h1 != h2 {
			t.Errorf("Expected Hash(%v) == Hash(%v); Got Hash(%[1]v) == %[3]v, Hash(%[2]v) == %[4]v", v1, v2, h1, h2)
		}

		if !math.Signbit(v2) {
			t.Errorf("Expected math.Signbit(v2) == true; Got false")
		}
	})

	t.Run("float32-NaN-different-hashes", func(t *testing.T) {
		v := float32(math.NaN())

		mh := ComparableMapHasher[float32]()

		if v == v {
			t.Errorf("Expected v != v; Got v == v (v: %v)", v)
		}

		if h1, h2 := mh(&v), mh(&v); h1 == h2 {
			t.Errorf("Expected Hash(%v) != Hash(%v); Got Hash(%[1]v) == %[3]v, Hash(%[2]v) == %[4]v", v, v, h1, h2)
		}
	})

	t.Run("float64-NaN-different-hashes", func(t *testing.T) {
		v := float64(math.NaN())

		mh := ComparableMapHasher[float64]()

		if h1, h2 := mh(&v), mh(&v); h1 == h2 {
			t.Errorf("Expected Hash(%v) != Hash(%v); Got Hash(%[1]v) == %[3]v, Hash(%[2]v) == %[4]v", v, v, h1, h2)
		}
	})
}

func TestComparableMapHasherPanicsForNonComparableDynamicTypes(t *testing.T) {
	defer func() {
		msg := recover()
		expected := "runtime error: hash of unhashable type func()"
		if msg == nil || fmt.Sprint(msg) != expected {
			t.Errorf(`Expected panic(%q); Got panic("%v")`, expected, msg)
		}
	}()
	mh := ComparableMapHasher[struct{ a any }]()
	mh(&struct{ a any }{a: func() {}})
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
	if h1, h2 := mh(&v1), mh(&v2); h1 != h2 {
		t.Errorf("Expected Hash(%v) == Hash(%v); Got Hash(%[1]v) == %[3]v, Hash(%[2]v) == %[4]v", v1, v2, h1, h2)
	}

	if v1.Equals(v3) || v3.Equals(v1) {
		t.Errorf("Expected v1.Equals(v3) == false; Got true (v1: %v, v3: %v)", v1, v3)
	}
	if h1, h2 := mh(&v1), mh(&v3); h1 == h2 {
		t.Errorf("Expected Hash(%v) != Hash(%v); Got Hash(%[1]v) == Hash(%[2]v) == %v", v1, v3, h1)
	}

	if v2.Equals(v3) || v3.Equals(v2) {
		t.Errorf("Expected v2.Equals(v3) == false; Got true (v2: %v, v3: %v)", v2, v3)
	}
	if h1, h2 := mh(&v2), mh(&v3); h1 == h2 {
		t.Errorf("Expected Hash(%v) != Hash(%v); Got Hash(%[1]v) == Hash(%[2]v) == %v", v2, v3, h1)
	}
}
