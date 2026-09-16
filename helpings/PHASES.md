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
- [ ] Known limitation: `Setattr` time-only requests (e.g. `touch` on an already-existing file) aren't supported yet
- [ ] Known limitation: birthtime not implemented (macOS-specific field, currently zero)

## Phase 3 — Persistence & Symlinks (done)

- [x] Persist state to disk (`gofs_data.json`), so data survives a process restart
- [x] Symlinks (`Symlink` / `Readlink`), including correct owner/timestamps via `Getattr`

## Phase 4 — Planned

- [ ] `touch` on an already-existing file (`Setattr` time-only requests, currently `ENOTSUP`)
- [ ] Birthtime support (macOS-specific field, currently zero)
- [ ] Store file content as bytes/chunks instead of one big string, so large files don't need a full copy on every write
- [ ] File permissions properly enforced (`chmod`) — currently mode is hardcoded (`0644`/`0755`)
- [ ] Disk usage / quota simulation (`df`, `du`)

## More upcoming (ideas, not yet scheduled to a phase)

- [ ] Hard links (same inode shared across multiple names, unlike symlinks which just point elsewhere)
- [ ] Per-user directory permissions/ownership (currently everything is owned by whoever runs the process)
- [ ] Concurrent access edge cases (e.g. a file being deleted while another handle is reading it)
- [ ] Extended attributes (xattrs) — custom metadata attached to files
- [ ] Better handling of odd edge cases (symlink loops like `a -> b -> a`, very long paths, etc.)
- [ ] Automated Go tests (`_test.go` files) — everything so far has been verified manually via `ls`/`cat`/etc.
