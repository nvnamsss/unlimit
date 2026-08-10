package algo

import "github.com/voidforge-studios/unlimit/utility"

// baseGraph contains common functionality for both directed and undirected graphs
type baseGraph[T comparable] struct {
	Nodes []*GraphNode[T]
	Edges []*GraphEdge[T]
}

// DirectedGraph represents a graph where edges have a direction
type DirectedGraph[T comparable] struct {
	baseGraph[T]
}

// UndirectedGraph represents a graph where edges have no direction
type UndirectedGraph[T comparable] struct {
	baseGraph[T]
}

// NewDirectedGraph creates a new empty directed graph
func NewDirectedGraph[T comparable]() *DirectedGraph[T] {
	return &DirectedGraph[T]{
		baseGraph: baseGraph[T]{
			Nodes: make([]*GraphNode[T], 0),
			Edges: make([]*GraphEdge[T], 0),
		},
	}
}

// NewUndirectedGraph creates a new empty undirected graph
func NewUndirectedGraph[T comparable]() *UndirectedGraph[T] {
	return &UndirectedGraph[T]{
		baseGraph: baseGraph[T]{
			Nodes: make([]*GraphNode[T], 0),
			Edges: make([]*GraphEdge[T], 0),
		},
	}
}

// NewGraph creates a new empty graph (for backward compatibility)
func NewGraph[T comparable](isDirected bool) Graph[T] {
	if isDirected {
		return NewDirectedGraph[T]()
	}
	return NewUndirectedGraph[T]()
}

// Graph is an interface that both DirectedGraph and UndirectedGraph implement
type Graph[T comparable] interface {
	AddNode(value T) *GraphNode[T]
	AddEdge(from, to *GraphNode[T], weight int) *GraphEdge[T]
	FindNode(value T) *GraphNode[T]
	BFS(start *GraphNode[T], visit func(*GraphNode[T]))
	DFS(start *GraphNode[T], visit func(*GraphNode[T]))
	GetNeighbors(node *GraphNode[T]) []*GraphNode[T]
	HasEdge(from, to *GraphNode[T]) bool
	GetEdge(from, to *GraphNode[T]) *GraphEdge[T]
	RemoveNode(node *GraphNode[T])
	RemoveEdge(from, to *GraphNode[T])
	IsCyclic() bool
	IsDirectedGraph() bool
}

// IsDirectedGraph returns whether this is a directed graph (always true for DirectedGraph)
func (g *DirectedGraph[T]) IsDirectedGraph() bool {
	return true
}

// IsDirectedGraph returns whether this is a directed graph (always false for UndirectedGraph)
func (g *UndirectedGraph[T]) IsDirectedGraph() bool {
	return false
}

// AddNode adds a node with the given value to the graph
func (g *baseGraph[T]) AddNode(value T) *GraphNode[T] {
	node := &GraphNode[T]{
		Value:    value,
		Children: make([]*GraphNode[T], 0),
	}
	g.Nodes = append(g.Nodes, node)
	return node
}

// Implementing AddNode for both graph types
func (g *DirectedGraph[T]) AddNode(value T) *GraphNode[T] {
	return g.baseGraph.AddNode(value)
}

func (g *UndirectedGraph[T]) AddNode(value T) *GraphNode[T] {
	return g.baseGraph.AddNode(value)
}

// AddEdge adds an edge between the two nodes in a directed graph
func (g *DirectedGraph[T]) AddEdge(from, to *GraphNode[T], weight int) *GraphEdge[T] {
	edge := &GraphEdge[T]{
		From:   from,
		To:     to,
		Weight: weight,
	}
	g.Edges = append(g.Edges, edge)

	// Add the destination node to the source node's children
	from.Children = append(from.Children, to)

	return edge
}

// AddEdge adds an edge between the two nodes in an undirected graph
func (g *UndirectedGraph[T]) AddEdge(from, to *GraphNode[T], weight int) *GraphEdge[T] {
	edge := &GraphEdge[T]{
		From:   from,
		To:     to,
		Weight: weight,
	}
	g.Edges = append(g.Edges, edge)

	// Add bidirectional connections
	from.Children = append(from.Children, to)
	to.Children = append(to.Children, from)

	return edge
}

// FindNode finds a node with the given value in the graph
func (g *baseGraph[T]) FindNode(value T) *GraphNode[T] {
	for _, node := range g.Nodes {
		if node.Value == value {
			return node
		}
	}
	return nil
}

// Implementing FindNode for both graph types
func (g *DirectedGraph[T]) FindNode(value T) *GraphNode[T] {
	return g.baseGraph.FindNode(value)
}

func (g *UndirectedGraph[T]) FindNode(value T) *GraphNode[T] {
	return g.baseGraph.FindNode(value)
}

