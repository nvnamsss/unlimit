package algo

import (
	"testing"
)

// TestDirectedGraph_New tests the creation of a new directed graph
func TestDirectedGraph_New(t *testing.T) {
	directedGraph := NewDirectedGraph[int]()
	if directedGraph == nil {
		t.Error("NewDirectedGraph() returned nil")
	}
	if !directedGraph.IsDirectedGraph() {
		t.Error("Expected directed graph, got undirected")
	}
	if len(directedGraph.Nodes) != 0 {
		t.Errorf("Expected 0 nodes, got %d", len(directedGraph.Nodes))
	}
	if len(directedGraph.Edges) != 0 {
		t.Errorf("Expected 0 edges, got %d", len(directedGraph.Edges))
	}
}

// TestUndirectedGraph_New tests the creation of a new undirected graph
func TestUndirectedGraph_New(t *testing.T) {
	undirectedGraph := NewUndirectedGraph[string]()
	if undirectedGraph == nil {
		t.Error("NewUndirectedGraph() returned nil")
	}
	if undirectedGraph.IsDirectedGraph() {
		t.Error("Expected undirected graph, got directed")
	}
}

// TestBackwardCompatibility tests the backward compatibility of NewGraph function
func TestBackwardCompatibility(t *testing.T) {
	directedGraph := NewGraph[int](true)
	if !directedGraph.IsDirectedGraph() {
		t.Error("Expected directed graph from backward compatibility function")
	}

	undirectedGraph := NewGraph[int](false)
	if undirectedGraph.IsDirectedGraph() {
		t.Error("Expected undirected graph from backward compatibility function")
	}
}

// TestDirectedGraph_AddNode tests adding nodes to a directed graph
func TestDirectedGraph_AddNode(t *testing.T) {
	g := NewDirectedGraph[int]()

	// Add a single node
	node1 := g.AddNode(1)
	if node1 == nil {
		t.Fatal("AddNode(1) returned nil")
	}
	if node1.Value != 1 {
		t.Errorf("Expected node value 1, got %d", node1.Value)
	}
	if len(g.Nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(g.Nodes))
	}

	// Add another node
	node2 := g.AddNode(2)
	if len(g.Nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(g.Nodes))
	}
	if node2.Value != 2 {
		t.Errorf("Expected node value 2, got %d", node2.Value)
	}
}

// TestUndirectedGraph_AddNode tests adding nodes to an undirected graph
func TestUndirectedGraph_AddNode(t *testing.T) {
	g := NewUndirectedGraph[int]()

	// Add a single node
	node1 := g.AddNode(1)
	if node1 == nil {
		t.Fatal("AddNode(1) returned nil")
	}
	if node1.Value != 1 {
		t.Errorf("Expected node value 1, got %d", node1.Value)
	}
	if len(g.Nodes) != 1 {
		t.Errorf("Expected 1 node, got %d", len(g.Nodes))
	}
}

// TestDirectedGraph_AddEdge tests adding edges to a directed graph
func TestDirectedGraph_AddEdge(t *testing.T) {
	g := NewDirectedGraph[int]()
	node1 := g.AddNode(1)
	node2 := g.AddNode(2)

	edge := g.AddEdge(node1, node2, 10)
	if edge == nil {
		t.Fatal("AddEdge returned nil")
	}
	if edge.Weight != 10 {
		t.Errorf("Expected edge weight 10, got %d", edge.Weight)
	}
	if edge.From != node1 || edge.To != node2 {
		t.Error("Edge endpoints do not match the nodes provided")
	}
	if len(g.Edges) != 1 {
		t.Errorf("Expected 1 edge, got %d", len(g.Edges))
	}
	if len(node1.Children) != 1 || node1.Children[0] != node2 {
		t.Error("Source node's children not updated correctly")
	}
	if len(node2.Children) != 0 {
		t.Error("Target node should not have children in a directed graph")
	}
}

// TestUndirectedGraph_AddEdge tests adding edges to an undirected graph
func TestUndirectedGraph_AddEdge(t *testing.T) {
	g := NewUndirectedGraph[int]()
	node1 := g.AddNode(1)
	node2 := g.AddNode(2)

	edge := g.AddEdge(node1, node2, 5)
	if edge == nil {
		t.Fatal("AddEdge returned nil")
	}
	if len(node1.Children) != 1 || node1.Children[0] != node2 {
		t.Error("Source node's children not updated correctly in undirected graph")
	}
	if len(node2.Children) != 1 || node2.Children[0] != node1 {
		t.Error("Target node's children not updated correctly in undirected graph")
	}
}

