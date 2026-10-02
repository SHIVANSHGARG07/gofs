# Phases

Tracking this project in phases instead of one big feature list, so it's easier to see what was actually done at each point.

## Phase 1 — Basic CRUD (done)

Goal: get a working in-memory filesystem where files and directories can be created, read, written, listed, and deleted.

- [x] Mount a folder as a FUSE filesystem
- [x] Read file content (`cat`)
- [x] Write file content, including append (`echo >`, `>>`)
- [x] Fix content-persistence bug (shared `FileData` pointer instead of copies)
- [x] Create new files (`touch`, `echo > newfile`)
- [x] Correct ownership/permissions on the mount root (`Getattr` on `RootNode`), so non-root users can write
- [x] Create directories, including nested ones (`Mkdir`)
- [x] Concurrency-safe access to the in-memory maps (`sync.Mutex`)
- [x] Delete files (`Unlink`)
- [x] Delete empty directories (`Rmdir`), refusing non-empty ones with `ENOTEMPTY`

## Phase 2 — Rename & Timestamps (done)

- [x] `Rename` — move/rename files and directories (same-dir and cross-dir)
- [x] Real timestamps (mtime/ctime) across `Create`, `Mkdir`, `Write`, `Setattr` (size), `Unlink`, `Rmdir`, `Rename`
- [x] `Setattr` time-only requests (e.g. `touch` on an already-existing file) — now handles `Size`, `Mtime`, `Atime` independently instead of rejecting when `Size` is absent
- [x] `Fsync` implemented as a no-op (data already persisted on `Write`/`Setattr` via `Save(globalRoot)`) — fixes vim `E667: Fsync failed` on save
- [x] Birthtime (crtime) support — see Phase 4

## Phase 3 — Persistence & Symlinks (done)

- [x] Persist state to disk (`gofs_data.json`), so data survives a process restart
- [x] Symlinks (`Symlink` / `Readlink`), including correct owner/timestamps via `Getattr`

## Phase 4 — Timestamps, Permissions & Quotas (done)

- [x] `touch` on an already-existing file (`Setattr` time-only requests) — done
- [x] Birthtime support (`btime`) — set on creation (`Create`/`Mkdir`/`Symlink`/fresh root), exposed via `Getattr` (`Crtime_`/`Crtimensec_`), persisted in `gofs_data.json`, and preserved (not touched) on `Write`/`Setattr`/`Rename`/existing-file `Create`. Known caveat: editors like `vim` (with `backupcopy=auto`) may swap in a new inode on save via temp-file+rename, which naturally resets birthtime — this is expected editor/OS behavior, verified even on real filesystems (APFS, Lustre), not a gofs bug.
- [x] File permissions (`chmod`) — `mode` stored per file/dir (`FileData`/`RootNode`), set on creation (`0644`/`0755` defaults), updated via `Setattr`'s `GetMode()`, exposed via `Getattr` (both `FileNode` and `RootNode`, plus both branches of `Create`), and persisted in `gofs_data.json`.
- [x] Disk usage / quota simulation (`df`, `du`) — via `Statfs`
- [x] Automated Go tests (`_test.go` files) — `mount_test.go` covers mount-based integration tests (CRUD, rename, timestamps, symlinks, permissions); CI (`.github/workflows/ci.yml`) runs `gofmt`, `go vet`, `go build`, `go test`

## Phase 5 — Links & Stable Inodes (done)

- [x] Permission enforcement — `FileNode.Open` rejects read/write based on the stored mode's owner bits, returning `syscall.EACCES` when disallowed. See `ARCHITECTURE.md` for details and the `ls -l` caching caveat.
- [x] Hard links (same inode shared across multiple names, unlike symlinks which just point elsewhere). Known caveat: after a process restart, hard-linked names are reloaded as separate `FileData` objects (same `Ino`/`Nlink` values, but no longer sharing the same pointer), so `chmod`/content changes stop staying in sync — tracked as a GitHub issue.
- [x] Directory/symlink stable inode numbers — `RootNode`/`SymLink` now carry their own `ino` (assigned via the same thread-safe `newIno()` counter used for files), set on `Mkdir`/`Symlink` and reused (not regenerated) on `Lookup`; persisted in `gofs_data.json` and resumed on restart via the tree scan in `Load()`. Verified: repeated `stat` across remounts/restarts no longer swaps inode numbers between directories.

## Phase 6 — Upcoming

- [ ] Concurrent access edge cases (e.g. a file being deleted while another handle is reading it)
- [ ] Extended attributes (xattrs) — custom metadata attached to files
- [ ] Better handling of odd edge cases (symlink loops like `a -> b -> a`, very long paths, etc.)

## More upcoming (ideas, not yet scheduled to a phase)

- [ ] Store file content as bytes/chunks instead of one big string, so large files don't need a full copy on every write — deferred until DB-backed storage is introduced (design will change anyway)
- [ ] Per-user directory permissions/ownership (currently single-user only — everything is owned by whoever runs the process; permission bits are stored/enforced per file/dir, but there's no concept of multiple distinct users/uids yet)
