package main

import (
	"fmt"
	"os"
	"sort"
)

type Graph struct {
	totalClickCount int
	nodeSlice       []Node
	nodeMap         map[string]Node
	edgeMap         map[string]Edge
}

type Node struct {
	Name       string
	Edges      []Edge
	VisitCount int
	Rank       float32
}

func (n Node) String() string {
	return fmt.Sprintf("node: %s (%0.f%%)", n.Name, n.Rank*100)
}

type Edge struct {
	Name        string
	Origin      string
	Destination string
	VisitCount  int
	Rank        float32
	LocalRank   float32
}

func (e Edge) String() string {
	return fmt.Sprintf("edge: %s -> %s (%0.f%%)", e.Origin, e.Destination, e.LocalRank*100)
}

func NewGraph(sessionMap map[string][]Log) Graph {
	nodeMap := createNodeMapWithVisitCount(sessionMap)
	edgeMap := createEdgeMapWithVisitCount(sessionMap)
	clickCount := calculateClickCount(edgeMap)
	enrichedNodeMap := enrichNodesWithRankAndEdges(nodeMap, edgeMap, clickCount)
	nodeSlice := nodeMapToSortedSlice(enrichedNodeMap)

	return Graph{
		totalClickCount: clickCount,
		nodeMap:         enrichedNodeMap,
		edgeMap:         edgeMap,
		nodeSlice:       nodeSlice,
	}
}

func (g Graph) String() string {
	return fmt.Sprintf("node-count:%d - total-click-count: %d", g.NodeCount(), g.totalClickCount)
}

func createNodeMapWithVisitCount(sessionMap map[string][]Log) map[string]Node {
	nodeMap := map[string]Node{}
	for _, sessionLogs := range sessionMap {
		for _, log := range sessionLogs {
			node, found := nodeMap[log.ScreenName]
			if !found {
				node = Node{
					Name:       log.ScreenName,
					Edges:      []Edge{},
					VisitCount: 0,
				}
			}
			node.VisitCount++
			nodeMap[log.ScreenName] = node
		}
	}
	return nodeMap
}

func calculateClickCount(edgeMap map[string]Edge) int {
	totalVisits := 0
	for _, edge := range edgeMap {
		if edge.Destination != "end" {
			totalVisits += edge.VisitCount
		}
	}
	return totalVisits
}

func createEdgeMapWithVisitCount(sessionMap map[string][]Log) map[string]Edge {
	// count visits for edges
	edgeMap := map[string]Edge{}
	for _, sessionLogs := range sessionMap {
		prevName := ""
		for _, log := range sessionLogs {
			if prevName != "" {
				edgeName := fmt.Sprintf("%s-%s", prevName, log.ScreenName)
				edge, found := edgeMap[edgeName]
				if !found {
					edge = Edge{
						Name:        edgeName,
						Origin:      prevName,
						Destination: log.ScreenName,
						VisitCount:  0,
					}
				}
				edge.VisitCount++
				edgeMap[edgeName] = edge
			}
			prevName = log.ScreenName
		}
	}
	return edgeMap
}

func enrichNodesWithRankAndEdges(nodeMap map[string]Node, edgeMap map[string]Edge, totalVisitCount int) map[string]Node {
	for _, edge := range edgeMap {
		origNode, exists := nodeMap[edge.Origin]
		if exists {
			origNode.Edges = append(origNode.Edges, edge)
			nodeMap[edge.Origin] = origNode
		}
	}
	for nodeName, node := range nodeMap {
		// Order descending on visit-count
		sort.Slice(node.Edges, func(i, j int) bool {
			return node.Edges[i].VisitCount > node.Edges[j].VisitCount
		})
		node.Rank = float32(node.VisitCount) / float32(totalVisitCount)
		for idx := range node.Edges {
			node.Edges[idx].Rank = float32(node.Edges[idx].VisitCount) / float32(totalVisitCount)
			node.Edges[idx].LocalRank = float32(node.Edges[idx].VisitCount) / float32(node.VisitCount)
		}
		nodeMap[nodeName] = node
	}
	return nodeMap
}

