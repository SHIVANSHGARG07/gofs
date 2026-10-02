package main

import (
	"encoding/json"
	"os"
	"time"
)

type SerializableFile struct {
	Content string
	Mtime   time.Time
	Ctime   time.Time
	Btime   time.Time
	Mode    uint32
	Nlink   uint32
	Ino     uint64
}

type SerializableDir struct {
	Files    map[string]SerializableFile
	Subdirs  map[string]SerializableDir
	Symlinks map[string]SerializableSymlink
	Mtime    time.Time
	Ctime    time.Time
	Btime    time.Time
	Mode     uint32
	Ino      uint64
}

type SerializableSymlink struct {
	Target string
	Mtime  time.Time
	Ctime  time.Time
	Btime  time.Time
	Ino    uint64
}

const dataFile = "gofs_data.json"

// to serializable
/**

Apply conversion to each files and subdirs of RootNode Dir
used during save data

**/
func toSerializable(r *RootNode) SerializableDir {

	dir := SerializableDir{
		Files:    map[string]SerializableFile{},
		Subdirs:  map[string]SerializableDir{},
		Symlinks: map[string]SerializableSymlink{},
		Mtime:    r.mtime,
		Ctime:    r.ctime,
		Btime:    r.btime,
		Mode:     r.mode,
		Ino:      r.ino,
	}

	for name, data := range r.files {
		dir.Files[name] = SerializableFile{
			Content: data.content,
			Mtime:   data.mtime,
			Ctime:   data.ctime,
			Btime:   data.btime,
			Mode:    data.mode,
			Nlink:   data.nlink,
			Ino:     data.ino,
		}
	}

	for name, sym := range r.symlinks {
		dir.Symlinks[name] = SerializableSymlink{
			Target: sym.target,
			Mtime:  sym.mtime,
			Ctime:  sym.ctime,
			Btime:  sym.btime,
			Ino:    sym.ino,
		}
	}

	// recursive
	for name, subDir := range r.subdirs {
		dir.Subdirs[name] = toSerializable(subDir)
	}

	return dir

}

/**

used when we have to load data

**/

func fromSerializable(s SerializableDir) *RootNode {

	r := &RootNode{
		files:    map[string]*FileData{},
		subdirs:  map[string]*RootNode{},
		symlinks: map[string]*SymLink{},
		mtime:    s.Mtime,
		ctime:    s.Ctime,
		btime:    s.Btime,
		mode:     s.Mode,
		ino:      s.Ino,
	}

	for name, sf := range s.Files {
		r.files[name] = &FileData{
			content: sf.Content,
			mtime:   sf.Mtime,
			ctime:   sf.Ctime,
			btime:   sf.Btime,
			mode:    sf.Mode,
			nlink:   sf.Nlink,
			ino:     sf.Ino,
		}
	}

	for name, sd := range s.Subdirs {
		r.subdirs[name] = fromSerializable(sd)
	}

	for name, ss := range s.Symlinks {
		r.symlinks[name] = &SymLink{
			target: ss.Target,
			mtime:  ss.Mtime,
			ctime:  ss.Ctime,
			btime:  ss.Btime,
			ino:    ss.Ino,
		}
	}

	return r
}

func Save(r *RootNode) error {

	dir := toSerializable(r)

	data, err := json.MarshalIndent(dir, "", " ")
	if err != nil {
		return err
	}

	return os.WriteFile(dataFile, data, 0644)

}

func Load() (*RootNode, error) {
	data, err := os.ReadFile(dataFile)

	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var dir SerializableDir

	if err := json.Unmarshal(data, &dir); err != nil {
		return nil, err
	}

	root := fromSerializable(dir)

	// resume the ino counter so restarts don't reuse old numbers
	var maxIno uint64 = 1
	var scan func(r *RootNode)
	scan = func(r *RootNode) {
		if r.ino > maxIno {
			maxIno = r.ino
		}
		for _, data := range r.files {
			if data.ino > maxIno {
				maxIno = data.ino
			}
		}
		for _, sym := range r.symlinks {
			if sym.ino > maxIno {
				maxIno = sym.ino
			}
		}
		for _, sub := range r.subdirs {
			scan(sub)
		}
	}
	scan(root)
	nextIno = maxIno

	return root, nil
}
