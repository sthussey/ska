package fs

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/sthussey/ska/graph"
)

// BuildGraph walks the directory tree starting at rootPath and builds a graph.
func BuildGraph(rootPath string) (graph.SkaffoldNode, error) {
	absRootPath, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path for %s: %w", rootPath, err)
	}

	// Get info about the root path
	info, err := os.Stat(absRootPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat root path %s: %w", absRootPath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("root path %s is not a directory", absRootPath)
	}

	// Create the root node using the base name of the absolute path
	rootNode := graph.NewDirectoryNode(filepath.Base(absRootPath))

	// Start the recursive walk
	err = walkDir(absRootPath, rootNode)
	if err != nil {
		return nil, err // Error already contains context from walkDir
	}

	return rootNode, nil
}

// walkDir recursively walks the directory structure under dirPath
// and adds nodes to the parentNode.
func walkDir(dirPath string, parentNode *graph.DirectoryNode) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	for _, entry := range entries {
		// Construct the full path for the current entry
		fullPath := filepath.Join(dirPath, entry.Name())

		if entry.IsDir() {
			// Create a new directory node
			dirNode := graph.NewDirectoryNode(entry.Name())

			// Set parent relationship (error ignored as SetParent currently always returns nil)
			_ = dirNode.SetParent(parentNode)
			_ = parentNode.AddChild(dirNode)

			// Recursively walk the subdirectory
			err = walkDir(fullPath, dirNode)
			if err != nil {
				return err // Propagate errors from deeper levels
			}
		} else {
			// Create a new file node
			fileNode := graph.NewFileNode(entry.Name())

			// Very naive, large files break here
			content, err := os.ReadFile(fullPath)
			// this eats errors for now. need to determine how fatal not being able to hash a file is
			if err == nil {
				fileNode.SetContent(content)
			}

			// Set parent relationship (error ignored as SetParent currently always returns nil)
			_ = fileNode.SetParent(parentNode)
			_ = parentNode.AddChild(fileNode)

			// Action is already set in NewFileNode based on extension
			// You could add more logic here later if needed (e.g., read content type)
		}
	}
	return nil
}