// TestDirectedGraph_FindNode tests finding a node by value
func TestDirectedGraph_FindNode(t *testing.T) {
	g := NewDirectedGraph[int]()
	g.AddNode(1)
	g.AddNode(2)
	g.AddNode(3)

	// Find existing node
	node := g.FindNode(2)
	if node == nil {
		t.Fatal("FindNode(2) returned nil for existing node")
	}
	if node.Value != 2 {
		t.Errorf("Expected node value 2, got %d", node.Value)
	}

	// Find non-existent node
	nonExistent := g.FindNode(99)
	if nonExistent != nil {
		t.Error("FindNode(99) should return nil for non-existent node")
	}
}

// TestDirectedGraph_BFS tests breadth-first search traversal on a directed graph
func TestDirectedGraph_BFS(t *testing.T) {
	g := createTestDirectedGraph()

	// Expected order depends on how nodes were connected
	var visited []int

	g.BFS(g.FindNode(1), func(node *GraphNode[int]) {
		visited = append(visited, node.Value)
	})

	if len(visited) != 5 {
		t.Fatalf("Expected 5 visited nodes, got %d", len(visited))
	}

	// First node should be the starting node
	if visited[0] != 1 {
		t.Errorf("First visited node should be 1, got %d", visited[0])
	}
}

// TestUndirectedGraph_BFS tests breadth-first search traversal on an undirected graph
func TestUndirectedGraph_BFS(t *testing.T) {
	g := createTestUndirectedGraph()

	var visited []int

	g.BFS(g.FindNode(1), func(node *GraphNode[int]) {
		visited = append(visited, node.Value)
	})

	if len(visited) != 5 {
		t.Fatalf("Expected 5 visited nodes, got %d", len(visited))
	}
}

// TestDirectedGraph_DFS tests depth-first search traversal on a directed graph
func TestDirectedGraph_DFS(t *testing.T) {
	g := createTestDirectedGraph()

	var visited []int

	g.DFS(g.FindNode(1), func(node *GraphNode[int]) {
		visited = append(visited, node.Value)
	})

	if len(visited) != 5 {
		t.Fatalf("Expected 5 visited nodes, got %d", len(visited))
	}

	// First node should be the starting node
	if visited[0] != 1 {
		t.Errorf("First visited node should be 1, got %d", visited[0])
	}
}

// TestUndirectedGraph_GetNeighbors tests getting neighbors of a node in an undirected graph
func TestUndirectedGraph_GetNeighbors(t *testing.T) {
	g := createTestUndirectedGraph()

	node1 := g.FindNode(1)
	neighbors := g.GetNeighbors(node1)

	if len(neighbors) != 2 {
		t.Errorf("Expected 2 neighbors for node 1, got %d", len(neighbors))
	}

	// Check if neighbors are the expected nodes
	neighborValues := []int{}
	for _, n := range neighbors {
		neighborValues = append(neighborValues, n.Value)
	}

	// Should contain 2 and 3
	if !contains(neighborValues, 2) || !contains(neighborValues, 3) {
		t.Errorf("Expected neighbors [2,3], got %v", neighborValues)
	}
}

// TestDirectedGraph_HasEdge tests edge existence check in a directed graph
func TestDirectedGraph_HasEdge(t *testing.T) {
	g := createTestDirectedGraph()

	node1 := g.FindNode(1)
	node2 := g.FindNode(2)
	node5 := g.FindNode(5)

	// Check existing edge
	if !g.HasEdge(node1, node2) {
		t.Error("HasEdge didn't find existing edge from 1 to 2")
	}

	// Check non-existent edge
	if g.HasEdge(node2, node5) {
		t.Error("HasEdge found non-existent edge from 2 to 5")
	}

	// For directed graph, check reverse direction (should not exist)
	if g.HasEdge(node2, node1) {
		t.Error("HasEdge found reverse edge in directed graph that shouldn't exist")
	}
}

