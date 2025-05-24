package graph

import (
	"fmt"
	"strings"
)

const NODETYPE_DIRECTORY = "DIRECTORY" //nolint:revive // ignore ST1003
const NODETYPE_FILE = "FILE"

type CollisionAction string

// A collision is defined as two identitically keyed nodes ocurring in the graph with the same prefix
// but non-reconciliable data. Two file nodes with the same name at the same prefix in the graph but
// differing content data.
var ErrorOnCollision = CollisionAction("ERROR")         // Abort and return error when merging nodes collide
var OverwriteOnCollision = CollisionAction("OVERWRITE") // The controlling graph node replaces other nodes
var YieldOnCollision = CollisionAction("YIELD")         // The controlling graph node yields to other nodes. If all nodes in the merge yield, the control node is chosen.
var DefaultOnCollision = CollisionAction("DEFAULT")     // The action is chosen based on the merge options specified

type SkaffoldNode interface {
	Children() []SkaffoldNode
	AddChild(child SkaffoldNode) error
	Parent() (SkaffoldNode, error)
	SetParent(parent SkaffoldNode) error
	Key() string
	Type() string
	CollisionAction() CollisionAction
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

type MergeOptions struct {
	DefaultCollisionAction CollisionAction
}

func Union(opts MergeOptions, control SkaffoldNode, add ...SkaffoldNode) (SkaffoldNode, error) {

}
