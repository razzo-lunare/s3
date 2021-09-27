package filesystem

import "github.com/razzo-lunare/s3/pkg/sync/betav1"

// Get gathers the content for incoming FileInfo and passes them
// to the next step
func (s *FileSystem) Get(<-chan *betav1.FileInfo) (<-chan *betav1.FileInfo, error) {
	// TODO implement this
	panic("not implemented")
}
