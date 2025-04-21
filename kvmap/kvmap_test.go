package kvmap

import (
	"fmt"
	"iter"
	"math/rand"
	"testing"
	"unsafe"
)

type testKey int

func (t testKey) HashBytes() []byte {
	size := int(unsafe.Sizeof(t))
	b := make([]byte, size)
	for i := 0; i < size; i++ {
		b[i] = byte(t & 0xff)
		t >>= 8
	}
	return b
}

func (t testKey) Equals(other testKey) bool {
	return t == other
}

func (t testKey) Before(other testKey) bool {
	return t < other
}

func TestKVMaps(t *testing.T) {
	tcs := []struct {
		name string
		m    IterableMap[testKey, string]
	}{
		{
			name: "ComparableLinkedHashMap",
			m:    NewComparableLinkedHashMap[testKey, string](Capacity(5), LoadFactor(1)),
		},
		{
			name: "HashableKeyLinkedHashMap",
			m:    NewHashableKeyLinkedHashMap[testKey, string](LoadFactor(.1)),
		},
		{
			name: "OrderedKeyTreeMap",
			m:    NewOrderedMap[testKey, string](),
		},
		{
			name: "OrderableKeyTreeMap",
			m:    NewOrderedMapWithOrderableKeys[testKey, string](),
		},
		{
			name: "MapWrapper",
			m:    NewMapWrapper[testKey, string](Capacity(0)),
		},
		{
			name: "BuiltinLinkedHashMap",
			m:    NewBuiltInLinkedHashMap[testKey, string](Capacity(5)),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			if !t.Run("Insertion", func(t *testing.T) {
				kvPairs := []struct {
					K testKey
					V string
				}{
					{-1, "negative one"},
					{2000000, "two million"},
					{5, "five"},
					{1, "one"},
					{16, "sixteen"},
					{-400, "negative four-hundred"},
					{-1, "minus one"},
					{80, "eighty"},
					{100, "one-hundred"},
					{0, "zero"},
				}

				for _, pair := range kvPairs {
					tc.m.Put(pair.K, pair.V)
					if !tc.m.Has(pair.K) {
						t.Errorf("Put(%d, %s); Want Has(%[1]d) == true, Got false", pair.K, pair.V)
					}
					if v, ok := tc.m.Get(pair.K); !ok || v != pair.V {
						t.Errorf("Put(%d, %s); Want Get(%[1]d) == (%s, true), Got (%s, %t)", pair.K, pair.V, v, ok)
					}
				}

				if l := tc.m.Len(); l != 9 {

					t.Errorf("Want Len() == 9, Got %d for map: %#v", l, tc.m)
				}
			}) {
				t.Skip("Insertion test failed... Skipping following tests")
			}

			t.Run("Deletion", func(t *testing.T) {
				keys := []testKey{5, -1, 5, 0}

				for _, k := range keys {
					tc.m.Delete(k)
					if tc.m.Has(k) {
						t.Errorf("Delete(%d); Want Has(%[1]d) == false, Got true", k)
					}
					if v, ok := tc.m.Get(k); ok || v != "" {
						t.Errorf(`Delete(%d); Want Get(%[1]d) == ("", false), Got (%q, %t)`, k, v, ok)
					}
				}

				if l := tc.m.Len(); l != 6 {
					t.Errorf("Want Len() == 6, Got %d; map: %#v", l, tc.m)
				}
			})
			t.Run("IteratorEntryValuesMutable", func(t *testing.T) {
				entriesInterface, ok := tc.m.(interface {
					Entries() iter.Seq[Entry[testKey, string]]
				})
				if !ok {
					t.Skipf("kvmap type %T does not support iterable Entries()", tc.m)
				}

				expected := map[testKey]string{
					2000000: "two million",
					1:       "one",
					16:      "sixteen",
					-400:    "negative four-hundred",
					80:      "eighty",
					100:     "one-hundred",
				}

				for entry := range entriesInterface.Entries() {
					k, v := entry.Key(), entry.Value()
					if v != expected[k] {
						t.Errorf("Want: Entry{%d, %s}, Got: {%[1]d, %[3]s}", k, expected[k], v)
					}
					entry.SetValue("mutated")
				}

				for key := range expected {
					if v, ok := tc.m.Get(key); !ok || v != "mutated" {
						t.Errorf(`Want Get(%d) == ("mutated", true), Got (%q, %t)`, key, v, ok)
					}
				}
			})
		})
	}
}

func BenchmarkLinkedHashMaps(b *testing.B) {
	for _, getNewMap := range []func() IterableMap[int, int]{
		func() IterableMap[int, int] { return NewComparableLinkedHashMap[int, int]() },
		func() IterableMap[int, int] { return NewBuiltInLinkedHashMap[int, int]() },
	} {
		b.Run(fmt.Sprintf("type-%T", getNewMap()), func(b *testing.B) {

			// Setup values to be inserted and deleted.
			rng := rand.New(rand.NewSource(0xDeadBeef))
			putVals := make([]int, b.N)
			delVals := make([]int, b.N)
			for i := range b.N {
				putVals[i] = rng.Intn(1000)
				delVals[i] = rng.Intn(1000)
			}

			testMap := getNewMap()
			b.ResetTimer()

			// Benchmark Put/Delete.
			for i := range b.N {
				testMap.Put(putVals[i], putVals[i])
				testMap.Delete(delVals[i])
			}
		})
	}
}
