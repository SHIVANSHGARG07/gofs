# Phases

Tracking this project in phases instead of one big feature list, so it's easier to see what was actually done at each point.

## Phase 1 — Basic CRUD (current)

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

## Phase 4 — Planned

- [x] `touch` on an already-existing file (`Setattr` time-only requests) — done
- [x] Birthtime support (`btime`) — set on creation (`Create`/`Mkdir`/`Symlink`/fresh root), exposed via `Getattr` (`Crtime_`/`Crtimensec_`), persisted in `gofs_data.json`, and preserved (not touched) on `Write`/`Setattr`/`Rename`/existing-file `Create`. Known caveat: editors like `vim` (with `backupcopy=auto`) may swap in a new inode on save via temp-file+rename, which naturally resets birthtime — this is expected editor/OS behavior, verified even on real filesystems (APFS, Lustre), not a gofs bug.
- [ ] Store file content as bytes/chunks instead of one big string, so large files don't need a full copy on every write — deferred until DB-backed storage is introduced (design will change anyway)
- [x] File permissions (`chmod`) — `mode` stored per file/dir (`FileData`/`RootNode`), set on creation (`0644`/`0755` defaults), updated via `Setattr`'s `GetMode()`, exposed via `Getattr` (both `FileNode` and `RootNode`, plus both branches of `Create`), and persisted in `gofs_data.json`. Note: this is storage/advertisement only (`ls -l` reflects the real mode) — actual enforcement (rejecting `Write`/`Read` based on mode/uid) is not implemented.
- [x] Disk usage / quota simulation (`df`, `du`) — via `Statfs`

## More upcoming (ideas, not yet scheduled to a phase)

- [ ] Hard links (same inode shared across multiple names, unlike symlinks which just point elsewhere)
- [ ] Per-user directory permissions/ownership (currently everything is owned by whoever runs the process)
- [ ] Concurrent access edge cases (e.g. a file being deleted while another handle is reading it)
- [ ] Extended attributes (xattrs) — custom metadata attached to files
- [ ] Better handling of odd edge cases (symlink loops like `a -> b -> a`, very long paths, etc.)
- [ ] Automated Go tests (`_test.go` files) — everything so far has been verified manually via `ls`/`cat`/etc.
