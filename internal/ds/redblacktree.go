package ds

import (
	"github.org/jccarlson/collections/compare"
)

type Direction int

const (
	Left Direction = iota
	Right
)

// TreeNode is a struct wrapping a element in a red-black tree, with pointers
// to the element's parent and children, if any.
type TreeNode[E any] struct {
	Elem E

	parent *TreeNode[E]
	child  [2]*TreeNode[E]

	black bool
}

func (n *TreeNode[E]) isRed() bool {
	return n != nil && !n.black
}

func (n *TreeNode[E]) isBlack() bool {
	return n == nil || n.black
}

func childDir[E any](n *TreeNode[E]) Direction {
	if n.parent.child[Left] == n {
		return Left
	}
	return Right
}

func (n *TreeNode[E]) Walk(d Direction) *TreeNode[E] {
	if n.child[d] != nil {
		// If n has a child in direction d, then if d == left the next in-order
		// node is the right-most descendant of n's left child, and vice-versa.
		t := n.child[d]
		for t.child[1-d] != nil {
			t = t.child[1-d]
		}
		return t
	}

	t := n
	for t.parent != nil && childDir(t) == d {
		// iterate up n's ancestors until one is a right child if d == left, or
		// vice versa. Then the parent is the previous in-order node. If there
		// is no parent, n is the first in-order node, so we return nil.
		t = t.parent
	}
	return t.parent
}

// RedBlackTree is a balanced binary tree of elements of type E.
type RedBlackTree[E any] struct {
	Ordering compare.Ordering[E]

	root        *TreeNode[E]
	first, last *TreeNode[E]
	size        int
}

func (m *RedBlackTree[E]) Put(elem E) {
	node := &TreeNode[E]{Elem: elem}
	m.putRecursive(&m.root, node, nil)
	if m.first == nil || m.Ordering(node.Elem, m.first.Elem) {
		m.first = node
	}
	if m.last == nil || m.Ordering(m.last.Elem, node.Elem) {
		m.last = node
	}
}

func (m *RedBlackTree[E]) putRecursive(root **TreeNode[E], e *TreeNode[E], parent *TreeNode[E]) {
	if *root == nil {
		*root = e
		e.parent = parent
		m.insertionRebalance(e)
		m.size++
		return
	}
	if m.Ordering(e.Elem, (*root).Elem) {
		m.putRecursive(&(*root).child[Left], e, *root)
		return

	}
	if m.Ordering((*root).Elem, e.Elem) {
		m.putRecursive(&(*root).child[Right], e, *root)
		return

	}
	(*root).Elem = e.Elem
}

func (m *RedBlackTree[E]) insertionRebalance(e *TreeNode[E]) {
	for parent := e.parent; parent != nil; parent = e.parent {
		if parent.isBlack() {
			// adding e as a red node doesn't violate the red-black tree
			// invariants if the parent node is black.
			return
		}
		grandparent := parent.parent
		if grandparent == nil {
			// parent is red, and the root, so we just switch parent to black.
			parent.black = true
			return
		}

		uncle, dir := grandparent.child[Right], Left
		if parent == uncle {
			uncle, dir = grandparent.child[Left], Right
		}
		if uncle.isBlack() {
			if e == parent.child[1-dir] {
				// parent is red, uncle is black, and e's order is between
				// parent and grandparent
				m.rotate(parent, dir)
				e = parent
				parent = grandparent.child[dir]
			}
			m.rotate(grandparent, 1-dir)
			parent.black = true
			grandparent.black = false
			return
		}
		// parent and uncle are red
		parent.black = true
		uncle.black = true
		grandparent.black = false
		e = grandparent
	}
}

// Rotates the sub-tree rooted at node e in direction dir, e.g.
//
// The sub-tree:
//
//							e
//	                      /   \
//	                    l       r
//	                  /   \   /   \
//	                 1     2 3     4
//
// with a 'right' rotation becomes:
//
//						    l
//	                      /   \
//	                    1       e
//	                          /   \
//	                         2     r
//	                             /   \
//	                            3     4
//
// and pointers to the 'root' are changed to point to the new root.
func (m *RedBlackTree[E]) rotate(e *TreeNode[E], dir Direction) {
	rootPtr := &m.root
	if e.parent != nil {
		rootPtr = &e.parent.child[Right]
		if e == e.parent.child[Left] {
			rootPtr = &e.parent.child[Left]
		}
	}

	*rootPtr = e.child[1-dir]
	(*rootPtr).parent = e.parent
	e.child[1-dir] = (*rootPtr).child[dir]
	if e.child[1-dir] != nil {
		e.child[1-dir].parent = e
	}
	(*rootPtr).child[dir] = e
	(*rootPtr).child[dir].parent = (*rootPtr)
}

func (m *RedBlackTree[E]) Get(elem E) (E, bool) {
	return getRecursive(m.root, elem, m.Ordering)
}

func (m *RedBlackTree[E]) Has(elem E) bool {
	_, ok := getRecursive(m.root, elem, m.Ordering)
	return ok
}

func getRecursive[E any](root *TreeNode[E], elem E, before compare.Ordering[E]) (value E, ok bool) {
	if root == nil {
		return
	}
	if before(elem, root.Elem) {
		return getRecursive(root.child[Left], elem, before)
	}
	if before(root.Elem, elem) {
		return getRecursive(root.child[Right], elem, before)
	}
	return root.Elem, true
}

func (m *RedBlackTree[E]) Delete(elem E) {
	m.deleteRecursive(&m.root, elem)
}

