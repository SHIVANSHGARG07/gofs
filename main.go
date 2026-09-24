/**

Components
1) import section
2) Main Function


**/

package main

// import section

import (
	"github.com/hanwen/go-fuse/v2/fs"
	"github.com/hanwen/go-fuse/v2/fuse"
	"log"
	"time"
)

var globalRoot *RootNode

// Main Function

/*
*
1) the file system is mounted by calling mount on the root of the tree
2) Pass low Level fuse operations, instead defualts will be passed
3) Waits until umount or ctrl+c

*
*/
func main() {

	root, err := Load()

	if err != nil {
		log.Fatal(err)
	}

	now := time.Now()

	if root == nil {
		root = &RootNode{
			files:    map[string]*FileData{},
			subdirs:  map[string]*RootNode{},
			symlinks: map[string]*SymLink{},
			btime:    now,
			mtime:    now,
			ctime:    now,
			mode:     0755,
		}
	}

	globalRoot = root

	server, err := fs.Mount("/Volumes/go-fs", root, &fs.Options{
		MountOptions: fuse.MountOptions{
			Name:  "gofs",
			Debug: false,
		},
	})

	if err != nil {
		log.Fatal(err)
	}

	log.Println(" Mounted at /Volumes/go-fs")
	server.Wait()

}
