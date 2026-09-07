package main

import (
	"context"
	"log"
	"syscall"

	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
)

// Open is called when opening the file for reading
func (f *FileNode) Open(ctx context.Context, flags uint32) (fs.FileHandle, uint32, syscall.Errno) {
	log.Println("📂 OPEN called!")
	return nil, 0, 0
}

// after open getattr should be called, else it considers filesie as 0
// Getattr returns file attributes (size, permissions, etc.)
func (f *FileNode) Getattr(ctx context.Context, fh fs.FileHandle, out *fuse.AttrOut) syscall.Errno {
	log.Println("📊 GETATTR called!")

	// Set file size to the length of our content
	out.Size = uint64(len(f.data.content))

	// Set file mode (permissions)
	out.Mode = fuse.S_IFREG | 0644 // Regular file, rw-r--r--

	out.Uid = uint32(syscall.Getuid())
	out.Gid = uint32(syscall.Getgid())

	log.Printf("   File size: %d bytes", out.Size)
	return 0
}

// Read is called when someone reads the file (cat)
// before this we need Open method
func (f *FileNode) Read(ctx context.Context, fh fs.FileHandle, dest []byte, off int64) (fuse.ReadResult, syscall.Errno) {

	log.Println("📖 READ CALLED!") // ← Add this

	// convert string to bytes
	data := []byte(f.data.content)

	// calculate and position
	end := off + int64(len(dest))
	if end > int64(len(data)) {
		end = int64(len(data))
	}

	return fuse.ReadResultData(data[off:end]), 0

}

/**
Write File

1) params
content: what to write (in bytes)
off: the position where to write, beacuse write is not for whole file, the os may split writing into several requests

2) Return ??
After storing the bytes, Write() responds how many bytes it accepted

macos tells i succesfully stored all 5 bytes

3) > -> means replace
>> -> means append

4) Rootnode and FileNode during Write()

RootNode Jobs: "Which file does the user mean?"

FileNode Jobs:"Where should these bytes go?"
"How should the content change?"

5) After writing the file size may change

GetAttr must be updated , ls -l must see new file size
**/

func (f *FileNode) Write(ctx context.Context, fh fs.FileHandle, data []byte, off int64) (uint32, syscall.Errno) {

	if off < 0 {
		return 0, syscall.EINVAL
	}

	current := []byte(f.data.content)
	end := int(off) + len(data)

	// handle case where write extends the file
	// create more empty spaces using bytes if needed
	// it only creates extra spaces
	if end > len(current) {
		current = append(current, make([]byte, end-len(current))...)
	}
	copy(current[int(off):end], data)
	f.data.content = string(current)

	return uint32(len(data)), 0

}

func (f *FileNode) Setattr(
	ctx context.Context,
	fh fs.FileHandle,
	in *fuse.SetAttrIn,
	out *fuse.AttrOut,
) syscall.Errno {

	size, ok := in.GetSize()
	if !ok {
		return syscall.ENOTSUP
	}

	current := []byte(f.data.content)

	if size < uint64(len(current)) {
		current = current[:size]
	} else if size > uint64(len(current)) {
		current = append(current,
			make([]byte, int(size)-len(current))...)
	}

	f.data.content = string(current)

	// Return the updated size and other attributes
	return f.Getattr(ctx, fh, out)
}