func nodeMapToSortedSlice(nodeMap map[string]Node) []Node {
	nodeList := []Node{}
	for _, node := range nodeMap {
		nodeList = append(nodeList, node)
	}
	sort.Slice(nodeList, func(i, j int) bool {
		// sort descending on visit-count
		return nodeList[i].VisitCount > nodeList[j].VisitCount
	})
	return nodeList
}

func nodeSliceToMap(nodeList []Node) map[string]Node {
	nodeMap := map[string]Node{}
	for _, node := range nodeList {
		nodeMap[node.Name] = node
	}
	return nodeMap
}

func (g Graph) ReduceTo(limit int) Graph {
	reducedNodeSlice := g.nodeSlice[0:min(limit, len(g.nodeSlice)-1)]
	reducedNodeMap := nodeSliceToMap(reducedNodeSlice)
	reducedEdgeMap := filterEdgeMap(g.edgeMap, reducedNodeMap)
	reducedClickCount := calculateClickCount(reducedEdgeMap)
	strippedNodeMap := stripNodeMap(reducedNodeMap, reducedEdgeMap)

	return Graph{
		totalClickCount: reducedClickCount,
		nodeSlice:       nodeMapToSortedSlice(strippedNodeMap),
		nodeMap:         strippedNodeMap,
		edgeMap:         reducedEdgeMap,
	}
}

func stripNodeMap(nodeMap map[string]Node, edgeMap map[string]Edge) map[string]Node {
	for nodeName, node := range nodeMap {
		reducedEdges := []Edge{}
		for _, edge := range node.Edges {
			_, exists := edgeMap[edge.Name]
			if exists {
				reducedEdges = append(reducedEdges, edge)
				node.Edges = reducedEdges
			}
		}
		nodeMap[nodeName] = node
	}
	return nodeMap
}

func filterEdgeMap(edgeMap map[string]Edge, nodeMap map[string]Node) map[string]Edge {
	reducedEdgeMap := map[string]Edge{}
	for edgeName, edge := range edgeMap {
		_, existsOrigin := nodeMap[edge.Origin]
		_, existsDestination := nodeMap[edge.Destination]

		if existsOrigin && existsDestination {
			reducedEdgeMap[edgeName] = edge
		}
	}
	return reducedEdgeMap
}

func factorRepresented(reducedClickCount, totalClickCount int) float32 {
	return float32(reducedClickCount) / float32(totalClickCount)
}

func (g Graph) NodeCount() int {
	return len(g.nodeSlice)
}

func (g Graph) Print(originalTotalClickCount int) {
	g.printGraphvizHeader()

	g.printNodesAndEdges(originalTotalClickCount)

	factor := factorRepresented(g.totalClickCount, originalTotalClickCount)

	g.printTitle(factor)

	g.printGraphvizFooter()

	g.printDebug(factor)

}

func (g Graph) printDebug(factorRepresented float32) {
	var nodeSum float32 = 0.0
	for _, node := range g.nodeSlice {
		fmt.Fprintf(os.Stderr, "%s\n", node)

		var sofar float32 = 0.0
		for _, edge := range node.Edges {
			sofar += edge.LocalRank
			if sofar <= factorRepresented {
				fmt.Fprintf(os.Stderr, "\t%s (%0.f)\n", edge, sofar*100)
			} else {
				fmt.Fprintf(os.Stderr, "\t(%s (%0.f))\n", edge, sofar*100)
			}
		}
		if node.Name != "start" && node.Name != "end" {
			nodeSum += node.Rank
		}
	}
	fmt.Fprintf(os.Stderr, "***Node sum: %d %0.f\n", g.totalClickCount, nodeSum*100)
}

func (g Graph) printGraphvizHeader() {
	fmt.Printf("\n\ndigraph mygraph {\n")
	fmt.Printf("\trankdir = \"TD\"")
}