// BFS performs a breadth-first search starting from the given node
func (g *baseGraph[T]) BFS(start *GraphNode[T], visit func(*GraphNode[T])) {
	if start == nil {
		return
	}

	visited := make(map[*GraphNode[T]]bool)
	queue := []*GraphNode[T]{start}

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		if visited[node] {
			continue
		}

		visited[node] = true
		visit(node)

		for _, child := range node.Children {
			if !visited[child] {
				queue = append(queue, child)
			}
		}
	}
}

// Implementing BFS for both graph types
func (g *DirectedGraph[T]) BFS(start *GraphNode[T], visit func(*GraphNode[T])) {
	g.baseGraph.BFS(start, visit)
}

func (g *UndirectedGraph[T]) BFS(start *GraphNode[T], visit func(*GraphNode[T])) {
	g.baseGraph.BFS(start, visit)
}

// DFS performs a depth-first search starting from the given node
func (g *baseGraph[T]) DFS(start *GraphNode[T], visit func(*GraphNode[T])) {
	if start == nil {
		return
	}

	visited := make(map[*GraphNode[T]]bool)
	g.dfsHelper(start, visited, visit)
}

// dfsHelper is a recursive helper function for DFS
func (g *baseGraph[T]) dfsHelper(node *GraphNode[T], visited map[*GraphNode[T]]bool, visit func(*GraphNode[T])) {
	if visited[node] {
		return
	}

	visited[node] = true
	visit(node)

	for _, child := range node.Children {
		if !visited[child] {
			g.dfsHelper(child, visited, visit)
		}
	}
}

// Implementing DFS for both graph types
func (g *DirectedGraph[T]) DFS(start *GraphNode[T], visit func(*GraphNode[T])) {
	g.baseGraph.DFS(start, visit)
}

func (g *UndirectedGraph[T]) DFS(start *GraphNode[T], visit func(*GraphNode[T])) {
	g.baseGraph.DFS(start, visit)
}

// GetNeighbors returns all neighbors of a node
func (g *baseGraph[T]) GetNeighbors(node *GraphNode[T]) []*GraphNode[T] {
	return node.Children
}

// Implementing GetNeighbors for both graph types
func (g *DirectedGraph[T]) GetNeighbors(node *GraphNode[T]) []*GraphNode[T] {
	return g.baseGraph.GetNeighbors(node)
}

func (g *UndirectedGraph[T]) GetNeighbors(node *GraphNode[T]) []*GraphNode[T] {
	return g.baseGraph.GetNeighbors(node)
}

// HasEdge checks if there is an edge between two nodes
func (g *baseGraph[T]) HasEdge(from, to *GraphNode[T]) bool {
	for _, edge := range g.Edges {
		if edge.From == from && edge.To == to {
			return true
		}
	}
	return false
}

// Implementing HasEdge for DirectedGraph
func (g *DirectedGraph[T]) HasEdge(from, to *GraphNode[T]) bool {
	return g.baseGraph.HasEdge(from, to)
}

// Implementing HasEdge for UndirectedGraph
func (g *UndirectedGraph[T]) HasEdge(from, to *GraphNode[T]) bool {
	// Check both directions for undirected graph
	return g.baseGraph.HasEdge(from, to) || g.baseGraph.HasEdge(to, from)
}

// GetEdge returns the edge between two nodes, or nil if no such edge exists
func (g *baseGraph[T]) GetEdge(from, to *GraphNode[T]) *GraphEdge[T] {
	for _, edge := range g.Edges {
		if edge.From == from && edge.To == to {
			return edge
		}
	}
	return nil
}

// Implementing GetEdge for DirectedGraph
func (g *DirectedGraph[T]) GetEdge(from, to *GraphNode[T]) *GraphEdge[T] {
	return g.baseGraph.GetEdge(from, to)
}

// Implementing GetEdge for UndirectedGraph
func (g *UndirectedGraph[T]) GetEdge(from, to *GraphNode[T]) *GraphEdge[T] {
	edge := g.baseGraph.GetEdge(from, to)
	if edge == nil {
		// Check the reverse direction for undirected graphs
		edge = g.baseGraph.GetEdge(to, from)
	}
	return edge
}

