package filesystem

import (
	"io/fs"
	"path/filepath"

	"github.com/razzo-lunare/s3/pkg/sync/betav1"
)

func (f *FileSystem) List() (<-chan *betav1.FileInfo, error) {
	// TODO implement this
	panic("not implemented")
	err := filepath.WalkDir(f.SyncDir, walk)
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func walk(s string, d fs.DirEntry, e error) error {
	if e != nil {
		return e
	}
	if !d.IsDir() {
		println(s)
	}
	return nil
}
