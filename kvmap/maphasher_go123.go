//go:build !go1.24 || forcepolyfill

package kvmap

// This file is a polyfill for maphash.Compareable for use with go v1.23. It is
// obsolete as of go v1.24.

import (
	"fmt"
	"hash/maphash"
	"math/rand"
	"reflect"
	"unsafe"
)

// ComparableMapHasher returns a MapHasher for comparable keys which is
// consistent with the == operator.
//
// Specifically, with
// hash := ComparableMapHasher[T]()
// for two values v1, v2 of type T:
// v1 == v2 -> hash(v1) == hash(v2)
func ComparableMapHasher[K comparable]() MapHasher[K] {
	seed := maphash.MakeSeed()
	toBytes := comparableHashBytesFunc[K]()
	return func(key *K) uint64 {
		return maphash.Bytes(seed, toBytes(key))
	}
}

var zeroArray = [8]byte{0, 0, 0, 0, 0, 0, 0, 0}

func bytes1(v unsafe.Pointer) []byte {
	return unsafe.Slice((*byte)(v), 1)
}

func bytes2(v unsafe.Pointer) []byte {
	return unsafe.Slice((*byte)(v), 2)
}

func bytes4(v unsafe.Pointer) []byte {
	return unsafe.Slice((*byte)(v), 4)
}

func bytes8(v unsafe.Pointer) []byte {
	return unsafe.Slice((*byte)(v), 8)
}

func bytesFuncForSize(size uintptr) func(unsafe.Pointer) []byte {
	switch size {
	case 1:
		return bytes1
	case 2:
		return bytes2
	case 4:
		return bytes4
	case 8:
		return bytes8
	default:
		return func(v unsafe.Pointer) []byte {
			return unsafe.Slice((*byte)(v), size)
		}
	}
}

func bytesf32(v unsafe.Pointer) []byte {
	f32 := (*float32)(v)
	if *f32 != *f32 {
		// NaN != NaN, so we randomize the hash
		r := rand.Uint32()
		return unsafe.Slice((*byte)(unsafe.Pointer(&r)), 4)
	}
	if *f32 == 0.0 {
		// -0.0 and 0.0 should compare and hash the same.
		return zeroArray[:4]
	}
	return unsafe.Slice((*byte)(unsafe.Pointer(f32)), 4)
}

func bytesf64(v unsafe.Pointer) []byte {
	f64 := (*float64)(v)
	if *f64 != *f64 {
		// NaN != NaN, so we randomize the hash
		r := rand.Uint64()
		return unsafe.Slice((*byte)(unsafe.Pointer(&r)), 8)
	}
	if *f64 == 0.0 {
		// -0.0 and 0.0 should compare and hash the same.
		return zeroArray[:8]
	}

	return unsafe.Slice((*byte)(unsafe.Pointer(f64)), 8)
}

func bytesc64(v unsafe.Pointer) []byte {
	c64 := (*complex64)(v)
	r, i := real(*c64), imag(*c64)
	result := make([]byte, 8)
	copy(result, bytesf32(unsafe.Pointer(&r)))
	copy(result[4:], bytesf32(unsafe.Pointer(&i)))
	return result
}

func bytesc128(v unsafe.Pointer) []byte {
	c128 := (*complex128)(v)
	r, i := real(*c128), imag(*c128)
	result := make([]byte, 16)
	copy(result, bytesf64(unsafe.Pointer(&r)))
	copy(result[8:], bytesf64(unsafe.Pointer(&i)))
	return result
}

func bytesEmpty(v unsafe.Pointer) []byte {
	return []byte{}
}

// fieldData stores data for converting struct fields to byte arrays for hashing.
type fieldData struct {
	offset    uintptr
	bytesFunc func(unsafe.Pointer) []byte
}

// dummy is a global variable to force pointers used as map keys to escape to
// the heap, since pointers on the stack can change if the stack grows.
var dummy struct {
	b bool
	v any
}

func escape[T comparable](v T) {
	if dummy.b {
		dummy.v = v
	}
}

func comparableHashBytesFunc[T comparable]() func(*T) []byte {
	var v T
	t := reflect.TypeOf(v)
	if t == nil {
		// T is an interface type, and we have to use reflection to hash.
		return func(v *T) []byte {
			t := reflect.TypeOf(*v)
			return defaultHashBytesFunc(t)(unsafe.Pointer(v))
		}
	}

	bytesFunc := defaultHashBytesFunc(t)
	if t.Kind() == reflect.Pointer {
		return func(v *T) []byte {
			escape(v)
			return bytesFunc(unsafe.Pointer(v))
		}
	}
	return func(v *T) []byte {
		return bytesFunc(unsafe.Pointer(v))
	}
}

// defaultHashBytesFunc returns a value to byte-slice function for values of
// type T which is consistent with the == operator. The functions should not
// be exposed and the returned byte slices should never be modified, as they
// are often the allocated memory of the key reinterpreted as a []byte.
func defaultHashBytesFunc(t reflect.Type) func(unsafe.Pointer) []byte {
	if t == nil {
		return bytesEmpty
	}

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
		reflect.Pointer:
		size := t.Size()
		return bytesFuncForSize(size)

	case reflect.Float32:
		return bytesf32
	case reflect.Float64:
		return bytesf64
	case reflect.Complex64:
		return bytesc64
	case reflect.Complex128:
		return bytesc128

	case reflect.UnsafePointer,
		reflect.Chan:
		return func(v unsafe.Pointer) []byte {
			ptrbytes := uintptr(reflect.NewAt(t, v).Elem().UnsafePointer())
			return unsafe.Slice((*byte)(unsafe.Pointer(&ptrbytes)), unsafe.Sizeof(uintptr(0)))
		}

	case reflect.String:
		return func(v unsafe.Pointer) []byte {
			s := (*string)(v)
			return unsafe.Slice(unsafe.StringData(*s), len(*s))
		}

	case reflect.Array:
		l := uintptr(t.Len())

		// Check for 0-length Array types
		if l == 0 {
			return bytesEmpty
		}

		eType := t.Elem()
		eSize := eType.Size()
		eBytesFunc := defaultHashBytesFunc(eType)

		return func(v unsafe.Pointer) []byte {
			bytes := make([]byte, 0, l*eSize)
			for i := uintptr(0); i < l; i++ {
				bytes = append(bytes, eBytesFunc(unsafe.Pointer(uintptr(v)+i*eSize))...)
			}
			return bytes
		}

	case reflect.Struct:
		fields := make([]fieldData, 0, t.NumField())

		size := t.Size()

		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.Name == "_" {
				continue
			}
			fields = append(fields, fieldData{
				offset:    field.Offset,
				bytesFunc: defaultHashBytesFunc(field.Type),
			})
		}

		// Check for empty structs or structs with only blank fields.
		if len(fields) == 0 {
			return bytesEmpty
		}

		return func(v unsafe.Pointer) []byte {
			r := make([]byte, 0, size)
			for _, field := range fields {
				r = append(r, field.bytesFunc(unsafe.Pointer(uintptr(unsafe.Pointer(v))+field.offset))...)
			}
			return r
		}

	case reflect.Interface:
		return func(v unsafe.Pointer) []byte {
			if v == nil {
				return []byte{}
			}
			val := reflect.NewAt(t, v).Elem()
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
			return defaultHashBytesFunc(val.Type())(ptrToValCopy.UnsafePointer())
		}
	}

	panic(fmt.Sprintf("runtime error: hash of unhashable type %v", t))
}
