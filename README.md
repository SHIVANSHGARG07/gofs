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
- Moving/renaming files and directories (`mv`), same-dir and cross-dir
- Real timestamps (`mtime`/`ctime`) that update correctly across creates, writes, truncates, deletes, and renames
- Correct file/folder ownership so `ls -l` shows the actual user, not root
- Safe concurrent access to the in-memory file/folder maps
- Symlinks (`ln -s`, `readlink`), with correct ownership/timestamps on `ls -l`
- Persistence to disk (`gofs_data.json`) — state survives a process restart
- Real timestamps including birthtime (`mtime`/`ctime`/`btime`), correct across creates, writes, truncates, deletes, renames, and chmod
- File permissions (`chmod`) — mode is stored per file/directory and persists across restarts

## What's next

- Permission enforcement (currently `chmod` only updates what `ls -l` shows; read/write isn't actually blocked based on mode)
- Hard links, extended attributes, and other ideas — see `PHASES.md`

See `PHASES.md` for how this is being built out phase by phase.
