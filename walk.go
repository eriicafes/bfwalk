package bfwalk

import (
	"io/fs"
	"path"
)

// WalkDir walks the file tree rooted at root, calling fn for each file or
// directory in the tree, including root.
//
// The traversal is breadth-first, meaning all entries at a given depth are
// visited before descending into subdirectories.
//
// All errors that arise visiting files and directories are filtered by fn:
// see the [fs.WalkDirFunc] documentation for details.
//
// The files are walked in lexical order, which makes the output deterministic
// but requires WalkDir to read an entire directory into memory before proceeding
// to walk that directory.
//
// WalkDir does not follow symbolic links found in directories,
// but if root itself is a symbolic link, its target will be walked.
func WalkDir(fsys fs.FS, root string, fn fs.WalkDirFunc) error {
	info, err := fs.Stat(fsys, root)
	if err != nil {
		err = fn(root, nil, err)
	} else {
		d := fs.FileInfoToDirEntry(info)
		err = fn(root, d, nil)
		// Walk root if it is a directory and err is nil
		if err == nil && d.IsDir() {
			entry := namedEntry{root, d}
			err = walkDir(fsys, []namedEntry{entry}, fn)
		}
	}
	if err == fs.SkipDir || err == fs.SkipAll {
		return nil
	}
	return err
}

type namedEntry struct {
	name string
	d    fs.DirEntry
}

// walkDir recursively descends path breadth first, calling walkDirFn.
func walkDir(fsys fs.FS, queue []namedEntry, walkDirFn fs.WalkDirFunc) error {
	if len(queue) == 0 {
		return nil
	}
	name, d := queue[0].name, queue[0].d
	queue = queue[1:] // Pop first entry

	dirs, err := fs.ReadDir(fsys, name)
	if err != nil {
		// Second call, to report ReadDir error.
		err = walkDirFn(name, d, err)
		if err != nil {
			if err == fs.SkipDir && d.IsDir() {
				err = nil
			}
			return err
		}
	}

	var subqueue []namedEntry
	for _, d1 := range dirs {
		name1 := path.Join(name, d1.Name())
		err := walkDirFn(name1, d1, nil)
		if err != nil {
			if err == fs.SkipAll {
				return err
			}
			if err == fs.SkipDir {
				if d1.IsDir() {
					continue // Skip current directory
				} else {
					subqueue = nil
					break // Skip parent directory
				}
			}
		}
		if d1.IsDir() {
			subqueue = append(subqueue, namedEntry{name1, d1})
		}
	}
	queue = append(queue, subqueue...)

	return walkDir(fsys, queue, walkDirFn)
}
