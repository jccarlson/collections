package kvmap

import (
	"fmt"
	"hash/maphash"
	"reflect"
	"unsafe"

	"github.org/jccarlson/collections/compare"
)

// A MapHasher wraps a key-serialization function, and is designed to be used
// to hash map keys.
type MapHasher[K any] struct {
	seed    maphash.Seed
	toBytes func(*K) []byte
}

func (m MapHasher[K]) Hash(key *K) uint64 {
	return maphash.Bytes(m.seed, m.toBytes(key))
}

// HashableKey is a compare.Equaler with a HashBytes() method. HashBytes()
// should return a byte-slice representation of the wrapped value which is
// consistent with Equals().
//
// Specifically, for two values v1, v2:
// v1.Equals(v2) -> bytes.Equal(v1.HashBytes(), v2.HashBytes())
// !bytes.Equal(v1.HashBytes(),v2.HashBytes()) -> !v1.Equals(v2)
type HashableKey[T any] interface {
	compare.Equalable[T]

	HashBytes() []byte
}

// HashableKeyMapHasher returns a MapHasher for HashableKey types.
func HashableKeyMapHasher[K HashableKey[K]]() MapHasher[K] {
	return MapHasher[K]{
		seed: maphash.MakeSeed(),
		toBytes: func(key *K) []byte {
			return (*key).HashBytes()
		},
	}
}

// ComparableMapHasher returns a MapHasher for comparable keys, where Hash()
// is consistent with the == operator.
func ComparableMapHasher[K comparable]() MapHasher[K] {
	return MapHasher[K]{
		seed:    maphash.MakeSeed(),
		toBytes: defaultHashBytesFunc[K](),
	}
}

// CustomMapHasher returns a MapHasher for any key type. Users must provide a
// serialization function that takes a pointer to a key and returns a
// byte-slice representation of the key which is consistent with the comparison
// implementation used.
func CustomMapHasher[K any](toBytes func(*K) []byte) MapHasher[K] {
	return MapHasher[K]{
		seed:    maphash.MakeSeed(),
		toBytes: toBytes,
	}
}

// isFixedSize returns true if values of comparable type t take a fixed-size
// contiguous block of memory for the purpose of hashing consistent with the ==
// operator for use as map keys.
func isFixedSize(t reflect.Type) bool {
	switch t.Kind() {
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128,
		reflect.Chan,
		// We consider pointers to be 'fixed size' because they compare equal
		// only if they point to the same variable, so we don't need to hash
		// what they point at, just the address.
		reflect.Pointer,
		reflect.UnsafePointer:
		return true

	case reflect.Array:
		// Array types are fixed size if their elements are fixed size.
		return isFixedSize(t.Elem())

	case reflect.Struct:
		// Structs are fixed size if all their fields are fixed size.
		for i := 0; i < t.NumField(); i++ {
			if !isFixedSize(t.Field(i).Type) {
				return false
			}
		}
		return true
	}
	// Strings are compared lexigraphically and interfaces can wrap various
	// dynamic types which have different memory layouts, so they are not
	// fixed size.
	// Other Kinds (Function, Map, Slice) are not comparable anyway so we don't
	// care.
	return false
}

