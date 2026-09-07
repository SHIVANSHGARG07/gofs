package main

import (
	"context"
	"log"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

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

	newData := &FileData{content: ""}

	r.files[name] = newData

	stable := fs.StableAttr{
		Mode: fuse.S_IFREG,
	}

	fileNode := &FileNode{data: newData}

	node := r.NewPersistentInode(ctx, fileNode, stable)

	out.Attr.Mode = fuse.S_IFREG | 0644
	out.Attr.Uid = uint32(syscall.Getuid())
	out.Attr.Gid = uint32(syscall.Getgid())

	return node, nil, 0, 0
}

func (r *RootNode) Getattr(ctx context.Context, f fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	out.Mode = fuse.S_IFDIR | 0755
	out.Uid = uint32(syscall.Getuid())
	out.Gid = uint32(syscall.Getgid())
	return 0
}

/*
*
1) Define locks
2) make a branch new Root node, new subfolder itself
3) put name and assign that to new Root Node
4) mark it as a regular dir
5) register as a real inode
6) send outputs : form out

*
*/
func (r *RootNode) Mkdir(ctx context.Context, name string, mode uint32, out *fuse.EntryOut) (*fs.Inode, syscall.Errno) {

	r.mu.Lock()
	defer r.mu.Unlock()

	newDir := &RootNode{
		files:   map[string]*FileData{},
		subdirs: map[string]*RootNode{},
	}

	r.subdirs[name] = newDir

	stable := fs.StableAttr{
		Mode: fuse.S_IFDIR,
	}

	node := r.NewPersistentInode(ctx, newDir, stable)

	out.Attr.Mode = fuse.S_IFDIR | 0755
	out.Attr.Uid = uint32(syscall.Getuid())
	out.Attr.Gid = uint32(syscall.Getgid())

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
	return 0

}
