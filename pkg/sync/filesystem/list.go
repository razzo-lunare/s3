package filesystem

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"k8s.io/klog/v2"

	"github.com/razzo-lunare/s3/pkg/asciiterm"
	"github.com/razzo-lunare/s3/pkg/sync/betav1"
)

// List all local files and passes them to the next step
func (f *FileSystem) List() (<-chan *betav1.FileInfo, error) {
	localFiles := make(chan *betav1.FileInfo, 10001)

	go listFiles(f.SyncDir, localFiles)

	return localFiles, nil
}

func listFiles(SyncDir string, localFiles chan *betav1.FileInfo) {
	klog.V(1).Info("listing files on the filesystem")

	walk := func(path string, d fs.DirEntry, e error) error {
		if e != nil {
			return fmt.Errorf("error in walk: %s", e)
		}
		if !d.IsDir() {
			fileOnDiskMd5, err := hashFileMd5(path)
			if err != nil {
				klog.Error("Error calculating m5d of file on disk while listing. ", path, err)
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
		klog.Errorf("walking directory error: %s", err)
	} else {
		asciiterm.PrintfInfo("Finished Listing Files\n")
	}

	close(localFiles)
}
