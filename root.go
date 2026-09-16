package main

import (
	"context"
	"log"
	"syscall"
	"time"
	"unsafe"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// RENAME_NOREPLACE mirrors the Linux renameat2() flag value (0x1).
// golang.org/x/sys/unix only defines this constant on Linux, but the
// FUSE protocol sends the same numeric flag value on all platforms.
const RENAME_NOREPLACE = 0x1

/**
1) Creates an empty slice in memory taht will hold directories entry
2) As we introduces Create operation, we need to add locks to file reading also to avoid crashing or race conditions
**/

func (r *RootNode) Readdir(ctx context.Context) (fs.DirStream, syscall.Errno) {

	r.mu.Lock()
	defer r.mu.Unlock()

	entries := make([]fuse.DirEntry, 0, len(r.files))

	for name := range r.files {
		entries = append(entries, fuse.DirEntry{
			Name: name,
			Mode: fuse.S_IFREG,
		})
	}

	for name := range r.subdirs {
		entries = append(entries, fuse.DirEntry{
			Name: name,
			Mode: fuse.S_IFDIR,
		})
	}

	for name := range r.symlinks {
		entries = append(entries, fuse.DirEntry{
			Name: name,
			Mode: fuse.S_IFLNK,
		})
	}

	return fs.NewListDirStream(entries), 0

}

func (r *RootNode) Lookup(ctx context.Context, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {
	log.Printf("🔍 Lookup called for: %s", name)

	r.mu.Lock()
	defer r.mu.Unlock()

	data, file_exists_or_not := r.files[name]

	if file_exists_or_not {
		//create a node for file
		stable := fs.StableAttr{
			Mode: fuse.S_IFREG,
		}

		// create a filenode with content
		fileNode := &FileNode{data: data}

		// alternate to newInode
		node := r.NewPersistentInode(ctx, fileNode, stable)
		return node, 0
	}

	subDir, subExists := r.subdirs[name]

	if subExists {
		stable := fs.StableAttr{
			Mode: fuse.S_IFDIR,
		}

		node := r.NewPersistentInode(ctx, subDir, stable)
		return node, 0
	}

	symlink, symlinkExists := r.symlinks[name]

	if symlinkExists {
		stable := fs.StableAttr{
			Mode: fuse.S_IFLNK,
		}

		node := r.NewPersistentInode(ctx, symlink, stable)
		return node, 0
	}

	return nil, syscall.ENOENT

}

/**
Params:

out:
	File size
	Permissions for files
	Who is owner

1) take lock and remove when defer
2) Create a brand new fileData wuth empty content
3) Put the new Filedata in map
4) Label new file type: Regular file
5) Wrap data in FileNode
6) Register as a real node
7) Fill the form for out
8) return things

**/

func (r *RootNode) Create(ctx context.Context, name string, flag uint32, mode uint32, out *fuse.EntryOut) (*fs.Inode, fs.FileHandle, uint32, syscall.Errno) {

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	newData := &FileData{content: "", mtime: now, ctime: now}

	r.files[name] = newData
	r.mtime = now
	r.ctime = now

	stable := fs.StableAttr{
		Mode: fuse.S_IFREG,
	}

	fileNode := &FileNode{data: newData}

	node := r.NewPersistentInode(ctx, fileNode, stable)

	// Sent back to kernel
	out.Attr.Mode = fuse.S_IFREG | 0644
	out.Attr.Uid = uint32(syscall.Getuid())
	out.Attr.Gid = uint32(syscall.Getgid())

	out.Attr.SetTimes(&now, &now, &now)

	Save(globalRoot)

	return node, nil, 0, 0
}

func (r *RootNode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = fuse.S_IFDIR | 0755
	out.Uid = uint32(syscall.Getuid())
	out.Gid = uint32(syscall.Getgid())

	out.SetTimes(&r.mtime, &r.mtime, &r.ctime)
	return 0
}

/*
*
1) Define locks
2) make a branch new Root node, new subfolder itself
3) put name and assign that to new Root Node
4) mark it as a regular dir
5) register as a real yeah minode
6) send outputs : form out

*
*/
func (r *RootNode) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	newDir := &RootNode{
		files:    map[string]*FileData{},
		subdirs:  map[string]*RootNode{},
		symlinks: map[string]*SymLink{},
		mtime:    now,
		ctime:    now,
	}

	r.subdirs[name] = newDir
	r.mtime = now
	r.ctime = now

	stable := fs.StableAttr{
		Mode: fuse.S_IFDIR,
	}

	node := r.NewPersistentInode(ctx, newDir, stable)

	out.Attr.Mode = fuse.S_IFDIR | 0755
	out.Attr.Uid = uint32(syscall.Getuid())
	out.Attr.Gid = uint32(syscall.Getgid())
	out.Attr.SetTimes(&now, &now, &now)

	Save(globalRoot)

	return node, 0

}

func (r *RootNode) Unlink(ctx context.Context, name string) syscall.Errno {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.files[name]

	if !exists {
		return syscall.ENOENT
	}

	delete(r.files, name)

	r.mtime = time.Now()
	r.ctime = time.Now()

	Save(globalRoot)

	return 0

}

func (r *RootNode) Rmdir(ctx context.Context, name string) syscall.Errno {

	r.mu.Lock()
	defer r.mu.Unlock()

	subDir, exists := r.subdirs[name]
	if !exists {
		return syscall.ENOENT
	}

	if len(subDir.files) > 0 || len(subDir.subdirs) > 0 {
		return syscall.ENOTEMPTY
	}

	delete(r.subdirs, name)

	r.mtime = time.Now()
	r.ctime = time.Now()

	Save(globalRoot)

	return 0

}

func (r *RootNode) Rename(ctx context.Context, name string, newParent fs.InodeEmbedder, newName string, flags uint32) syscall.Errno {

	newDir := newParent.(*RootNode)

	if r == newDir {
		r.mu.Lock()
		defer r.mu.Unlock()
	} else {
		first, second := r, newDir

		if uintptr(unsafe.Pointer(r)) > uintptr(unsafe.Pointer(newDir)) {
			first, second = newDir, r
		}

		first.mu.Lock()
		defer first.mu.Unlock()
		second.mu.Lock()
		defer second.mu.Unlock()
	}

	if flags&RENAME_NOREPLACE != 0 {
		if _, exists := newDir.files[newName]; exists {
			return syscall.EEXIST
		}
		if _, exists := newDir.subdirs[newName]; exists {
			return syscall.EEXIST
		}
	}

	now := time.Now()

	if data, ok := r.files[name]; ok {
		// it is a file
		delete(r.files, name)
		data.ctime = now
		newDir.files[newName] = data
	} else if subDir, ok := r.subdirs[name]; ok {
		delete(r.subdirs, name)
		subDir.ctime = now
		newDir.subdirs[newName] = subDir
	} else {
		return syscall.ENOENT
	}

	r.mtime = now
	r.ctime = now
	newDir.mtime = now
	newDir.ctime = now

	ok := r.MvChild(name, &newDir.Inode, newName, true)
	if !ok {
		return syscall.EIO
	}

	Save(globalRoot)

	return 0

}

func (r *RootNode) Symlink(ctx context.Context, target, name string, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {

	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()

	newSymlink := &SymLink{target: target, mtime: now, ctime: now}

	r.symlinks[name] = newSymlink
	r.mtime = now
	r.ctime = now

	stable := fs.StableAttr{
		Mode: fuse.S_IFLNK,
	}

	node := r.NewPersistentInode(ctx, newSymlink, stable)

	out.Attr.Mode = fuse.S_IFLNK | 0777
	out.Attr.Uid = uint32(syscall.Getuid())
	out.Attr.Gid = uint32(syscall.Getgid())
	out.Attr.SetTimes(&now, &now, &now)

	Save(globalRoot)

	return node, 0
}

func (s *SymLink) Readlink(ctx context.Context) ([]byte, syscall.Errno) {
	return []byte(s.target), 0
}

func (s *SymLink) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Attr.Mode = fuse.S_IFLNK | 0777
	out.Attr.Uid = uint32(syscall.Getuid())
	out.Attr.Gid = uint32(syscall.Getgid())
	out.SetTimes(&s.mtime, &s.mtime, &s.ctime)
	return 0
}
