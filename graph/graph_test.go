package graph

import (
	"testing"
)

// Helper function to find a child by key
func findChildByKey(node SkaffoldNode, key string) SkaffoldNode {
	if node == nil {
		return nil
	}
	for _, child := range node.Children() {
		if child.Key() == key {
			return child
		}
	}
	return nil
}

func TestUnion_MergeTwoGraphs(t *testing.T) {
	// Graph 1 (control)
	// root/
	//   dir1/
	//     file1_control.txt
	//   file_control_root.txt

	file1Control := NewFileNode("file1_control.txt")
	dir1Control := NewDirectoryNode("dir1")
	dir1Control.AddChild(file1Control)

	fileControlRoot := NewFileNode("file_control_root.txt")
	graph1Root := NewDirectoryNode("root")
	graph1Root.AddChild(dir1Control)
	graph1Root.AddChild(fileControlRoot)

	// Graph 2 (add)
	// root/
	//   dir1/  (common directory name)
	//     file2_add.txt
	//   dir2_add/
	//     file3_add_in_dir2.txt
	//   file_add_root.txt

	file2Add := NewFileNode("file2_add.txt")
	dir1Add := NewDirectoryNode("dir1") // Same key as dir1Control
	dir1Add.AddChild(file2Add)

	file3AddInDir2 := NewFileNode("file3_add_in_dir2.txt")
	dir2Add := NewDirectoryNode("dir2_add")
	dir2Add.AddChild(file3AddInDir2)

	fileAddRoot := NewFileNode("file_add_root.txt")
	graph2Root := NewDirectoryNode("root") // Same key as graph1Root
	graph2Root.AddChild(dir1Add)
	graph2Root.AddChild(dir2Add)
	graph2Root.AddChild(fileAddRoot)

	// Merge options
	opts := MergeOptions{DefaultCollisionAction: ErrorOnCollision} // Note: CollisionAction not fully used by current Union for content

	// Perform the Union
	mergedGraph, err := Union(opts, graph1Root, graph2Root)

	if err != nil {
		tFatalf(t, "Union returned an error: %v", err)
	}

	if mergedGraph == nil {
		tFatalf(t, "Union returned a nil graph")
	}

	// --- Assertions on the merged graph structure ---
	// Expected:
	// root/
	//   dir1/ (merged)
	//     file1_control.txt (from graph1)
	//     file2_add.txt     (from graph2)
	//   file_control_root.txt (from graph1)
	//   dir2_add/           (from graph2)
	//     file3_add_in_dir2.txt
	//   file_add_root.txt   (from graph2)

	if mergedGraph.Key() != "root" {
		t.Errorf("Expected merged graph root key 'root', got '%s'", mergedGraph.Key())
	}
	if len(mergedGraph.Children()) != 4 {
		t.Errorf("Expected merged graph to have 4 children at root, got %d", len(mergedGraph.Children()))
		for _, child := range mergedGraph.Children() {
			t.Logf("Child: %s (Type: %s)", child.Key(), child.Type())
		}
	}

	// Check dir1
	mergedDir1 := findChildByKey(mergedGraph, "dir1")
	if mergedDir1 == nil {
		tFatalf(t, "Expected to find 'dir1' in merged graph")
	}
	if mergedDir1.Type() != NODETYPE_DIRECTORY {
		t.Errorf("Expected 'dir1' to be a DIRECTORY, got %s", mergedDir1.Type())
	}
	if len(mergedDir1.Children()) != 2 {
		t.Errorf("Expected 'dir1' to have 2 children, got %d", len(mergedDir1.Children()))
	}
	if findChildByKey(mergedDir1, "file1_control.txt") == nil {
		t.Errorf("Expected 'file1_control.txt' in 'dir1'")
	}
	if findChildByKey(mergedDir1, "file2_add.txt") == nil {
		t.Errorf("Expected 'file2_add.txt' in 'dir1'")
	}

	// Check other root children
	if findChildByKey(mergedGraph, "file_control_root.txt") == nil {
		t.Errorf("Expected 'file_control_root.txt' at root")
	}
	mergedDir2Add := findChildByKey(mergedGraph, "dir2_add")
	if mergedDir2Add == nil || findChildByKey(mergedDir2Add, "file3_add_in_dir2.txt") == nil {
		t.Errorf("Expected 'dir2_add/file3_add_in_dir2.txt' structure")
	}
	if findChildByKey(mergedGraph, "file_add_root.txt") == nil {
		t.Errorf("Expected 'file_add_root.txt' at root")
	}
}

func tFatalf(t *testing.T, format string, args ...interface{}) {
	t.Helper()
	t.Fatalf(format, args...)
}

