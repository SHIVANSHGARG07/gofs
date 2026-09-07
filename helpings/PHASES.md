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

## Phase 2 — Planned

- [ ] `Rename` — move/rename files and directories
- [ ] Real timestamps (currently zero-valued)

## Phase 3 — Planned

- [ ] Persist state to disk, so data survives a process restart
- [ ] Symlinks (`Symlink` / `Readlink`)
