package plugin

import (
	"os"
)

const T_DIR = 1
const T_FILE = 2
const T_SYMLINK = 3

type VFSFile struct {
	Type     int         `json:"type"`
	Name     string      `json:"name"`
	Files    []VFSFile   `json:"files"`
	Contents string      `json:"contents"`
	Mode     os.FileMode `json:"mode"`
}

func (f *VFSFile) Find(name string, t int) (string, *VFSFile) {
	if name == f.Name && f.Type == t {
		return "/" + f.Name, f
	}

	if f.Type == T_DIR {
		for _, file := range f.Files {
			if filePath, foundFile := file.Find(name, t); foundFile != nil {
				return "/" + f.Name + filePath, foundFile
			}
		}
	}

	return "", nil
}

type VFS struct {
	Root VFSFile `json:"root"`
}

func (vfs *VFS) Find(name string, t int) (string, *VFSFile) {
	return vfs.Root.Find(name, t)
}
