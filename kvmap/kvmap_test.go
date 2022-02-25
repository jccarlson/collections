package kvmap

import (
	"testing"
	"unsafe"
)

func init() {
	baseTableCap = 1 << 3 // lower this to test resizing easier.
}

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

func TestKVMaps(t *testing.T) {
	tcs := []struct {
		name string
		m    Interface[testKey, string]
	}{{
		name: "LinkedHashMap",
		m:    NewComparableLinkedHashMap[testKey, string](),
	},
		{
			name: "TreeMap",
			m:    NewOrderedTreeMap[testKey, string](),
		},
		{
			name: "MapWrapper",
			m:    MapWrapper[testKey, string](make(map[testKey]string)),
		}}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
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
					t.Errorf("Want Len() == 9, Got %d", l)
				}
			}) {
				t.Skip("Insertion test failed... Skipping deletion test")
			}

			t.Run("Deletion", func(t *testing.T) {
				keys := []testKey{5, -1, 0}

				for _, k := range keys {
					tc.m.Delete(k)
					if tc.m.Has(k) {
						t.Errorf("Delete(%d); Want Has(%[1]d) == false, Got true", k)
					}
					if v, ok := tc.m.Get(k); ok || v != "" {
						t.Errorf(`Delete(%d); Want Get(%[1]d) == ("", false), Got (%s, %t)`, k, v, ok)
					}
				}

				if l := tc.m.Len(); l != 6 {
					t.Errorf("Want Len() == 6, Got %d", l)
				}
			})
		})
	}
}