// RemoveNode removes a node and all its connected edges from the graph
func (g *DirectedGraph[T]) RemoveNode(node *GraphNode[T]) {
	// Remove all edges connected to this node
	newEdges := utility.Filter(g.Edges, func(edge *GraphEdge[T]) bool {
		return edge.From != node && edge.To != node
	})

	g.Edges = newEdges

	// Remove node from children lists of other nodes
	for _, n := range g.Nodes {
		if n == node {
			continue
		}

		newChildren := make([]*GraphNode[T], 0)
		for _, child := range n.Children {
			if child != node {
				newChildren = append(newChildren, child)
			}
		}
		n.Children = newChildren
	}

	// Remove node from graph's nodes list
	newNodes := make([]*GraphNode[T], 0)
	for _, n := range g.Nodes {
		if n != node {
			newNodes = append(newNodes, n)
		}
	}
	g.Nodes = newNodes
}

// RemoveNode for UndirectedGraph uses same logic as DirectedGraph
func (g *UndirectedGraph[T]) RemoveNode(node *GraphNode[T]) {
	// Same implementation as DirectedGraph
	// Remove all edges connected to this node
	newEdges := make([]*GraphEdge[T], 0)
	for _, edge := range g.Edges {
		if edge.From != node && edge.To != node {
			newEdges = append(newEdges, edge)
		}
	}
	g.Edges = newEdges

	// Remove node from children lists of other nodes
	for _, n := range g.Nodes {
		if n == node {
			continue
		}

		newChildren := make([]*GraphNode[T], 0)
		for _, child := range n.Children {
			if child != node {
				newChildren = append(newChildren, child)
			}
		}
		n.Children = newChildren
	}

	// Remove node from graph's nodes list
	newNodes := make([]*GraphNode[T], 0)
	for _, n := range g.Nodes {
		if n != node {
			newNodes = append(newNodes, n)
		}
	}
	g.Nodes = newNodes
}

// RemoveEdge removes an edge between two nodes in a directed graph
func (g *DirectedGraph[T]) RemoveEdge(from, to *GraphNode[T]) {
	// Remove edge from edges list
	newEdges := utility.Filter(g.Edges, func(edge *GraphEdge[T]) bool {
		return !(edge.From == from && edge.To == to)
	})
	g.Edges = newEdges

	// Remove destination from source's children
	newChildren := make([]*GraphNode[T], 0)
	for _, child := range from.Children {
		if child != to {
			newChildren = append(newChildren, child)
		}
	}
	from.Children = newChildren
}

// RemoveEdge removes an edge between two nodes in an undirected graph
func (g *UndirectedGraph[T]) RemoveEdge(from, to *GraphNode[T]) {
	// Remove edge from edges list (checking both directions)
	newEdges := utility.Filter(g.Edges, func(edge *GraphEdge[T]) bool {
		return !((edge.From == from && edge.To == to) || (edge.From == to && edge.To == from))
	})
	g.Edges = newEdges

	// Remove each node from the other's children list
	removeFromChildren := func(node, target *GraphNode[T]) {
		newChildren := make([]*GraphNode[T], 0)
		for _, child := range node.Children {
			if child != target {
				newChildren = append(newChildren, child)
			}
		}
		node.Children = newChildren
	}

	removeFromChildren(from, to)
	removeFromChildren(to, from)
}

// IsCyclic checks if the directed graph contains a cycle
func (g *DirectedGraph[T]) IsCyclic() bool {
	visited := make(map[*GraphNode[T]]bool)
	recStack := make(map[*GraphNode[T]]bool)

	for _, node := range g.Nodes {
		if !visited[node] {
			if g.isCyclicDFS(node, visited, recStack) {
				return true
			}
		}
	}
	return false
}

// isCyclicDFS is a recursive helper function for cycle detection in directed graphs
func (g *DirectedGraph[T]) isCyclicDFS(node *GraphNode[T], visited, recStack map[*GraphNode[T]]bool) bool {
	visited[node] = true
	recStack[node] = true

	for _, child := range node.Children {
		if !visited[child] {
			if g.isCyclicDFS(child, visited, recStack) {
				return true
			}
		} else if recStack[child] {
			return true
		}
	}

	recStack[node] = false
	return false
}

// IsCyclic checks if the undirected graph contains a cycle
func (g *UndirectedGraph[T]) IsCyclic() bool {
	visited := make(map[*GraphNode[T]]bool)

	for _, node := range g.Nodes {
		if !visited[node] {
			if g.isCyclicUndirectedDFS(node, visited, nil) {
				return true
			}
		}
	}
	return false
}

// isCyclicUndirectedDFS is a recursive helper function for cycle detection in undirected graphs
func (g *UndirectedGraph[T]) isCyclicUndirectedDFS(node *GraphNode[T], visited map[*GraphNode[T]]bool, parent *GraphNode[T]) bool {
	visited[node] = true

	for _, child := range node.Children {
		// Skip the edge back to parent
		if child == parent {
			continue
		}

		if visited[child] {
			return true
		}

		if g.isCyclicUndirectedDFS(child, visited, node) {
			return true
		}
	}
	return false
}
