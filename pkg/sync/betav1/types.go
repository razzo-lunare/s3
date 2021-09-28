package betav1

import "io"

// SyncObject is the interface for the goroutine pipeline to dowload/upload s3 objects
type SyncObject interface {
	// Returns a channel of files
	List() (<-chan *FileInfo, error)

	// Verify checks to see if the FileInfo needs to be downloaded by;
	// - checking if the resource exists in the destination
	// - the md5 matches in the source and destination
	Verify(<-chan *FileInfo) (<-chan *FileInfo, error)

	// Get downloads the contents of the object or file
	Get(<-chan *FileInfo) (<-chan *FileInfo, error)

	// Create accepts a channel of files to create and writes them to the destination
	Create(<-chan *FileInfo) (<-chan *FileInfo, error)
}

// FileInfo is the data that is passed between pipeline steps
type FileInfo struct {
	Name string
	MD5  string

	// Content is defined in the 'Get()' func
	Content     io.Reader
	ContentSize int64
}
