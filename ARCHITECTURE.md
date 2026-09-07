# Architecture

I wrote this mainly to keep track of *why* things are structured the way they are, since a lot of it wasn't obvious to me until I hit the actual bugs.

## Where does a command like `cat file.txt` even end up in my code?

It doesn't come to `gofs` directly. There's a whole chain before it:

```
shell command (ls, cat, touch, ...)
        |
        v
   OS syscall (open, read, mkdir, unlink, ...)
        |
        v
   Linux/macOS kernel VFS layer
        |
        v
   FUSE kernel module
        |
        v
   gofs (this program, user-space)
```

So the kernel is the one deciding which of my Go functions to call, based on the syscall and whatever it already has cached. My job is just to answer those specific calls correctly — I never see a raw syscall myself.

## Two kinds of nodes, and why

I split things into `RootNode` and `FileNode` because they answer two different questions:

- `RootNode` answers "what exists in this folder?" — it holds the map of file names and subfolder names. Anything that adds/removes/lists names goes here: `Lookup`, `Readdir`, `Create`, `Mkdir`, `Unlink`, `Rmdir`, and `Getattr` for the folder itself.
- `FileNode` answers "what's inside this one file?" — it only knows about one file's bytes (`FileData`). `Read`, `Write`, `Getattr`/`Setattr` for that file live here.

The way I like to think about it: `RootNode` is a receptionist who keeps a list of drawers and can add/remove drawers. `FileNode` is one drawer that's already open in your hand — it doesn't get to decide whether it exists, it just holds content.

A subdirectory is literally just another `RootNode`. That's why nested folders "just worked" once `Mkdir` was in place — I didn't have to write anything extra for nesting.

**Rule I use when adding a new method:** if it changes what names exist in a directory, it goes in `root.go`. If it changes the content/attributes of a file that's already been found, it goes in `file.go`.

## What's actually stored

```
RootNode
├── files   map[string]*FileData     // file name -> shared, mutable content
├── subdirs map[string]*RootNode     // dir name  -> nested RootNode
└── mu      sync.Mutex               // guards both maps above
```

`FileData` is a pointer on purpose — I hit a bug early on where every `Lookup` created a fresh copy of the content, so writes seemed to vanish. Sharing the pointer across lookups fixed that; now every `FileNode` for the same file is looking at the same underlying data.

`mu` exists because `Create`/`Mkdir`/`Unlink`/`Rmdir` all mutate these maps, while `Lookup`/`Readdir` read them — without a lock, that's a data race. Every method that touches `files` or `subdirs` locks first.

## The `Getattr` bug that took a while to figure out

For a while, creating a file (`touch`/`echo >`) kept failing with "permission denied", even though `Create` itself looked correct. Turned out the kernel checks write access on the *parent folder* before it ever calls `Create` — and since `RootNode` had no `Getattr`, the folder defaulted to being "owned by root" with no real permissions. So the access check failed before my code even ran. Adding `Getattr` to `RootNode` that reports the actual current user (`syscall.Getuid()`/`Getgid()`) as the owner, with mode `0755`, fixed it.

## How delete works

`Unlink` and `Rmdir` just remove an entry from the map and return success — go-fuse takes care of removing the node from the kernel's tree after that. `Rmdir` also checks the folder is empty first and returns `ENOTEMPTY` if not, same as real `rmdir`. I'm not implementing recursive delete myself — `rm -rf` is the shell doing many individual `Unlink`/`Rmdir` calls, not something the filesystem needs to handle.

## What's not done yet

- No disk persistence — everything disappears when the process stops.
- No `Rename`, no symlinks, no real timestamps (they're all zero right now).
- File content is one big string per file in memory, so this wouldn't hold up for large files.
