package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	filename, limit := processArgs()
	limit += 2

	csvFile, err := os.Open(filename)
	if err != nil {
		log.Fatalf("Error opening file %s: %s", filename, err)
	}
	defer csvFile.Close()

	logs, err := ReadFromCSV(csvFile)
	if err != nil {
		log.Fatalf("Error reading file %s: %s", filename, err)
	}
	fmt.Fprintf(os.Stderr, "Found %d visits, ", len(logs))

	sessionMap := logs.ToSessions()
	fmt.Fprintf(os.Stderr, "divided over %d sessions, ", len(sessionMap))
	fmt.Fprintf(os.Stderr, "so on average %0.00f clicks per session.\n", float32(len(logs))/float32(len(sessionMap)))

	graph := NewGraph(sessionMap)
	fmt.Fprintf(os.Stderr, "Entire data-set contains %d nodes. ", graph.NodeCount())

	reducedGraph := graph.ReduceTo(limit)
	fmt.Fprintf(os.Stderr, "After reducing to top %d, ", reducedGraph.NodeCount()-2)

	fmt.Fprintf(os.Stderr, "%0.f %% of all clicks is still represented\n",
		(float32(reducedGraph.totalClickCount)/float32(graph.totalClickCount))*100)

	reducedGraph.Print(graph.totalClickCount)
}

func processArgs() (string, int) {
	filename := flag.String("input-filename", "logs.csv", "CSV file to read")
	nodeCount := flag.Int("limit", 10, "Amount of nodes to display")

	flag.Parse()

	return *filename, *nodeCount
}
