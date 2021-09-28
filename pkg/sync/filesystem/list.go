package filesystem

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"

	"github.com/razzo-lunare/s3/pkg/sync/betav1"
)

// List all local files and passes them to the next step
func (f *FileSystem) List() (<-chan *betav1.FileInfo, error) {
	localFiles := make(chan *betav1.FileInfo, 10001)

	go listFiles(f.SyncDir, localFiles)

	return localFiles, nil
}

// TODO write this list function similar to the others so it's fast by using a pool of threads
func listFiles(SyncDir string, localFiles chan *betav1.FileInfo) {
	walk := func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return fmt.Errorf("error in walk: %s", e)
		}
		if !d.IsDir() {
			fileOnDiskMd5, err := hashFileMd5(path)
			if err != nil {
				klog.Errorf("Error calculating m5d of file on disk while listing: ", err)
				return err
			}

			localFiles <- &betav1.FileInfo{
				Name: strings.TrimPrefix(path, SyncDir),
				MD5:  fileOnDiskMd5,
			}
		}
		return nil
	}

	err := filepath.WalkDir(SyncDir, walk)
	if err != nil {
		klog.Errorf("walking directory error: ", err)
	}

	close(localFiles)
}