// defaultHashBytesFunc returns a value to byte-slice function for values of
// type T which is consistent with the == operator. The functions should not
// be exposed and the returned byte slices should never be modified, as they
// are often the allocated memory of the key reinterpreted as a []byte.
func defaultHashBytesFunc[T comparable]() func(*T) []byte {
	var v T
	t := reflect.TypeOf(v)

	if t == nil {
		// T is an interface type, and we have to do reflection to hash.
		return deepHashBytes[T]
	}

	// T is a concrete type
	switch t.Kind() {
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128,
		reflect.Pointer:
		size := t.Size()
		return func(v *T) []byte {
			return unsafe.Slice((*byte)(unsafe.Pointer(v)), size)
		}

	case reflect.UnsafePointer,
		reflect.Chan:
		return func(v *T) []byte {
			ptrbytes := uintptr(reflect.ValueOf(v).Elem().UnsafePointer())
			return unsafe.Slice((*byte)(unsafe.Pointer(&ptrbytes)), unsafe.Sizeof(uintptr(0)))
		}

	case reflect.String:
		return func(v *T) []byte {
			s := (*string)(unsafe.Pointer(v))
			return unsafe.Slice(unsafe.StringData(*s), len(*s))
		}

	case reflect.Array:
		// Check for 0-length Array types
		if t.Len() == 0 {
			return func(v *T) []byte {
				return []byte{}
			}
		}
		// If the array's elements are fixed-size, we can just reinterpret the
		// whole array as a byte array.
		if isFixedSize(t) {
			size := t.Size()
			return func(v *T) []byte {
				return unsafe.Slice((*byte)(unsafe.Pointer(v)), size)
			}
		}
		// Otherwise (e.g. for string or interface elements), we need to do a
		// deep hash via reflection.
		return deepHashBytes[T]

	case reflect.Struct:
		// Check for empty struct types
		if t.Size() == 0 || t.NumField() == 0 {
			return func(v *T) []byte {
				return []byte{}
			}
		}
		// If the structs's fields are fixed-size, we can just reinterpret the
		// whole struct as a byte array.
		if isFixedSize(t) {
			size := t.Size()
			return func(v *T) []byte {
				return unsafe.Slice((*byte)(unsafe.Pointer(v)), size)
			}
		}
		// Otherwise (e.g. for string or interface fields), we need to do a
		// deep hash via reflection.
		return deepHashBytes[T]
	}
	panic("T is not a comparable type")
}

func deepHashBytes[T comparable](v *T) []byte {
	if v == nil {
		return []byte{}
	}

	// We use reflect.ValueOf(v).Elem() instead of reflect.ValueOf(*v) so that
	// all recursed values are addressable.
	return deepHashBytesRecur(reflect.ValueOf(v).Elem())
}

func deepHashBytesRecur(val reflect.Value) []byte {
	switch val.Kind() {
	case reflect.Bool,
		reflect.Int,
		reflect.Int8,
		reflect.Int16,
		reflect.Int32,
		reflect.Int64,
		reflect.Uint,
		reflect.Uint8,
		reflect.Uint16,
		reflect.Uint32,
		reflect.Uint64,
		reflect.Uintptr,
		reflect.Float32,
		reflect.Float64,
		reflect.Complex64,
		reflect.Complex128,
		reflect.Pointer:
		size := val.Type().Size()
		return unsafe.Slice((*byte)(val.Addr().UnsafePointer()), size)

	case reflect.UnsafePointer,
		reflect.Chan:
		ptrbytes := uintptr(val.Addr().UnsafePointer())
		return unsafe.Slice((*byte)(unsafe.Pointer(&ptrbytes)), unsafe.Sizeof(uintptr(0)))

	case reflect.String:
		s := val.String()
		return unsafe.Slice(unsafe.StringData(s), len(s))

	case reflect.Array:
		b := []byte{}
		for i := 0; i < val.Len(); i++ {
			b = append(b, deepHashBytesRecur(val.Index(i))...)
		}
		return b

	case reflect.Struct:
		b := []byte{}
		for i := 0; i < val.NumField(); i++ {
			b = append(b, deepHashBytesRecur(val.Field(i))...)
		}
		return b

	case reflect.Interface:
		if val.IsNil() {
			return []byte{}
		}

		// val is addressable, but may be derived from an unexported struct
		// field. If so, we force it to be settable.
		if !val.CanSet() {
			val = reflect.NewAt(val.Type(), val.Addr().UnsafePointer()).Elem()
		}

		// Values contained in interfaces aren't addressable, so we create a
		// pointer to a value of val's dynamic type, then copy val into it, so
		// that the recursed value remains addressable.
		val = val.Elem()
		ptrToValCopy := reflect.New(val.Type())
		ptrToValCopy.Elem().Set(val)
		return deepHashBytesRecur(ptrToValCopy.Elem())
	}
	panic(fmt.Sprintf("Dynamic type %T is not comparable", val.Interface()))
}