func (g Graph) printNodesAndEdges(originalTotalCount int) {
	for _, node := range g.nodeSlice {
		// print node
		if node.Name == "start" {
			fmt.Printf("\n\t\"start\" [shape=circle, style=filled, color=black, fontcolor=white];\n")
		} else if node.Name == "end" {
			fmt.Printf("\n\t\"end\" [shape=circle, style=filled, color=black, fontcolor=white];\n")
		} else {
			factor := float32(node.VisitCount) / float32(originalTotalCount)
			fmt.Printf(
				"\n\t\"%s\" [label=\"%s\\n%0.f%% (%d)\", penwidth=%0.00f, color=%s, href=\"%s\"];\n",
				node.Name,
				node.Name,
				factor*100,
				node.VisitCount,
				determineNodePenwidth(factor),
				determineNodeColor(factor),
				createURL(node.Name))
		}

		// print edges
		factorRepresented := factorRepresented(g.totalClickCount, originalTotalCount)
		reducedEdges := reduceEdgesUpTo(node.Edges, factorRepresented)
		for _, edge := range reducedEdges {
			fmt.Printf(
				"\t\"%s\" -> \"%s\" [label=\"%0.f%% (%d)\", penwidth=%0.00f, color=%s];\n",
				edge.Origin,
				edge.Destination,
				edge.LocalRank*100,
				edge.VisitCount,
				determineEdgePenwidth(edge.Rank),
				determineEdgeColor(edge.Rank))
		}
	}
}

func determineNodeColor(factor float32) string {
	if isBetween(factor, 0.00, 0.1) {
		return "grey"
	}
	if isBetween(factor, 0.1, 0.2) {
		return "black"
	}
	if isBetween(factor, 0.2, 0.3) {
		return "orange"
	}

	return "red"
}

func determineNodePenwidth(factor float32) float32 {
	if isBetween(factor, 0.00, 0.1) {
		return 1.0
	}
	if isBetween(factor, 0.1, 0.2) {
		return 2.0
	}
	if isBetween(factor, 0.2, 0.3) {
		return 3.0
	}
	if isBetween(factor, 0.3, 0.4) {
		return 4.0
	}
	return 6.0
}

func createURL(nodeName string) string {
	return fmt.Sprintf("%s/ca/ca/%s", "https://ca-test.adyen.com", nodeName)
}

func determineEdgePenwidth(factor float32) float32 {
	if isBetween(factor, 0.00, 0.01) {
		return 1.0
	}
	if isBetween(factor, 0.01, 0.02) {
		return 2.0
	}
	if isBetween(factor, 0.02, 0.03) {
		return 3.0
	}
	if isBetween(factor, 0.03, 0.04) {
		return 4.0
	}

	return 6.0
}

func determineEdgeColor(factor float32) string {
	if isBetween(factor, 0.0, 0.01) {
		return "grey"
	}
	if isBetween(factor, 0.01, 0.02) {
		return "black"
	}
	if isBetween(factor, 0.02, 0.03) {
		return "orange"
	}
	return "red"
}

func isBetween(value, lowerInclusive, upperExclusive float32) bool {
	return (value >= lowerInclusive && value < upperExclusive)
}

func reduceEdgesUpTo(edges []Edge, upto float32) []Edge {
	reducedEdges := []Edge{}
	var sofar float32 = 0.0
	for _, edge := range edges {
		sofar += edge.LocalRank
		if sofar <= upto {
			reducedEdges = append(reducedEdges, edge)
		}
	}
	return reducedEdges
}

func (g Graph) printTitle(factorClickRepresented float32) {
	fmt.Printf("\n\tfontsize = \"40\"\n")
	fmt.Printf("\tlabel=\"Top %d screens represent %0.f %% of CA clicks\"\n", g.NodeCount()-2, factorClickRepresented*100)
	fmt.Printf("\tlabelloc=\"t\"\n\n")
}

func (g Graph) printGraphvizFooter() (int, error) {
	return fmt.Printf("}\n")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