func (m *RedBlackTree[E]) deleteRecursive(root **TreeNode[E], elem E) {
	if *root == nil {
		// elem not in the tree, delete nothing.
		return
	}
	before := m.Ordering
	if before(elem, (*root).Elem) {
		m.deleteRecursive(&(*root).child[Left], elem)
		return

	}
	if before((*root).Elem, elem) {
		m.deleteRecursive(&(*root).child[Right], elem)
		return

	}
	// Found the node, but if it has 2 non-nil children, we need to swap it
	// with it's in-order successor.

	if (*root).child[Left] != nil && (*root).child[Right] != nil {
		t := &(*root).child[Right]
		for (*t).child[Left] != nil {
			t = &(*t).child[Left]
		}
		(*root).Elem = (*t).Elem
		root = t
	}

	// root now references the parent's child pointer to the node to be
	// deleted. *root has at most 1 non-nil child.

	if (*root).isRed() || ((*root).parent == nil && (*root).child[Left] == nil && (*root).child[Right] == nil) {
		// *root can simply be deleted if:
		//     - *root is red (guaranteed to have no children).
		//     - *root is the actual root and has no children.
		*root = nil
		m.size--
		return
	}

	// *root is black, with at most one child.
	// If *root has one child, it must be red, so replace *root with the child
	// and paint the child black.
	if (*root).child[Right] != nil {
		(*root).child[Right].parent = (*root).parent
		*root = (*root).child[Right]
		(*root).black = true
		m.size--
		return
	}
	if (*root).child[Left] != nil {
		(*root).child[Left].parent = (*root).parent
		*root = (*root).child[Left]
		(*root).black = true
		m.size--
		return
	}

	// *root is black, with no children, and is not the root of the tree.
	m.balanceBlackLeafForDeletion(*root)

	// Update first and last pointers if needed.
	if m.first == *root {
		m.first = (*root).Walk(Right)
	}
	if m.last == *root {
		m.last = (*root).Walk(Left)
	}
	*root = nil
	m.size--
}

// balanceBlackLeafFOrDeletion iterates up and modifies m so that n's black
// height is 1 greater than the black height of the rest of the leaf nodes, so
// deleting it will result in a balanced tree.
func (m *RedBlackTree[E]) balanceBlackLeafForDeletion(n *TreeNode[E]) {
	type balanceCase int
	const (
		parentNil balanceCase = iota
		siblingRed
		parentRedSiblingAndNephewsBlack
		siblingFarNephewBlackCloseNephewRed
		siblingBlackFarNephewRed
	)

	var dir Direction
	var parent, sibling, closeNephew, farNephew *TreeNode[E]
	nextCase := parentNil

	for parent = n.parent; parent != nil; parent = n.parent {
		dir = Left
		if parent.child[Left] != n {
			dir = Right
		}
		sibling = parent.child[1-dir]
		closeNephew = sibling.child[dir]
		farNephew = sibling.child[1-dir]

		if sibling.isRed() {
			nextCase = siblingRed
			break
		}
		if farNephew.isRed() {
			nextCase = siblingBlackFarNephewRed
			break
		}
		if closeNephew.isRed() {
			nextCase = siblingFarNephewBlackCloseNephewRed
			break
		}
		if parent.isRed() {
			nextCase = parentRedSiblingAndNephewsBlack
			break
		}
		// parent, sibling, and nephews are black. change the sibling to red
		// so that sibling's black height is n's black height minus 1, then
		// iterate up the tree (because the parent's sibling may now be
		// imbalanced).
		sibling.black = false
		n = parent
	}

	for {
		switch nextCase {

		case parentNil:
			// We iterated up the entire tree, so every leaf has the same black
			// height except the original n, which is +1.
			return

		case siblingRed:
			// sibling is red ==> parent, closeNephew, and farNephew are black.
			// Rotate so sibling is now n's grandparent, and set parent to red
			// and sibling to black.
			m.rotate(parent, dir)
			parent.black = false
			sibling.black = true

			// n's new sibling is it's old closeNephew (which is black).
			sibling = closeNephew
			farNephew = sibling.child[1-dir]

			// Now one of the below cases applies.
			if farNephew.isRed() {
				nextCase = siblingBlackFarNephewRed
				break
			}
			closeNephew = sibling.child[dir]
			if closeNephew.isRed() {
				nextCase = siblingFarNephewBlackCloseNephewRed
				break
			}
			fallthrough

		case parentRedSiblingAndNephewsBlack:
			// If the parent is red and sibling and nephews are all black, we
			// can make parent black and sibling red. Now n's black height is
			// +1 and sibling's black height is the same, so we can delete
			// the original n.
			sibling.black = false
			parent.black = true
			return

		case siblingFarNephewBlackCloseNephewRed:
			// If the sibling and far nephew are black and the close nephew is
			// red, we rotate sibling away from n and make closeNephew black
			// and sibling red. Now closeNephew is the new (black) sibling and
			// the old sibling is the new (red) farNephew, so we fallthrough
			// to the next case.
			m.rotate(sibling, 1-dir)
			sibling.black = false
			closeNephew.black = true
			farNephew = sibling
			sibling = closeNephew
			fallthrough

		case siblingBlackFarNephewRed:
			// If the sibling is black and the far nephew is red, we rotate
			// parent towards n. We color sibling the same as parent, then
			// color parent and far nephew black. Now n has black height +1 and
			// nephew's black height's are unchanged, so we can delete the
			// original n.
			m.rotate(parent, dir)
			sibling.black = parent.black
			parent.black = true
			farNephew.black = true
			return
		}
	}
}

func (m *RedBlackTree[E]) Len() int {
	return m.size
}

func (m *RedBlackTree[E]) First() *TreeNode[E] {
	return m.first
}

func (m *RedBlackTree[E]) Last() *TreeNode[E] {
	return m.first
}
