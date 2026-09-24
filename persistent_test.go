package main

import "testing"

func TestSaveLoadRoundTrip(t *testing.T) {
	root := &RootNode{
		files:    map[string]*FileData{"a.txt": {content: "hello", mode: 0644}},
		subdirs:  map[string]*RootNode{},
		symlinks: map[string]*SymLink{},
	}

	dir := toSerializable(root)
	restored := fromSerializable(dir)

	if restored.files["a.txt"].content != "hello" {
		t.Errorf("content mismatch after round-trip")
	}
}