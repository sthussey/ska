package ska

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const NODETYPE_DIRECTORY = "DIRECTORY" //nolint:revive // ignore ST1003
const NODETYPE_FILE = "FILE"

type SkaffoldNode interface {
	Children() []SkaffoldNode
	AddChild(child SkaffoldNode) error
	Parent() (SkaffoldNode, error)
	SetParent(parent SkaffoldNode) error
	Key() string
	Type() string
}

type DirectoryNode struct {
	name     string         // Name of the file or directory
	children []SkaffoldNode // Child nodes (nil for files, populated for directories)
	parent   SkaffoldNode   // Optional: Pointer to the parent node, might be useful later
}

// NewDirectoryNode creates a new DirectoryNode.
func NewDirectoryNode(name string) *DirectoryNode {
	return &DirectoryNode{
		name:     name,
		children: make([]SkaffoldNode, 0), // Initialize slice
	}
}

func NewDirectoryNodeWithParent(name string, parent SkaffoldNode) *DirectoryNode {
	n := NewDirectoryNode(name)
	n.parent = parent
	return n
}

func (d *DirectoryNode) Children() []SkaffoldNode {
	return d.children
}

func (d *DirectoryNode) AddChild(child SkaffoldNode) error {
	d.children = append(d.children, child)
	return nil
}

func (d *DirectoryNode) Parent() (SkaffoldNode, error) {
	if d.parent == nil {
		return nil, fmt.Errorf("node %s has no parent", d.name)
	}
	return d.parent, nil
}

func (d *DirectoryNode) SetParent(parent SkaffoldNode) error {
	d.parent = parent
	return nil
}

func (d *DirectoryNode) Key() string {
	return d.name // Assuming Name is unique enough for a key within its context
}

func (d *DirectoryNode) Type() string {
	return NODETYPE_DIRECTORY
}

const FILEACTION_COPY = "COPY"
const FILEACTION_TEMPLATE = "TEMPLATE"

type FileNode struct {
	name         string
	action       string
	data         []byte
	content_type string
	parent       SkaffoldNode
}

// NewFileNode creates a new FileNode.
func NewFileNode(name string) *FileNode {
	// Default action to COPY, can be overridden
	action := FILEACTION_COPY
	// Simple check for template files
	if strings.HasSuffix(name, ".tmpl") {
		action = FILEACTION_TEMPLATE
	}
	return &FileNode{
		name:   name,
		action: action,
	}
}

func NewFileNodeWithParent(name string, parent SkaffoldNode) *FileNode {
	n := NewFileNode(name)
	n.parent = parent
	return n
}

func (f *FileNode) Children() []SkaffoldNode {
	return []SkaffoldNode{}
}

func (f *FileNode) AddChild(child SkaffoldNode) error {
	return fmt.Errorf("cannot add child to a file node %s", f.name)
}

func (f *FileNode) Parent() (SkaffoldNode, error) {
	if f.parent == nil {
		return nil, fmt.Errorf("node %s has no parent", f.name)
	}
	return f.parent, nil
}

func (f *FileNode) SetParent(parent SkaffoldNode) error {
	f.parent = parent
	return nil
}

func (f *FileNode) Key() string {
	return f.name // Assuming Name is unique enough for a key within its context
}

func (f *FileNode) Type() string {
	return NODETYPE_FILE
}

func (f *FileNode) Action() string {
	return f.action
}

func (f *FileNode) SetAction(action string) error {
	if action != FILEACTION_COPY && action != FILEACTION_TEMPLATE {
		return fmt.Errorf("invalid action %s for file %s", action, f.name)
	}
	f.action = action
	return nil
}

func (f *FileNode) ContentType() string {
	return f.content_type
}

// BuildGraph walks the directory tree starting at rootPath and builds a graph.
func BuildGraph(rootPath string) (SkaffoldNode, error) {
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
	rootNode := NewDirectoryNode(filepath.Base(absRootPath))

	// Start the recursive walk
	err = walkDir(absRootPath, rootNode)
	if err != nil {
		return nil, err // Error already contains context from walkDir
	}

	return rootNode, nil
}

// walkDir recursively walks the directory structure under dirPath
// and adds nodes to the parentNode.
func walkDir(dirPath string, parentNode *DirectoryNode) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %w", dirPath, err)
	}

	for _, entry := range entries {
		// Construct the full path for the current entry
		fullPath := filepath.Join(dirPath, entry.Name())

		if entry.IsDir() {
			// Create a new directory node
			dirNode := NewDirectoryNode(entry.Name())

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
			fileNode := NewFileNode(entry.Name())

			// Set parent relationship (error ignored as SetParent currently always returns nil)
			_ = fileNode.SetParent(parentNode)
			_ = parentNode.AddChild(fileNode)

			// Action is already set in NewFileNode based on extension
			// You could add more logic here later if needed (e.g., read content type)
		}
	}
	return nil
}
