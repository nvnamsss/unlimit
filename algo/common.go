package algo

// DoublyNode represents a DoublyNode in a doubly linked list for DequeList
type DoublyNode[T any] struct {
	Value T
	Prev  *DoublyNode[T]
	Next  *DoublyNode[T]
}

// ListNode represents a node in a linked list
type ListNode[T any] struct {
	Data T
	Next *ListNode[T]
}

// TreeNode represents a node in a binary tree
type TreeNode[T any] struct {
	Value       T
	Left, Right *TreeNode[T]
}

// GraphNode represents a node in a graph
type GraphNode[T any] struct {
	Value    T
	Children []*GraphNode[T]
}

type GraphEdge[T any] struct {
	From, To *GraphNode[T]
	Weight   int
}
