package betav1

type SyncObject interface {
	// Returns a channel of files
	List() (<-chan *FileInfo, error)

	// Verify checks to see if the FileInfo exists
	Verify(<-chan *FileInfo) (<-chan *FileInfo, error)

	// Create accepts a channel of files to create
	Create(<-chan *FileInfo) (<-chan *FileInfo, error)
}

type FileInfo struct {
	Name string
	MD5  string
}
