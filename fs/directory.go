package fs

type Directory interface {
	Create(nodeName string) (File, error)
	Hardlink(targetPath, linkName string) error
	Mkdir(dirName string) (FileHandle, error)
	Remove(nodeName string)
	Rename(newName string, newDir Directory) error
	Symlink(targetPath, linkName string) error
}
