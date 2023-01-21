package plugin

import (
	"encoding/json"
	"os"
)

const T_DIR = 1
const T_FILE = 2
const T_SYMLINK = 3
const T_ANY = 4

type VFSFile struct {
	Type     int         `json:"type"`
	Name     string      `json:"name"`
	Files    []VFSFile   `json:"files"`
	Contents string      `json:"contents"`
	Mode     os.FileMode `json:"mode"`
}

// Find will check the current file or its descendants for the file with name `name` and type `t`.
func (f *VFSFile) Find(name string, t int) (string, *VFSFile) {
	if name == f.Name && (f.Type == t || t == T_ANY) {
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

// Find tries to find a specific file in the virtual file system.
func (vfs *VFS) Find(name string, t int) (string, *VFSFile) {
	return vfs.Root.Find(name, t)
}

// ReadVFSJSONFile reads the JSON file which contains the the virtual file system model.
func ReadVFSJSONFile(path string) (*VFS, error) {
	data, err := os.ReadFile(path)

	if err != nil {
		return nil, err
	}

	vfs := &VFS{}

	if err = json.Unmarshal(data, vfs); err != nil {
		return nil, err
	}

	return vfs, nil
}