// TestUndirectedGraph_HasEdge tests edge existence check in an undirected graph
func TestUndirectedGraph_HasEdge(t *testing.T) {
	g := createTestUndirectedGraph()

	node1 := g.FindNode(1)
	node2 := g.FindNode(2)
	node5 := g.FindNode(5)

	// Check existing edge
	if !g.HasEdge(node1, node2) {
		t.Error("HasEdge didn't find existing edge from 1 to 2")
	}

	// Check non-existent edge
	if g.HasEdge(node2, node5) {
		t.Error("HasEdge found non-existent edge from 2 to 5")
	}

	// For undirected graph, check reverse direction (should exist)
	if !g.HasEdge(node2, node1) {
		t.Error("HasEdge didn't find reverse edge in undirected graph")
	}
}

// TestDirectedGraph_GetEdge tests retrieving edge between nodes in a directed graph
func TestDirectedGraph_GetEdge(t *testing.T) {
	g := createTestDirectedGraph()

	node1 := g.FindNode(1)
	node2 := g.FindNode(2)

	// Get existing edge
	edge := g.GetEdge(node1, node2)
	if edge == nil {
		t.Fatal("GetEdge returned nil for existing edge")
	}
	if edge.Weight != 10 {
		t.Errorf("Expected edge weight 10, got %d", edge.Weight)
	}

	// Get non-existent edge
	nonExistentEdge := g.GetEdge(node2, node1)
	if nonExistentEdge != nil {
		t.Error("GetEdge found non-existent reverse edge in directed graph")
	}
}

// TestUndirectedGraph_GetEdge tests retrieving edge between nodes in an undirected graph
func TestUndirectedGraph_GetEdge(t *testing.T) {
	g := createTestUndirectedGraph()

	node1 := g.FindNode(1)
	node2 := g.FindNode(2)

	// Get existing edge in forward direction
	edge := g.GetEdge(node1, node2)
	if edge == nil {
		t.Fatal("GetEdge returned nil for existing edge")
	}

	// Get existing edge in reverse direction
	reverseEdge := g.GetEdge(node2, node1)
	if reverseEdge == nil {
		t.Fatal("GetEdge returned nil for existing edge in reverse direction")
	}
}

// TestDirectedGraph_RemoveEdge tests removing an edge from a directed graph
func TestDirectedGraph_RemoveEdge(t *testing.T) {
	g := createTestDirectedGraph()

	node1 := g.FindNode(1)
	node2 := g.FindNode(2)

	initialEdgeCount := len(g.Edges)
	initialChildren := len(node1.Children)

	// Remove edge
	g.RemoveEdge(node1, node2)

	// Check edge count
	if len(g.Edges) != initialEdgeCount-1 {
		t.Errorf("Expected %d edges after removal, got %d", initialEdgeCount-1, len(g.Edges))
	}

	// Check children
	if len(node1.Children) != initialChildren-1 {
		t.Errorf("Expected %d children after removal, got %d", initialChildren-1, len(node1.Children))
	}

	// Check edge existence
	if g.HasEdge(node1, node2) {
		t.Error("Edge still exists after removal")
	}
}

// TestUndirectedGraph_RemoveEdge tests removing an edge from an undirected graph
func TestUndirectedGraph_RemoveEdge(t *testing.T) {
	g := createTestUndirectedGraph()

	node1 := g.FindNode(1)
	node2 := g.FindNode(2)

	initialEdgeCount := len(g.Edges)

	// Remove edge
	g.RemoveEdge(node1, node2)

	// Check edge count
	if len(g.Edges) != initialEdgeCount-1 {
		t.Errorf("Expected %d edges after removal, got %d", initialEdgeCount-1, len(g.Edges))
	}

	// Check edge existence in both directions
	if g.HasEdge(node1, node2) || g.HasEdge(node2, node1) {
		t.Error("Edge still exists after removal in undirected graph")
	}

	// Check that both nodes' children lists were updated
	for _, child := range node1.Children {
		if child == node2 {
			t.Error("node2 still in node1's children after edge removal")
		}
	}

	for _, child := range node2.Children {
		if child == node1 {
			t.Error("node1 still in node2's children after edge removal")
		}
	}
}

