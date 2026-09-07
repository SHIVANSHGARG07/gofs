package main

import (
	"github.com/hanwen/go-fuse/v2/fs"
	"sync"
)

// Define the root node

/*
*

1) Why fs.Inode ?? embeddings
Embedding means RootNode automatically gets ALL the methods that  fs.Inode  has
It's like saying "RootNode IS an Inode, plus anything extra I want to add"
You get methods like  NewInode() ,  Parent() ,  Children()  for FREE without writing them
This is how go-fuse knows your RootNode is a valid filesystem node
*
*/

type FileData struct {
	content string
}
type RootNode struct {
	fs.Inode
	files   map[string]*FileData
	subdirs map[string]*RootNode
	mu      sync.Mutex
}

type FileNode struct {
	fs.Inode
	data *FileData
}
