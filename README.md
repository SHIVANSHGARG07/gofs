# gofs

I'm building a small in-memory filesystem in Go using [FUSE](https://github.com/hanwen/go-fuse), mainly to actually understand how a filesystem talks to the OS instead of just reading about it. Once mounted, it behaves like a real folder — `ls`, `cat`, `touch`, `mkdir`, `rm`, `rmdir` all work against it — except everything lives in RAM, not on disk.

## Prerequisites

- Go
- [macFUSE](https://osxfuse.github.io/) (needed on macOS to mount FUSE filesystems at all)

## Running it

```bash
go run .
```

This mounts the filesystem at `/Volumes/go-fs`. Open another terminal and just use it like a normal folder:

```bash
ls /Volumes/go-fs
echo "hello" > /Volumes/go-fs/test.txt
cat /Volumes/go-fs/test.txt
mkdir /Volumes/go-fs/newfolder
rm /Volumes/go-fs/test.txt
```

`Ctrl+C` stops it, or unmount manually if it's stuck:

```bash
umount /Volumes/go-fs
```

## How it's put together

See [ARCHITECTURE.md](./ARCHITECTURE.md) for the actual design notes and the bugs I ran into along the way (the short version: a `RootNode` tracks what files/folders exist, a `FileNode` holds one file's actual content, and subfolders are just more `RootNode`s).

## What works right now

- Reading and writing file content (`cat`, `echo >`, `>>`)
- Creating files (`touch`, `echo > newfile`)
- Listing a directory (`ls`)
- Creating directories, including nested ones (`mkdir`)
- Deleting files (`rm`)
- Deleting empty directories (`rmdir`) — refuses non-empty ones like real `rmdir` does
- Correct file/folder ownership so `ls -l` shows the actual user, not root
- Safe concurrent access to the in-memory file/folder maps

## What's next

- `Rename` / `mv` support
- Actually persist to disk, so restarting the program doesn't wipe everything
- Real timestamps (right now they're all zero)
- Symlinks

See `PHASES.md` for how this is being built out phase by phase.
