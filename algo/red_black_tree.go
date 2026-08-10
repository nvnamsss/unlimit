package algo

import (
	"cmp"
	"fmt"
)

// Color represents the color of a Red-Black Tree node
type Color bool

const (
	Red   Color = false
	Black Color = true
)

// RBNode represents a node in the Red-Black Tree
type RBNode[K cmp.Ordered, V any] struct {
	Key    K
	Value  V
	Color  Color
	Left   *RBNode[K, V]
	Right  *RBNode[K, V]
	Parent *RBNode[K, V]
}

// RedBlackTree represents the Red-Black Tree
type RedBlackTree[K cmp.Ordered, V any] struct {
	Root *RBNode[K, V]
	Nil  *RBNode[K, V] // Sentinel node
}

// NewRedBlackTree creates a new Red-Black Tree
func NewRedBlackTree[K cmp.Ordered, V any]() *RedBlackTree[K, V] {
	nil := &RBNode[K, V]{Color: Black}
	return &RedBlackTree[K, V]{
		Root: nil,
		Nil:  nil,
	}
}

// leftRotate performs left rotation
func (rbt *RedBlackTree[K, V]) leftRotate(x *RBNode[K, V]) {
	y := x.Right
	x.Right = y.Left
	if y.Left != rbt.Nil {
		y.Left.Parent = x
	}
	y.Parent = x.Parent
	if x.Parent == rbt.Nil {
		rbt.Root = y
	} else if x == x.Parent.Left {
		x.Parent.Left = y
	} else {
		x.Parent.Right = y
	}
	y.Left = x
	x.Parent = y
}

// rightRotate performs right rotation
func (rbt *RedBlackTree[K, V]) rightRotate(x *RBNode[K, V]) {
	y := x.Left
	x.Left = y.Right
	if y.Right != rbt.Nil {
		y.Right.Parent = x
	}
	y.Parent = x.Parent
	if x.Parent == rbt.Nil {
		rbt.Root = y
	} else if x == x.Parent.Right {
		x.Parent.Right = y
	} else {
		x.Parent.Left = y
	}
	y.Right = x
	x.Parent = y
}

// Insert inserts a key-value pair into the tree
func (rbt *RedBlackTree[K, V]) Insert(key K, value V) {
	z := &RBNode[K, V]{
		Key:    key,
		Value:  value,
		Color:  Red,
		Left:   rbt.Nil,
		Right:  rbt.Nil,
		Parent: rbt.Nil,
	}

	y := rbt.Nil
	x := rbt.Root

	for x != rbt.Nil {
		y = x
		if z.Key < x.Key {
			x = x.Left
		} else {
			x = x.Right
		}
	}

	z.Parent = y
	if y == rbt.Nil {
		rbt.Root = z
	} else if z.Key < y.Key {
		y.Left = z
	} else {
		y.Right = z
	}

	rbt.insertFixup(z)
}

// insertFixup maintains Red-Black Tree properties after insertion
func (rbt *RedBlackTree[K, V]) insertFixup(z *RBNode[K, V]) {
	for z.Parent.Color == Red {
		if z.Parent == z.Parent.Parent.Left {
			y := z.Parent.Parent.Right
			if y.Color == Red {
				z.Parent.Color = Black
				y.Color = Black
				z.Parent.Parent.Color = Red
				z = z.Parent.Parent
			} else {
				if z == z.Parent.Right {
					z = z.Parent
					rbt.leftRotate(z)
				}
				z.Parent.Color = Black
				z.Parent.Parent.Color = Red
				rbt.rightRotate(z.Parent.Parent)
			}
		} else {
			y := z.Parent.Parent.Left
			if y.Color == Red {
				z.Parent.Color = Black
				y.Color = Black
				z.Parent.Parent.Color = Red
				z = z.Parent.Parent
			} else {
				if z == z.Parent.Left {
					z = z.Parent
					rbt.rightRotate(z)
				}
				z.Parent.Color = Black
				z.Parent.Parent.Color = Red
				rbt.leftRotate(z.Parent.Parent)
			}
		}
	}
	rbt.Root.Color = Black
}

// Search finds a node with the given key
func (rbt *RedBlackTree[K, V]) Search(key K) *RBNode[K, V] {
	x := rbt.Root
	for x != rbt.Nil && key != x.Key {
		if key < x.Key {
			x = x.Left
		} else {
			x = x.Right
		}
	}
	if x == rbt.Nil {
		return nil
	}
	return x
}

