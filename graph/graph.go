package graph

import (
	"crypto/md5"
	"fmt"
	"slices"
	"strings"

	"github.com/h2non/filetype"
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

type LinkType string

var RegularLink = LinkType("REGULAR")

type SkaffoldNode interface {
	Children() []SkaffoldNode
	AddChild(child SkaffoldNode) error
	Parent() (SkaffoldNode, error)
	SetParent(parent SkaffoldNode) error
	Key() string
	Type() string
	CollisionAction() CollisionAction
}

type SkaffoldLink struct {
	Target   SkaffoldNode
	LinkType LinkType
	Name     string
}
type DirectoryNode struct {
	name     string         // Name of the file or directory
	children []SkaffoldLink // Child nodes (nil for files, populated for directories)
	parent   SkaffoldNode   // Optional: Pointer to the parent node, might be useful later
}

// NewDirectoryNode creates a new DirectoryNode.
func NewDirectoryNode(name string) *DirectoryNode {
	return &DirectoryNode{
		name:     name,
		children: make([]SkaffoldLink, 0), // Initialize slice
	}
}

func NewDirectoryNodeWithParent(name string, parent SkaffoldNode) *DirectoryNode {
	n := NewDirectoryNode(name)
	n.parent = parent
	return n
}

func (d *DirectoryNode) CollisionAction() CollisionAction {
	return DefaultOnCollision
}

func (d *DirectoryNode) Children() []SkaffoldNode {
	nodes := make([]SkaffoldNode, len(d.children))
	for i, link := range d.children {
		nodes[i] = link.Target
	}
	return nodes
}

func (d *DirectoryNode) AddChild(child SkaffoldNode) error {
	// Potentially check for duplicate keys or handle existing child with same key
	link := SkaffoldLink{
		Target:   child,
		LinkType: RegularLink, // Assuming RegularLink as default
		Name:     child.Key(), // Using child's key as the link name
	}
	d.children = append(d.children, link)
	// Consider child.SetParent(d) if parent pointers should be actively managed here
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
	datahash     []byte
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

func (f *FileNode) SetContent(src []byte) error {
	if len(src) == 0 {
		return nil
	}

	kind, _ := filetype.Match(src)
	if kind != filetype.Unknown {
		f.content_type = kind.MIME.Value
	}

	md5sum := md5.Sum(src)
	f.datahash = md5sum[:]
	return nil
}
func (f *FileNode) CollisionAction() CollisionAction {
	return DefaultOnCollision
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
	u := control

	if len(add) == 0 {
		return control, nil
	}

	for _, n := range add {
		if n.Key() != control.Key() {
			return nil, fmt.Errorf("mismatched roots: %s does not match %s", n.Key(), control.Key())
		}
	}

	mergedKeys := make([]string, 0)

	for _, c := range control.Children() {
		mergedKeys = append(mergedKeys, c.Key())

		matchingChildrenFromAdd := make([]SkaffoldNode, 0)
		for _, n_graph := range add { // n_graph is one of the 'add' graphs
			for _, child_of_n_graph := range n_graph.Children() {
				if child_of_n_graph.Key() == c.Key() {
					matchingChildrenFromAdd = append(matchingChildrenFromAdd, child_of_n_graph)
				}
			}
		}

		// Recursively merge 'c' (from control) with all matching children from 'add' graphs.
		// This call modifies 'c' in-place if 'c' is a DirectoryNode and has further children to merge.
		// The returned 'm' is the (potentially modified) 'c'.
		_, err := Union(opts, c, matchingChildrenFromAdd...)
		if err != nil {
			return nil, fmt.Errorf("error merging children: %w", err)
		}
		// DO NOT u.AddChild(m) here. 'c' (which is 'm') is already a child of 'u' (which is 'control')
		// and has been modified in-place by the recursive Union call.
	}

	// Append all children in the graphs in add that do not appear in control
	var addChildren SkaffoldNode
	var err error

	if len(add) == 0 {
		// No 'add' graphs, so no new children to add from them.
		// The initial check 'if len(add) == 0 { return control, nil }' handles the case of no 'add' graphs entirely.
	} else if len(add) > 1 {
		addChildren, err = Union(opts, add[0], add[1:]...)
		if err != nil {
			return nil, fmt.Errorf("error merging external children: %w", err)
		}
	} else {
		addChildren = add[0]
	}

	if addChildren != nil {
		for _, childFromAdd := range addChildren.Children() {
			if !slices.Contains(mergedKeys, childFromAdd.Key()) {
				u.AddChild(childFromAdd) // Add new children from the 'add' side to 'u'
			}
		}
	}
	return u, nil
}
