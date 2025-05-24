package console

import (
	"fmt"
	"strings"

	"github.com/sthussey/ska/graph"
)

// PrintGraph recursively prints a graph node and its children with indentation
func PrintGraph(node graph.SkaffoldNode, level int) {
	// Create indentation based on level
	indent := strings.Repeat("  ", level)

	// Print current node
	nodeType := ""
	if node.Type() == graph.NODETYPE_DIRECTORY {
		nodeType = "[DIR]"
	} else if node.Type() == graph.NODETYPE_FILE {
		// Try to cast to FileNode to get action
		if fileNode, ok := node.(interface{ Action() string }); ok {
			nodeType = fmt.Sprintf("[FILE:%s]", fileNode.Action())
		} else {
			nodeType = "[FILE]"
		}
	}

	fmt.Printf("%s%s %s\n", indent, nodeType, node.Key())

	// Print children recursively
	for _, child := range node.Children() {
		PrintGraph(child, level+1)
	}
}