// transplant replaces subtree rooted at u with subtree rooted at v
func (rbt *RedBlackTree[K, V]) transplant(u, v *RBNode[K, V]) {
	if u.Parent == rbt.Nil {
		rbt.Root = v
	} else if u == u.Parent.Left {
		u.Parent.Left = v
	} else {
		u.Parent.Right = v
	}
	v.Parent = u.Parent
}

// minimum finds the minimum node in subtree rooted at x
func (rbt *RedBlackTree[K, V]) minimum(x *RBNode[K, V]) *RBNode[K, V] {
	for x.Left != rbt.Nil {
		x = x.Left
	}
	return x
}

// Delete removes a node with the given key
func (rbt *RedBlackTree[K, V]) Delete(key K) bool {
	z := rbt.Search(key)
	if z == nil {
		return false
	}

	y := z
	yOriginalColor := y.Color
	var x *RBNode[K, V]

	if z.Left == rbt.Nil {
		x = z.Right
		rbt.transplant(z, z.Right)
	} else if z.Right == rbt.Nil {
		x = z.Left
		rbt.transplant(z, z.Left)
	} else {
		y = rbt.minimum(z.Right)
		yOriginalColor = y.Color
		x = y.Right
		if y.Parent == z {
			x.Parent = y
		} else {
			rbt.transplant(y, y.Right)
			y.Right = z.Right
			y.Right.Parent = y
		}
		rbt.transplant(z, y)
		y.Left = z.Left
		y.Left.Parent = y
		y.Color = z.Color
	}

	if yOriginalColor == Black {
		rbt.deleteFixup(x)
	}
	return true
}

// deleteFixup maintains Red-Black Tree properties after deletion
func (rbt *RedBlackTree[K, V]) deleteFixup(x *RBNode[K, V]) {
	for x != rbt.Root && x.Color == Black {
		if x == x.Parent.Left {
			w := x.Parent.Right
			if w.Color == Red {
				w.Color = Black
				x.Parent.Color = Red
				rbt.leftRotate(x.Parent)
				w = x.Parent.Right
			}
			if w.Left.Color == Black && w.Right.Color == Black {
				w.Color = Red
				x = x.Parent
			} else {
				if w.Right.Color == Black {
					w.Left.Color = Black
					w.Color = Red
					rbt.rightRotate(w)
					w = x.Parent.Right
				}
				w.Color = x.Parent.Color
				x.Parent.Color = Black
				w.Right.Color = Black
				rbt.leftRotate(x.Parent)
				x = rbt.Root
			}
		} else {
			w := x.Parent.Left
			if w.Color == Red {
				w.Color = Black
				x.Parent.Color = Red
				rbt.rightRotate(x.Parent)
				w = x.Parent.Left
			}
			if w.Right.Color == Black && w.Left.Color == Black {
				w.Color = Red
				x = x.Parent
			} else {
				if w.Left.Color == Black {
					w.Right.Color = Black
					w.Color = Red
					rbt.leftRotate(w)
					w = x.Parent.Left
				}
				w.Color = x.Parent.Color
				x.Parent.Color = Black
				w.Left.Color = Black
				rbt.rightRotate(x.Parent)
				x = rbt.Root
			}
		}
	}
	x.Color = Black
}

// InOrderTraversal performs in-order traversal
func (rbt *RedBlackTree[K, V]) InOrderTraversal() []K {
	var result []K
	rbt.inOrderHelper(rbt.Root, &result)
	return result
}

func (rbt *RedBlackTree[K, V]) inOrderHelper(node *RBNode[K, V], result *[]K) {
	if node != rbt.Nil {
		rbt.inOrderHelper(node.Left, result)
		*result = append(*result, node.Key)
		rbt.inOrderHelper(node.Right, result)
	}
}

// PrintTree prints the tree structure (for debugging)
func (rbt *RedBlackTree[K, V]) PrintTree() {
	rbt.printHelper(rbt.Root, "", true)
}

func (rbt *RedBlackTree[K, V]) printHelper(node *RBNode[K, V], prefix string, isLast bool) {
	if node != rbt.Nil {
		fmt.Print(prefix)
		if isLast {
			fmt.Print("└── ")
			prefix += "    "
		} else {
			fmt.Print("├── ")
			prefix += "│   "
		}

		color := "R"
		if node.Color == Black {
			color = "B"
		}
		fmt.Printf("%v(%s)\n", node.Key, color)

		if node.Left != rbt.Nil || node.Right != rbt.Nil {
			rbt.printHelper(node.Left, prefix, node.Right == rbt.Nil)
			rbt.printHelper(node.Right, prefix, true)
		}
	}
}
