package collections

import "testing"

func TestDeque(t *testing.T) {
	deque := &Deque[int]{}

	// initialize a Deque with the numbers 0..30, in order.
	for i := 0; i < 31; i++ {
		deque.AddLast(i)
	}

	t.Run("verifyInitialDeque", func(t *testing.T) {
		if s := deque.Size(); s != 31 {
			t.Errorf("Want deque.Size() == 31; Got %v", s)
		}
		if e, err := deque.Peek(); e != 0 || err != nil {
			t.Errorf("Want deque.Peek() == 0, nil; Got %v, %v", e, err)
		}
		if e, err := deque.PeekLast(); e != 30 || err != nil {
			t.Errorf("Want deque.PeekLast() == 31, nil; Got %v, %v", e, err)
		}
		for i := 0; i < 31; i++ {
			if e, err := deque.ElementAt(i); e != i || err != nil {
				t.Errorf("Want deque.ElementAt(%v) == %v, nil; Got %v, %v", i, i, e, err)
			}
		}
	})

	t.Run("Remove", func(t *testing.T) {
		for i := 0; i < 4; i++ {
			if e, err := deque.RemoveFirst(); e != i || err != nil {
				t.Errorf("Want deque.RemoveFirst() == %v, nil; Got %v, %v", i, e, err)
			}
		}
		for i := 30; i > 26; i-- {
			if e, err := deque.RemoveLast(); e != i || err != nil {
				t.Errorf("Want deque.RemoveLast() == %v, nil; Got %v, %v", i, e, err)
			}
		}
		if s := deque.Size(); s != 23 {
			t.Errorf("Want deque.Size() == 23; Got %v", s)
		}
	})

	t.Run("Add", func(t *testing.T) {
		for i := 0; i < 6; i++ {
			deque.AddFirst(i)
			if e, err := deque.Peek(); e != i || err != nil {
				t.Errorf("Want deque.Peek() == %v, nil; Got %v, %v", i, e, err)
			}
		}
		for i := -2; i < 0; i++ {
			deque.AddLast(i)
			if e, err := deque.PeekLast(); e != i || err != nil {
				t.Errorf("Want deque.PeekLast() == %v, nil; Got %v, %v", i, e, err)
			}
		}
		if s := deque.Size(); s != 31 {
			t.Errorf("Want deque.Size() == 31; Got %v", s)
		}
	})

	expectedState := [31]int{5, 4, 3, 2, 1, 0, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26, -2, -1}
	t.Run("Iterators", func(t *testing.T) {
		itElems := [31]int{}
		i := 0
		for e := range deque.All() {
			itElems[i] = e
			i++
		}
		if itElems != expectedState {
			t.Errorf("Want deque.All() == %v; Got %v", expectedState, itElems)
		}

		itElems = [31]int{}
		i = 30
		for e := range deque.Backwards() {
			itElems[i] = e
			i--
		}
		if itElems != expectedState {
			t.Errorf("Want deque.Backwards() == %v; Got %v", expectedState, itElems)
		}
	})

	t.Run("Errors", func(t *testing.T) {
		if s := deque.Size(); s != 31 {
			t.Errorf("Want deque.Size() == 31; Got %v", s)
		}

		for i := 0; i < 31; i++ {
			deque.RemoveLast()
		}

		if e, err := deque.Peek(); e != 0 || err == nil || err.Error() != "empty Deque" {
			t.Errorf("Want deque.Peek() == 0, error(\"empty Deque\"); Got %v, %v", e, err)
		}
		if e, err := deque.PeekLast(); e != 0 || err == nil || err.Error() != "empty Deque" {
			t.Errorf("Want deque.PeekLast() == 0, error(\"empty Deque\"); Got %v, %v", e, err)
		}
		if e, err := deque.RemoveFirst(); e != 0 || err == nil || err.Error() != "empty Deque" {
			t.Errorf("Want deque.RemoveFirst() == 0, error(\"empty Deque\"); Got %v, %v", e, err)
		}
		if e, err := deque.RemoveLast(); e != 0 || err == nil || err.Error() != "empty Deque" {
			t.Errorf("Want deque.RemoveLast() == 0, error(\"empty Deque\"); Got %v, %v", e, err)
		}
		if e, err := deque.ElementAt(1); e != 0 || err == nil || err.Error() != "index out of bounds: 1" {
			t.Errorf("Want deque.ElementAt(1) == 0, error(\"index out of bounds: 1\"); Got %v, %v", e, err)
		}
		if e, err := deque.ElementAt(-1); e != 0 || err == nil || err.Error() != "index out of bounds: -1" {
			t.Errorf("Want deque.ElementAt(-1) == 0, error(\"index out of bounds: -1\"); Got %v, %v", e, err)
		}

		if s := deque.Size(); s != 0 {
			t.Errorf("Want deque.Size() == 0; Got %v", s)
		}
	})
}