func TestUnion_MergeThreeGraphs(t *testing.T) {
	// Graph 1 (control)
	// root/
	//   common_dir/
	//     control_file1.txt
	//   control_only_file.txt
	cf1 := NewFileNode("control_file1.txt")
	cd1 := NewDirectoryNode("common_dir")
	cd1.AddChild(cf1)
	controlRootFile := NewFileNode("control_only_file.txt")
	graph1Root := NewDirectoryNode("root")
	graph1Root.AddChild(cd1)
	graph1Root.AddChild(controlRootFile)

	// Graph 2 (add1)
	// root/
	//   common_dir/  (common directory name)
	//     add1_file2.txt
	//   add1_only_dir/
	//     add1_file3.txt
	add1f2 := NewFileNode("add1_file2.txt")
	add1d1 := NewDirectoryNode("common_dir") // Same key as cd1
	add1d1.AddChild(add1f2)
	add1f3 := NewFileNode("add1_file3.txt")
	add1d2 := NewDirectoryNode("add1_only_dir")
	add1d2.AddChild(add1f3)
	graph2Root := NewDirectoryNode("root") // Same key as graph1Root
	graph2Root.AddChild(add1d1)
	graph2Root.AddChild(add1d2)

	// Graph 3 (add2)
	// root/
	//   common_dir/ (common directory name)
	//     add2_file4.txt
	//   add2_only_file.txt
	add2f4 := NewFileNode("add2_file4.txt")
	add2d1 := NewDirectoryNode("common_dir") // Same key as cd1
	add2d1.AddChild(add2f4)
	add2RootFile := NewFileNode("add2_only_file.txt")
	graph3Root := NewDirectoryNode("root") // Same key as graph1Root
	graph3Root.AddChild(add2d1)
	graph3Root.AddChild(add2RootFile)

	// Merge options
	opts := MergeOptions{DefaultCollisionAction: ErrorOnCollision}

	// Perform the Union
	mergedGraph, err := Union(opts, graph1Root, graph2Root, graph3Root)

	if err != nil {
		tFatalf(t, "Union returned an error: %v", err)
	}

	if mergedGraph == nil {
		tFatalf(t, "Union returned a nil graph")
	}

	// --- Assertions on the merged graph structure ---
	// Expected:
	// root/
	//   common_dir/ (merged)
	//     control_file1.txt (from graph1)
	//     add1_file2.txt    (from graph2)
	//     add2_file4.txt    (from graph3)
	//   control_only_file.txt (from graph1)
	//   add1_only_dir/        (from graph2)
	//     add1_file3.txt
	//   add2_only_file.txt    (from graph3)

	if mergedGraph.Key() != "root" {
		t.Errorf("Expected merged graph root key 'root', got '%s'", mergedGraph.Key())
	}
	// common_dir, control_only_file.txt, add1_only_dir, add2_only_file.txt
	if len(mergedGraph.Children()) != 4 {
		t.Errorf("Expected merged graph to have 4 children at root, got %d", len(mergedGraph.Children()))
		for _, child := range mergedGraph.Children() {
			t.Logf("Root Child: %s (Type: %s)", child.Key(), child.Type())
		}
	}

	// Check common_dir
	mergedCommonDir := findChildByKey(mergedGraph, "common_dir")
	if mergedCommonDir == nil {
		tFatalf(t, "Expected to find 'common_dir' in merged graph")
	}
	if mergedCommonDir.Type() != NODETYPE_DIRECTORY {
		t.Errorf("Expected 'common_dir' to be a DIRECTORY, got %s", mergedCommonDir.Type())
	}
	// control_file1.txt, add1_file2.txt, add2_file4.txt
	if len(mergedCommonDir.Children()) != 3 {
		t.Errorf("Expected 'common_dir' to have 3 children, got %d", len(mergedCommonDir.Children()))
	}
	if findChildByKey(mergedCommonDir, "control_file1.txt") == nil {
		t.Errorf("Expected 'control_file1.txt' in 'common_dir'")
	}
	if findChildByKey(mergedCommonDir, "add1_file2.txt") == nil {
		t.Errorf("Expected 'add1_file2.txt' in 'common_dir'")
	}
	if findChildByKey(mergedCommonDir, "add2_file4.txt") == nil {
		t.Errorf("Expected 'add2_file4.txt' in 'common_dir'")
	}

	// Check other root children
	if findChildByKey(mergedGraph, "control_only_file.txt") == nil {
		t.Errorf("Expected 'control_only_file.txt' at root")
	}
	mergedAdd1OnlyDir := findChildByKey(mergedGraph, "add1_only_dir")
	if mergedAdd1OnlyDir == nil || findChildByKey(mergedAdd1OnlyDir, "add1_file3.txt") == nil {
		t.Errorf("Expected 'add1_only_dir/add1_file3.txt' structure")
	}
	if findChildByKey(mergedGraph, "add2_only_file.txt") == nil {
		t.Errorf("Expected 'add2_only_file.txt' at root")
	}
}
