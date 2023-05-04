package ds

import (
	"fmt"
	"math/rand"
	"testing"

	"github.org/jccarlson/collections/compare"
)

func validateTree(n *TreeNode[int]) (blackHeight int, err error) {
	if n == nil {
		return 1, nil
	}

	if n.isRed() {
		// ensure neither of n's children are red.
		if n.child[Left].isRed() {
			return 0, fmt.Errorf("Node @ %p with elem: %v is red with red left child @ %p with elem: %v", n, n.Elem, n.child[Left], n.child[Left].Elem)
		}
		if n.child[Right].isRed() {
			return 0, fmt.Errorf("Node @ %p with elem: %v is red with red right child @ %p with elem: %v", n, n.Elem, n.child[Right], n.child[Right].Elem)
		}
	}

	// Validate child-parent links and ordering.
	if n.child[Left] != nil {
		if n.Elem <= n.child[Left].Elem {
			return 0, fmt.Errorf("Node @ %p with elem: %v has left child @ %p with elem: %v, which is >= %[2]v", n, n.Elem, n.child[Left], n.child[Left].Elem)
		}
		if n.child[Left].parent != n {
			return 0, fmt.Errorf("Node @ %p with elem: %v has left child @ %p with elem: %v with parent @ %p with elem %v", n, n.Elem, n.child[Left], n.child[Left].Elem, n.child[Left].parent, n.child[Left].parent.Elem)
		}
	}
	if n.child[Right] != nil {
		if n.Elem >= n.child[Right].Elem {
			return 0, fmt.Errorf("Node @ %p with elem: %v has right child @ %p with elem: %v, which is <= %[2]v", n, n.Elem, n.child[Right], n.child[Right].Elem)
		}
		if n.child[Right].parent != n {
			return 0, fmt.Errorf("Node @ %p with elem: %v has right child @ %p with elem: %v with parent @ %p with elem %v", n, n.Elem, n.child[Right], n.child[Right].Elem, n.child[Right].parent, n.child[Right].parent.Elem)
		}
	}

	// Validate subtrees.
	bhLeft, err := validateTree(n.child[Left])
	if err != nil {
		return
	}
	bhRight, err := validateTree(n.child[Right])
	if err != nil {
		return
	}

	// Validate black-height of subtrees.
	if bhLeft != bhRight {
		return 0, fmt.Errorf("Node @ %p with elem: %v has left sub-tree with black-height == %v and right sub-tree with black-height == %v", n, n.Elem, bhLeft, bhRight)
	}

	// Return n's black-height.
	blackHeight = bhLeft
	if n.isBlack() {
		blackHeight++
	}
	return
}

func TestAllBlackPerfectTreeDelete(t *testing.T) {
	// Manually construct a perfect binary tree with all black nodes. By
	// definition, this is a valid red-black tree.
	rbTree := &RedBlackTree[int]{Ordering: compare.Less[int]}
	rbTree.root = &TreeNode[int]{Elem: 4, black: true}
	rbTree.root.child[Left] = &TreeNode[int]{Elem: 2, parent: rbTree.root, black: true}
	rbTree.root.child[Left].child[Left] = &TreeNode[int]{Elem: 1, parent: rbTree.root.child[Left], black: true}
	rbTree.root.child[Left].child[Right] = &TreeNode[int]{Elem: 3, parent: rbTree.root.child[Left], black: true}
	rbTree.root.child[Right] = &TreeNode[int]{Elem: 6, parent: rbTree.root, black: true}
	rbTree.root.child[Right].child[Left] = &TreeNode[int]{Elem: 5, parent: rbTree.root.child[Right], black: true}
	rbTree.root.child[Right].child[Right] = &TreeNode[int]{Elem: 7, parent: rbTree.root.child[Right], black: true}

	rbTree.first, rbTree.last = rbTree.root.child[Left].child[Left], rbTree.root.child[Right].child[Right]

	_, err := validateTree(rbTree.root)
	if err != nil {
		t.Error(err.Error())
	}

	for _, e := range []int{4, 2, 1, 5, 3, 7, 6} {
		rbTree.Delete(e)
		_, err := validateTree(rbTree.root)
		if err != nil {
			t.Error(err.Error())
		}
	}
}

func TestRedBlackConstraints(t *testing.T) {
	rbTree := &RedBlackTree[int]{Ordering: compare.Less[int]}
	rng := rand.New(rand.NewSource(0xDeadBeef))

	if !t.Run("EmptyTree", func(t *testing.T) {
		_, err := validateTree(rbTree.root)
		if err != nil {
			t.Error(err.Error())
		}
	}) {
		t.Skip("EmptyTree failed, skipping remaining tests...")
	}

	if !t.Run("Put1000Times", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			t.Logf("iteration: %v", i)
			t.Logf("Currently, rbTree.Len() == %v", rbTree.Len())

			e := rng.Intn(1000)
			t.Logf("Put(%v)", e)
			rbTree.Put(e)

			bh, err := validateTree(rbTree.root)
			if err != nil {
				t.Error(err.Error())
				return
			}
			t.Logf("rbTree is valid with black-height: %v", bh)
		}
	}) {
		t.Skip("Put1000Times failed, skipping remaining tests...")
	}

	t.Run("PutDelete1000Times", func(t *testing.T) {
		for i := 0; i < 1000; i++ {
			t.Logf("iteration: %v", i)
			t.Logf("Currently, rbTree.Len() == %v", rbTree.Len())

			e := rng.Intn(1000)
			t.Logf("Put(%v)", e)
			rbTree.Put(e)

			bh, err := validateTree(rbTree.root)
			if err != nil {
				t.Error(err.Error())
				return
			}
			t.Logf("rbTree is valid with black-height: %v", bh)

			e = rng.Intn(1000)
			t.Logf("Delete(%v)", e)
			rbTree.Delete(e)

			bh, err = validateTree(rbTree.root)
			if err != nil {
				t.Error(err.Error())
				return
			}
			t.Logf("rbTree is valid with black-height: %v", bh)
		}
	})
}
