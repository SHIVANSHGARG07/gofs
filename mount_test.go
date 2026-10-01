package main

import (
	"os"
	"syscall"
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

func TestMountAndRmdir(t *testing.T) {
	mntDir := mountTestFS(t)

	dirPath := mntDir + "/emptydir"

	if err := os.Mkdir(dirPath, 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	if err := os.Remove(dirPath); err != nil {
		t.Fatalf("Rmdir (via Remove) failed: %v", err)
	}

	if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
		t.Errorf("expected dir to be gone, got err=%v", err)
	}
}

func TestMountAndRmdirNonEmpty(t *testing.T) {
	mntDir := mountTestFS(t)

	dirPath := mntDir + "/nonempty"

	if err := os.Mkdir(dirPath, 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	if err := os.WriteFile(dirPath+"/inside.txt", []byte("x"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	if err := os.Remove(dirPath); err == nil {
		t.Errorf("expected Remove on non-empty dir to fail, but it succeeded")
	}
}

func TestMountAndUnlinkNonExistent(t *testing.T) {
	mntDir := mountTestFS(t)

	filePath := mntDir + "/ghost.txt"

	if err := os.Remove(filePath); !os.IsNotExist(err) {
		t.Errorf("expected ENOENT-like error removing non-existent file, got %v", err)
	}
}

func TestMountAndTruncate(t *testing.T) {
	mntDir := mountTestFS(t)

	filePath := mntDir + "/trunc.txt"

	if err := os.WriteFile(filePath, []byte("hello world"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	if err := os.Truncate(filePath, 5); err != nil {
		t.Fatalf("Truncate failed: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	if string(data) != "hello" {
		t.Errorf("expected content %q, got %q", "hello", string(data))
	}
}

func TestMountAndSymlink(t *testing.T) {
	mntDir := mountTestFS(t)

	targetPath := mntDir + "/target.txt"
	linkPath := mntDir + "/link.txt"

	if err := os.WriteFile(targetPath, []byte("real data"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	if err := os.Symlink(targetPath, linkPath); err != nil {
		t.Fatalf("Symlink failed: %v", err)
	}

	resolved, err := os.Readlink(linkPath)
	if err != nil {
		t.Fatalf("Readlink failed: %v", err)
	}

	if resolved != targetPath {
		t.Errorf("expected readlink %q, got %q", targetPath, resolved)
	}

	data, err := os.ReadFile(linkPath)
	if err != nil {
		t.Fatalf("ReadFile via symlink failed: %v", err)
	}

	if string(data) != "real data" {
		t.Errorf("expected content %q, got %q", "real data", string(data))
	}
}

func TestMountAndNestedDir(t *testing.T) {
	mntDir := mountTestFS(t)

	subDir := mntDir + "/nested"
	filePath := subDir + "/inner.txt"

	if err := os.Mkdir(subDir, 0755); err != nil {
		t.Fatalf("Mkdir failed: %v", err)
	}

	if err := os.WriteFile(filePath, []byte("nested content"), 0644); err != nil {
		t.Fatalf("WriteFile in subdir failed: %v", err)
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("ReadFile in subdir failed: %v", err)
	}

	if string(data) != "nested content" {
		t.Errorf("expected content %q, got %q", "nested content", string(data))
	}
}

func TestMountAndPermissionDenied(t *testing.T) {
	mntDir := mountTestFS(t)

	filePath := mntDir + "/readonly.txt"

	if err := os.WriteFile(filePath, []byte("secret"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Create() currently always stores new files as 0644 regardless of the
	// requested mode, so set the restrictive permission explicitly via
	// Chmod (same pattern as TestMountAndChmod).
	if err := os.Chmod(filePath, 0400); err != nil {
		t.Fatalf("Chmod failed: %v", err)
	}

	// write-only open should fail: file has no write bit (0400)
	f, err := os.OpenFile(filePath, os.O_WRONLY, 0)
	if err == nil {
		f.Close()
		t.Fatalf("expected OpenFile(O_WRONLY) to fail on a read-only file, but it succeeded")
	}
	if !os.IsPermission(err) {
		t.Errorf("expected permission error, got %v", err)
	}

	// read-only open should still succeed
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("expected read to succeed on a read-only file, got %v", err)
	}
	if string(data) != "secret" {
		t.Errorf("expected content %q, got %q", "secret", string(data))
	}

	// make it write-only (no read bit) and verify read is denied
	if err := os.Chmod(filePath, 0200); err != nil {
		t.Fatalf("Chmod failed: %v", err)
	}

	if _, err := os.ReadFile(filePath); !os.IsPermission(err) {
		t.Errorf("expected permission error reading write-only file, got %v", err)
	}
}

func TestMountAndHardLink(t *testing.T) {
	mntDir := mountTestFS(t)

	srcPath := mntDir + "/src.txt"
	linkPath := mntDir + "/link.txt"

	if err := os.WriteFile(srcPath, []byte("shared"), 0644); err != nil {
		t.Fatalf("WriteFile failed: %v", err)
	}

	// Force a fresh Lookup of srcPath before linking. On some FUSE
	// implementations (observed on Linux CI runners, not reproduced on
	// macOS/macFUSE), the kernel dentry cache isn't guaranteed warm right
	// after WriteFile, and linkat() can return ENOENT without this.
	if _, err := os.Stat(srcPath); err != nil {
		t.Fatalf("Stat on src before Link failed: %v", err)
	}

	if err := os.Link(srcPath, linkPath); err != nil {
		t.Fatalf("Link failed: %v", err)
	}

	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		t.Fatalf("Stat on src failed: %v", err)
	}
	linkInfo, err := os.Stat(linkPath)
	if err != nil {
		t.Fatalf("Stat on link failed: %v", err)
	}

	srcStat := srcInfo.Sys().(*syscall.Stat_t)
	linkStat := linkInfo.Sys().(*syscall.Stat_t)

	if srcStat.Ino != linkStat.Ino {
		t.Errorf("expected same inode for hard-linked files, got %d and %d", srcStat.Ino, linkStat.Ino)
	}

	if linkInfo.Sys().(*syscall.Stat_t).Nlink != 2 {
		t.Errorf("expected Nlink=2 after hard link, got %d", linkStat.Nlink)
	}

	// content/mode changes should be visible through either name, since
	// they share the same underlying FileData
	if err := os.Chmod(linkPath, 0600); err != nil {
		t.Fatalf("Chmod via link failed: %v", err)
	}

	srcInfo2, err := os.Stat(srcPath)
	if err != nil {
		t.Fatalf("Stat on src after chmod failed: %v", err)
	}
	if srcInfo2.Mode().Perm() != 0600 {
		t.Errorf("expected chmod via link to be visible on src, got mode %o", srcInfo2.Mode().Perm())
	}

	// removing one name should leave the other intact with Nlink decremented
	if err := os.Remove(linkPath); err != nil {
		t.Fatalf("Remove link failed: %v", err)
	}

	srcInfo3, err := os.Stat(srcPath)
	if err != nil {
		t.Fatalf("expected src to survive after removing link, got %v", err)
	}
	if srcInfo3.Sys().(*syscall.Stat_t).Nlink != 1 {
		t.Errorf("expected Nlink=1 after removing one hard link, got %d", srcInfo3.Sys().(*syscall.Stat_t).Nlink)
	}
}

func TestMountDirInodeStableAcrossLookups(t *testing.T) {
	mntDir := mountTestFS(t)

	dirA := mntDir + "/dirA"
	dirB := mntDir + "/dirB"

	if err := os.Mkdir(dirA, 0755); err != nil {
		t.Fatalf("Mkdir dirA failed: %v", err)
	}
	if err := os.Mkdir(dirB, 0755); err != nil {
		t.Fatalf("Mkdir dirB failed: %v", err)
	}

	infoA1, err := os.Stat(dirA)
	if err != nil {
		t.Fatalf("Stat dirA failed: %v", err)
	}
	infoB1, err := os.Stat(dirB)
	if err != nil {
		t.Fatalf("Stat dirB failed: %v", err)
	}

	inoA1 := infoA1.Sys().(*syscall.Stat_t).Ino
	inoB1 := infoB1.Sys().(*syscall.Stat_t).Ino

	if inoA1 == inoB1 {
		t.Fatalf("expected dirA and dirB to have distinct inodes, both got %d", inoA1)
	}

	// Look them up again, in reverse order. Stable inodes should not
	// depend on lookup order (this was the bug: Ino:0 let go-fuse assign
	// a new number per Lookup call, causing values to swap).
	infoB2, err := os.Stat(dirB)
	if err != nil {
		t.Fatalf("Stat dirB (2nd) failed: %v", err)
	}
	infoA2, err := os.Stat(dirA)
	if err != nil {
		t.Fatalf("Stat dirA (2nd) failed: %v", err)
	}

	if infoA2.Sys().(*syscall.Stat_t).Ino != inoA1 {
		t.Errorf("dirA inode changed across lookups: %d -> %d", inoA1, infoA2.Sys().(*syscall.Stat_t).Ino)
	}
	if infoB2.Sys().(*syscall.Stat_t).Ino != inoB1 {
		t.Errorf("dirB inode changed across lookups: %d -> %d", inoB1, infoB2.Sys().(*syscall.Stat_t).Ino)
	}
}
