package main

import (
	"os"
	"testing"

	"github.com/hanwen/go-fuse/v2/fs"
)

// mountTestFS mounts a fresh RootNode at a temp directory and returns
// the mount point path. It registers cleanup to unmount automatically
// when the test finishes.
func mountTestFS(t *testing.T) string {
	t.Helper()

	mntDir := t.TempDir()

	root := &RootNode{
		files:    map[string]*FileData{},
		subdirs:  map[string]*RootNode{},
		symlinks: map[string]*SymLink{},
		mode:     0755,
	}

	// Create/Setattr/etc. call Save(globalRoot) internally, so it must
	// point at the same root we're mounting, otherwise Save panics on nil.
	globalRoot = root

	server, err := fs.Mount(mntDir, root, &fs.Options{})
	if err != nil {
		t.Fatalf("mount failed: %v", err)
	}

	t.Cleanup(func() {
		if err := server.Unmount(); err != nil {
			t.Logf("unmount failed: %v", err)
		}
	})

	return mntDir
}

func TestMountAndCreate(t *testing.T) {
	mntDir := mountTestFS(t)

	filePath := mntDir + "/test.txt"

	if err := os.WriteFile(filePath, []byte("hello"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(data) != "hello" {
		t.Errorf("expected content %q, got %q", "hello", string(data))
	}
}

func TestMountAndChmod(t *testing.T) {
	mntDir := mountTestFS(t)

	filePath := mntDir + "/perm.txt"

	if err := os.WriteFile(filePath, []byte("data"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	if err := os.Chmod(filePath, 0600); err != nil {
		t.Fatalf("Chmod failed: %v", err)
	}

	info, err := os.Stat(filePath)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if info.Mode().Perm() != 0600 {
		t.Errorf("expected mode %o, got %o", 0600, info.Mode().Perm())
	}
}

func TestMountAndMkdir(t *testing.T) {
	mntDir := mountTestFS(t)

	dirPath := mntDir + "/subdir"

	if err := os.Mkdir(dirPath, 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	info, err := os.Stat(dirPath)
	if err != nil {
		t.Fatalf("Stat failed: %v", err)
	}

	if !info.IsDir() {
		t.Errorf("expected %s to be a directory", dirPath)
	}
}

func TestMountAndUnlink(t *testing.T) {
	mntDir := mountTestFS(t)

	filePath := mntDir + "/todelete.txt"

	if err := os.WriteFile(filePath, []byte("bye"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	if err := os.Remove(filePath); err != nil {
		t.Fatalf("Remove failed: %v", err)
	}

	if _, err := os.Stat(filePath); !os.IsNotExist(err) {
		t.Errorf("expected file to be gone, got err=%v", err)
	}
}

func TestMountAndRename(t *testing.T) {
	mntDir := mountTestFS(t)

	oldPath := mntDir + "/old.txt"
	newPath := mntDir + "/new.txt"

	if err := os.WriteFile(oldPath, []byte("content"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	if err := os.Rename(oldPath, newPath); err != nil {
		t.Fatalf("Rename failed: %v", err)
	}

	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Errorf("expected old path to be gone, got err=%v", err)
	}

	data, err := os.ReadFile(newPath)
	if err != nil {
		t.Fatalf("ReadFile on new path failed: %v", err)
	}

	if string(data) != "content" {
		t.Errorf("expected content %q, got %q", "content", string(data))
	}
}