// TestDirectedGraph_RemoveNode tests removing a node from a directed graph
func TestDirectedGraph_RemoveNode(t *testing.T) {
	g := createTestDirectedGraph()

	node3 := g.FindNode(3)
	initialNodeCount := len(g.Nodes)
	initialEdgeCount := len(g.Edges)

	// Count edges connected to node3
	connectedEdges := 0
	for _, edge := range g.Edges {
		if edge.From == node3 || edge.To == node3 {
			connectedEdges++
		}
	}

	// Remove node
	g.RemoveNode(node3)

	// Check node count
	if len(g.Nodes) != initialNodeCount-1 {
		t.Errorf("Expected %d nodes after removal, got %d", initialNodeCount-1, len(g.Nodes))
	}

	// Check edge count
	if len(g.Edges) != initialEdgeCount-connectedEdges {
		t.Errorf("Expected %d edges after removal, got %d", initialEdgeCount-connectedEdges, len(g.Edges))
	}

	// Check node existence
	if g.FindNode(3) != nil {
		t.Error("Node still exists after removal")
	}
}

// TestDirectedGraph_IsCyclic tests cycle detection in a directed graph
func TestDirectedGraph_IsCyclic(t *testing.T) {
	// Test acyclic directed graph
	acyclicGraph := NewDirectedGraph[int]()
	node1 := acyclicGraph.AddNode(1)
	node2 := acyclicGraph.AddNode(2)
	node3 := acyclicGraph.AddNode(3)

	acyclicGraph.AddEdge(node1, node2, 1)
	acyclicGraph.AddEdge(node2, node3, 1)

	if acyclicGraph.IsCyclic() {
		t.Error("Acyclic directed graph incorrectly identified as cyclic")
	}

	// Test cyclic directed graph
	cyclicGraph := NewDirectedGraph[int]()
	cNode1 := cyclicGraph.AddNode(1)
	cNode2 := cyclicGraph.AddNode(2)
	cNode3 := cyclicGraph.AddNode(3)

	cyclicGraph.AddEdge(cNode1, cNode2, 1)
	cyclicGraph.AddEdge(cNode2, cNode3, 1)
	cyclicGraph.AddEdge(cNode3, cNode1, 1) // Creates a cycle

	if !cyclicGraph.IsCyclic() {
		t.Error("Cyclic directed graph not identified as cyclic")
	}
}

// TestUndirectedGraph_IsCyclic tests cycle detection in an undirected graph
func TestUndirectedGraph_IsCyclic(t *testing.T) {
	// Test acyclic undirected graph (a tree)
	acyclicGraph := NewUndirectedGraph[int]()
	node1 := acyclicGraph.AddNode(1)
	node2 := acyclicGraph.AddNode(2)
	node3 := acyclicGraph.AddNode(3)
	node4 := acyclicGraph.AddNode(4)

	acyclicGraph.AddEdge(node1, node2, 1)
	acyclicGraph.AddEdge(node1, node3, 1)
	acyclicGraph.AddEdge(node3, node4, 1)

	if acyclicGraph.IsCyclic() {
		t.Error("Acyclic undirected graph incorrectly identified as cyclic")
	}

	// Test cyclic undirected graph
	cyclicGraph := NewUndirectedGraph[int]()
	cNode1 := cyclicGraph.AddNode(1)
	cNode2 := cyclicGraph.AddNode(2)
	cNode3 := cyclicGraph.AddNode(3)

	cyclicGraph.AddEdge(cNode1, cNode2, 1)
	cyclicGraph.AddEdge(cNode2, cNode3, 1)
	cyclicGraph.AddEdge(cNode3, cNode1, 1) // Creates a cycle

	if !cyclicGraph.IsCyclic() {
		t.Error("Cyclic undirected graph not identified as cyclic")
	}
}

// Helper function to create a test directed graph for reuse
func createTestDirectedGraph() *DirectedGraph[int] {
	g := NewDirectedGraph[int]()

	node1 := g.AddNode(1)
	node2 := g.AddNode(2)
	node3 := g.AddNode(3)
	node4 := g.AddNode(4)
	node5 := g.AddNode(5)

	g.AddEdge(node1, node2, 10)
	g.AddEdge(node1, node3, 5)
	g.AddEdge(node2, node4, 1)
	g.AddEdge(node3, node5, 2)

	return g
}

// Helper function to create a test undirected graph for reuse
func createTestUndirectedGraph() *UndirectedGraph[int] {
	g := NewUndirectedGraph[int]()

	node1 := g.AddNode(1)
	node2 := g.AddNode(2)
	node3 := g.AddNode(3)
	node4 := g.AddNode(4)
	node5 := g.AddNode(5)

	g.AddEdge(node1, node2, 10)
	g.AddEdge(node1, node3, 5)
	g.AddEdge(node2, node4, 1)
	g.AddEdge(node3, node5, 2)

	return g
}

// Helper function to check if a slice contains a value
func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
